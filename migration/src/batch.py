import datetime
from src.db import get_db_connection
from src.api import fetch_stock_data, parse_data
from src.signals import calculate_pricemovement, generate_signal

def run_daily_batch():
    conn = get_db_connection()
    # Use dictionary cursor for column access by name
    cursor = conn.cursor(dictionary=True)
    
    print("Starting daily batch processing...")
    
    # 1. Get all tickers from watch_list
    cursor.execute("SELECT ticker FROM watch_list")
    rows = cursor.fetchall()
    tickers = [row['ticker'] for row in rows]
    
    if not tickers:
        print("No tickers found in watch_list.")
        conn.close()
        return

    today_str = datetime.date.today().strftime('%Y-%m-%d')
    
    for ticker in tickers:
        print(f"Processing {ticker}...")
        
        # 2. Fetch data
        raw_data = fetch_stock_data(ticker)
        if not raw_data:
            print(f"Skipping {ticker} due to fetch error.")
            continue
            
        data = parse_data(raw_data)
        
        # 3. Calculate indicators
        current_price = data.get('currentPrice')
        prev_close_val = data.get('previousCloseVal')
        
        if current_price is None or prev_close_val is None:
             print(f"Skipping {ticker}: Missing price data.")
             continue
             
        movement = calculate_pricemovement(current_price, prev_close_val)
        signal = generate_signal(movement)
        
        # 4. Update watch_list
        # MySQL uses %s placeholder
        cursor.execute('''
            UPDATE watch_list
            SET currentPrice = %s,
                previousClose = %s,
                dividendYield = %s,
                per = %s,
                pbr = %s,
                marketCap = %s,
                volume = %s,
                pricemovement = %s,
                `signal_val` = %s
            WHERE ticker = %s
        ''', (
            current_price,
            data.get('previousClose'),
            data.get('dividendYield'),
            data.get('per'),
            data.get('pbr'),
            data.get('marketCap'),
            data.get('volume'),
            movement,
            signal,
            ticker
        ))
        
        # 5. Insert into watch_detail
        # Use REPLACE INTO for MySQL (similar to INSERT OR REPLACE)
        cursor.execute('''
            REPLACE INTO watch_detail (
                ticker, yymmdd, companyName, currentPrice, previousClose,
                dividendYield, per, pbr, marketCap, volume,
                pricemovement, `signal_val`, memo
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        ''', (
            ticker,
            today_str,
            data.get('companyName'),
            current_price,
            data.get('previousClose'),
            data.get('dividendYield'),
            data.get('per'),
            data.get('pbr'),
            data.get('marketCap'),
            data.get('volume'),
            movement,
            signal,
            "" # memo
        ))
        
    conn.commit()
    conn.close()
    print("Daily batch processing completed.")

if __name__ == '__main__':
    run_daily_batch()
