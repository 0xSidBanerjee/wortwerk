import path from 'node:path'
import { fileURLToPath } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import sirv from 'sirv'
import { defineConfig, type Plugin } from 'vite'

const repoRoot = fileURLToPath(new URL('..', import.meta.url))

function contentStatic(): Plugin {
  const mount = sirv(path.join(repoRoot, 'content'), {
    dev: true,
    etag: true,
  })
  return {
    name: 'wortwerk-content-static',
    configureServer(server) {
      server.middlewares.use('/content', mount)
    },
    configurePreviewServer(server) {
      server.middlewares.use('/content', mount)
    },
  }
}

export default defineConfig({
  plugins: [react(), tailwindcss(), contentStatic()],
  server: {
    fs: { allow: [repoRoot] },
  },
})
