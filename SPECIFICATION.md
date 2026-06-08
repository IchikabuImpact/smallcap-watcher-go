# JPX Smallcap Watcher (Node.js Specification)

## 概要
JPX（日本取引所グループ）の株価データを毎日取得し、監視・分析するためのツールです。監視対象の銘柄リストをもとに株価を取得し、MySQL へ保存したうえでシグナル判定を行い、`public/` に静的 HTML レポートを生成します。

Node.js 実装では、`bin/smallcap-watcher.js` の CLI から各処理を実行し、`lib/` 配下のロジックでデータ取得・保存・HTML生成を行います。

## システム構成
- **バッチ実行**: `bin/smallcap-watcher.js` の CLI で各処理を実行 (`--init`, `--batch`, `--gen`, `--seed`).
- **データ取得 API**: `SCRAPER_BASE_URL/scrape?ticker={ticker}`。`SCRAPER_BASE_URL` は `env.config` または環境変数で切り替え（デフォルト: `http://localhost:8085`）。
- **データ保存**: MySQL 8.0 互換 DB。
- **HTML生成**: `lib/render.js` で HTML をレンダリングし、`public/` に静的 HTML を出力。
- **配信**: Apache など任意の Web サーバーで `public/` を静的配信。

## データフロー
1. **監視対象の取得**: `watch_list` テーブルから全銘柄を読み込む。
2. **API取得**: 銘柄ごとに `fetchStockData` を呼び出す。
3. **パース**: `parseNumeric` と `parsePreviousClose` で文字列の数値を正規化し、計算用の `previousClose` を生成。
4. **指標計算**: 変動率とシグナルを算出。
5. **DB更新**:
   - `watch_list` を最新値で `INSERT ... ON DUPLICATE KEY UPDATE` で更新。
   - `watch_detail` に `REPLACE INTO` で日次レコードを保存（主キーは `ticker + yymmdd`）。
6. **HTML生成**: `watch_list` と `watch_detail` から一覧/詳細ページを生成。
7. **鮮度確認**: `index.html` のサイズ、最大経過時間、詳細ページとの mtime 整合を検証。

## API仕様 (取得データ)
`/scrape` のレスポンスから以下のキーを利用します。
- `ticker`: 銘柄コード
- `companyName`: 企業名
- `currentPrice`: 現在値 (例: `"2,493.0円"`)
- `previousClose`: 前日終値＋日付 (例: `"2,496.5 (12/29)"`)
- `dividendYield`: 配当利回り (文字列, `%`付き)
- `per`: PER (文字列)
- `pbr`: PBR (例: `"1.23倍"`)
- `marketCap`: 時価総額 (例: `"29兆5,862億円"`)
- `volume`: 出来高 (例: `"1,234,000"`)

### 数値正規化ルール
`parseNumeric` で以下を処理します。
- カンマ除去、`円`/`倍`/`%`/`株` などの単位を除去して数値化。
- `兆`/`億`/`万` があれば倍率を掛けて数値化（`兆`=1e12, `億`=1e8, `万`=1e4）。
- 変換できない場合は無効値として扱い、呼び出し側で文字列保持。
- `previousClose` は「数値部分」だけを抽出して計算に利用。

## DB仕様
- **DBMS**: MySQL 8.0 互換。
- **接続情報**: `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` を `env.config` または環境変数で指定。
- **取得API切替**: `SCRAPER_BASE_URL` で取得先を切り替える。
- **スキーマ**: `watch_list`, `watch_detail` の 2 テーブル。
- **スキーマ初期化**: `--init` が `CREATE TABLE IF NOT EXISTS` を実行。

### watch_list (最新状態)
| Column | Type | Notes |
| --- | --- | --- |
| ticker | VARCHAR(10) | PK. 監視銘柄コード |
| companyName | VARCHAR(255) | 企業名 |
| currentPrice | DECIMAL(10,2) | 現在値 |
| previousClose | VARCHAR(20) | 前日終値 (日付付き文字列) |
| dividendYield | VARCHAR(20) | 配当利回り（文字列） |
| per | VARCHAR(20) | PER（文字列） |
| pbr | DECIMAL(5,2) | PBR |
| marketCap | VARCHAR(50) | 時価総額（文字列） |
| volume | INT | 出来高 |
| pricemovement | VARCHAR(50) | 変動率（文字列/数値混在） |
| signal_val | VARCHAR(50) | シグナル (`Buy`, `Sell`, `Neutral`) |
| memo | TEXT | メモ |

### watch_detail (日次履歴)
- `PRIMARY KEY (ticker, yymmdd)`。
- それ以外のカラムは `watch_list` と同型（`yymmdd` のみ追加）。

## シグナル仕様
- 変動率は `((currentPrice - previousClose) / previousClose) * 100`。
- **Buy**: 3.0% 超の上昇
- **Sell**: -3.0% 未満の下落
- **Neutral**: 上記以外

## HTML生成仕様
- **出力先**: `public/`（`OUTPUT_DIR` で変更可）。
- **一覧ページ**: `public/index.html`。
- **詳細ページ**: `public/detail/{ticker}.html`。
- **スタイル/画像**: `static/` を `public/static/` へコピー（`public/static/` は生成物扱いで Git 管理しない）。
- 一覧行はクリックで詳細ページへ遷移する。

## マスターデータ
- 監視対象の銘柄は `src/tickers1.tsv` で管理。
- `--seed` で `watch_list` へ upsert (INSERT ... ON DUPLICATE KEY UPDATE)。

## 実行コマンド（Node.js）
事前に `env.config` を用意し、依存関係をインストールします。

```bash
cd /var/www/jpx-smallcap-watcher
cp env.config.sample env.config
npm ci --omit=dev

# 初回のみ
node ./bin/smallcap-watcher.js --init
node ./bin/smallcap-watcher.js --seed

# バッチ実行
node ./bin/smallcap-watcher.js --batch

# バッチ + HTML生成
node ./bin/smallcap-watcher.js --batch --gen
```

## データ初期化・生成・デイリーバッチ
- **データ初期化**: `--init` でテーブル作成後、`--seed` で `src/tickers1.tsv` を投入。
- **データ生成**: `--batch --gen` で取得・更新・HTML生成を一括実行。
- **デイリーバッチ**: cron などで `scripts/run-daily-batch.sh` を定期実行。

## public ディレクトリ仕様
- `public/` は静的 HTML の出力先で、`index.html` と `detail/*.html` を生成する。
- `public/static/` に `style.css`, `favicon.svg`, 背景画像を実行時コピーする。
- Apache 側では `public/` を DocumentRoot または Alias として公開する。
- リポジトリには生成物や重複バイナリを含めず、`public/.gitkeep` のみを管理する。

## 技術スタック
- **Node.js**: 22 以上。
- **主要ライブラリ**: `mysql2`、Node.js 標準ライブラリ、Node.js 標準 `fetch`。
- **インフラ**: MySQL 8.0 互換 DB。
- **配信**: Apache など任意の静的 Web サーバー。
