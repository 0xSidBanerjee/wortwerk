# Contributing to Wortwerk

Thank you for your interest in Wortwerk. This project is built in focused milestones; see [docs/project-plan.md](docs/project-plan.md) for scope and architecture.

## Development setup

1. Install [Go 1.27+](https://go.dev/dl/) and [Node 26](https://nodejs.org/) (see `.nvmrc`).
2. Clone the repo and run `make install` to install frontend dependencies.
3. Run `make test` and `make lint` before opening a pull request.

Ingestion and audio generation need optional local tools (LLM API key, Piper, ffmpeg) documented in the README. You can run the web app using committed `content/compiled/` output without an API key.

## Commits

We use [Conventional Commits](https://www.conventionalcommits.org/):

- Types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `build`, `ci`, `chore`
- Scopes: `web`, `ingest`, `dict`, `audio`, `server`, `schema`, `content`, `ci`
- One logical change per commit; imperative subject, ≤ 72 characters

## Pull requests

- Use a short-lived branch (`feat/…`, `fix/…`).
- Fill out the PR template checklist (bundle size, dependencies, schema, tests, privacy).
- Code is MIT; changes under `content/` are CC BY-SA 4.0.

## Schema changes

Edit `schema/story.schema.json` first, bump `schemaVersion` on breaking changes, regenerate TypeScript types (`make generate-types`), and extend Go validation tests.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
