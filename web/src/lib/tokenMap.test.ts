import { readFileSync } from 'node:fs'
import { join, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

import { buildTokenUnitMap, unitTokenKey } from './tokenMap'
import type { WortwerkCompiledStory } from '../types/story.generated'

const story = JSON.parse(
  readFileSync(
    join(dirname(fileURLToPath(import.meta.url)), '../../../content/compiled/fixture.json'),
    'utf8',
  ),
) as WortwerkCompiledStory

describe('buildTokenUnitMap', () => {
  it('maps separable verb tokens to one unit', () => {
    const map = buildTokenUnitMap(story)
    const a = map.get(unitTokenKey('s0', 1))
    const b = map.get(unitTokenKey('s0', 3))
    expect(a).toBeDefined()
    expect(b).toBeDefined()
    expect(a?.unit.lex).toBe('aufstehen')
    expect(a?.unit).toEqual(b?.unit)
  })

  it('maps compound noun', () => {
    const map = buildTokenUnitMap(story)
    const ref = map.get(unitTokenKey('s2', 1))
    expect(ref?.unit.lex).toBe('handschuh')
    expect(story.lexicon.handschuh.parts).toEqual(['Hand', 'Schuh'])
  })

  it('maps contraction zum', () => {
    const map = buildTokenUnitMap(story)
    const ref = map.get(unitTokenKey('s1', 2))
    expect(ref?.unit.lex).toBe('zum')
  })

  it('returns undefined for unknown token', () => {
    const map = buildTokenUnitMap(story)
    expect(map.get(unitTokenKey('s0', 99))).toBeUndefined()
  })
})
