# JPX Smallcap Watcher (Specification)

## 概要
JPX（日本取引所グループ）の株価データを毎日取得し、監視・分析するためのツールです。監視対象の銘柄リストをもとに株価を取得し、MySQLへ保存したうえでシグナル判定を行い、静的HTMLレポートを生成します。

移植（Node.js/Go）時は、本仕様と `src/` 配下の実装を参照し、同等のデータ取得・保存・HTML生成の振る舞いを再現してください。

## システム構成
- **バッチ実行**: `main.py` で各処理を実行 (`--init`, `--batch`, `--gen`).
- **データ取得 API**: `https://jpx-indicator.pinkgold.space/scrape?ticker={ticker}` (`src/api.py`).
- **データ保存**: MySQL 8.0 (Dockerコンテナまたはローカル環境).
- **HTML生成**: Jinja2 テンプレート (`templates/`) から `output/` に静的HTMLを出力.
- **配信**: Docker構成では Nginx が `output/` を配信.

## データフロー
1. **監視対象の取得**: `watch_list` テーブルから全銘柄を読み込む (`src/batch.py`).
2. **API取得**: 銘柄ごとに `fetch_stock_data` を呼び出す.
3. **パース**: `parse_data` で文字列の数値を正規化し、計算用の `previousCloseVal` を生成.
4. **指標計算**: `calculate_pricemovement` と `generate_signal` で変動率とシグナルを算出.
5. **DB更新**:
   - `watch_list` を最新値で `UPDATE`.
   - `watch_detail` に `REPLACE INTO` で日次レコードを保存（主キーは `ticker + yymmdd`）.
6. **HTML生成**: `watch_list` と `watch_detail` から一覧/詳細ページを生成.

## API仕様 (取得データ)
`/scrape` のレスポンスから以下のキーを利用します（`src/api.py`）。
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
`parse_val` で以下を処理します。
- カンマ除去、`円`/`倍`/`%` などの単位を除去して数値化.
- `兆`/`億`/`万` があれば倍率を掛けて数値化（`兆`=1e12, `億`=1e8, `万`=1e4）.
- 変換できない場合は元の文字列を保持.
- `previousClose` は「数値部分」だけを抽出して `previousCloseVal` として計算に利用.

## DB仕様
- **DBMS**: MySQL 8.0
- **接続情報**: 環境変数 `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` (デフォルトは `src/db.py` を参照).
- **スキーマ**: `watch_list`, `watch_detail` の2テーブル.
- **公式スキーマ出力**: `schema/mysqldump_jpx_data_nodata.sql` (mysqldump --no-data 相当).

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
- RSI 等の拡張指標は未実装（将来拡張を考慮）.

## HTML生成仕様
- **出力先**: `output/`.
- **一覧ページ**: `output/index.html` (`templates/list.html`).
- **詳細ページ**: `output/detail/{ticker}.html` (`templates/detail.html`).
- **スタイル**: `static/style.css`.
- 一覧行はクリックで詳細ページへ遷移する.

## マスターデータ
- 監視対象の銘柄は `src/tickers1.tsv` で管理.
- `seed.py` で `watch_list` へ upsert (INSERT ... ON DUPLICATE KEY UPDATE).

## 実行コマンド
- `python main.py --init` : DB初期化 (テーブル作成).
- `python main.py --batch` : データ取得とDB更新.
- `python main.py --gen` : HTML生成.
- `python seed.py` : 監視銘柄の初期投入/更新.

## 技術スタック
- **Python**: 3.12 (要件は 3.10+).
- **主要ライブラリ**: `requests`, `jinja2`, `mysql-connector-python`.
- **インフラ**: Docker / Docker Compose.
- **配信**: Nginx (静的HTML配信).
