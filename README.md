# smallcap-watcher-go
smallcap watcher by  go language

# 移植前のREADME
# JPX Smallcap Watcher


<img width="70％" height="70%" alt="1cb98e84-bbc0-40ef-a84a-4cc34959d6e1" src="https://github.com/user-attachments/assets/0daff931-9322-4980-84a1-3ddd35b2376d" />

JPX（日本取引所グループ）の株価データを毎日取得し、監視・分析するためのツールです。
DockerまたはPython環境で動作し、収集したデータを元に静的HTMLレポートを生成します。

## 必要要件 (Prerequisites)

*   **Docker & Docker Compose** (推奨)
*   または
*   **Python 3.10+** (ローカルで動かす場合)
    *   MySQL Server 8.0+ が別途必要です。

## セットアップ手順

### 1. `/etc/httpd/conf.d/virtualhost.conf` (HTTP)

Certbot実行前に、まずはHTTPでアクセスできるようにします。
`DocumentRoot` を指定することで、Certbotの認証がスムーズに行えます。

```apache
<VirtualHost *:80>
    ServerName smallcap.pinkgold.space
    DocumentRoot /var/www/jpx-smallcap-watcher/output

    ErrorLog "/var/log/httpd/smallcap_error_log"
    CustomLog "/var/log/httpd/smallcap_access_log" combined

    <Directory "/var/www/jpx-smallcap-watcher/output">
        Options -Indexes +FollowSymLinks
        AllowOverride all
        Require all granted
    </Directory>

    # Dockerコンテナ(Nginx:8181)へのプロキシ設定
    # 実際にはProxyPassが優先して処理されますが、Certbot等のためにDocumentRootも定義します
    ProxyPreserveHost On
    ProxyPass / http://localhost:8181/
    ProxyPassReverse / http://localhost:8181/
</VirtualHost>
```

### 2. `/etc/httpd/conf.d/virtualhost-le-ssl.conf` (HTTPS)

Certbot実行後に有効化される設定です。

```apache
<VirtualHost *:443>
    ServerName smallcap.pinkgold.space
    DocumentRoot /var/www/jpx-smallcap-watcher/output

    ErrorLog "/var/log/httpd/smallcap_pinkgold_space_error_log"
    CustomLog "/var/log/httpd/smallcap_pinkgold_space_access_log" combined

    <Directory "/var/www/jpx-smallcap-watcher/output">
        Options -Indexes +FollowSymLinks
        AllowOverride all
        Require all granted
    </Directory>

    ProxyPreserveHost On
    ProxyPass / http://localhost:8181/
    ProxyPassReverse / http://localhost:8181/

    # SSL設定
    Include /etc/letsencrypt/options-ssl-apache.conf
    SSLCertificateFile /etc/letsencrypt/live/smallcap.pinkgold.space/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/smallcap.pinkgold.space/privkey.pem
</VirtualHost>
```

### 3. SSL証明書の取得 (Certbot)

HTTPでのアクセス確認ができたら、Certbotを実行してHTTPS化します。

```bash
sudo certbot --apache -d smallcap.pinkgold.space
```

**重要:**
Certbot実行後、`/etc/httpd/conf.d/virtualhost-le-ssl.conf` (または生成されたSSL設定ファイル) を確認してください。
`ProxyPass / http://localhost:8181/` などのリバースプロキシ設定が自動生成に含まれていない場合は、**手動で追記**する必要があります（「2. ... (HTTPS)」の設定例を参照）。

### A. Docker環境 (推奨)

最も簡単に環境を構築できます。MySQLとアプリ、Webサーバーが自動で構成されます。

1.  **リポジトリのクローン**
    ```bash
    git clone https://github.com/YourUsername/jpx-smallcap-watcher.git
    cd jpx-smallcap-watcher
    ```

2.  **設定ファイルの準備**
    `.env` ファイルがプロジェクトルートにあることを確認してください（なければ作成）。
    ```env
    DB_HOST=mysql
    DB_USER=jpx_user
    DB_PASSWORD=jpx_password
    DB_NAME=jpx_data
    ```

3.  **起動と初期化**
    ```bash
    # ビルドとデーモン起動
    docker-compose up -d --build

    # データベースの初期化と初期データ投入
    docker-compose exec app python main.py --init
    docker-compose exec app python seed.py
    ```

