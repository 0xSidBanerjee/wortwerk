import type { Unit, WortwerkCompiledStory } from '../types/story.generated'

export type UnitRef = {
  sentenceId: string
  unit: Unit
}

/** Map key is `${sentenceId}:${tokenId}` for O(1) tap lookup. */
export function buildTokenUnitMap(story: WortwerkCompiledStory): Map<string, UnitRef> {
  const map = new Map<string, UnitRef>()
  for (const sentence of story.sentences) {
    for (const unit of sentence.units) {
      for (const tokenId of unit.tokens) {
        map.set(`${sentence.id}:${tokenId}`, { sentenceId: sentence.id, unit })
      }
    }
  }
  return map
}

export function unitTokenKey(sentenceId: string, tokenId: number): string {
  return `${sentenceId}:${tokenId}`
}
