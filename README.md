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

特に、取得先のスクレイパーは `SCRAPER_BASE_URL` で切り替えできます。

- ローカル開発: `http://127.0.0.1:8082`
- Docker からホスト上のスクレイパーにアクセスする場合: `http://host.docker.internal:8082`

### 2. コンテナ起動（DB + Web）

```bash
docker compose up -d
```

`http://localhost:8183` を開くとレポートを確認できます（生成後）。

### 3. 初回のみ: DB初期化と監視銘柄の投入

```bash
docker compose run --rm app --init
docker compose run --rm app --seed
```

### 4. 日次/都度の更新

```bash
# 取得のみ
docker compose run --rm app --batch

# 取得 + HTML生成
docker compose run --rm app --batch --gen
```

### 5. 取得APIの疎通確認（コンテナ内から）

```bash
docker compose run --rm app sh -c 'set -a; . /app/env.config; set +a; curl -sS "$SCRAPER_BASE_URL/scrape?ticker=5020"'
```

200 応答で JSON が返ることを確認してください。

## リバースプロキシ運用の補足

同一ホスト上で複数の Docker Compose を動かしている場合、リバースプロキシの upstream 指定で 502 が発生しがちです。
プロキシがコンテナ内なら service 名、ホスト上なら公開ポートに向けてください。
