import argparse
import sys
from src.db import init_db
from src.batch import run_daily_batch
from src.generator import generate_html

def main():
    parser = argparse.ArgumentParser(description="JPX Smallcap Watcher Automation")
    parser.add_argument('--init', action='store_true', help='Initialize database')
    parser.add_argument('--batch', action='store_true', help='Run daily batch process')
    parser.add_argument('--gen', action='store_true', help='Generate static HTML site')
    
    if len(sys.argv) == 1:
        parser.print_help(sys.stderr)
        sys.exit(1)
        
    args = parser.parse_args()
    
    if args.init:
        init_db()
        
    if args.batch:
        run_daily_batch()
        
    if args.gen:
        generate_html()

if __name__ == '__main__':
    main()
