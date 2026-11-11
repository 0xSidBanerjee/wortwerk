import { LibraryPage } from './pages/LibraryPage'
import { ReaderPage } from './pages/ReaderPage'
import { storySlugFromPath, usePath } from './lib/usePath'

function App() {
  const { path, navigate } = usePath()
  const slug = storySlugFromPath(path)

  if (slug) {
    return <ReaderPage slug={slug} navigate={navigate} />
  }

  if (path === '/' || path === '') {
    return <LibraryPage navigate={navigate} />
  }

  return (
    <div className="mx-auto max-w-lg px-4 py-10">
      <p className="text-zinc-600 dark:text-zinc-300">Page not found.</p>
      <button type="button" className="mt-4 text-sm underline" onClick={() => navigate('/')}>
        Library
      </button>
    </div>
  )
}

export default App
