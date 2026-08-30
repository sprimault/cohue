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

## [Non publié]

Rien de jouable. Le dépôt porte la conception, les outils de fabrication des
ressources et l'échafaudage ; l'étape 1 de la feuille de route est la simulation
nue.

### Ajouté

- Le format des lieux et des pièces, en JSON, `version_format` 1. Un lieu ne
  porte que des identifiants de pièces et leurs positions, et ne peut embarquer
  ni image ni son. La clé `$comment` y est admise partout et sert de
  commentaire ; toute autre clé inconnue fait refuser le fichier.
- Le coût de traversée d'une tuile de décor, `cout_traversee`. La flaque, le sol
  sale et le sol fissuré ralentissent qui les traverse — le joueur comme les
  créatures, qui les contournent plutôt que d'y passer. Il est exigé sur ce qui
  se franchit et refusé sur ce qui bloque.
- Un lieu livré, `assets/lieux/place.json`, chargé au lancement par le même
  chemin qu'un lieu écrit par un tiers. Les ressources sont embarquées dans le
  binaire : l'exécutable se suffit à lui-même.

***

Nothing playable. The repository holds the design, the asset-building tools and
the scaffolding; step 1 of the roadmap is the bare simulation.

### Added

- The level and room format, in JSON, `version_format` 1. A level carries only
  room identifiers and their positions, and can embed neither images nor sounds.
  The `$comment` key is allowed anywhere and serves as a comment; any other
  unknown key causes the file to be rejected.
- The traversal cost of a decor tile, `cout_traversee`. Puddles, dirty ground
  and cracked ground slow down whoever crosses them — the player as much as the
  creatures, which go around rather than through. It is required on whatever can
  be crossed and rejected on whatever blocks.
- A shipped level, `assets/lieux/place.json`, loaded at startup through the same
  path as a level written by a third party. Assets are embedded in the binary:
  the executable stands on its own.
