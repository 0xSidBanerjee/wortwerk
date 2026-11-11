# Wortwerk

Distraction-free German reading for beginners. Open a short A1 story, tap any word or phrase, and see the English meaning instantly — no accounts, no tracking, no network call on tap.

## Privacy

All dynamic reading state stays in your browser (`localStorage`). The app does not load third-party scripts or fonts. Word audio is fetched only when you press listen.

## Quick start

**Requirements:** Go 1.27+, Node 26+

```bash
git clone https://github.com/0xSidBanerjee/wortwerk.git
cd wortwerk
make install
make build
make serve    # single binary server (after M6)
# or, during development:
make dev      # Vite dev server
```

Run tests and linters:

```bash
make test
make lint
```

Ingestion (author machine):

```bash
wortwerk dict build                    # testdata sample, or path to kaikki JSONL
wortwerk ingest --offline fixture     # no API key (uses testdata/ingest/)
wortwerk validate
```

Live ingest needs `WORTWERK_LLM_*` in `.env`. Piper/ffmpeg for audio (M5) — see [docs/project-plan.md](docs/project-plan.md).

## Project layout

| Path | Purpose |
| --- | --- |
| `cmd/wortwerk/` | CLI: `serve`, `ingest`, `validate`, `audio`, `dict` |
| `internal/` | Ingestion, dictionary, audio, static server |
| `schema/` | Compiled story JSON Schema (contract) |
| `web/` | Vite + React + TypeScript reader |
| `content/` | Stories, compiled JSON, audio (CC BY-SA 4.0) |

## Licenses

- **Source code:** [MIT](LICENSE)
- **Stories, compiled JSON, audio:** [CC BY-SA 4.0](content/LICENSE) — see [content/README.md](content/README.md) for Wiktionary and Thorsten-Voice attribution

## Documentation

- [Project plan](docs/project-plan.md)
- [Contributing](CONTRIBUTING.md)

## Repository

https://github.com/0xSidBanerjee/wortwerk
