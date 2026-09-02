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

La section en cours s'écrit `## [Non publié]`, sans date ni titre, et chaque lot
y ajoute ce qu'il change : écrite au fil de l'eau, elle est relue en même temps
que le code qu'elle décrit, alors qu'une section rédigée le jour du tag résume
de mémoire un mois de travail. **Elle prend son numéro et sa date au moment du
tag**, faute de quoi la publication cherche une section qui n'existe pas et
s'arrête.

Chaque section est **bilingue, français d'abord, séparé par `***`** — les notes
de version sont ce que lit un auteur de niveau étranger avant de savoir s'il
doit reprendre son travail. Ce préambule reste en français : il n'est jamais
publié, et explique les conventions du dépôt à qui y contribue.

## [Non publié]

**On y meurt, et on relance.** La horde qui convergeait sans effet blesse
désormais au contact : collé sans se dégager, le joueur tombe en cinq secondes.
L'écran de mort laisse voir ce qui a eu raison de lui, et une touche remonte une
partie neuve sur le même lieu.

Les créatures abattues laissent des gemmes, qui font monter de niveau — et une
montée met le jeu en pause pour offrir **trois cartes**, dont on prend une.

### Ajouté

- **La vie du joueur et les dégâts au contact.** Une créature collée blesse en
  continu plutôt qu'en un coup, et le total encaissé par seconde est plafonné
  quel que soit le nombre d'ennemis au contact : l'encerclement reste très
  dangereux, mais il laisse une fenêtre pour se dégager plutôt que de tuer d'un
  bloc.
- **La mort fige la scène, une touche relance.** Le décor et la horde restent à
  l'écran, assombris, et Espace remonte la partie sans rien redemander — même
  lieu, aucune sélection à refaire.
- **Les gemmes.** Une créature laisse, à l'endroit exact où elle meurt, le
  nombre de gemmes que son profil déclare ; le joueur les ramasse en passant à
  courte portée.
- **L'expérience et la montée de niveau.** Les gemmes ramassées font monter d'un
  niveau, à un seuil qui croît d'un niveau au suivant : c'est le seuil qui monte,
  une gemme vaut la même chose du début à la fin de la partie. Un niveau est de
  plus donné toutes les quarante-cinq secondes sans rien ramasser, et il ne
  retire rien de ce qui est déjà acquis.
- **Le bandeau de la partie.** La vie, l'expérience et le temps écoulé
  s'affichent en haut de l'écran, du premier tick à la mort.
- **`assets/progression/manifeste.json`.** Les seuils de niveau, le plancher de
  temps et ce qu'une gemme rapporte se règlent dans ce fichier, tenu à la main
  comme la table d'armes.

- **Les trois cartes de la montée de niveau.** La horde se fige, trois cartes
  s'affichent en bas de l'écran, et les touches 1, 2 et 3 en prennent une. Deux
  axes améliorent l'arme — cadence et portée, six paliers chacun — et une
  troisième carte rend de la vie quand rien d'autre n'est disponible. La table
  se règle dans `assets/armes/manifeste.json`.

### Modifié

- **Les lieux vivent désormais dans une campagne**, `assets/campagnes/<nom>/`,
  et c'est elle qu'on compose et qu'on partagera : un dossier de salles dont le
  descripteur `campagne.json` dit par laquelle on commence. Le dossier existe
  pour cloisonner les noms, comme celui d'un lieu le fait pour ses pièces — sans
  lui, deux auteurs appellent tous les deux une salle « parking ».
- **Les fichiers d'un lieu portent des noms fixes.** Le jeu de pièces s'appelle
  `jeu.json` et les pièces vivent dans `pieces/`, là où deux noms libres au même
  niveau ne disaient pas lequel était une palette et lequel un plan. Un lieu
  existant se reprend en renommant son jeu de pièces et en déplaçant ses pièces.
- **Ce qu'une gemme rapporte a quitté `assets/objets/manifeste.json`** pour le
  manifeste de progression : une valeur vit à côté de ce qu'elle alimente, et
  celle-ci alimente les seuils. Le manifeste de progression nomme l'objet qu'elle
  décrit, et le contrôle des ressources exige qu'il existe.

### À savoir

- **Les gemmes tombées au contact se ramassent aussitôt**, la horde mourant
  collée au joueur. On en voit peu au sol, et c'est normal tant que rien ne tue
  à distance.
- **L'écran de choix ne s'atteint pas en jouant**, et ce n'est pas un défaut de
  la progression. Le semis provisoire pose trois cents créatures d'un coup :
  elles convergent en cinq secondes, et le joueur tombe vers la sixième avec au
  plus six gemmes des dix que le premier niveau demande. La courbe de pression,
  qui achète les créatures dans un budget croissant, est ce qui rendra la
  première montée atteignable.

***

**You can die here, and start over.** The horde that used to converge without
effect now hurts on contact: caught and unable to break free, the player falls
in five seconds. The death screen leaves in view whatever got the better of
them, and one key brings up a fresh run on the same place.

