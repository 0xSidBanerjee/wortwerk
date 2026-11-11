import { memo, useMemo } from 'react'

import { isWordToken } from '../lib/isWordToken'
import type { Sentence } from '../types/story.generated'

type Props = {
  sentence: Sentence
  /** Token ids to highlight in this sentence; null if none. */
  highlightIds: number[] | null
  onTokenTap: (sentenceId: string, tokenId: number) => void
  sentenceRef?: (el: HTMLParagraphElement | null) => void
}

function SentenceBlockInner({ sentence, highlightIds, onTokenTap, sentenceRef }: Props) {
  const highlighted = useMemo(() => new Set(highlightIds ?? []), [highlightIds])

  return (
    <p
      ref={sentenceRef}
      data-sentence-id={sentence.id}
      className="mb-4 text-lg leading-relaxed"
    >
      {sentence.tokens.map((token) => {
        const word = isWordToken(token.text)
        const hl = word && highlighted.has(token.id)
        if (!word) {
          return (
            <span key={token.id}>
              {token.text}
              {token.space ? ' ' : ''}
            </span>
          )
        }
        return (
          <span
            key={token.id}
            role="button"
            tabIndex={0}
            data-s={sentence.id}
            data-t={token.id}
            className={
              hl
                ? 'cursor-pointer rounded-sm bg-amber-200/80 px-0.5 text-zinc-900 select-none [-webkit-touch-callout:none] dark:bg-amber-500/35 dark:text-zinc-50'
                : 'cursor-pointer rounded-sm px-0.5 text-zinc-900 select-none [-webkit-touch-callout:none] hover:bg-zinc-200/80 dark:text-zinc-100 dark:hover:bg-zinc-800'
            }
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                onTokenTap(sentence.id, token.id)
              }
            }}
          >
            {token.text}
            {token.space ? ' ' : ''}
          </span>
        )
      })}
    </p>
  )
}

function sameIds(a: number[] | null, b: number[] | null): boolean {
  if (a === b) return true
  if (a === null || b === null) return false
  if (a.length !== b.length) return false
  for (let i = 0; i < a.length; i++) {
    if (a[i] !== b[i]) return false
  }
  return true
}

export const SentenceBlock = memo(SentenceBlockInner, (prev, next) => {
  return (
    prev.sentence === next.sentence &&
    sameIds(prev.highlightIds, next.highlightIds) &&
    prev.onTokenTap === next.onTokenTap &&
    prev.sentenceRef === next.sentenceRef
  )
})
