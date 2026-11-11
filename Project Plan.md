# Wortwerk: Project Plan

*Version 1 · November 10, 2026 · MVP scope locked*

## 1. What we are building

Wortwerk is a distraction-free web app for learning German by reading. A beginner opens a short A1 story (about 2 to 5 minutes), taps any word or phrase, and instantly sees the English meaning. A tap highlights the whole unit (a phrase, a separable verb like *steht … auf*, or a compound word) and opens a small popup with a button to hear it spoken. The reader remembers where you stopped.

**Principles (these drive every decision below):**

1. **Fast above all.** Tapping a word is an in-memory lookup. No network call, no parsing. No performance trade-offs.
2. **Minimal UI.** Library list, reader, popup. Nothing else.
3. **Private by design.** No accounts, no analytics, no third-party requests. All user state lives in the browser (localStorage).
4. **Heavy work happens once, offline, on the author's machine.** Users download small static files.
5. **Open-source ready.** Permissive code license, clear data licenses, reproducible pipeline, committed outputs so contributors can run the app without an LLM key.

## 2. Scope

**MVP**

- 5 A1 stories: Pinocchio, The Ugly Duckling, The Pied Piper of Hamelin, The Three Little Pigs, Jack and the Beanstalk ("The Fox Who Sold the Sunrise" is dropped).
- Library screen and reader screen.
- Tap a word or phrase: highlight the unit, show a popup with meaning, part of speech, and a listen button. A "Show more" accordion holds base form, gender and article, plural, the word's grammatical role, and the full-sentence translation.
- Compound words show their parts (*Hand + Schuh = Handschuh*). Separable verbs highlight both pieces together.
- Pre-generated audio for each word or phrase unit (listen button only; no story narration).
- Continue where you left off.
- Mobile-friendly responsive web, light and dark following the system setting.
- Go ingestion CLI and a single-binary server with the React app embedded.

**v2.0:** story narration with word-by-word highlighting, and reading speed control.

**Later:** saved words (browser-only), add-your-own story, BYOK for better TTS and dictionary, level picker once there is more than one level, more languages, optional PWA/offline.

**Explicitly out of scope for MVP:** accounts, analytics, saved words, narration, level picker UI, native apps, a story review UI (review happens through git diffs).

## 3. Architecture

There is no runtime backend logic. Everything dynamic is computed once at ingestion.

```
  AUTHOR TIME (your machine, once per story)
  story.md --> parse/split/tokenize (Go) --> dictionary candidates (local index)
           --> LLM: group units + pick senses + translate
           --> validate (Go) --> compiled JSON  --> Piper TTS --> audio files

  RUN TIME (user's browser)
  index.json --> story.json (one file, ~15-30 KB gzipped) --> in-memory token->unit map
  tap word --> map lookup --> popup (no network)
  tap listen --> fetch one small audio file --> play
```

**Why static files:** a story is about 1,000 tokens. Even verbose JSON is roughly 100 KB raw and about 15 to 25 KB gzipped, and parsing it takes milliseconds once per story open. Growth in story count does not affect per-user cost, since only the opened story is fetched. The cost of growth lands on the author side (LLM calls, review, audio), not on the runtime.

**Go's role:** the ingestion CLI plus a small static server that embeds the built frontend. For production you can ignore the binary and host the build output on any CDN.

## 4. Repository layout

```
wortwerk/
  CLAUDE.md                 # rules for Claude Code (see section 12)
  README.md  LICENSE        # code: MIT
  schema/story.schema.json  # single source of truth for the compiled format
  cmd/wortwerk/main.go      # subcommands: serve, ingest, validate, audio, dict
  internal/
    ingest/                 # split, tokenize, candidates, llm, validate, cache, compile
    dict/                   # Wiktextract import, SQLite index (tooling only)
    audio/                  # Piper + ffmpeg wrappers
    server/                 # embed, caching headers, SPA fallback
  web/                      # Vite + React + TS + Tailwind + shadcn/ui
  content/                  # data license: CC BY-SA 4.0 (see section 14)
    stories/*.md            # authored sources
    compiled/index.json, *.json   # generated, committed
    audio/*.mp3             # generated, committed, hashed names
  data/                     # gitignored: dict.sqlite, LLM response cache
  .github/workflows/        # CI: tests, build, bundle-size check
```

