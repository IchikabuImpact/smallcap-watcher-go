const MAX_RETRIES = 6;
const RETRY_BASE_DELAY_MS = 1000;
const MAX_RETRY_DELAY_MS = 4000;
const RETRY_JITTER_MAX_MS = 250;
const RETRY_STATUSES = new Set([502, 503, 504]);

export async function fetchStockData(baseUrl, ticker) {
  const url = `${baseUrl}/scrape?ticker=${encodeURIComponent(ticker)}`;
  let lastError;
  for (let attempt = 0; attempt <= MAX_RETRIES; attempt += 1) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 15_000);
    try {
      const response = await fetch(url, { signal: controller.signal });
      const body = await response.text();
      if (!response.ok) {
        lastError = new Error(`unexpected status: ${response.status} ${response.statusText} for ${url} (body: ${body.slice(0, 2048).trim()})`);
        if (!shouldRetry(attempt, response.status)) {
          break;
        }
        await sleep(retryDelayMs(attempt, response.headers.get('Retry-After')));
        continue;
      }
      return JSON.parse(body);
    } catch (error) {
      lastError = error;
      if (!shouldRetry(attempt, 0, error)) {
        break;
      }
      await sleep(retryDelayMs(attempt));
    } finally {
      clearTimeout(timeout);
    }
  }
  throw lastError ?? new Error(`request failed for ${url}`);
}

function shouldRetry(attempt, statusCode, error) {
  if (attempt >= MAX_RETRIES) {
    return false;
  }
  if (statusCode) {
    return RETRY_STATUSES.has(statusCode);
  }
  return error?.name === 'AbortError' || ['ETIMEDOUT', 'ECONNRESET', 'ECONNREFUSED', 'EAI_AGAIN'].includes(error?.cause?.code);
}

function retryDelayMs(attempt, retryAfterHeader) {
  const retryAfterMs = parseRetryAfterMs(retryAfterHeader);
  const exponential = Math.min(RETRY_BASE_DELAY_MS * 2 ** attempt, MAX_RETRY_DELAY_MS);
  return Math.max(exponential, retryAfterMs) + Math.floor(Math.random() * (RETRY_JITTER_MAX_MS + 1));
}

function parseRetryAfterMs(value) {
  if (!value) {
    return 0;
  }
  const seconds = Number(value.trim());
  if (!Number.isNaN(seconds) && seconds > 0) {
    return seconds * 1000;
  }
  const dateMs = Date.parse(value);
  return Number.isNaN(dateMs) ? 0 : Math.max(0, dateMs - Date.now());
}

export function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
