# Journal des versions

Le format suit [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/).

SemVer avec la clause du zéro : **en `0.x`, rien n'est imposé**. Le mineur
marque une étape de la feuille de route, pas une rupture d'API — tout le reste
s'accumule en correctif, correctifs, fonctionnalités et ruptures confondus.

Trois numéros à ne pas confondre :

| Numéro | Où | Ce qu'il suit |
|---|---|---|
| version du dépôt | tag git | le binaire |
| `version_format` | chaque niveau et chaque pièce | le format de fichier |
| `empreinte_jeu_pieces` | chaque niveau | l'état réel du jeu de pièces |

Les deux derniers ne suivent pas SemVer. `version_format` est un entier :
ajouter un champ optionnel ne l'incrémente pas, tout le reste l'incrémente, et
un incrément oblige à écrire la migration des niveaux existants.
`empreinte_jeu_pieces` n'est pas une version mais une somme de contrôle. Une
version peut sortir sans qu'ils bougent ; ils ne bougent jamais sans version.

Un titre de section s'écrit `## [version] — date — titre`. La publication en
tire le nom et les notes de la version : ce qui est relu ici est ce qui sera lu
sur la page des versions, et il n'y a rien à recopier ensuite. Le titre est
facultatif ; sans lui, la version se nomme par son tag.

**Une section absente arrête la publication.** Une version sans notes ne dit ni
ce qui change, ni ce qu'un auteur de niveau doit reprendre.

Chaque section est **bilingue, français d'abord, séparé par `***`** — les notes
de version sont ce que lit un auteur de niveau étranger avant de savoir s'il
doit reprendre son travail. Ce préambule reste en français : il n'est jamais
publié, et explique les conventions du dépôt à qui y contribue.

## [0.2.0] — 2026-08-31 — La horde à l'écran

**Cette version se regarde et se traverse, elle ne se joue pas encore.** La
fenêtre s'ouvre sur un lieu, on s'y déplace au clavier, et la horde converge en
contournant les obstacles — mais on ne peut pas mourir, donc pas recommencer.

Il n'y a donc ni écran de mort, ni porte, ni caisse, ni arme à ramasser, ni
montée de niveau, ni enchaînement de lieux. Le décor est fait de rectangles
colorés : les images existent dans le binaire, mais rien ne les charge encore.
C'est le comportement attendu, pas un défaut.

### Ajouté

- **Le lieu s'affiche, et se parcourt au clavier** : les flèches, ou le carré de
  touches à gauche — WASD sur un clavier américain, ZQSD sur un clavier français,
  puisque c'est la place d'une touche qui compte et non la lettre inscrite
  dessus. La caméra suit le joueur, se bloque aux bords du lieu plutôt que de
  découvrir du vide, et centre une fois pour toutes un lieu qui tient dans
  l'écran.
- Le sol montre ce que la simulation lit d'une case — franchissable, coûteuse ou
  mur — et non le décor, qui viendra avec les images.
- **La horde est à l'écran et converge**, en contournant les obstacles. Elle est
  semée au lancement et n'arrive pas par vagues : le scénario de pression viendra
  plus tard. Les créatures se recouvrent dans l'ordre de leur profondeur, et le
  joueur passe devant celles qui partagent la sienne.

### À savoir

- **La fenêtre agrandit l'image d'un facteur qui n'est pas entier**, si on la
  redimensionne : les pixels deviennent alors inégaux. Le tampon interne, lui,
  est bien fixe. L'agrandissement en entier est un réglage d'affichage, et il
  arrivera avec les écrans de réglages.

***

**This release can be looked at and walked through, it cannot be played yet.**
The window opens on a level, you move through it with the keyboard, and the
horde closes in around obstacles — but you cannot die, so you cannot start over.

There is therefore no death screen, no door, no crate, no weapon to pick up, no
level-up, no chain of levels. The scenery is coloured rectangles: the images are
in the binary, but nothing loads them yet. This is the expected behaviour, not a
defect.

### Added

- **The level is drawn, and can be walked with the keyboard**: the arrow keys, or
  the square of keys on the left — WASD on a US layout, ZQSD on a French one,
  since what counts is where a key sits and not the letter printed on it. The
  camera follows the player, stops at the level edges rather than revealing empty
  space, and centres a level small enough to fit the screen once and for all.
- The floor shows what the simulation reads from a tile — walkable, costly or
  wall — and not the scenery, which comes with the images.
