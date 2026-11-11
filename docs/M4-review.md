# M4 content review (native speaker gate)

Five A1 stories are in `content/stories/` with compiled JSON in `content/compiled/`.

## Before release

- [ ] Read each story in the app (library → story → tap checks).
- [ ] Review `git diff` on `content/compiled/*.json` for German, glosses, and unit grouping.
- [ ] Confirm length: target **250–500** words (`wordCount` in JSON). *Pied Piper (226) and Pinocchio (241) may need a short expansion pass.*
- [ ] Re-run after edits: `wortwerk ingest --offline <slug>` or live `wortwerk ingest <slug>` with `WORTWERK_LLM_*`.

## Stories

| Slug | Title | wordCount |
| --- | --- | --- |
| pinocchio | Pinocchio | 241 |
| ugly-duckling | Das hässliche Entlein | 250 |
| pied-piper | Der Rattenfänger von Hameln | 226 |
| three-little-pigs | Die drei kleinen Schweine | 248 |
| jack-beanstalk | Jack und die Bohnenranke | 252 |

## Re-ingest

```bash
wortwerk validate
wortwerk ingest --offline <slug>   # uses testdata/ingest/<slug>-annotation.json
```

Live annotation: `wortwerk ingest <slug>` with API key and `data/dict.sqlite` built from Wiktextract.
