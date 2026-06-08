#!/usr/bin/env node
import { loadConfig } from '../lib/config.js';
import { initSchema, openDatabase, seedWatchList } from '../lib/db.js';
import { verifyIndexFreshness } from '../lib/freshness.js';
import { generateHtml, runBatch } from '../lib/service.js';

const FLAGS = new Set(['--init', '--batch', '--gen', '--seed']);
const HELP_FLAGS = new Set(['--help', '-h']);
const args = process.argv.slice(2);
const selected = args.filter((arg) => FLAGS.has(arg));

if (args.some((arg) => HELP_FLAGS.has(arg))) {
  console.log('Usage: node ./bin/smallcap-watcher.js --init | --seed | --batch | --gen');
  process.exit(0);
}

if (selected.length === 0 || args.some((arg) => !FLAGS.has(arg))) {
  console.error('Usage: node ./bin/smallcap-watcher.js --init | --seed | --batch | --gen');
  process.exit(1);
}

let db;
try {
  const config = loadConfig();
  db = await openDatabase(config);

  if (args.includes('--init')) {
    await initSchema(db);
    console.log('schema initialized');
  }
  if (args.includes('--seed')) {
    await seedWatchList(db, 'src/tickers1.tsv');
    console.log('watch list seeded');
  }
  if (args.includes('--batch')) {
    await runBatch(db, config.scraperBaseUrl, config.scraperRequestIntervalMs);
    console.log('batch completed');
  }
  if (args.includes('--gen')) {
    await generateHtml(db, config.outputDir);
    await verifyIndexFreshness(config.outputDir, config.indexMaxAgeMs);
    console.log('HTML generated');
  }
} catch (error) {
  console.error(error?.stack || error?.message || String(error));
  process.exitCode = 1;
} finally {
  if (db) {
    await db.end();
  }
}
