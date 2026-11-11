import { useCallback, useEffect, useState } from 'react'

export function usePath(): { path: string; navigate: (to: string) => void } {
  const [path, setPath] = useState(() => window.location.pathname)

  useEffect(() => {
    const onPop = () => setPath(window.location.pathname)
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])

  const navigate = useCallback((to: string) => {
    window.history.pushState(null, '', to)
    setPath(to)
  }, [])

  return { path, navigate }
}

export function storySlugFromPath(path: string): string | null {
  const m = /^\/s\/([^/]+)\/?$/.exec(path)
  return m ? decodeURIComponent(m[1]) : null
}
