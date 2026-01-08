import mysql.connector
import os
import time

def get_db_connection():
    # Retry logic or wait logic might be needed if not using depends_on healthy,
    # but docker-compose healthy check should handle init.
    
    # Get config from env or defaults
    db_host = os.environ.get('DB_HOST', 'localhost')
    db_user = os.environ.get('DB_USER', 'jpx_user')
    db_password = os.environ.get('DB_PASSWORD', 'jpx_password')
    db_name = os.environ.get('DB_NAME', 'jpx_data')
    
    conn = mysql.connector.connect(
        host=db_host,
        user=db_user,
        password=db_password,
        database=db_name
    )
    return conn

def init_db():
    # For MySQL, we might need to create the DB first if it doesn't exist?
    # The docker image creates MYSQL_DATABASE on startup.
    # So we just need to create tables.
    
    conn = get_db_connection()
    cursor = conn.cursor()

    # watch_list table
    # MySQL types: VARCHAR, DECIMAL(10,2), INT, TEXT are fine.
    # watch_list table
    # MySQL types: VARCHAR, DECIMAL(10,2), INT, TEXT are fine.
    # Re-writing with backticks for column names to be safe.
    # Renamed signal to signal_val or use backticks? 
    # Let's use backticks to minimize code changes elsewhere, 
    # BUT python multiline string with backticks is easy.
    # Actually 'SIGNAL' is a reserved word in MySQL 5.5+.
    # Let's quote it as `signal`.
    
    # Wait, 'signal' column.
    # Let's try to stick to `signal` with backticks.
    
    # Re-writing with backticks for column names to be safe.
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS watch_list (
            `ticker` VARCHAR(10) PRIMARY KEY,
            `companyName` VARCHAR(255),
            `currentPrice` DECIMAL(10, 2),
            `previousClose` VARCHAR(20),
            `dividendYield` VARCHAR(20),
            `per` VARCHAR(20),
            `pbr` DECIMAL(5, 2),
            `marketCap` VARCHAR(50),
            `volume` INT,
            `pricemovement` VARCHAR(50),
            `signal_val` VARCHAR(50),
            `memo` TEXT
        )
    ''')

    # watch_detail table
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS watch_detail (
            `ticker` VARCHAR(10),
            `yymmdd` DATE,
            `companyName` VARCHAR(255),
            `currentPrice` DECIMAL(10, 2),
            `previousClose` VARCHAR(20),
            `dividendYield` VARCHAR(20),
            `per` VARCHAR(20),
            `pbr` DECIMAL(5, 2),
            `marketCap` VARCHAR(50),
            `volume` INT,
            `pricemovement` VARCHAR(50),
            `signal_val` VARCHAR(50),
            `memo` TEXT,
            PRIMARY KEY (`ticker`, `yymmdd`)
        )
    ''')
    
    conn.commit()
    conn.close()
    print("Database initialized successfully (MySQL).")

if __name__ == '__main__':
    init_db()
