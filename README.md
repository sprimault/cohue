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

**The game opens and can be walked, it cannot be played yet.** Steps 1 and 2 are
delivered: the window shows a level, you move through it with the keyboard, and
the horde closes in around obstacles.

What is missing for this to be a game: dying and restarting, doors to open,
crates to break, weapons to pick up, level-ups, and the chain of levels. The
scenery is coloured rectangles — the images exist, but nothing loads them yet.

What is settled: the simulation running three hundred creatures with no
allocation per tick, the isometric rendering with its projection, camera and
depth sort, the scenery generated with its manifest, and the characters
generated as well — one body type per family, eight orientations, and the
gameplay values in the same manifest as the rendering.

- [`docs/conception.md`](docs/conception.md) — the full design (French)
- [`ROADMAP.md`](ROADMAP.md) — the steps, and what is out of scope for v1

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
