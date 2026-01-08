import requests
import json
import re

BASE_URL = "https://jpx.pinkgold.space/scrape"

def fetch_stock_data(ticker):
    """
    Fetches stock data for a given ticker from the custom API.
    """
    try:
        response = requests.get(f"{BASE_URL}?ticker={ticker}", timeout=10)
        response.raise_for_status()
        return response.json()
    except requests.RequestException as e:
        print(f"Error fetching data for {ticker}: {e}")
        return None

def parse_val(val_str):
    """
    Parses numeric values from strings like '2,493.0円', '13.5倍', '29兆5,862億円'.
    Returns a float or int, or the original string if parsing fails.
    """
    if not isinstance(val_str, str):
        return val_str
    
    # Remove commas
    clean_str = val_str.replace(',', '')
    
    # helper for Japanese units
    multipliers = {'兆': 1000000000000, '億': 100000000, '万': 10000}
    
    # Check for units
    multiplier = 1
    for unit, mult in multipliers.items():
        if unit in clean_str:
            multiplier = mult
            clean_str = clean_str.replace(unit, '')
            break  # Assume only one major unit like 兆 or 億 usually appears or handle sequentially if needed (rare for this simple format)
            
    # Remove other non-numeric chars (except dot and minus)
    # Using regex to extract the number part
    match = re.search(r'-?\d+(\.\d+)?', clean_str)
    if match:
        num_str = match.group()
        try:
            val = float(num_str)
            result = val * multiplier
            # Return int if it's a whole number (optional, but cleaner)
            if result.is_integer():
                return int(result)
            return result
        except ValueError:
            pass
            
    return val_str

def parse_data(json_data):
    """
    Parses the raw JSON response into a cleaner dictionary with numeric types where appropriate.
    """
    if not json_data:
        return None
        
    parsed = {}
    
    parsed['ticker'] = json_data.get('ticker')
    parsed['companyName'] = json_data.get('companyName')
    
    # Numeric fields
    parsed['currentPrice'] = parse_val(json_data.get('currentPrice'))
    parsed['dividendYield'] = json_data.get('dividendYield') # Keep as string properly? or parse? Plan said "clean strings". Yield is usually %, let's keep string for display or parse.
    # The requirement says "dividendYield" type VARCHAR(20) in DB, so keeping as string or parsed string is fine. 
    # But usually we might want to store raw text for display or numeric for sorting. 
    # The prompt DB design says VARCHAR(20), so let's keep it simple, maybe just strip '％' if we want calculation, 
    # but for now I will keep it as is or just basic cleanup if needed. DB schema is VARCHAR.
    
    parsed['per'] = json_data.get('per')
    parsed['pbr'] = parse_val(json_data.get('pbr')) # decimal in DB
    parsed['marketCap'] = json_data.get('marketCap') # string in DB? Wait, schema said VARCHAR(50). 
    # "marketCap": "29兆5,862億円" -> This is complex to store as number without big int support or just store display string.
    # Schema says VARCHAR(50), so I will store the raw string but maybe clean it up for display if needed.
    # Wait, previousClose is VARCHAR, but currentPrice is DECIMAL.
    # Let's stick to the schema types.
    
    parsed['volume'] = parse_val(json_data.get('volume')) # INT in DB
    
    # previousClose comes as "2,496.5 (12/29)" -> we might want just the number for calculation
    prev_close_raw = json_data.get('previousClose', '')
    if prev_close_raw:
        # Extract the number part before any space or parenthesis
        prev_close_val_str = prev_close_raw.split()[0] # "2,496.5"
        parsed['previousCloseVal'] = parse_val(prev_close_val_str) # for calculation
        parsed['previousClose'] = prev_close_raw # for display storage
    else:
        parsed['previousClose'] = None
        parsed['previousCloseVal'] = None

    return parsed
