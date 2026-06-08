import { numericValue } from './parser.js';

export function renderList(items) {
  return `<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>JPX Smallcap Watcher</title>
  <link rel="stylesheet" href="static/style.css">
  <link rel="icon" type="image/svg+xml" href="static/favicon.svg">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
  <div class="container">
    <header>
      <h1>JPX Smallcap Watcher</h1>
      <p class="subtitle">Small Cap Stock Tracking Dashboard</p>
    </header>

    <main>
      <div class="card">
        <table id="watch-table">
          <thead>
            <tr>
              <th data-type="text">Ticker</th>
              <th data-type="text">Company</th>
              <th data-type="number">Price</th>
              <th data-type="text">Prev Close</th>
              <th class="col-volume" data-type="number">Vol</th>
              <th data-type="number">Change</th>
              <th class="col-signal" data-type="text">Signal</th>
              <th data-type="number">PER</th>
              <th data-type="number">PBR</th>
              <th data-type="number">Yield</th>
            </tr>
          </thead>
          <tbody>
${items.map(renderListRow).join('\n')}
          </tbody>
        </table>
      </div>
    </main>

    <footer>
      <p>&copy; 2026 IchikabuImpact</p>
    </footer>
  </div>

  <script>
    const table = document.getElementById('watch-table');
    const headers = table.querySelectorAll('th');
    let sortIndex = -1;
    let sortAsc = true;

    headers.forEach((header, index) => {
      header.addEventListener('click', () => {
        const type = header.dataset.type || 'text';
        sortAsc = sortIndex === index ? !sortAsc : true;
        sortIndex = index;

        const rows = Array.from(table.tBodies[0].rows);
        rows.sort((a, b) => {
          const aCell = a.cells[index];
          const bCell = b.cells[index];
          const aValue = aCell.dataset.value || aCell.textContent.trim();
          const bValue = bCell.dataset.value || bCell.textContent.trim();

          if (type === 'number') {
            const aNum = aValue === '' ? NaN : Number(aValue);
            const bNum = bValue === '' ? NaN : Number(bValue);
            if (Number.isNaN(aNum) && Number.isNaN(bNum)) return 0;
            if (Number.isNaN(aNum)) return 1;
            if (Number.isNaN(bNum)) return -1;
            return sortAsc ? aNum - bNum : bNum - aNum;
          }

          const aText = aValue.toLowerCase();
          const bText = bValue.toLowerCase();
          if (aText === bText) return 0;
          return sortAsc ? (aText > bText ? 1 : -1) : (aText > bText ? -1 : 1);
        });

        rows.forEach((row) => table.tBodies[0].appendChild(row));
        headers.forEach((h) => h.classList.remove('sorted-asc', 'sorted-desc'));
        header.classList.add(sortAsc ? 'sorted-asc' : 'sorted-desc');
      });
    });
  </script>
</body>
</html>
`;
}

export function renderDetail(view) {
  return `<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>${escapeHtml(view.ticker)} - JPX Smallcap Watcher</title>
  <link rel="stylesheet" href="../static/style.css">
  <link rel="icon" type="image/svg+xml" href="../static/favicon.svg">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body>
  <div class="container">
    <header>
      <a class="back-link" href="../index.html">&larr; Back to List</a>
      <h1>${escapeHtml(view.companyName)} <span class="ticker-header">${escapeHtml(view.ticker)}</span></h1>
    </header>

    <main>
      <div class="summary-grid">
        <div class="card summary-item">
          <h3>Current Price</h3>
          <p class="price-large">${formatFloat(view.currentPrice)}</p>
        </div>
        <div class="card summary-item">
          <h3>Prev Close</h3>
          <p>${escapeHtml(view.previousClose)}</p>
        </div>
        <div class="card summary-item">
          <h3>Signal</h3>
          <span class="badge ${signalClass(view.signal)}">${escapeHtml(view.signal || '-')}</span>
        </div>
        <div class="card summary-item">
          <h3>Market Cap</h3>
          <p>${escapeHtml(view.marketCap)}</p>
        </div>
      </div>

      <div class="card">
        <h2>History</h2>
        <table>
          <thead>
            <tr>
              <th>Date</th>
              <th>Price</th>
              <th>Change</th>
              <th>Signal</th>
              <th>Volume</th>
            </tr>
          </thead>
          <tbody>
${view.items.map(renderDetailRow).join('\n')}
          </tbody>
        </table>
      </div>
    </main>

    <footer>
      <p>&copy; 2026 IchikabuImpact</p>
    </footer>
  </div>
</body>
</html>
`;
}

