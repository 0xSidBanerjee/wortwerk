import type { ReactNode } from 'react'

import type { LexiconEntry, Unit } from '../types/story.generated'

type Selection = {
  sentenceId: string
  unit: Unit
  entry: LexiconEntry
  sentenceTranslation: string
}

type Props = {
  open: boolean
  selection: Selection | null
  mobile: boolean
  onClose: () => void
}

export function UnitPanel({ open, selection, mobile, onClose }: Props) {
  if (!open || !selection) return null

  const { unit, entry, sentenceTranslation } = selection
  const headline = entry.lemma
  const meaning = unit.gloss?.trim() || entry.meaning

  return (
    <PanelOverlay onClose={onClose}>
      <div
        className={
          mobile
            ? 'relative max-h-[70dvh] w-full overflow-y-auto rounded-t-2xl border border-zinc-200 bg-white px-4 py-5 shadow-lg dark:border-zinc-700 dark:bg-zinc-900'
            : 'relative w-full max-w-sm overflow-y-auto rounded-xl border border-zinc-200 bg-white px-4 py-5 shadow-xl dark:border-zinc-700 dark:bg-zinc-900'
        }
        role="dialog"
        aria-modal="true"
        aria-label={headline}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => e.stopPropagation()}
      >
        <PanelHeader headline={headline} onClose={onClose} />
        <p className="text-base text-zinc-800 dark:text-zinc-100">{meaning}</p>
        <span className="mt-2 inline-block rounded-full bg-zinc-100 px-2 py-0.5 text-xs text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300">
          {entry.pos}
        </span>

        {entry.kind === 'compound' && entry.parts && entry.parts.length > 0 && (
          <p className="mt-3 text-sm text-zinc-600 dark:text-zinc-400">
            {entry.parts.join(' + ')} = {entry.lemma}
          </p>
        )}

        <button
          type="button"
          disabled
          className="mt-4 w-full rounded-lg border border-zinc-200 py-2 text-sm font-medium text-zinc-400 dark:border-zinc-700"
          aria-disabled="true"
        >
          Listen (M5)
        </button>

        <details className="mt-4">
          <summary className="cursor-pointer text-sm font-medium text-zinc-600 dark:text-zinc-400">
            Show more
          </summary>
          <dl className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            <ShowRow label="Base form" value={entry.lemma} />
            {entry.article && entry.gender && (
              <ShowRow label="Article" value={`${entry.article} (${entry.gender})`} />
            )}
            {entry.plural !== undefined && (
              <ShowRow label="Plural" value={entry.plural === null ? '—' : entry.plural} />
            )}
            {unit.form && <ShowRow label="Role here" value={unit.form} />}
            <ShowRow label="Sentence" value={sentenceTranslation} />
          </dl>
        </details>
      </div>
    </PanelOverlay>
  )
}

function PanelOverlay({
  children,
  onClose,
}: {
  children: ReactNode
  onClose: () => void
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center md:items-center md:p-4"
      role="presentation"
      onClick={onClose}
      onKeyDown={(e) => {
        if (e.key === 'Escape') onClose()
      }}
    >
      <div className="absolute inset-0 bg-black/40 md:bg-black/25" aria-hidden />
      {children}
    </div>
  )
}

function PanelHeader({ headline, onClose }: { headline: string; onClose: () => void }) {
  return (
    <div className="mb-3 flex items-start justify-between gap-2">
      <h2 className="text-lg font-semibold leading-snug">{headline}</h2>
      <button
        type="button"
        className="rounded-md px-2 py-1 text-sm text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
        onClick={onClose}
        aria-label="Close"
      >
        ✕
      </button>
    </div>
  )
}

function ShowRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs uppercase tracking-wide text-zinc-500">{label}</dt>
      <dd className="mt-0.5">{value}</dd>
    </div>
  )
}
