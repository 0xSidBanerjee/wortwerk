import { useEffect, useState } from 'react'

import type { WortwerkStoryIndex } from './types/index.generated'

function App() {
  const [index, setIndex] = useState<WortwerkStoryIndex | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    void (async () => {
      try {
        const res = await fetch('/content/compiled/index.json')
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`)
        }
        const data = (await res.json()) as WortwerkStoryIndex
        if (!cancelled) {
          setIndex(data)
        }
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : 'Failed to load library')
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="mx-auto min-h-dvh max-w-lg px-4 py-10">
      <header className="mb-10">
        <h1 className="text-2xl font-semibold tracking-tight">Wortwerk</h1>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Read German. Tap any word for its meaning.
        </p>
      </header>

      <main>
        <h2 className="mb-4 text-sm font-medium uppercase tracking-wide text-zinc-500 dark:text-zinc-400">
          Library
        </h2>

        {error !== null && (
          <p className="text-sm text-red-600 dark:text-red-400" role="alert">
            {error}
          </p>
        )}

        {index === null && error === null && (
          <p className="text-sm text-zinc-500">Loading stories…</p>
        )}

        {index !== null && index.stories.length === 0 && (
          <p className="text-sm text-zinc-600 dark:text-zinc-300">
            No stories yet. The first compiled stories arrive in milestone M1/M4.
          </p>
        )}

        {index !== null && index.stories.length > 0 && (
          <ul className="divide-y divide-zinc-200 dark:divide-zinc-800">
            {index.stories.map((story) => (
              <li key={story.slug} className="py-4 first:pt-0">
                <span className="font-medium">{story.title}</span>
                <span className="ml-2 text-xs text-zinc-500">
                  {story.level} · {story.minutes} min
                </span>
              </li>
            ))}
          </ul>
        )}
      </main>
    </div>
  )
}

export default App
