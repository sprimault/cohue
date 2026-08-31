# Security

Français : [SECURITY.fr.md](SECURITY.fr.md)

## Supported versions

The latest published release. In `0.x` there is no maintenance branch: a fix
ships in the next version.

## Reporting a vulnerability

Through a private security advisory on the GitHub repository, never a public
issue. Expect a reply within a few days.

## In scope

The game loads files written by third parties: levels and campaigns. That is the
only real attack surface, and the one that matters.

- A level that crashes the game, loops forever, or exhausts memory while loading.
- A file path built from a level that would escape the intended directory.
- Anything that would execute code from shared content — **nothing in the format
  allows it, and that is an invariant**: a level holds only identifiers and
  positions, no binary, no script.

## Out of scope

Editing your own save files to cheat offline. The game is single-player: there is
nothing to protect against its owner.
