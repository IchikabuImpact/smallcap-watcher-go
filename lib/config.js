import fs from 'node:fs';

const DURATION_UNITS = {
  ms: 1,
  s: 1000,
  m: 60_000,
  h: 3_600_000,
};

export function loadConfig(path = 'env.config') {
  loadConfigFile(path);
  return {
    dbHost: getEnv('DB_HOST', 'localhost:3306'),
    dbUser: getEnv('DB_USER', 'jpx_user'),
    dbPassword: getEnv('DB_PASSWORD', 'jpx_password'),
    dbName: getEnv('DB_NAME', 'jpx_data'),
    scraperBaseUrl: stripTrailingSlash(getEnv('SCRAPER_BASE_URL', 'http://localhost:8085')),
    scraperRequestIntervalMs: getEnvDurationMs('SCRAPER_REQUEST_INTERVAL', 3000),
    outputDir: getEnv('OUTPUT_DIR', 'public'),
    indexMaxAgeMs: getEnvDurationMs('INDEX_MAX_AGE', 36 * 3_600_000),
  };
}

export function loadConfigFile(path) {
  if (!fs.existsSync(path)) {
    return;
  }
  const content = fs.readFileSync(path, 'utf8');
  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith('#')) {
      continue;
    }
    const index = line.indexOf('=');
    if (index < 0) {
      continue;
    }
    const key = line.slice(0, index).trim();
    const value = line.slice(index + 1).trim();
    if (key && process.env[key] === undefined) {
      process.env[key] = value;
    }
  }
}

export function parseDurationMs(value) {
  const trimmed = String(value ?? '').trim();
  if (!trimmed) {
    return undefined;
  }
  const match = trimmed.match(/^(\d+(?:\.\d+)?)(ms|s|m|h)$/);
  if (!match) {
    return undefined;
  }
  return Number(match[1]) * DURATION_UNITS[match[2]];
}

function getEnv(key, fallback) {
  return process.env[key] && process.env[key].trim() !== '' ? process.env[key] : fallback;
}

function getEnvDurationMs(key, fallback) {
  return parseDurationMs(process.env[key]) ?? fallback;
}

function stripTrailingSlash(value) {
  return value.trim().replace(/\/+$/, '');
}
