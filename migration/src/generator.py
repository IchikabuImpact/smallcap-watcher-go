import os
from jinja2 import Environment, FileSystemLoader
from src.db import get_db_connection

OUTPUT_DIR = 'output'
TEMPLATE_DIR = 'templates'

def render_template(template_name, context):
    env = Environment(loader=FileSystemLoader(TEMPLATE_DIR))
    template = env.get_template(template_name)
    return template.render(context)

def generate_html():
    print("Generating HTML report...")
    if not os.path.exists(OUTPUT_DIR):
        os.makedirs(OUTPUT_DIR)
        
    conn = get_db_connection()
    cursor = conn.cursor(dictionary=True)
    
    # 1. Generate List Page
    cursor.execute("SELECT * FROM watch_list ORDER BY ticker")
    stocks = cursor.fetchall()
    
    list_html = render_template('list.html', {'stocks': stocks})
    with open(os.path.join(OUTPUT_DIR, 'index.html'), 'w', encoding='utf-8') as f:
        f.write(list_html)
        
    # 2. Generate Detail Pages
    detail_dir = os.path.join(OUTPUT_DIR, 'detail')
    if not os.path.exists(detail_dir):
        os.makedirs(detail_dir)
        
    for stock in stocks:
        ticker = stock['ticker']
        # MySQL uses %s placeholder
        cursor.execute("SELECT * FROM watch_detail WHERE ticker = %s ORDER BY yymmdd DESC", (ticker,))
        history = cursor.fetchall()
        
        detail_html = render_template('detail.html', {'stock': stock, 'history': history})
        with open(os.path.join(detail_dir, f'{ticker}.html'), 'w', encoding='utf-8') as f:
            f.write(detail_html)
            
    conn.close()
    print(f"HTML generation completed. Output in {OUTPUT_DIR}/")

if __name__ == '__main__':
    generate_html()
