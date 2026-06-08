import fs from 'node:fs/promises';
import path from 'node:path';
import { fetchStockData, sleep } from './api.js';
import { parseNumeric, parsePreviousClose } from './parser.js';
import { renderDetail, renderList } from './render.js';

export async function runBatch(db, scraperBaseUrl, requestIntervalMs) {
  const [rows] = await db.query('SELECT ticker FROM watch_list');
  for (let i = 0; i < rows.length; i += 1) {
    const ticker = rows[i].ticker;
    if (i > 0 && requestIntervalMs > 0) {
      await sleep(requestIntervalMs);
    }
    try {
      const payload = await fetchStockData(scraperBaseUrl, ticker);
      await updateStock(db, payload);
    } catch (error) {
      console.error(`failed to update ${ticker}: ${error.message}`);
    }
    await sleep(batchDelayMs());
  }
}

export async function generateHtml(db, outputDir) {
  await fs.mkdir(path.join(outputDir, 'detail'), { recursive: true });
  await fs.mkdir(path.join(outputDir, 'static'), { recursive: true });
  await copyStaticAssets('static', path.join(outputDir, 'static'));

  const items = await loadListItems(db);
  await fs.writeFile(path.join(outputDir, 'index.html'), renderList(items), 'utf8');
  const indexInfo = await fs.stat(path.join(outputDir, 'index.html'));
  console.log(`index generated path=${path.join(outputDir, 'index.html')} mtime=${indexInfo.mtime.toISOString()} size=${indexInfo.size}`);

  for (const item of items) {
    const detailItems = await loadDetailItems(db, item.ticker);
    const html = renderDetail({
      ticker: item.ticker,
      companyName: item.companyName,
      currentPrice: item.currentPrice,
      previousClose: item.previousClose,
      signal: item.signal_val,
      marketCap: item.marketCap,
      items: detailItems,
    });
    await fs.writeFile(path.join(outputDir, 'detail', `${item.ticker}.html`), html, 'utf8');
  }

  const newestDetail = await newestDetailFileInfo(outputDir);
  if (indexInfo.mtimeMs < newestDetail.info.mtimeMs - 60_000) {
    throw new Error(`generated index is older than latest detail page (index=${indexInfo.mtime.toISOString()} detail=${newestDetail.info.mtime.toISOString()} detail_path=${newestDetail.path})`);
  }
}

async function updateStock(db, payload) {
  const current = parseNumeric(payload.currentPrice);
  const pbr = parseNumeric(payload.pbr);
  const volume = parseNumeric(payload.volume);
  const previousClose = parsePreviousClose(payload.previousClose);

  let pricemovement = '';
  let signal = 'Neutral';
  if (current.ok && previousClose.ok && previousClose.value !== 0) {
    const movement = ((current.value - previousClose.value) / previousClose.value) * 100;
    pricemovement = `${movement.toFixed(2)}%`;
    signal = generateSignal(movement);
  }

  const yymmdd = formatYYMMDD(new Date());
  const values = [
    payload.ticker,
    payload.companyName,
    current.ok ? current.value : null,
    payload.previousClose,
    payload.dividendYield,
    payload.per,
    pbr.ok ? pbr.value : null,
    payload.marketCap,
    volume.ok ? Math.trunc(volume.value) : null,
    pricemovement,
    signal,
  ];

  await db.execute(
    `INSERT INTO watch_list (
      ticker, companyName, currentPrice, previousClose, dividendYield, per, pbr, marketCap, volume, pricemovement, signal_val
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON DUPLICATE KEY UPDATE
      companyName=VALUES(companyName), currentPrice=VALUES(currentPrice), previousClose=VALUES(previousClose),
      dividendYield=VALUES(dividendYield), per=VALUES(per), pbr=VALUES(pbr), marketCap=VALUES(marketCap),
      volume=VALUES(volume), pricemovement=VALUES(pricemovement), signal_val=VALUES(signal_val)`,
    values,
  );

  await db.execute(
    `REPLACE INTO watch_detail (
      ticker, yymmdd, companyName, currentPrice, previousClose, dividendYield, per, pbr, marketCap, volume, pricemovement, signal_val
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    [payload.ticker, yymmdd, ...values.slice(1)],
  );
}

async function loadListItems(db) {
  const [rows] = await db.query('SELECT ticker, companyName, currentPrice, previousClose, dividendYield, per, pbr, marketCap, volume, pricemovement, signal_val FROM watch_list ORDER BY ticker');
  return rows.map(normalizeRow);
}

async function loadDetailItems(db, ticker) {
  const [rows] = await db.execute(
    'SELECT yymmdd, currentPrice, previousClose, dividendYield, per, pbr, marketCap, volume, pricemovement, signal_val FROM watch_detail WHERE ticker = ? ORDER BY yymmdd DESC LIMIT 90',
    [ticker],
  );
  return rows.map((row) => ({ ...normalizeRow(row), date: formatDisplayDate(row.yymmdd) }));
}

async function copyStaticAssets(srcDir, destDir) {
  try {
    await fs.cp(srcDir, destDir, { recursive: true, force: true });
  } catch (error) {
    if (error.code !== 'ENOENT') {
      throw error;
    }
  }
}

async function newestDetailFileInfo(outputDir) {
  const detailDir = path.join(outputDir, 'detail');
  const entries = await fs.readdir(detailDir, { withFileTypes: true });
  const files = entries.filter((entry) => entry.isFile() && entry.name.endsWith('.html'));
  if (files.length === 0) {
    throw new Error(`no detail files found under ${detailDir}`);
  }
  let newest;
  for (const file of files) {
    const filePath = path.join(detailDir, file.name);
    const info = await fs.stat(filePath);
    if (!newest || info.mtimeMs > newest.info.mtimeMs) {
      newest = { path: filePath, info };
    }
  }
  return newest;
}

function normalizeRow(row) {
  return {
    ticker: row.ticker ?? '',
    companyName: row.companyName ?? '',
    currentPrice: row.currentPrice ?? null,
    previousClose: row.previousClose ?? '',
    dividendYield: row.dividendYield ?? '',
    per: row.per ?? '',
    pbr: row.pbr ?? null,
    marketCap: row.marketCap ?? '',
    volume: row.volume ?? null,
    pricemovement: row.pricemovement ?? '',
    signal_val: row.signal_val ?? '',
  };
}

function generateSignal(movement) {
  if (movement > 3.0) return 'Buy';
  if (movement < -3.0) return 'Sell';
  return 'Neutral';
}

function batchDelayMs() {
  return 2000 + Math.floor(Math.random() * 701);
}

function formatYYMMDD(date) {
  return `${String(date.getFullYear()).slice(2)}${String(date.getMonth() + 1).padStart(2, '0')}${String(date.getDate()).padStart(2, '0')}`;
}

function formatDisplayDate(yymmdd) {
  const value = String(yymmdd ?? '');
  return value.length === 6 ? `20${value.slice(0, 2)}-${value.slice(2, 4)}-${value.slice(4, 6)}` : value;
}
