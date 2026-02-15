# JPX Smallcap Watcher (Go)

独自の小型株のデーターを取得し、MySQLに保存したうえで静的HTMLレポートを生成するバッチツールです。Go版の実装になります。
監視銘柄をsrc/tickers1.tsvに入れて起動するだけ　crontabで終値の出てくる15時半以降に設定すると終値ベースで眺めることができます。

## 主な機能

- JPXデータ取得 → DB保存 → シグナル判定 → HTML生成
- 監視銘柄は `src/tickers1.tsv` で管理
- 生成物は `output/` に出力（`output/index.html` / `output/detail/*.html`）

## 必要要件

- **Docker** / **Docker Compose**

## Dockerのインストール（Ubuntu / RedHat系 Linux）

### Ubuntu

```bash
sudo snap install docker
sudo apt install docker.io
sudo apt install podman-docker
```

#### Ubuntu 24.04 LTS (WSL) の注意点

WSL環境で `podman-docker` を入れると Docker CLI が Podman エミュレーションになり、
`docker compose` が外部の `docker-compose` (v1) を呼び出して `FileNotFoundError` が出る場合があります。
その場合は以下を確認してください。

```bash
# Docker Desktop for Windows を使うか、WSL 内で docker を有効化する
sudo systemctl enable --now docker

# podman-docker を外して Docker 公式パッケージ + Compose プラグインを入れる
sudo apt remove podman-docker
sudo apt update
sudo apt install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo \"$VERSION_CODENAME\") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install docker-ce docker-ce-cli docker-compose-plugin

# docker compose が v2 系か確認する
docker compose version
```

### RedHat系 Linux（RHEL / CentOS / Rocky / Alma など）

```bash
sudo dnf install docker
sudo systemctl enable --now docker
```

## 使い方（Docker Compose）

### 1. 設定ファイルの準備

```bash
cp env.config.sample env.config
```

必要に応じて `env.config` を編集してください。`env.config` が存在しない場合はデフォルト値で動作します。Docker実行時は `env.config` をコンテナへマウントします。

特に、取得先のスクレイパーは `SCRAPER_BASE_URL` で切り替えできます。`SCRAPER_REQUEST_INTERVAL` を指定すると、スクレイパーへのアクセス間隔を調整できます（例: `3s`）。

- ローカル開発: `http://host.docker.internal:8085`
- Docker からホスト上のスクレイパーにアクセスする場合: `http://host.docker.internal:8085`

### 2. コンテナ起動（DB + Web）

```bash
env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD docker compose --env-file env.config up -d
```

`http://localhost:8183` を開くとレポートを確認できます（生成後）。

### 3. 初回のみ: DB初期化と監視銘柄の投入

```bash
env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD docker compose --env-file env.config run --rm app --init
env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD docker compose --env-file env.config run --rm app --seed
```

### 4. 日次/都度の更新

```bash
# 取得のみ
env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD docker compose --env-file env.config run --rm app --batch

# 取得 + HTML生成
env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD docker compose --env-file env.config run --rm app --batch --gen
```

### 5. 取得APIの疎通確認（コンテナ内から）

```bash
env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD docker compose --env-file env.config run --rm app sh -c 'set -a; . /app/env.config; set +a; curl -sS "$SCRAPER_BASE_URL/scrape?ticker=5020"'
```

200 応答で JSON が返ることを確認してください。


## トラブルシュート

### `Access denied for user` が出る場合

`Error 1045 (28000): Access denied for user ...` は、**ポート競合よりも認証情報の不一致**で起きることが多いです。
特に、`mysql-data` ボリュームを使っていると、MySQL のユーザー情報は初回作成時の値が保持されます。
また、シェルに `DB_USER` などの環境変数が既に export されていると、`docker compose` の変数展開で `env.config` より優先され、意図しないユーザー（例: `jpx`）で接続してしまうことがあります。

そのため README のコマンドは `env -u ... docker compose --env-file env.config ...` 形式にしており、DB 関連の環境変数を一度クリアしてから `env.config` の値を確実に使うようにしています。

以下で `env.config` の値に合わせて DB ユーザー権限を再設定できます。

```bash
./scripts/db-repair-auth.sh
```

その後、再実行してください。

```bash
env -u DB_HOST -u DB_USER -u DB_PASSWORD -u DB_NAME -u MYSQL_ROOT_PASSWORD docker compose --env-file env.config run --rm app --batch --gen
```

### VPS 側でポート競合を確認したい場合

このプロジェクトの MySQL はホスト `3312` を使います（`3312:3306`）。
競合確認は以下でできます。

```bash
ss -ltnp | grep 3312
```

既に別プロセスが使っている場合は、`docker-compose.yml` の左側ポート（`3312`）を別番号に変更してください。

## リバースプロキシ運用の補足

同一ホスト上で複数の Docker Compose を動かしている場合、リバースプロキシの upstream 指定で 502 が発生しがちです。
プロキシがコンテナ内なら service 名、ホスト上なら公開ポートに向けてください。
