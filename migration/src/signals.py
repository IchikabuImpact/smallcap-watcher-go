def calculate_pricemovement(current_price, previous_close):
    """
    Calculates percentage change.
    Formula: ((currentPrice - previousClose) / previousClose) * 100
    """
    if not previous_close or previous_close == 0:
        return 0.0
    
    diff = current_price - previous_close
    movement = (diff / previous_close) * 100
    return round(movement, 2)

def generate_signal(pricemovement, rsi=None):
    """
    Generates a signal (Buy, Neutral, Sell) based on indicators.
    For now, simplified logic based on pricemovement since RSI implementation requires historical data which we might not have yet.
    
    Logic:
    - Buy: Price movement > 3.0% (Strong upward momentum)
    - Sell: Price movement < -3.0% (Strong downward momentum)
    - Neutral: Otherwise
    """
    if pricemovement > 3.0:
        return "Buy"
    elif pricemovement < -3.0:
        return "Sell"
    else:
        return "Neutral"
