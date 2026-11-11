import { readdirSync, readFileSync, statSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { gzipSync } from 'node:zlib'

const webRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const assetsDir = path.join(webRoot, 'dist', 'assets')
const BUDGET_BYTES = 100 * 1024

if (!statSync(path.join(webRoot, 'dist'), { throwIfNoEntry: false })) {
  console.error('Run `npm run build` before check-bundle-size')
  process.exit(1)
}

let totalJsGzip = 0
for (const file of readdirSync(assetsDir)) {
  if (!file.endsWith('.js')) continue
  const full = path.join(assetsDir, file)
  totalJsGzip += gzipSync(readFileSync(full)).length
}

console.log(`Initial JS (gzip): ${totalJsGzip} bytes (budget ${BUDGET_BYTES})`)
if (totalJsGzip > BUDGET_BYTES) {
  process.exit(1)
}