## 5. Content formats

### 5.1 Story source (markdown)

```markdown
---
title: Die drei kleinen Schweine
slug: three-little-pigs
level: A1
lang: de
source: The Three Little Pigs (public domain folk tale)
---

Es waren einmal drei kleine Schweine. ...
```

Markdown is fine for this: it is easy to write, review and diff. Allowed body content is **plain paragraphs only** (no lists, images or headings), which the linter enforces. Text must be Unicode NFC normalized (umlauts and ß can otherwise arrive in decomposed form and break matching).

### 5.2 Compiled story (JSON)

The compiled file stores the text as **tokens** (ground truth, includes punctuation and spacing) with **units** layered on top as stand-off annotations that point at token ids. A unit is a set of token ids, so it can be one word, adjacent words, or non-adjacent words (separable verbs).

| Level | Field | Notes |
| --- | --- | --- |
| story | `schemaVersion`, `id`, `title`, `lang`, `glossLang`, `level`, `wordCount` | `lang`/`glossLang` keep the format language-agnostic |
| sentence | `id`, `p` (paragraph index), `tokens[]`, `units[]`, `translation` | resume position = sentence id |
| token | `id`, `text`, `space` (whitespace follows) | lets the app re-render the text exactly |
| unit | `tokens[]`, `lex`, optional `form`, optional `gloss` | `form` = short role of this occurrence (e.g. "3rd person singular, present"); `gloss` overrides the lexicon meaning in context |
| lexicon entry | `lemma`, `kind` (word, phrase, separable, compound, name), `pos`, `meaning`, optional `gender`, `article`, `plural`, `separable{prefix,stem}`, `parts[]`, `audio` | each story embeds its own small lexicon for the MVP |

Sample fragment:

```json
{
  "id": "s1", "p": 0,
  "tokens": [
    {"id":0,"text":"Pinocchio","space":true},
    {"id":1,"text":"steht","space":true},
    {"id":2,"text":"früh","space":true},
    {"id":3,"text":"auf","space":false},
    {"id":4,"text":".","space":true}
  ],
  "units": [
    {"tokens":[0],"lex":"pinocchio"},
    {"tokens":[1,3],"lex":"aufstehen","form":"3rd person singular, present"},
    {"tokens":[2],"lex":"früh"}
  ],
  "translation": "Pinocchio gets up early."
}
```

**Runtime use:** on load, build `Map<sentenceId:tokenId, unit>` once. A tap reads two data attributes from the clicked span, looks up the unit, highlights all its tokens, and opens the popup. Event delegation on the text container (one handler, not one per word) keeps rendering fast.

**Rule: every word token belongs to exactly one unit.** That keeps the tap-to-unit mapping unambiguous. Interleaving is allowed (the words between *steht* and *auf* are their own units).

**Default grouping rules (given to the LLM and checked in review):**

- Fixed phrases and idioms whose meaning is not word-by-word (*Guten Morgen*, *zum Beispiel*, *es gibt*) are one unit.
- Separable verbs are always one unit, even when split across the sentence.
- Reflexive verbs include the reflexive pronoun (*freut sich*).
- Compounds are one unit with `parts` shown in the popup. Linking elements (Fugen-s etc.) are allowed in the parts.
- Contractions (*zum*, *im*, *ins*) are one unit with `parts` (zu + dem).
- Articles are separate tappable units. The noun's article and gender appear under "Show more" on the noun.
- Names are `kind: name` with a short note.

### 5.3 Index file

