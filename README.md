# JPX Smallcap Watcher (Node.js)

独自の小型株データを取得し、MySQL に保存したうえで `public/` に静的 HTML レポートを生成する Node.js バッチツールです。
Docker 常駐コンテナを前提にせず、VPS 上の `/var/www/jpx-smallcap-watcher` で Node.js と Apache だけを使う構成にしています。

監視銘柄は `src/tickers1.tsv` で管理します。crontab で終値が出てくる 15:30 以降に `--batch --gen` を実行すると、終値ベースのダッシュボードを `public/index.html` と `public/detail/*.html` に生成できます。

## 主な機能

- JPX データ取得 → DB 保存 → シグナル判定 → 静的 HTML 生成
- 監視銘柄は `src/tickers1.tsv` で管理
- 生成物は `public/` に出力（`public/index.html` / `public/detail/*.html`）
- 既存の `env.config` 形式を継続利用
- Node.js v22 系で動作（VPS 確認済み: `v22.22.0`）

## 必要要件

- Node.js 22 以上
- npm
- MySQL 8.0 互換のデータベース
- スクレイパー API（`SCRAPER_BASE_URL/scrape?ticker=5020` のように JSON を返すもの）

## VPS への配置前提

この README のコマンド例は、プロジェクトが以下に配置されている前提です。

```bash
/var/www/jpx-smallcap-watcher
```

Apache 側の DocumentRoot / Alias / cron 設定は運用側で調整してください。静的サイトとして公開する対象は `public/` です。

## 1. セットアップ

```bash
cd /var/www/jpx-smallcap-watcher
cp env.config.sample env.config
npm ci --omit=dev
```

ローカルでテストも実行したい場合は、dev 依存はありませんが通常の `npm ci` でも問題ありません。

```bash
npm ci
npm test
```

## 2. 設定ファイル

`env.config` は既存形式をそのまま読み込みます。プロセス環境変数が設定済みの場合は、環境変数の値が優先されます。

```env
DB_HOST=localhost:3306
DB_USER=jpx_user
DB_PASSWORD=jpx_password
DB_NAME=jpx_data
SCRAPER_BASE_URL=http://localhost:8085
SCRAPER_REQUEST_INTERVAL=3s
OUTPUT_DIR=public
INDEX_MAX_AGE=36h
```

主な項目:

- `DB_HOST`: MySQL の接続先。ホスト上で動かすため通常は `localhost:3306`。
- `SCRAPER_BASE_URL`: スクレイパー API のベース URL。Docker の `host.docker.internal` ではなく、VPS ホスト上では通常 `http://localhost:8085`。
- `SCRAPER_REQUEST_INTERVAL`: スクレイパー API へのアクセス間隔。
- `OUTPUT_DIR`: 静的 HTML の生成先。Apache で公開する `public` を推奨。
- `INDEX_MAX_AGE`: 生成後ヘルスチェックで許容する `index.html` の最大経過時間。

## 3. 初回のみ: DB 初期化と監視銘柄の投入

```bash
cd /var/www/jpx-smallcap-watcher
node ./bin/smallcap-watcher.js --init
node ./bin/smallcap-watcher.js --seed
```

## 4. 日次/都度の更新

```bash
# 取得のみ
node ./bin/smallcap-watcher.js --batch

# HTML生成のみ
node ./bin/smallcap-watcher.js --gen

# 取得 + HTML生成
node ./bin/smallcap-watcher.js --batch --gen
```

`npm` script でも実行できます。

```bash
npm run batch
npm run gen
npm start -- --batch --gen
```

## 5. crontab 設定例

`scripts/run-daily-batch.sh` を使うと、`--batch --gen` 実行と鮮度チェックを 1 コマンドに固定できます。

```cron
# JPX Smallcap Watcher (Node.js)
# 平日 15:30〜15:49 のどこか1回（ランダムディレイ + flock で多重起動防止）
30-49 15 * * 1-5 cd /var/www/jpx-smallcap-watcher && /usr/bin/flock -n /tmp/jpx-smallcap-watcher.lock bash -lc 'sleep $((RANDOM % 60)); ./scripts/run-daily-batch.sh /var/www/jpx-smallcap-watcher' >> /var/www/jpx-smallcap-watcher/cron.log 2>&1
```

## 6. 生成物の鮮度ガード

`--gen` 実行後に `index.html` の鮮度チェックを行い、以下の場合は終了コード 1 で異常終了します。

- `index.html` が存在しない
- `index.html` のサイズが 0
- `INDEX_MAX_AGE` より古い
- `detail/*.html` の最新更新時刻より `index.html` が 60 秒以上古い

追加したヘルスチェックスクリプトでも同じ検証が可能です。

```bash
./scripts/check-index-freshness.sh public
```

## 7. 取得 API の疎通確認

```bash
set -a; . ./env.config; set +a
curl -sS "$SCRAPER_BASE_URL/scrape?ticker=5020"
```

200 応答で JSON が返ることを確認してください。

## 8. Apache 公開ディレクトリ

`public/` は実行時に生成される出力先です（リポジトリには `.gitkeep` だけを置き、HTML/CSS/画像は `--gen` 実行時に作成/コピーします）。Apache 側では以下のどちらかの形で公開できます。

- `DocumentRoot /var/www/jpx-smallcap-watcher/public`
- 既存 VirtualHost から `/var/www/jpx-smallcap-watcher/public` へ Alias

`public/static/` は `--gen` 実行時に `static/` からコピーされます。PR に生成物やバイナリ資産の重複を含めないため、`public/static/` は Git 管理しません。

## 9. 開発者向け

```bash
npm test
node ./bin/smallcap-watcher.js --help
```

Node.js 版の主要ファイル:

- `bin/smallcap-watcher.js`: CLI エントリポイント
- `lib/config.js`: `env.config` 読み込みと duration パース
- `lib/db.js`: MySQL 接続、スキーマ初期化、銘柄 seed
- `lib/api.js`: スクレイパー API クライアント
- `lib/service.js`: batch 更新と HTML 生成
- `lib/render.js`: 静的 HTML レンダリング
- `lib/freshness.js`: 生成物ヘルスチェック