- **The horde is on screen and closes in**, going around obstacles. It is seeded
  at startup and does not arrive in waves: the pressure schedule comes later.
  Creatures overlap in depth order, and the player is drawn in front of those
  sharing the same depth.

### Worth knowing

- **The window scales the image by a factor that is not an integer** if you
  resize it: pixels then come out uneven. The internal buffer itself is fixed.
  Integer scaling is a display setting, and it will arrive with the settings
  screens.

## [0.1.0] — 2026-08-31 — La simulation nue

**Cette version ne se joue pas et ne montre rien.** Le binaire charge ses
ressources, monte le monde et s'arrête sur « à implémenter : étape 2 » : c'est le
comportement attendu, pas un défaut. L'étape 1 de la feuille de route est la
simulation, sans rendu par construction ; la 0.2.0 sera la première à ouvrir une
fenêtre.

Il n'y a donc ni affichage, ni clavier, ni porte, ni caisse, ni arme à ramasser,
ni montée de niveau, ni enchaînement de lieux. Ce qui existe se mesure en test.

### Ajouté

- Le format des lieux et des pièces, en JSON, `version_format` 1. Un lieu ne
  porte que des identifiants de pièces et leurs positions, et ne peut embarquer
  ni image ni son. La clé `$comment` y est admise partout et sert de
  commentaire ; toute autre clé inconnue fait refuser le fichier.
- Le coût de traversée d'une tuile de décor, `cout_traversee`. La flaque, le sol
  sale et le sol fissuré ralentissent qui les traverse — le joueur comme les
  créatures, qui les contournent plutôt que d'y passer. Il est exigé sur ce qui
  se franchit et refusé sur ce qui bloque.
- Un lieu livré, `assets/lieux/place/`, chargé au lancement par le même chemin
  qu'un lieu écrit par un tiers. Les ressources sont embarquées dans le binaire :
  l'exécutable se suffit à lui-même.
- **Un lieu est un dossier** : son `lieu.json`, son jeu de pièces et ses pièces.
  Les noms de pièces sont ainsi locaux à un lieu, et deux auteurs peuvent nommer
  chacun la leur `carrefour`. Le nom du dossier et le champ `identifiant` doivent
  s'accorder, faute de quoi le lieu est refusé.
- La simulation : bassins d'entités préalloués et références qui survivent à la
  suppression, champ de flux pondéré par les coûts de terrain, grille de densité
  qui desserre la horde, et tir automatique visant le plus proche à portée. Trois
  cents poursuivants convergent vers une cible mobile en contournant les
  obstacles, sans une allocation par tick.
- Le manifeste des armes, `assets/armes/manifeste.json`, le seul de `assets/`
  tenu à la main. Le tireur porte les valeurs de son tir — cadence, portée,
  dégâts, nombre de projectiles, vitesse — et un projectile ne porte que son
  apparence.

***

**This release does not play and shows nothing.** The binary loads its assets,
builds the world and stops on "à implémenter : étape 2": that is the expected
behaviour, not a defect. Step 1 of the roadmap is the simulation, without
rendering by construction; 0.2.0 will be the first to open a window.

So there is no display, no keyboard, no door, no crate, no weapon to pick up, no
level-up and no chaining of places. What exists is measured in tests.

### Added

- The level and room format, in JSON, `version_format` 1. A level carries only
  room identifiers and their positions, and can embed neither images nor sounds.
  The `$comment` key is allowed anywhere and serves as a comment; any other
  unknown key causes the file to be rejected.
- The traversal cost of a decor tile, `cout_traversee`. Puddles, dirty ground
  and cracked ground slow down whoever crosses them — the player as much as the
  creatures, which go around rather than through. It is required on whatever can
  be crossed and rejected on whatever blocks.
- A shipped level, `assets/lieux/place/`, loaded at startup through the same path
  as a level written by a third party. Assets are embedded in the binary: the
  executable stands on its own.
- **A level is a directory**: its `lieu.json`, its room set and its rooms. Room
  names are thus local to a level, and two authors can each name theirs
  `carrefour`. The directory name and the `identifiant` field must agree, or the
  level is rejected.
- The simulation: preallocated entity pools with references that survive
  removal, a flow field weighted by terrain costs, a density grid that loosens
  the horde, and automatic fire aiming at the nearest target in range. Three
  hundred pursuers close in on a moving target while going around obstacles,
  without a single allocation per tick.
- The weapon manifest, `assets/armes/manifeste.json`, the only one in `assets/`
  kept by hand. The shooter carries the values of its shot — cooldown, range,
  damage, projectile count, speed — and a projectile carries only its appearance.
