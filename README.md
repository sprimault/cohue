# Cohue

Français : [README.fr.md](README.fr.md)

An urban isometric action-roguelite under horde pressure. Chained rooms,
automatic fire, a build you assemble in fifteen minutes.

Apache 2.0 — [`LICENSE`](LICENSE), and [`CREDITS.md`](CREDITS.md) for graphic
assets.

The player only controls movement. They level up by collecting gems, break
crates by walking through them, read the scenery's signage to find the exit, and
move on to the next place — car park, supermarket, neighbourhood, cinema,
station.

The design goal is not difficulty, it is the restart: one key, under a second,
same configuration.

## Status

**You die here, or you get out — and start over.** One place, one weapon, a horde
that rises: the window opens, you move with the keyboard, and seven kinds of
creature enter one after another — the one that charges, the one that shoots from
afar, the one you cannot walk through, the one that explodes, the one that heals
the others. Bystanders come and go without attacking anyone. Gems raise your
level, every level up offers three cards, and crates break when you walk through
them. The door opens once you have killed what the place asks for: touching it is
enough to leave. The scenery is coloured rectangles — the images exist, but
nothing loads them yet.

The project follows a sixteen-step roadmap, published at each one. **What is done
and what remains are not listed here**: both lists would live beside the
documents that carry them, and would drift from them.

- [`ROADMAP.md`](ROADMAP.md) — the steps, which are passed, and what is out of
  scope for v1
- [`CHANGELOG.md`](CHANGELOG.md) — what each version brought, dated
- [`docs/conception.md`](docs/conception.md) — the full design (French)
- [`docs/regles-du-jeu.md`](docs/regles-du-jeu.md) — how to play, and what each
  creature does (French)

## Assets

Scenery is **generated** by `outils/decor_iso.py`, deterministically: a shape is
fixed in the script, never in the PNG.

```
make decors      # the places, six themes
make figurines   # the creatures, six body types, colour variants
make objets      # what gets picked up or fired
make sons        # the sound effects, by synthesis
```

Scenery and characters are **generated and versioned**: they come from the same
isometric primitives, depend on no third-party source, and the images are
embedded in the executable. A fresh clone builds with nothing else to install.

That is also what makes the creatures contributable: changing a profile's
proportions or adding a body type happens in code reviewable in a pull request,
not in a PNG.

## Levels and sharing

A level is a list of placed rooms, not a map: a few hundred bytes, copyable as
base64 in a message.

**A shared level contains no image and no sound.** It only references what the
binary provides — rooms, objects, profiles, events. That is what makes
distribution trivial and removes any rights question about what passes through
the game.

## Building

```
make build     # local binary in .tmp/
make test
make lint
```

Ebitengine is pure Go on Windows: no C compiler for day-to-day development. The
Linux and macOS targets require native builds and go through CI.

[`docs/construction.md`](docs/construction.md) covers the rest (in French): local
targets, asset regeneration and the build matrix.
