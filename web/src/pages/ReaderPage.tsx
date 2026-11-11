import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'

import { SentenceBlock } from '../components/SentenceBlock'
import { UnitPanel } from '../components/UnitPanel'
import { loadProgress, saveProgress } from '../lib/progress'
import { buildTokenUnitMap, unitTokenKey } from '../lib/tokenMap'
import type { Sentence, Unit, WortwerkCompiledStory } from '../types/story.generated'
import type { WortwerkStoryIndex } from '../types/index.generated'

type ActiveSelection = {
  sentenceId: string
  tokenIds: number[]
  unit: Unit
} | null

type Props = {
  slug: string
  navigate: (to: string) => void
}

export function ReaderPage({ slug, navigate }: Props) {
  const [story, setStory] = useState<WortwerkCompiledStory | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [selection, setSelection] = useState<ActiveSelection>(null)
  const [panelOpen, setPanelOpen] = useState(false)
  const [mobile, setMobile] = useState(() =>
    typeof window !== 'undefined' ? window.matchMedia('(max-width: 767px)').matches : false,
  )

  const tokenMap = useMemo(() => (story ? buildTokenUnitMap(story) : null), [story])
  const sentenceById = useMemo(() => {
    const m = new Map<string, Sentence>()
    if (story) {
      for (const s of story.sentences) m.set(s.id, s)
    }
    return m
  }, [story])

  const sentenceRefs = useRef(new Map<string, HTMLParagraphElement>())
  const observerRef = useRef<IntersectionObserver | null>(null)
  const visibleRef = useRef(new Map<string, IntersectionObserverEntry>())
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)

  useEffect(() => {
    const mq = window.matchMedia('(max-width: 767px)')
    const onChange = () => setMobile(mq.matches)
    mq.addEventListener('change', onChange)
    return () => mq.removeEventListener('change', onChange)
  }, [])

  useEffect(() => {
    let cancelled = false
    void (async () => {
      try {
        const indexRes = await fetch('/content/compiled/index.json')
        if (!indexRes.ok) throw new Error(`index HTTP ${indexRes.status}`)
        const index = (await indexRes.json()) as WortwerkStoryIndex
        const meta = index.stories.find((s) => s.slug === slug)
        if (!meta) throw new Error('Story not found')
        const storyRes = await fetch(
          `/content/compiled/${slug}.json?v=${encodeURIComponent(meta.rev)}`,
        )
        if (!storyRes.ok) throw new Error(`story HTTP ${storyRes.status}`)
        const data = (await storyRes.json()) as WortwerkCompiledStory
        if (!cancelled) setStory(data)
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Failed to load story')
      }
    })()
    return () => {
      cancelled = true
    }
  }, [slug])

  useEffect(() => {
    if (!story) return
    const saved = loadProgress().stories[slug]?.sentenceId
    if (!saved) return
    requestAnimationFrame(() => {
      sentenceRefs.current.get(saved)?.scrollIntoView({ block: 'start' })
    })
  }, [story, slug])

  const flushProgress = useCallback(() => {
    const visible = visibleRef.current
    let best: { id: string; ratio: number } | null = null
    for (const [id, entry] of visible) {
      if (!entry.isIntersecting) continue
      if (!best || entry.intersectionRatio > best.ratio) {
        best = { id, ratio: entry.intersectionRatio }
      }
    }
    if (!best) {
      for (const [id, entry] of visible) {
        if (entry.boundingClientRect.top >= 0 && entry.boundingClientRect.top < 200) {
          saveProgress({ slug, sentenceId: id })
          return
        }
      }
      return
    }
    saveProgress({ slug, sentenceId: best.id })
  }, [slug])

  useEffect(() => {
    if (!story) return
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          const id = (entry.target as HTMLElement).dataset.sentenceId
          if (!id) continue
          visibleRef.current.set(id, entry)
        }
        clearTimeout(saveTimerRef.current)
        saveTimerRef.current = setTimeout(flushProgress, 400)
      },
      { root: null, threshold: [0, 0.25, 0.5, 0.75, 1], rootMargin: '-20% 0px -60% 0px' },
    )
    observerRef.current = observer
    for (const el of sentenceRefs.current.values()) {
      observer.observe(el)
    }
    return () => {
      clearTimeout(saveTimerRef.current)
      observer.disconnect()
      observerRef.current = null
      visibleRef.current.clear()
    }
  }, [story, slug, flushProgress])

  const attachSentenceRef = useCallback((sentenceId: string, el: HTMLParagraphElement | null) => {
    const observer = observerRef.current
    const prev = sentenceRefs.current.get(sentenceId)
    if (prev && observer) observer.unobserve(prev)
    if (el) {
      sentenceRefs.current.set(sentenceId, el)
      observer?.observe(el)
    } else {
      sentenceRefs.current.delete(sentenceId)
    }
  }, [])

  const onTokenTap = useCallback(
    (sentenceId: string, tokenId: number) => {
      if (!tokenMap) return
      const ref = tokenMap.get(unitTokenKey(sentenceId, tokenId))
      if (!ref) return
      const tokenIds = [...ref.unit.tokens].sort((a, b) => a - b)
      setSelection({ sentenceId, tokenIds, unit: ref.unit })
      setPanelOpen(true)
    },
    [tokenMap],
  )

  const panelSelection = useMemo(() => {
    if (!selection || !story || !panelOpen) return null
    const sentence = sentenceById.get(selection.sentenceId)
    const entry = story.lexicon[selection.unit.lex]
    if (!sentence || !entry) return null
    return {
      sentenceId: selection.sentenceId,
      unit: selection.unit,
      entry,
      sentenceTranslation: sentence.translation,
    }
  }, [selection, story, panelOpen, sentenceById])

  if (error) {
    return (
      <div className="mx-auto max-w-prose px-4 py-10">
        <p className="text-red-600" role="alert">
          {error}
        </p>
        <button type="button" className="mt-4 text-sm underline" onClick={() => navigate('/')}>
          Back to library
        </button>
      </div>
    )
  }

  if (!story) {
    return (
      <div className="mx-auto max-w-prose px-4 py-10">
        <p className="text-zinc-500">Loading…</p>
      </div>
    )
  }

  const paragraphs = new Map<number, Sentence[]>()
  for (const s of story.sentences) {
    const list = paragraphs.get(s.p) ?? []
    list.push(s)
    paragraphs.set(s.p, list)
  }

  return (
    <article className="mx-auto min-h-dvh max-w-prose px-4 py-8 [max-width:38ch]">
      <header className="mb-8">
        <button
          type="button"
          className="mb-4 text-sm text-zinc-500 hover:text-zinc-800 dark:hover:text-zinc-200"
          onClick={() => navigate('/')}
        >
          ← Library
        </button>
        <h1 className="text-2xl font-semibold tracking-tight">{story.title}</h1>
        <p className="mt-1 text-sm text-zinc-500">
          {story.level} · {story.wordCount} words
        </p>
      </header>

      <StoryBody
        paragraphs={paragraphs}
        selection={selection}
        onTokenTap={onTokenTap}
        attachSentenceRef={attachSentenceRef}
      />

      <UnitPanel
        open={panelOpen}
        selection={panelSelection}
        mobile={mobile}
        onClose={() => setPanelOpen(false)}
      />
    </article>
  )
}

