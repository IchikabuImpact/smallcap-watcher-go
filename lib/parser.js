export function parseNumeric(value) {
  let clean = String(value ?? '').trim();
  if (!clean) {
    return { value: 0, ok: false };
  }

  let multiplier = 1;
  if (clean.includes('兆')) {
    multiplier = 1e12;
    clean = clean.replaceAll('兆', '');
  } else if (clean.includes('億')) {
    multiplier = 1e8;
    clean = clean.replaceAll('億', '');
  } else if (clean.includes('万')) {
    multiplier = 1e4;
    clean = clean.replaceAll('万', '');
  }

  clean = clean
    .replaceAll('円', '')
    .replaceAll('%', '')
    .replaceAll('倍', '')
    .replaceAll('株', '')
    .replaceAll(',', '')
    .trim();

  if (!clean) {
    return { value: 0, ok: false };
  }
  const numeric = Number(clean);
  if (Number.isNaN(numeric)) {
    return { value: 0, ok: false };
  }
  return { value: numeric * multiplier, ok: true };
}

export function parsePreviousClose(value) {
  const match = String(value ?? '').match(/([0-9][0-9,]*(?:\.[0-9]+)?)/);
  return match ? parseNumeric(match[1]) : { value: 0, ok: false };
}

export function numericValue(value) {
  const parsed = parseNumeric(value);
  return parsed.ok ? String(parsed.value) : '';
}
