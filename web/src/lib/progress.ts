export const PROGRESS_STORAGE_KEY = 'wortwerk:v1:progress'

export type StoryProgress = {
  sentenceId: string
  ts: number
}

export type ReaderProgress = {
  lastStory: string
  stories: Record<string, StoryProgress>
}

const empty: ReaderProgress = { lastStory: '', stories: {} }

export function loadProgress(): ReaderProgress {
  try {
    const raw = localStorage.getItem(PROGRESS_STORAGE_KEY)
    if (!raw) return { ...empty }
    const parsed = JSON.parse(raw) as ReaderProgress
    if (!parsed || typeof parsed !== 'object' || !parsed.stories) {
      return { ...empty }
    }
    return {
      lastStory: typeof parsed.lastStory === 'string' ? parsed.lastStory : '',
      stories: parsed.stories,
    }
  } catch {
    return { ...empty }
  }
}

export function saveProgress(update: Partial<ReaderProgress> & { slug: string; sentenceId: string }): void {
  try {
    const current = loadProgress()
    const next: ReaderProgress = {
      lastStory: update.slug,
      stories: {
        ...current.stories,
        [update.slug]: { sentenceId: update.sentenceId, ts: Date.now() },
      },
    }
    localStorage.setItem(PROGRESS_STORAGE_KEY, JSON.stringify(next))
  } catch {
    // private mode / quota — degrade silently
  }
}

export function getStoryProgress(slug: string): StoryProgress | undefined {
  return loadProgress().stories[slug]
}
