import csv
import os
from src.db import get_db_connection
from src.api import parse_val

TSV_PATH = os.path.join(os.path.dirname(__file__), 'src', 'tickers1.tsv')

def seed_db():
    conn = get_db_connection()
    cursor = conn.cursor()
    
    print(f"Importing master data from {TSV_PATH}...")
    
    try:
        with open(TSV_PATH, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f, delimiter='\t')
            count = 0
            for row in reader:
                # Basic fields
                ticker = row['ticker']
                company_name = row['companyName']
                memo = row.get('memo', '')
                
                # We can also populate initial data if needed, but primarily we want the clean ticker list.
                # Let's clean up some fields to insert rich initial data.
                current_price = parse_val(row.get('currentPrice', '0'))
                # previousClose might be "408 (12/29)", parse_val handles strings but let's be careful. api.py's parse_val is good.
                
                # Clean pricemovement: "-0.74%" -> -0.74
                pm_str = row.get('pricemovement', '').replace('%', '')
                try:
                    pricemovement = float(pm_str) if pm_str and pm_str != '－' else None
                except ValueError:
                    pricemovement = None
                    
                # Clean other fields
                pbr = parse_val(row.get('pbr'))
                volume = parse_val(row.get('volume'))
                
                # MySQL uses INSERT ... ON DUPLICATE KEY UPDATE
                # Values: %s placeholders
                cursor.execute('''
                    INSERT INTO watch_list (
                        ticker, companyName, currentPrice, previousClose, dividendYield, 
                        per, pbr, marketCap, volume, pricemovement, `signal_val`, memo
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON DUPLICATE KEY UPDATE
                        companyName=VALUES(companyName),
                        currentPrice=VALUES(currentPrice),
                        previousClose=VALUES(previousClose),
                        dividendYield=VALUES(dividendYield),
                        per=VALUES(per),
                        pbr=VALUES(pbr),
                        marketCap=VALUES(marketCap),
                        volume=VALUES(volume),
                        pricemovement=VALUES(pricemovement),
                        `signal_val`=VALUES(`signal_val`),
                        memo=VALUES(memo)
                ''', (
                    ticker, 
                    company_name, 
                    current_price if isinstance(current_price, (int, float)) else None,
                    row.get('previousClose'),
                    row.get('dividendYield'),
                    row.get('per'),
                    pbr if isinstance(pbr, (int, float)) else None,
                    row.get('marketCap'),
                    volume if isinstance(volume, (int, float)) else None,
                    pricemovement,
                    row.get('signal'),
                    memo
                ))
                count += 1
                
        conn.commit()
        print(f"Imported {count} tickers from master data.")
        
    except FileNotFoundError:
        print(f"Error: {TSV_PATH} not found.")
    except Exception as e:
        print(f"Error during import: {e}")
        
    conn.close()

if __name__ == '__main__':
    seed_db()
