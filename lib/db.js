import fs from 'node:fs/promises';
import mysql from 'mysql2/promise';

export async function openDatabase(config) {
  const { host, port } = parseHostPort(config.dbHost);
  return mysql.createConnection({
    host,
    port,
    user: config.dbUser,
    password: config.dbPassword,
    database: config.dbName,
    charset: 'utf8mb4',
    timezone: 'local',
  });
}

export async function initSchema(db) {
  const statements = [
    `CREATE TABLE IF NOT EXISTS watch_list (
      ticker VARCHAR(10) PRIMARY KEY,
      companyName VARCHAR(255),
      currentPrice DECIMAL(10,2),
      previousClose VARCHAR(20),
      dividendYield VARCHAR(20),
      per VARCHAR(20),
      pbr DECIMAL(5,2),
      marketCap VARCHAR(50),
      volume INT,
      pricemovement VARCHAR(50),
      signal_val VARCHAR(50),
      memo TEXT
    )`,
    `CREATE TABLE IF NOT EXISTS watch_detail (
      ticker VARCHAR(10) NOT NULL,
      yymmdd CHAR(6) NOT NULL,
      companyName VARCHAR(255),
      currentPrice DECIMAL(10,2),
      previousClose VARCHAR(20),
      dividendYield VARCHAR(20),
      per VARCHAR(20),
      pbr DECIMAL(5,2),
      marketCap VARCHAR(50),
      volume INT,
      pricemovement VARCHAR(50),
      signal_val VARCHAR(50),
      memo TEXT,
      PRIMARY KEY (ticker, yymmdd)
    )`,
  ];
  for (const statement of statements) {
    await db.execute(statement);
  }
}

export async function seedWatchList(db, path = 'src/tickers1.tsv') {
  const content = await fs.readFile(path, 'utf8');
  let dataLineNo = 0;
  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line) {
      continue;
    }
    dataLineNo += 1;
    if (dataLineNo === 1 && line.toLowerCase().startsWith('ticker')) {
      continue;
    }
    const [tickerRaw, companyNameRaw = ''] = line.split('\t');
    const ticker = tickerRaw.trim();
    const companyName = companyNameRaw.trim();
    if (!ticker) {
      throw new Error(`empty ticker at data line ${dataLineNo}`);
    }
    await db.execute(
      'INSERT INTO watch_list (ticker, companyName) VALUES (?, ?) ON DUPLICATE KEY UPDATE companyName=VALUES(companyName)',
      [ticker, companyName],
    );
  }
}

function parseHostPort(dbHost) {
  const [host, portRaw] = dbHost.split(':');
  return {
    host: host || 'localhost',
    port: portRaw ? Number(portRaw) : 3306,
  };
}
