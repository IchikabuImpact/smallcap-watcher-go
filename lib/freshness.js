import fs from 'node:fs/promises';
import path from 'node:path';

export async function verifyIndexFreshness(outputDir, maxAgeMs) {
  const indexPath = path.join(outputDir, 'index.html');
  const indexInfo = await fs.stat(indexPath).catch((error) => {
    throw new Error(`stat index ${indexPath}: ${error.message}`);
  });
  const detail = await newestDetailFileInfo(outputDir);
  const indexAgeMs = Date.now() - indexInfo.mtimeMs;

  console.log(`output healthcheck index_path=${indexPath} index_mtime=${indexInfo.mtime.toISOString()} index_size=${indexInfo.size} index_age=${Math.round(indexAgeMs / 1000)}s detail_path=${detail.path} detail_mtime=${detail.info.mtime.toISOString()} detail_size=${detail.info.size}`);

  if (indexInfo.size === 0) {
    throw new Error(`index file is empty: ${indexPath}`);
  }
  if (maxAgeMs > 0 && indexAgeMs > maxAgeMs) {
    throw new Error(`index file too old (age=${Math.round(indexAgeMs / 1000)}s max_age=${Math.round(maxAgeMs / 1000)}s path=${indexPath})`);
  }
  if (indexInfo.mtimeMs < detail.info.mtimeMs - 60_000) {
    throw new Error(`index is older than latest detail page (index=${indexInfo.mtime.toISOString()} detail=${detail.info.mtime.toISOString()})`);
  }
}

async function newestDetailFileInfo(outputDir) {
  const detailDir = path.join(outputDir, 'detail');
  const entries = await fs.readdir(detailDir, { withFileTypes: true }).catch((error) => {
    throw new Error(`read detail dir ${detailDir}: ${error.message}`);
  });
  const htmlFiles = entries.filter((entry) => entry.isFile() && entry.name.endsWith('.html'));
  if (htmlFiles.length === 0) {
    throw new Error(`no detail html files in ${detailDir}`);
  }
  let newest;
  for (const file of htmlFiles) {
    const filePath = path.join(detailDir, file.name);
    const info = await fs.stat(filePath);
    if (!newest || info.mtimeMs > newest.info.mtimeMs) {
      newest = { path: filePath, info };
    }
  }
  return newest;
}
