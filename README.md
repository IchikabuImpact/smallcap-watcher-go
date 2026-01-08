# JPX Smallcap Watcher (Go)

独自の小型株のデーターを取得し、MySQLに保存したうえで静的HTMLレポートを生成するバッチツールです。Go版の実装になります。

## 主な機能

- JPXデータ取得 → DB保存 → シグナル判定 → HTML生成
- 監視銘柄は `src/tickers1.tsv` で管理
- 生成物は `output/` に出力（`output/index.html` / `output/detail/*.html`）

## 必要要件

- **Docker** / **Docker Compose**（推奨）
- もしくは **Go 1.22+** と **MySQL 8.0+**

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

## 使い方（Docker）

### 1. 設定ファイルの準備

```bash
cp env.config.sample env.config
```

必要に応じて `env.config` を編集してください。`env.config` が存在しない場合はデフォルト値で動作します。Docker実行時は `env.config` をコンテナへマウントします。

### 2. コンテナ起動（DB + Web）

```bash
docker compose up -d
```

`http://localhost:8282` を開くとレポートを確認できます（生成後）。
HTML 出力は Docker ボリューム `output-data` に保存されます。

### 3. 初期化・シード・バッチ実行

```bash
# スキーマ作成
docker compose exec app /app/smallcap-watcher --init

# 監視銘柄の投入
docker compose exec app /app/smallcap-watcher --seed

# データ取得 + HTML生成
docker compose exec app /app/smallcap-watcher --batch --gen
```

## 使い方（ローカル）

### 1. 設定ファイルの準備

```bash
cp env.config.sample env.config
```

必要に応じて `env.config` を編集してください。`env.config` が存在しない場合はデフォルト値で動作します。

### 2. 実行

```bash
go run ./cmd/smallcap-watcher --init
go run ./cmd/smallcap-watcher --seed
go run ./cmd/smallcap-watcher --batch --gen
```

## CLIオプション

- `--init` : DBスキーマ初期化
- `--seed` : `src/tickers1.tsv` から監視銘柄を投入
- `--batch` : データ取得とDB更新
- `--gen` : 静的HTML生成

## ディレクトリ構成

- `cmd/smallcap-watcher/` : エントリポイント
- `internal/` : API取得、DB操作、HTML生成ロジック
- `templates/` : HTMLテンプレート
- `static/` : CSS
- `output/` : 生成HTML（コンテナ起動後に作成）
- `src/tickers1.tsv` : 監視銘柄マスタ

## 補足

- 詳細な仕様は `SPECIFICATION.md` を参照してください。
- 本ツールはバッチ実行を想定しています。定期実行する場合は `cron` などをご利用ください。