function StoryBody({
  paragraphs,
  selection,
  onTokenTap,
  attachSentenceRef,
}: {
  paragraphs: Map<number, Sentence[]>
  selection: ActiveSelection
  onTokenTap: (sentenceId: string, tokenId: number) => void
  attachSentenceRef: (sentenceId: string, el: HTMLParagraphElement | null) => void
}) {
  return (
    <TapDelegate
      onTokenTap={onTokenTap}
      className="text-[1.0625rem] leading-loose text-zinc-900 select-none dark:text-zinc-100"
    >
      {[...paragraphs.entries()]
        .sort(([a], [b]) => a - b)
        .map(([p, sentences]) => (
          <section key={p} className="mb-6">
            {sentences.map((sentence) => {
              const highlightIds =
                selection && selection.sentenceId === sentence.id ? selection.tokenIds : null
              return (
                <SentenceBlock
                  key={sentence.id}
                  sentence={sentence}
                  highlightIds={highlightIds}
                  onTokenTap={onTokenTap}
                  sentenceRef={(el) => attachSentenceRef(sentence.id, el)}
                />
              )
            })}
          </section>
        ))}
    </TapDelegate>
  )
}

function TapDelegate({
  children,
  className,
  onTokenTap,
}: {
  children: ReactNode
  className?: string
  onTokenTap: (sentenceId: string, tokenId: number) => void
}) {
  return (
    <div
      className={className}
      onClick={(e) => {
        const target = (e.target as HTMLElement).closest<HTMLElement>('[data-s][data-t]')
        if (!target?.dataset.s || target.dataset.t === undefined) return
        onTokenTap(target.dataset.s, Number(target.dataset.t))
      }}
    >
      {children}
    </div>
  )
}