`index.json` lists each story: slug, title, level, word count, estimated minutes (words ÷ 100, rounded up), and a content `rev`. Story files are fetched as `/content/compiled/<slug>.json?v=<rev>` so they can be cached as immutable.

## 6. Ingestion pipeline (Go CLI)

Code does everything deterministic. The LLM does only what needs judgment, and **never outputs the tokens**.

| # | Stage | Who | Notes |
| --- | --- | --- | --- |
| 0 | `dict build` (one-time) | code | Import German Wiktextract data into a local SQLite index (tooling only, gitignored) |
| 1 | Parse and lint markdown | code | Frontmatter, plain paragraphs, NFC check, no abbreviations (style guide) |
| 2 | Sentence split | code | Rules for German direct speech: do not split before a lowercase continuation (*„Hallo!“, sagte er.*) or after known abbreviations |
| 3 | Tokenize | code | Letters, digits and intra-word hyphens stay in a token; punctuation and „ “ quotes are separate tokens; record `space` flags |
| 4 | Candidate lookup | code | For each token: form → lemma(s) → top 3 senses, POS, gender, plural, separable info from the local index |
| 5 | LLM annotation | LLM | Input: sentences with token ids plus candidates. Output (structured JSON, schema-constrained): units, lexicon entries, contextual senses, sentence translations. Chunked by paragraph group (about 8 to 12 sentences) so a retry is cheap |
| 6 | Validate | code | See 6.1. On failure, errors go back to the LLM for that chunk only (max 2 retries), else the run fails with a report |
| 7 | Compile | code | Deterministic key order and formatting so git diffs are clean; regenerate `index.json` |
| 8 | Audio | code | See section 8 |

**Caching:** LLM responses are cached under `data/` keyed by hash(sentence text + prompt version + model). Editing one sentence re-runs only that sentence's chunk. Changing the prompt version re-runs everything deliberately.

**Provider abstraction:** a small Go interface with one implementation to start. Model and key come from environment variables. Run a pilot on one story with two cheap models and compare validator pass rate plus manual review before standardizing. The CLI prints token usage and an estimated cost per run.

**Cost expectation (estimate, to be confirmed in the pilot):** roughly 10k input and 10k to 12k output tokens per story, which is cents with a cheap model and around twenty cents with a premium one. A one-time author-side cost, never paid per user.

### 6.1 Validation rules

- Joining tokens (respecting `space`) reproduces the original sentence text exactly.
- Every word token is in exactly one unit; unit token ids are valid, unique and sorted.
- Every `lex` reference exists; no orphan lexicon entries.
- Nouns have gender and a matching article; plural is present or explicitly null.
- Separable verbs: prefix + stem equals the lemma, and the lemma exists in the dictionary index (warn if not).
- Compound parts concatenate (case-insensitively, allowing linking elements) to the lemma; warn otherwise.
- Meanings and translations are non-empty; meanings stay short (about 60 characters).
- Lint warnings: sentence longer than about 14 tokens, abbreviations, non-NFC text.

These checks catch the typical LLM failures (dropped words, wrong ids, invented entries) automatically, so your review time goes to the German and the glosses, not the structure.

### 6.2 Review workflow (no UI)

Run ingest, then review the `git diff` of `content/compiled/`. An optional `wortwerk review <slug>` prints a terminal table (sentence, translation, units, glosses) for faster reading. A native German speaker should review each story once before it ships.

## 7. Dictionary data

**Primary source:** Wiktionary data, extracted by Wiktextract and published as machine-readable German data on kaikki.org. It carries gender, plural, inflected forms (inflected forms exist as their own entries pointing to the lemma, which helps resolve *steht* to *stehen*), and separable-verb information, which matches the "Show more" fields. The raw German extract is large (the full JSONL is on the order of 186 MB), which is exactly why we never ship it: the local index is tooling-only and each story carries only the entries it uses.

**Fallback:** FreeDict's German-English for plain headword lookups if Wiktionary coverage is thin for a word. Verify its license when adding it.

