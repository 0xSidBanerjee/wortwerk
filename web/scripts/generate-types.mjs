import { mkdir, writeFile } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import { compileFromFile } from 'json-schema-to-typescript'

const webRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const repoRoot = path.join(webRoot, '..')
const outDir = path.join(webRoot, 'src', 'types')

await mkdir(outDir, { recursive: true })

const banner =
  '/* eslint-disable */\n/** Auto-generated — run `npm run generate-types`. Do not edit. */\n\n'

for (const [schema, outName] of [
  ['story.schema.json', 'story.generated.ts'],
  ['index.schema.json', 'index.generated.ts'],
]) {
  const ts = await compileFromFile(path.join(repoRoot, 'schema', schema), {
    bannerComment: banner,
    additionalProperties: false,
  })
  await writeFile(path.join(outDir, outName), ts)
}

console.log('Wrote generated types to src/types/')
