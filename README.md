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

### 2. コンテナ起動（DB + Web）

```bash
docker compose up -d
```

`http://localhost:8083` を開くとレポートを確認できます（生成後）。

## リバースプロキシ運用時の 502 対策（重要）

同一ホスト上で複数の Docker Compose を動かしている場合、リバースプロキシの upstream 指定ミスで 502 が発生しがちです。
以下の原則に沿って upstream を指定してください。

### プロキシが「コンテナ内」で動いている場合

- **原則：Docker Compose の service 名で接続する**
  - 例: `proxy_pass http://web:80;`（この Compose 内の `web` サービス）
  - 例: `proxy_pass http://app:8085;`（この Compose 内の `app` サービス）
- **`host.docker.internal` を使う必要がある場合**
  - Linux では名前解決できないことがあるため、`extra_hosts` に以下を追加します。

```yaml
services:
  proxy:
    extra_hosts:
      - "host.docker.internal:host-gateway"
```

### プロキシが「ホスト OS 上」で動いている場合

- **原則：公開ポート（ports で publish しているポート）に向ける**
  - 例: `proxy_pass http://127.0.0.1:8183;`（`web` サービスの公開ポート）
  - 例: `proxy_pass http://127.0.0.1:8085;`（別スタックの公開ポート）

### 502 を解消するための確認手順（例）

```bash
docker compose ps
docker compose logs -n 200 web
docker compose logs -n 200 app

# プロキシがホスト上の場合（nginx / apache）
sudo nginx -T
sudo apachectl -S
sudo apachectl -t -D DUMP_VHOSTS

# プロキシがコンテナ内の場合
docker exec -it <proxy_container> sh
curl -v http://<upstream>:<port>/

# ホストから公開ポート到達を確認
curl -v http://127.0.0.1:<published_port>/
```
