# Contributing

Français : [CONTRIBUTING.fr.md](CONTRIBUTING.fr.md)

## How this project is written

The code is written alongside an assistant, under this repository's rules. It is
not a footnote: the method is the project's second subject, and the way the rules
are laid down follows from it.

- **Rules come before code**, they are not inferred from it.
  [`docs/conception.md`](docs/conception.md) is authoritative, and a disagreement
  between the document and the code is a defect in the code.
- **A decision carries what it rules out.** Every trade-off keeps the reason the
  rejected options were rejected, so it can be reopened without being replayed.
- **What can be measured is measured.** Several decisions were corrected by a
  measurement that contradicted an argument; the measurement decides.
- **Commit messages do not narrate how the work was done.** They say what changes
  and why — the rest is in the diff.

None of this applies differently to an outside contribution: same checks, same
style rules, same bilingual format.

## Before writing code

Open an issue first for anything beyond a fix. The design is written down in
[`docs/conception.md`](docs/conception.md); changing it is a discussion, not a
patch.

## What is discussed before it is written

A pull request touching one of these paths without prior discussion will be sent
back to an issue, whatever its quality — not on principle, but because these are
the places where one change tips others over.

- **[`docs/conception.md`](docs/conception.md)** is authoritative. The code
  conforms to it, so changing a line changes what the code must do.
- **The level and piece format** is a public contract: a level shared today must
  still load tomorrow. An added field is optional; a removed or renamed field
  breaks everything already in circulation.
- **`assets/` and `outils/`** go together. The scenery is generated: a shape is
  fixed in `outils/decor_iso.py`, never in the PNG — a manual edit would be
  overwritten by the next run without anyone noticing.
- **`.github/workflows/`** decides what gets checked. Branch protection requires
  checks by name, not by content: a modified workflow can turn green a check that
  no longer verifies anything.

Everything else — code, tests, accompanying documentation — can be proposed
directly.

## What a contribution is judged on

Code conventions and testing doctrine are in [`docs/go.md`](docs/go.md). What
follows is the enforceable summary.

- `make fmt && make lint && make test && make race && make vulncheck && make sec` pass.
- Every declaration is documented. Comments say *why*; they never paraphrase the
  next line.
- No banners, no decorative emoji, in code, logs or commit messages.
- **No allocation in the update loop.** Preallocated pools, reused slices,
  swap-remove. No pointer escapes a pool.
- **Determinism is preserved.** Nothing reads the clock or system entropy: random
  values come from the generator seeded by the run's seed. A change that makes a
  replay diverge is a defect even if every test passes.
- **Data stays data.** A new enemy profile is a table row, not a branch in a
  function.
- An added dependency goes into `THIRD-PARTY-NOTICES`. An added graphic or sound
  asset goes into [`CREDITS.md`](CREDITS.md), **in the same commit**.
- Nothing in `internal/game` imports the renderer. CI runners are headless; a test
  that needs a window does not belong in the default suite.

## Delivery

**One batch, one branch, one commit.** The branch starts from an up-to-date
`master` and is named `<type>/<subject>`, where the type is its commit's
conventional prefix: `feat/`, `fix/`, `docs/`, `chore/`, `test/`, `refactor/`.
Do not chain two batches on one branch — each must stay readable and revertible
on its own.

It returns to `master` **through a pull request**, never a local merge: the PR is
what records what was delivered, and merging it deletes the branch on both sides.

**Check before pushing, not after:**

```
make fmt && make lint && make test && make race && make vulncheck && make sec
```

Two more apply as soon as a change touches `assets/`, `outils/`, or the shape of
a file:

```
make ressources-verif && make entetes
```

`govulncheck` queries its advisory database **live**: a job green in the morning
can be red in the afternoon on exactly the same code. Do not rely on CI alone,
which validates once the branch is already pushed.

**The `CHANGELOG` section ships with the batch**, not at tagging time: it is
reviewed in the pull request, when it matters. Publication takes the release name
and notes from it, and a missing section stops it.

**Documentation ships with the change.** Before committing, check what the change
makes false elsewhere: the state announced in the README, a decision in
[`docs/conception.md`](docs/conception.md), a step in
[`ROADMAP.md`](ROADMAP.md). Specific to this project: a design decision that
changes makes the document false, and that document is authoritative — the code
never runs ahead of it.

**A message says what changes and why**, in a few lines. The default is the title
alone: a body exists only if it carries something the title does not say and the
diff does not show.

## Fixing a vulnerability without creating another

Do not adopt a version released **the same day**, even a fix. Look for the oldest
one that suffices:

```
go list -m -versions <module>
```

A version published within the hour is the typical profile of a compromised
maintainer account.

A pin is explained: a `require` held below the latest available carries a
trailing comment saying why, and **when to remove it**.

## Three numbers, not to be confused

| Number | Where | What it tracks |
|---|---|---|
| repository version | git tag | the binary |
| `version_format` | every level and every piece | the file format |
| `empreinte_jeu_pieces` | every level | the actual state of the piece set |

The last two do not follow SemVer. `version_format` is an integer: adding an
optional field does not bump it, anything else does, and a bump requires writing
the migration for existing levels.

`empreinte_jeu_pieces` is not a version but a checksum: it changes as soon as a
shipped piece moves. Without it, a level would load silently with geometry
different from the one its author built.

The repository follows SemVer with the zero clause, defined in
[`CHANGELOG.md`](CHANGELOG.md): **in `0.x`, nothing is guaranteed.** The minor marks a [`ROADMAP.md`](ROADMAP.md) step, not an API
break; everything else accumulates in the patch. Direct consequence: **the number
warns of nothing**, and it is the release notes that must say what a level author
needs to revisit.

## Language

**Identifiers are in English** — directories, files, packages, types, functions,
fields: `Enemy`, `FlowField`, `Tile`. **Documentation is in French**: godoc,
comments, error messages and logs. The API reads in English because it is code;
the reasoning reads in French because it is thought.

Commit messages in French first, English second, in a single text separated by
`***`. Never `---`: `git am` treats it as a patch separator and truncates
everything after it.

Contributions in English are welcome and are not subject to the bilingual rule.

## Shared levels

A level is data, never code: it references only the vocabulary the binary
provides — pieces, objects, profiles, events.

**No binary file, under any extension.** No image, no sound, no executable. This
is a mechanical rule, not a judgement: it removes every question of provenance
and rights over what passes through the game.
