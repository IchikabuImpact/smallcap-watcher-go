import assert from 'node:assert/strict';
import { test } from 'node:test';
import { parseDurationMs } from '../lib/config.js';
import { parseNumeric, parsePreviousClose } from '../lib/parser.js';
import { escapeHtml, formatFloat, formatInt } from '../lib/render.js';

test('parseNumeric handles Japanese market units and decorations', () => {
  assert.deepEqual(parseNumeric('1,234円'), { value: 1234, ok: true });
  assert.deepEqual(parseNumeric('12.3億'), { value: 1_230_000_000, ok: true });
  assert.deepEqual(parseNumeric('4.5万株'), { value: 45_000, ok: true });
  assert.deepEqual(parseNumeric('-'), { value: 0, ok: false });
});

test('parsePreviousClose extracts the first numeric value', () => {
  assert.deepEqual(parsePreviousClose('前日終値 1,234円'), { value: 1234, ok: true });
});

test('duration parsing mirrors env.config values', () => {
  assert.equal(parseDurationMs('500ms'), 500);
  assert.equal(parseDurationMs('3s'), 3000);
  assert.equal(parseDurationMs('36h'), 129_600_000);
  assert.equal(parseDurationMs('bad'), undefined);
});

test('HTML formatting helpers handle empty values safely', () => {
  assert.equal(formatFloat('1234'), '1,234.00');
  assert.equal(formatInt(1234.9), '1,234');
  assert.equal(formatFloat(null), '-');
  assert.equal(escapeHtml('<x>&"\''), '&lt;x&gt;&amp;&quot;&#39;');
});
