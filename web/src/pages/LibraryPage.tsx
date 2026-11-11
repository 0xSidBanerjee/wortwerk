import { useEffect, useState } from 'react'

import { loadProgress } from '../lib/progress'
import type { WortwerkStoryIndex } from '../types/index.generated'

type Props = {
  navigate: (to: string) => void
}

export function LibraryPage({ navigate }: Props) {
  const [index, setIndex] = useState<WortwerkStoryIndex | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [lastStory, setLastStory] = useState('')

  useEffect(() => {
    setLastStory(loadProgress().lastStory)
    let cancelled = false
    void (async () => {
      try {
        const res = await fetch('/content/compiled/index.json')
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = (await res.json()) as WortwerkStoryIndex
        if (!cancelled) setIndex(data)
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Failed to load library')
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
          <p className="text-sm text-zinc-600 dark:text-zinc-300">No stories yet.</p>
        )}

        {index !== null && index.stories.length > 0 && (
          <ul className="divide-y divide-zinc-200 dark:divide-zinc-800">
            {index.stories.map((story) => {
              const showContinue = lastStory === story.slug
              return (
                <li key={story.slug} className="flex items-baseline justify-between gap-3 py-4 first:pt-0">
                  <button
                    type="button"
                    className="text-left font-medium underline-offset-2 hover:underline"
                    onClick={() => navigate(`/s/${story.slug}`)}
                  >
                    {story.title}
                  </button>
                  <div className="shrink-0 text-right text-xs text-zinc-500">
                    <span>
                      {story.level} · {story.minutes} min
                    </span>
                    {showContinue && (
                      <span className="mt-0.5 block font-medium text-zinc-700 dark:text-zinc-300">
                        Continue
                      </span>
                    )}
                  </div>
                </li>
              )
            })}
          </ul>
        )}
      </main>
    </div>
  )
}