function renderListRow(item) {
  return `            <tr onclick="window.location.href='detail/${escapeAttribute(item.ticker)}.html'" class="clickable-row">
              <td class="ticker" data-value="${escapeAttribute(item.ticker)}">${escapeHtml(item.ticker)}</td>
              <td class="company" data-value="${escapeAttribute(item.companyName)}">${escapeHtml(item.companyName)}</td>
              <td class="price" data-value="${dataNumber(item.currentPrice)}">${formatFloat(item.currentPrice)}</td>
              <td data-value="${escapeAttribute(item.previousClose)}">${escapeHtml(item.previousClose)}</td>
              <td class="col-volume" data-value="${dataNumber(item.volume)}">${formatInt(item.volume)}</td>
              <td class="change ${changeClass(item.pricemovement)}" data-value="${escapeAttribute(numericValue(item.pricemovement))}">${formatChange(item.pricemovement)}</td>
              <td class="signal col-signal ${signalClass(item.signal_val)}" data-value="${escapeAttribute(item.signal_val)}">
                <span class="badge">${escapeHtml(item.signal_val)}</span>
              </td>
              <td data-value="${escapeAttribute(numericValue(item.per))}">${formatPER(item.per)}</td>
              <td data-value="${dataNumber(item.pbr)}">${formatPBR(item.pbr)}</td>
              <td data-value="${escapeAttribute(numericValue(item.dividendYield))}">${formatYield(item.dividendYield)}</td>
            </tr>`;
}

function renderDetailRow(item) {
  return `            <tr>
              <td>${escapeHtml(item.date)}</td>
              <td>${formatFloat(item.currentPrice)}</td>
              <td class="change ${changeClass(item.pricemovement)}">${formatChange(item.pricemovement)}</td>
              <td class="signal ${signalClass(item.signal_val)}"><span class="badge">${escapeHtml(item.signal_val)}</span></td>
              <td>${formatInt(item.volume)}</td>
            </tr>`;
}

export function formatFloat(value) {
  if (value === null || value === undefined || value === '') {
    return '-';
  }
  const numeric = Number(value);
  return Number.isNaN(numeric) ? '-' : numeric.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

export function formatInt(value) {
  if (value === null || value === undefined || value === '') {
    return '-';
  }
  const numeric = Number(value);
  return Number.isNaN(numeric) ? '-' : Math.trunc(numeric).toLocaleString('en-US');
}

function formatPBR(value) {
  return formatFloat(value);
}

function formatPER(value) {
  const val = String(value ?? '').trim();
  return val || '-';
}

function formatYield(value) {
  const val = String(value ?? '').trim();
  return val || '-';
}

function formatChange(value) {
  const text = String(value ?? '').trim();
  if (!text) {
    return '';
  }
  const numeric = Number(text.replace('%', '').replace(/^\+/, '').trim());
  if (Number.isNaN(numeric)) {
    return escapeHtml(text);
  }
  if (numeric > 0) {
    return `+${numeric.toFixed(2)}%`;
  }
  if (numeric < 0) {
    return `${numeric.toFixed(2)}%`;
  }
  return '+0.00%';
}

function changeClass(value) {
  const numeric = Number(String(value ?? '').replace('%', '').replace(/^\+/, '').trim());
  if (Number.isNaN(numeric) || numeric === 0) {
    return 'neutral';
  }
  return numeric > 0 ? 'positive' : 'negative';
}

function signalClass(value) {
  switch (String(value ?? '').toLowerCase()) {
    case 'buy':
      return 'buy';
    case 'sell':
      return 'sell';
    default:
      return 'neutral';
  }
}

function dataNumber(value) {
  return value === null || value === undefined || value === '' ? '' : escapeAttribute(String(Number(value)));
}

export function escapeHtml(value) {
  return String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

function escapeAttribute(value) {
  return escapeHtml(value);
}
