import { afterEach, describe, expect, it, vi } from 'vitest'

import { PROGRESS_STORAGE_KEY, loadProgress, saveProgress } from './progress'

describe('progress', () => {
  afterEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('reads empty when missing', () => {
    expect(loadProgress()).toEqual({ lastStory: '', stories: {} })
  })

  it('round-trips save and load', () => {
    saveProgress({ slug: 'fixture', sentenceId: 's1' })
    const p = loadProgress()
    expect(p.lastStory).toBe('fixture')
    expect(p.stories.fixture?.sentenceId).toBe('s1')
  })

  it('degrades on corrupt JSON', () => {
    localStorage.setItem(PROGRESS_STORAGE_KEY, '{not json')
    expect(loadProgress()).toEqual({ lastStory: '', stories: {} })
  })

  it('save does not throw when storage throws', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('quota')
    })
    expect(() => saveProgress({ slug: 'fixture', sentenceId: 's0' })).not.toThrow()
  })
})