Killed creatures drop gems, and those gems raise the level — and a level up
pauses the game to offer **three cards**, one of which you take.

### Added

- **Player health and contact damage.** A creature pressed against the player
  hurts continuously rather than in a single blow, and the total taken per second
  is capped whatever the number of enemies in contact: being surrounded stays
  very dangerous, but it leaves a window to break free instead of killing
  outright.
- **Death freezes the scene, one key restarts.** The place and the horde stay on
  screen, darkened, and Space brings the run back with nothing to re-enter — same
  place, no selection to redo.
- **Gems.** A creature drops, exactly where it dies, as many gems as its profile
  declares; the player picks them up by passing within a short range.
- **Experience and levelling up.** Gems picked up raise the level, at a threshold
  that grows from one level to the next: it is the threshold that rises, a gem is
  worth the same from the start of a run to its end. A level is further granted
  every forty-five seconds without picking anything up, and it takes nothing away
  from what is already banked.
- **The run panel.** Health, experience and elapsed time are shown at the top of
  the screen, from the first tick to death.
- **`assets/progression/manifeste.json`.** Level thresholds, the time floor and
  what a gem is worth are tuned in this file, hand-held like the weapon table.

- **The three level-up cards.** The horde freezes, three cards appear at the
  bottom of the screen, and keys 1, 2 and 3 take one. Two axes improve the
  weapon — fire rate and range, six tiers each — and a third card restores
  health when nothing else is available. The table is tuned in
  `assets/armes/manifeste.json`.

### Changed

- **Places now live inside a campaign**, `assets/campagnes/<name>/`, and the
  campaign is what you compose and will share: a folder of rooms whose
  `campagne.json` descriptor says which one a run starts in. The folder exists to
  scope names, as a place's folder already does for its rooms — without it, two
  authors both call a room "parking".
- **A place's files carry fixed names.** The room set is `jeu.json` and rooms
  live in `pieces/`, where two free names at the same level did not say which was
  a palette and which a plan. An existing place is brought over by renaming its
  room set and moving its rooms.
- **What a gem is worth has left `assets/objets/manifeste.json`** for the
  progression manifest: a value lives next to what it feeds, and this one feeds
  the thresholds. The progression manifest names the object it describes, and the
  resource check requires that object to exist.

### Good to know

- **Gems dropped in contact are picked up at once**, since the horde dies
  pressed against the player. Few are seen on the ground, and that is expected
  as long as nothing kills at a distance.
- **The choice screen cannot be reached by playing**, and that is not a defect
  of progression. The provisional seeding places three hundred creatures at
  once: they converge in five seconds, and the player falls around the sixth
  with at most six of the ten gems the first level asks for. The pressure curve,
  which buys creatures from a growing budget, is what will make the first level
  up reachable.

## [0.2.1] — 2026-09-01 — Ce que la relecture a corrigé

**Rien ne change en jeu.** Un binaire 0.2.1 se comporte exactement comme un
0.2.0 : cette version corrige un défaut interne qu'on ne voit pas et remet la
documentation d'accord avec le code. Aucun lieu n'est à reprendre.

### Corrigé

- Le champ de flux allouait de la mémoire au fil de ses reconstructions : ses
  files naissaient vides et grandissaient à mesure que le joueur traversait le
  lieu, là où le budget d'allocation l'interdit. Elles reçoivent leur place au
  montage.
- Le contrôle des mentions de licence sautait tous les fichiers JSON, en
  confiant les manifestes à un contrôle qui ne regarde que `assets/` : ceux du
  reste du dépôt n'étaient donc lus par personne.

### Modifié

- La documentation dit ce que le code fait — l'ordre d'un pas de simulation, le
  domicile des plafonds de bassin, la clé du tri en profondeur, la structure
  d'une entité et l'origine d'une tuile. L'agrandissement du tampon vers la
  fenêtre est inscrit à l'étape des écrans de réglages, où il se décide.
- Les identifiants publics et les clés de journalisation sont en anglais, ce que
  les règles du dépôt demandaient déjà.

***

**Nothing changes in game.** A 0.2.1 binary behaves exactly like a 0.2.0 one:
this release fixes an internal defect no one can see and brings the
documentation back in line with the code. No level needs revisiting.

### Fixed

- The flow field allocated memory across its rebuilds: its queues started empty
  and grew as the player crossed the level, where the allocation budget forbids
  it. They now get their room at setup.
- The licence notice check skipped every JSON file, leaving manifests to a check
  that only looks at `assets/`: those elsewhere in the repository were read by
  no one.

### Changed

- The documentation states what the code does — the order of a simulation step,
  where the pool caps live, the depth sort key, an entity's layout and a tile's
  origin. Scaling the buffer to the window is recorded under the settings
  screens step, where it gets decided.
- Public identifiers and logging keys are in English, as the repository's own
  rules already required.

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