4.  **動作確認**
    ブラウザで [http://localhost:8181](http://localhost:8181) にアクセスします。最初はデータがないため空です。

---

### B. Python仮想環境 (venv)

ローカルのPythonで開発・実行する場合の手順です。**別途MySQLサーバーがローカル(localhost:3309など)で稼働している必要があります。**
※ DockerのMySQLをポートマッピングで使うことも可能です。

1.  **仮想環境の作成**
    ```bash
    python -m venv .venv
    source .venv/bin/activate  # Windows: .venv\Scripts\activate
    ```

2.  **依存ライブラリのインストール**
    ```bash
    pip install -r requirements.txt
    ```

3.  **環境変数の設定** (MySQLへの接続情報)
    Windows (PowerShell) の例:
    ```powershell
    $env:DB_HOST="localhost"
    $env:DB_USER="jpx_user"
    $env:DB_PASSWORD="jpx_password"
    $env:DB_NAME="jpx_data"
    ```

---

### C. Mamba / Conda 環境

Mamba (またはConda) を使用する場合の手順です。

1.  **環境の作成と有効化**
    ```bash
    mamba create -n jpx-watcher python=3.12
    mamba activate jpx-watcher
    ```

2.  **ライブラリのインストール**
    `requirements.txt` を使うか、mamba installで入れます。
    ```bash
    # pipを使う場合 (推奨 - シンプル)
    pip install -r requirements.txt
    
    # または mamba/conda で入れる場合
    mamba install requests jinja2 pandas mysql-connector-python
    ```

## 使い方 (Usage)

### 1. 手動実行 (Manual Execution)

以下のコマンドで、「データ取得」「DB更新」「HTML生成」を一括で行います。

**Dockerの場合:**
```bash
docker-compose exec app python main.py --batch --gen
```

**ローカル(Python/Mamba)の場合:**
```bash
python main.py --batch --gen
```

### 2. マスターデータの変更 (Add/Update Tickers)

監視対象の銘柄を追加・変更したい場合は、`src/tickers1.tsv` を編集します。

1.  `src/tickers1.tsv` をエディタで開き、TSV形式（タブ区切り）で銘柄情報を追記・修正します。
2.  `seed.py` を実行してデータベースに反映させます。
    ```bash
    # Docker
    docker-compose exec app python seed.py
    
    # Local
    python seed.py
    ```
    ※ 既存の銘柄情報は更新され、新しい銘柄は追加されます（Upsert処理）。

### 3. 自動実行の設定 (Cron)

毎日決まった時間にバッチを実行するには、cron (Linux/WSL) を使用します。
例: 月曜から金曜の 15:30 (日本市場クローズ後) に実行する場合。

1.  crontabを開く
    ```bash
    crontab -e
    ```

2.  設定を追記 (Dockerを使用する場合の実装例)
    ```cron
    # 月曜から金曜の 15:30 に実行し、ログを cron.log に出力
    30 15 * * 1-5 cd /path/to/jpx-smallcap-watcher && /usr/bin/docker-compose exec -T app python main.py --batch --gen >> /path/to/jpx-smallcap-watcher/cron.log 2>&1
    ```
    *※ `/path/to/...` は実際の絶対パスに置き換えてください。*
    *※ `-T` オプションはTTYを無効化し、cron実行時のエラーを防ぐために重要です。*

### 4. データの初期化・リセット (Data Reset)

データベースを完全に削除し、初期状態から再構築する場合の手順です。

```bash
# 1. コンテナとボリュームの削除 (データが完全に消えます)
docker-compose down -v --remove-orphans

# 2. ビルドと起動 (キャッシュなし推奨: コード変更を確実に反映させるため)
docker-compose build --no-cache
docker-compose up -d

# 3. 初期データ投入
docker-compose exec app python main.py --init
docker-compose exec app python seed.py

# 4. バッチ実行
docker-compose exec app python main.py --batch --gen
```

## ディレクトリ構成

*   `src/`: ソースコード (API, DB, ロジック)
*   `templates/`: HTMLテンプレート (Jinja2)
*   `static/`: CSSファイル
*   `output/`: 生成されたHTMLファイル (Nginxのドキュメントルートにマウント)
*   `docker-compose.yml`: Docker構成定義





# JPX Smallcap Watcher (Specification)

## 概要
JPX（日本取引所グループ）の株価データを毎日取得し、監視・分析するためのツールです。監視対象の銘柄リストをもとに株価を取得し、MySQLへ保存したうえでシグナル判定を行い、静的HTMLレポートを生成します。

移植（Node.js/Go）時は、本仕様と `src/` 配下の実装を参照し、同等のデータ取得・保存・HTML生成の振る舞いを再現してください。

## システム構成
- **バッチ実行**: `main.py` で各処理を実行 (`--init`, `--batch`, `--gen`).
- **データ取得 API**: `https://jpx.pinkgold.space/scrape?ticker={ticker}` (`src/api.py`).
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