**License:** Wiktionary text is dual-licensed CC BY-SA and GFDL. Plan for attribution (a small About/Credits page and README section linking Wiktionary and kaikki.org, and citing the Wiktextract paper as the data's authors request for research use). Treat the compiled lexicon as derived data and apply ShareAlike thinking to the `content/` directory (see section 14). Confirm exact attribution wording before release.

## 8. Audio (word and phrase pronunciation)

- **Engine:** Piper (local, fast, offline) with the German *thorsten* voice. The voice is released under CC0 (public domain), so it has no attribution requirements, though crediting Thorsten Müller is good form.
- **Tooling license caveat:** the piper-tts Python package is GPL. Keep Piper as an external build-time tool invoked by the CLI; do not link it into or ship it with the app. The generated audio files are the only artifacts distributed. Confirm this reading before open-sourcing.
- **What gets spoken:** the surface text of each unit (e.g. *steht auf* for a separable verb, joined in order). This is a decision point; see section 15.
- **Pipeline:** normalize text, hash it (text + voice + settings) for the filename, skip if the file exists, synthesize, trim silence and normalize loudness with ffmpeg, encode to mono MP3 (or AAC) at about 48 to 64 kbps for universal browser support. Each file is a few KB. Identical words across stories reuse the same file.
- **Playback:** the listen button fetches its file on demand; prefetch it when the popup opens so the tap feels instant. Browsers require playback to begin from a user gesture, which the button is, but test on iOS Safari specifically.
- **Known risk:** very short isolated words can sound clipped or oddly intonated. Mitigation: do a listening pass on every generated file, add trailing punctuation or a short carrier if needed, and revisit in v2 when sentence-level audio plus alignment may let us slice words from natural speech.
- **Repo size:** five stories means a few MB of audio, fine in git. If the library grows large, move audio to Git LFS or release assets.

## 9. Frontend

**Stack:** Vite, React, TypeScript (strict), Tailwind, shadcn/ui. Use only the components we need: Popover (desktop), Drawer or bottom sheet (mobile), Accordion, Button. Router: a tiny one (wouter) or hand-rolled, since there are two routes (`/` and `/s/:slug`).

**Self-host everything:** fonts and assets are served from the same origin. No Google Fonts, no analytics, no third-party scripts. This keeps the privacy promise real and the performance predictable. Add a strict Content-Security-Policy in the server.

**Screens**

1. **Library:** a list of stories with title, level badge, estimated minutes, and a "Continue" affordance on the story you last opened. A tiny progress indicator is optional; cut it if it feels like clutter.
2. **Reader:** the story title and paragraphs of tokens. Comfortable measure (about 65 characters), generous line height, system light/dark. No toolbars.
3. **Popup (on tap):** main view shows the contextual meaning, a part-of-speech chip, and the listen button. For compounds, one extra line shows the split. For separable verbs the title is the full verb (*aufstehen*: to get up). "Show more" reveals base form, article and gender, plural, grammatical role of this occurrence, and the full-sentence translation. Popover on desktop, bottom drawer on mobile.

**Interaction details**

- Words are plain spans with `data-s` and `data-t`; punctuation is not tappable. Selected unit gets a highlight class on all its tokens, including non-adjacent ones.
- Memoize per sentence so a selection only re-renders the affected sentence.
- Suppress text-selection UI on tokens (`user-select: none`, no touch callout) so taps do not trigger the OS selection handles.
- Basic keyboard support: arrow keys move between words, Enter opens the popup, Escape closes. (Budget this in the polish milestone.)

**Resume**

- localStorage key `wortwerk:v1:progress` with `{ lastStory, stories: { slug: { sentenceId, ts } } }`.
- Save the sentence id of the first sentence near the top of the viewport (debounced scroll or IntersectionObserver). Sentence ids survive reflow on different screen sizes; pixel offsets do not.
- Wrap all storage access in try/catch and degrade silently when storage is unavailable (private mode). Version the key so future schema changes can migrate.

**Performance budget (enforced in CI)**

| Metric | Target |
| --- | --- |
| Initial JS, gzipped | 100 KB or less |
| Story JSON, gzipped | 30 KB or less |
| Tap to popup visible | within one frame (under 50 ms) on a mid-range phone |
| Largest contentful paint on throttled 4G | under 2 s |
| Lighthouse performance | 95 or more |

## 10. Go server and single binary

- `net/http` from the standard library. `go:embed` bundles `web/dist` and `content/` into the executable; `wortwerk serve` is the whole app.
- Cache headers: hashed JS/CSS assets `immutable` for a year; `index.json` revalidated; versioned story and audio URLs `immutable`.
- Precompress assets at build time (gzip and brotli) and serve with content negotiation, since the standard library does not do brotli.
- SPA fallback to `index.html` for the `/s/:slug` route.
- Static hosting is the preferred production path. A `make site` target assembles `web/dist` plus `content/` for any CDN or static host.
- A Dockerfile is optional and trivial (scratch or distroless image with the binary).

## 11. Milestones

Each milestone has a clear definition of done. Sizes are relative (S, M, L), not calendar estimates.

| # | Milestone | Size | Done when |
| --- | --- | --- | --- |
| M0 | Foundations | S | Repo, Go module, Vite app, CI, `CLAUDE.md`, JSON Schema v1, generated TS types |
| M1 | Data spine | S | A hand-annotated one-paragraph fixture validates against the schema and loads in the app |
| M2 | Reader UI (on fixture) | M | Tap, highlight (including separable and compound), popup with Show more, library, resume, mobile layout |
| M3 | Ingestion pipeline | L | `dict build`, splitter, tokenizer, candidates, LLM stage, validator, cache, `ingest` and `validate` commands; one story passes end to end |
| M4 | Content | M | Five A1 stories written, ingested, reviewed by a German speaker, fixed |
| M5 | Audio | M | Piper + ffmpeg stage, hashed files, listen button with prefetch; listening pass done |
| M6 | Server and packaging | S | Embedded single binary, cache headers, precompression, release build in CI |
| M7 | Polish and release | M | Perf budget green, keyboard nav, real-device QA (iOS Safari, Android Chrome), Credits page, README, license files |

M2 and M3 can run in parallel after M1 because the schema is the contract between them. Building the UI against a hand-written fixture first de-risks the interaction design before any LLM spend.

## 12. Vibe-coding workflow

**`CLAUDE.md` should state, at minimum:**

- Principles from section 1, especially the performance budget, no tracking, and no third-party requests.
- Dependency policy: Go standard library first; a new runtime dependency (frontend or backend) needs a one-line justification and a size check.
- The schema is the contract: change `schema/story.schema.json` first, bump `schemaVersion` when breaking, regenerate types.
- Conventions: TypeScript strict, table-driven Go tests, deterministic output from the compiler.
- Never ship the full dictionary or any LLM key to the client.

**Order of work with Claude Code:** one milestone per session. Start each with: read `CLAUDE.md` and this plan's section for the milestone, propose a short task list, then implement. Write the validator and splitter tests **before** the implementation (German edge cases: direct speech, hyphens, ß and umlauts, NFC, separable verbs, abbreviations).

**Your architectural review gate (per PR):**

1. Bundle size delta against the budget.
2. Any new dependency justified?
3. Schema changes versioned and types regenerated?
4. Validator and splitter tests cover the change?
5. Does anything add a network request or storage the user did not expect?

**Contract-first tip:** keep the validator pure and dependency-free so it can later compile to WASM or be ported to TypeScript for the add-your-own-story feature.

## 13. Testing

- **Go:** table tests for splitter, tokenizer and validator; golden-file test that ingesting a fixture story with a recorded LLM response produces byte-identical JSON (stability).
- **Web:** Vitest for the token-to-unit map and resume logic; Playwright smoke test on a mobile viewport (open story, tap a word, popup appears, Show more expands, listen button plays, reload resumes position).
- **Perf:** Lighthouse CI against the budget in section 9; bundle-size check in CI.
- **Content:** the German-speaker review in M4 and the listening pass in M5 are release gates, not optional.

## 14. Licensing and attribution checklist

This is a planning checklist, not legal advice. Decide these before the first public release.

- **Code:** MIT (or Apache-2.0 if you want an explicit patent grant). Default: MIT.
- **Content directory:** stories are LLM-written retellings of public-domain tales, but the lexicon derives from Wiktionary (CC BY-SA, GFDL). Simplest consistent choice: license `content/` under CC BY-SA 4.0 and keep it clearly separate from the MIT code.
- **Audio:** Thorsten voice is CC0. Piper stays an external build-time tool (GPL package not distributed).
- **Credits page:** Wiktionary and kaikki.org (with a link), Thorsten-Voice, and the LLM used for annotation if you want transparency.
- **Fonts:** pick an open font (e.g. OFL) and self-host.
- **Name:** check domain, GitHub org, and trademark availability for "Wortwerk".

## 15. Risks and open decisions

**Risks**

| Risk | Mitigation |
| --- | --- |
| LLM mistakes on German morphology or separable verbs | Validator, dictionary cross-checks, native-speaker review |
| Wrong word sense chosen | Contextual `gloss` on the unit, review of glosses in the diff |
| Isolated word audio sounds unnatural | Listening pass, carrier trick, v2 slicing from sentence audio |
| ShareAlike implications for data | Decide `content/` license early (section 14) |
| iOS Safari audio quirks | Test early in M5 on a real device |
| Scope creep from "just one more feature" | The scope list in section 2 is the contract |

**Decisions to make (defaults in bold)**

1. **Narration tense for A1 stories:** **present tense (historical present)**, with Perfekt or Präteritum only for *sein*, *haben* and modals. Fairy tales are traditionally in Präteritum, which is beyond A1 and would hurt the "beginner" promise.
2. **Story length target:** **250 to 500 words** (about 2 to 5 minutes at beginner reading speed).
3. **Audio text for separable verbs:** **speak the joined surface (*steht auf*)**.
4. **Router:** **wouter**, or hand-rolled if even that feels heavy.
5. **Data license:** **CC BY-SA 4.0 for `content/`**.
6. **Browser support:** **last two versions of evergreen browsers**, including iOS Safari.

## 16. After the MVP

- **v2.0 narration:** generate full-story audio, align it to tokens (forced alignment or TTS timestamps), highlight the current word, add speed control with `playbackRate` and pitch preserved.
- **Saved words:** browser-only, with export and import of a file so users keep full control and nothing leaves the device.
- **Add your own story:** run the pipeline in the browser with the user's own API key (BYOK), store the result locally, reuse the same schema and validator.
- **Levels:** show the level picker once stories exist beyond A1.
- **Other languages:** the schema already carries `lang` and `glossLang`; the pipeline needs a per-language tokenizer, dictionary source and voice.
- **Optional:** PWA and offline support, BYOK for premium TTS, a shared build-time lexicon cache if ingestion cost ever matters.

## Appendix: A1 story writing guidelines

- Present tense narration; short main clauses; mostly subject-verb-object order.
- At most about 10 to 12 words per sentence; one idea per sentence.
- Avoid Genitiv, Konjunktiv, Passiv and long subordinate chains. Simple *weil*, *dass* and *wenn* clauses are acceptable sparingly.
- High-frequency vocabulary and deliberate repetition of key words so beginners meet them several times.
- Avoid abbreviations, numerals in digits, and idioms that need cultural explanation (unless the idiom is the point and is grouped as a phrase).
- Direct speech uses German quotes („ “) and short lines.
- Keep the story's beginning, middle and end clear; no moral lecture at the end unless it is one short sentence.
