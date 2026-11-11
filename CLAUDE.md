# Wortwerk — agent rules

Read [docs/project-plan.md](docs/project-plan.md) for full scope. Implement one milestone per session.

## Principles

1. **Fast above all.** Tap-to-gloss is an in-memory lookup — no network, no parsing on tap.
2. **Minimal UI.** Library, reader, popup only.
3. **Private by design.** No accounts, analytics, or third-party requests. User state in `localStorage` only.
4. **Heavy work offline on the author machine.** Ship small static files.
5. **Open-source ready.** Committed compiled content so the app runs without an LLM key.

## Performance budget (CI-enforced)

| Metric | Target |
| --- | --- |
| Initial JS (gzip) | ≤ 100 KB |
| Story JSON (gzip) | ≤ 30 KB |
| Tap → popup | < 50 ms on a mid-range phone |
| Lighthouse performance | ≥ 95 |

## Dependencies

- Go: standard library first. New runtime deps need a one-line justification in the PR and size impact.
- Web: justify new packages against the 100 KB gzip budget.

## Contract

- `schema/story.schema.json` is the contract between ingestion and the web app.
- Change schema first; bump `schemaVersion` when breaking; run `make generate-types`.
- Compiler output must be deterministic (stable key order, formatting).
- Go validator stays pure (no I/O) where possible.

## Conventions

- TypeScript `strict`.
- Table-driven Go tests; write splitter/validator tests before implementation for German edge cases.
- Never ship `data/dict.sqlite`, full dictionary JSON, or LLM keys to the client.

## PR review gates

1. Bundle size delta vs budget
2. New dependency justified?
3. Schema versioned and types regenerated?
4. Validator/splitter tests updated?
5. Any unexpected network request or storage?
