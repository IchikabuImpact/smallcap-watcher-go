# JPX Smallcap Watcher (Go Specification)

## 概要
JPX（日本取引所グループ）の株価データを毎日取得し、監視・分析するためのツールです。監視対象の銘柄リストをもとに株価を取得し、MySQLへ保存したうえでシグナル判定を行い、静的HTMLレポートを生成します。

本Go実装では、`cmd/smallcap-watcher` 配下のCLIから各処理を実行し、`internal/` 配下のロジックで同等のデータ取得・保存・HTML生成の振る舞いを提供します。

## システム構成
- **バッチ実行**: `cmd/smallcap-watcher/main.go` のCLIで各処理を実行 (`--init`, `--batch`, `--gen`, `--seed`).
- **データ取得 API**: `SCRAPER_BASE_URL/scrape?ticker={ticker}`。`SCRAPER_BASE_URL` は環境変数で切り替え（デフォルト: `http://127.0.0.1:8085`）。 (`internal/api/client.go`).
- **データ保存**: MySQL 8.0 (Docker Compose で起動するコンテナ).
- **HTML生成**: Go `html/template` (`templates/`) から `output/` に静的HTMLを出力.
- **配信**: 任意のWebサーバーで `output/` を静的配信.

## データフロー
1. **監視対象の取得**: `watch_list` テーブルから全銘柄を読み込む (`internal/service/batch.go`).
2. **API取得**: 銘柄ごとに `FetchStockData` を呼び出す (`internal/api/client.go`).
3. **パース**: `parse.ParseNumeric` と `parse.ParsePreviousClose` で文字列の数値を正規化し、計算用の `previousCloseVal` を生成 (`internal/parse/parse.go`).
4. **指標計算**: `generateSignal` で変動率とシグナルを算出 (`internal/service/batch.go`).
5. **DB更新**:
   - `watch_list` を最新値で `INSERT ... ON DUPLICATE KEY UPDATE` で更新.
   - `watch_detail` に `REPLACE INTO` で日次レコードを保存（主キーは `ticker + yymmdd`）.
6. **HTML生成**: `watch_list` と `watch_detail` から一覧/詳細ページを生成 (`internal/service/generate.go`).

## API仕様 (取得データ)
`/scrape` のレスポンスから以下のキーを利用します（`internal/api/client.go`）。
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
`parse.ParseNumeric` で以下を処理します。
- カンマ除去、`円`/`倍`/`%`/`株` などの単位を除去して数値化.
- `兆`/`億`/`万` があれば倍率を掛けて数値化（`兆`=1e12, `億`=1e8, `万`=1e4）.
- 変換できない場合は `false` を返却し、呼び出し側で文字列保持.
- `previousClose` は「数値部分」だけを抽出して `previousCloseVal` として計算に利用.

## DB仕様
- **DBMS**: MySQL 8.0
- **接続情報**: 環境変数 `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` (デフォルトは `internal/config/config.go` を参照).
- **取得API切替**: 環境変数 `SCRAPER_BASE_URL` で取得先を切り替える.
- **スキーマ**: `watch_list`, `watch_detail` の2テーブル.
- **スキーマ初期化**: `db.InitSchema` が `CREATE TABLE IF NOT EXISTS` を実行 (`internal/db/db.go`).

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
- `PRIMARY KEY (ticker, yymmdd)`.
- それ以外のカラムは `watch_list` と同型（`yymmdd` のみ追加）.

## シグナル仕様
- 変動率は `((currentPrice - previousCloseVal) / previousCloseVal) * 100`.
- **Buy**: 3.0% 超の上昇
- **Sell**: -3.0% 未満の下落
- **Neutral**: 上記以外

## HTML生成仕様
- **出力先**: `output/`.
- **一覧ページ**: `output/index.html` (`templates/list.html`).
- **詳細ページ**: `output/detail/{ticker}.html` (`templates/detail.html`).
- **スタイル**: `static/style.css` を `output/static/style.css` へコピー.
- **背景画像**: `static/background.png` を `output/static/background.png` へコピー.
- 一覧行はクリックで詳細ページへ遷移する.

## マスターデータ
- 監視対象の銘柄は `src/tickers1.tsv` で管理.
- `SeedWatchList` で `watch_list` へ upsert (INSERT ... ON DUPLICATE KEY UPDATE).

## 実行コマンド（Docker Compose）
事前に `env.config` を用意し、Docker Compose を起動してから実行します。

```bash
# 設定ファイル（任意）
cp env.config.sample env.config

# DB + Web 起動
docker compose up -d

# 初回のみ
docker compose run --rm app --init
docker compose run --rm app --seed

# バッチ実行
docker compose run --rm app --batch

# バッチ + HTML生成
docker compose run --rm app --batch --gen
```

## データ初期化・生成・デイリーバッチ
- **データ初期化**: `--init` でテーブル作成後、`--seed` で `src/tickers1.tsv` を投入.
- **データ生成**: `--batch --gen` で取得・更新・HTML生成を一括実行.
- **デイリーバッチ**: cron などで `--batch --gen` を定期実行.

## output ディレクトリ仕様
- `output/` は静的HTMLの出力先で、`index.html` と `detail/*.html` を生成する.
- `output/static/` に `style.css` と `background.png` を配置する.
- Docker利用時はホストの `output/` をバインドマウントし、ファイル配置時の Permission エラーを避けるためにパーミッションを調整する.

## 技術スタック
- **Go**: 1.25+ (go.mod 参照).
- **主要ライブラリ**: `github.com/go-sql-driver/mysql`, 標準ライブラリの `net/http`, `html/template`.
- **インフラ**: MySQL 8.0.
- **配信**: 任意の静的Webサーバー.
