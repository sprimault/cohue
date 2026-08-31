# Conventions de code

Ce que ce dépôt attend d'un changement en Go, et pourquoi. La conception est
dans [`conception.md`](conception.md) ; ce document ne parle que du code.

---

## 1. Bibliothèques

Bibliothèque standard par défaut. Pas de framework de ligne de commande, pas de
bibliothèque d'assertion de test, pas de gestionnaire d'état. La question avant
d'ajouter quoi que ce soit : « qu'est-ce que ça m'évite d'écrire », pas « est-ce
que c'est répandu ».

Les exceptions connues d'avance :

| Module | Ce qu'il apporte |
|---|---|
| `github.com/hajimehoshi/ebiten/v2` | rendu, entrées, son |

Un seul candidat non tranché : une bibliothèque d'interface pour l'éditeur,
`ebitenui` ou tout dessiner à la main.

Le lecteur TOML n'en est plus un. Tout ce que le jeu lit est en JSON, y compris
les pièces, que le TOML rendrait pourtant plus agréables à écrire : un lieu
partagé est du JSON compact compressé, et un second format sur le même objet
imposerait une conversion, donc deux représentations qui divergent. La
bibliothèque standard suffit.

Une dépendance ajoutée entre dans `THIRD-PARTY-NOTICES`, qui accompagne le
binaire dans chaque archive. Les ressources graphiques suivent un chemin
distinct : [`../CREDITS.md`](../CREDITS.md), tenu dans le même commit que
l'asset.

**Une dépendance qui ne vit que dans des fichiers de test ne se juge pas au même
critère.** Elle n'entre dans aucun binaire, ne peut pas introduire de cgo dans
une cible publiée, et disparaît du graphe dès qu'on retire le test.

## 2. cgo n'est pas uniforme, et c'est la seule entorse

| Cible | `CGO_ENABLED` |
|---|---|
| windows/amd64 | `0` — Ebitengine y est en Go pur |
| js/wasm | `0` |
| linux, darwin | `1` — Ebitengine y passe par les API système |

Le mettre à `0` sur darwin ne produit pas une erreur claire mais un échec de
liaison obscur. C'est une variable de matrice, jamais une valeur globale.

Conséquence à connaître : le binaire Linux est lié à la glibc, il n'y a pas de
construction statique. Sans importance pour un jeu de bureau, mais ça interdit
une image `scratch`.

**Aucune autre dépendance n'a le droit d'introduire cgo.** Ebitengine
est choisi pour cela : une seconde source rendrait les cibles Windows et
WebAssembly impossibles. C'est aussi pourquoi `js/wasm` continue d'être compilé
en intégration continue sans être publié : il empêche une dépendance d'en
ajouter par surprise, comme `windows/amd64` qui est compilé sans cgo lui aussi.

## 3. La langue des identifiants

**La règle est dans [`../CONTRIBUTING.fr.md`](../CONTRIBUTING.fr.md)** :
identifiants en anglais, documentation en français. Ce qui suit est ce qu'elle
laisse ouvert, et qui se tranche en écrivant du Go.

Une variable locale ou un paramètre garde le mot du domaine quand
[`conception.md`](conception.md) l'a nommé en français : `graine`, `essai`,
`trame`. Ce ne sont pas des noms d'API mais les mots du document, et les traduire
imposerait un aller-retour à chaque lecture croisée. `Generate(graine int64, p
Settings)` est donc correct : la signature se lit en anglais, le paramètre porte
le mot qu'on relira dans la conception.

Règle d'arbitrage quand les deux se rencontrent : **la cohérence prime sur la
langue.** On ne mélange pas `Enemy` et `Bassin` dans la même structure — le type
décide, pas la préférence du moment.

Un cas à part, quand il arrivera : les libellés d'interface viennent du
dictionnaire de traduction et existent dans les deux langues. Leurs clés —
`pref_audio`, `arme_lourde` — ne sont pas du code et relèvent de la règle des
données.

## 4. En-tête de fichier

Tout fichier source commence par deux lignes, dans la syntaxe de commentaire de
son format :

```go
// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0
```

`#` en Python et en YAML. Une ligne vide sépare l'en-tête de ce qui suit, faute
de quoi le copyright devient la godoc du paquet — ou, en Python, le docstring de
module cesse d'en être un.

JSON n'a pas de commentaires : l'en-tête y est un `$comment` en **première
clé**, le mot-clé que JSON Schema réserve à cet usage et que les validateurs
ignorent. C'est ce que portent les manifestes de `assets/`, écrits par
[`../outils/manifestes.py`](../outils/manifestes.py). La même clé sert de
commentaire ordinaire **partout ailleurs dans le document** : c'est ce qui rend
le JSON tenable pour des pièces écrites à la main. Un fichier partagé se lit en
refusant les clés inconnues — sinon une faute de frappe prend la valeur par
défaut sans rien dire — et `$comment` est la seule clé exemptée de ce refus.

**La liste des dispensés est close** :

| Dispensé | Pourquoi |
|---|---|
| `.editorconfig`, `.gitattributes`, `.gitignore`, `.golangci.yml`, `Makefile` | outillage que le dépôt ne redistribue dans aucune archive |
| `go.mod`, `go.sum` | réécrits par la chaîne Go, qui n'y préserverait rien |
| `LICENSE`, `NOTICE`, `THIRD-PARTY-NOTICES` | ce sont les mentions de licence, elles ne s'en préfixent pas |
| les documents Markdown | ils s'adressent à un lecteur, pas à un compilateur ; la licence du dépôt est déclarée en tête du `README` et dans `LICENSE` |
| les images et les sons | un format binaire ne porte pas de commentaire ; le manifeste de leur lot porte la mention pour eux |

Tout le reste en porte un, y compris un fichier d'une ligne. `make entetes` le
vérifie, et la vérification tourne en intégration continue : une règle que rien
ne contrôle est une règle qu'on découvre violée six mois plus tard, sur des
fichiers que plus personne ne relit.

## 5. Structure

- Une struct de dépendances par paquet, construite dans `cmd/cohue/main.go`.
- Constructeurs `New…` rendant `(*T, error)` si l'initialisation peut échouer.
- Interfaces définies côté consommateur, pas côté implémentation.
- Une interface de plus de trois méthodes signale en général deux
  responsabilités mélangées. Une exception se justifie en godoc, méthode par
  méthode.

Un fichier porte le nom du type qu'il déclare — `enemy.go` pour `Enemy`,
`flowfield.go` pour `FlowField`. C'est ce qui permet de trouver une déclaration
sans chercher.

**Un paquet ne se coupe pas en sous-paquets pour cause de volume.** En Go un
sous-répertoire est un paquet distinct : découper `internal/game` forcerait à
exporter des champs privés et créerait un cycle entre les bassins et la boucle.
Le découpage se fait par fichier, un sujet par fichier.

## 6. Erreurs

```go
if err != nil {
    return fmt.Errorf("chargement du niveau %s: %w", nom, err)
}
```

- Message en français, minuscules, sans ponctuation finale ni accent — la
  convention Go pour la casse, et le `grep` reste simple pour les accents.
- Sentinelles exportées quand l'appelant doit distinguer les cas :
  `ErrUnknownRoom`, `ErrDisconnectedLevel`.
- Pas de `panic` hors de `main`, pas de `log.Fatal` dans `internal/`.
- **Pas de `panic("à implémenter")`.** Un bout non écrit rend
  `errors.New("à implémenter : étape N")`, où N renvoie à
  [`ROADMAP.md`](../ROADMAP.md) : la compilation passe, le programme échoue
  proprement, et un `grep` retrouve la liste. C'est la mesure d'avancement du
  projet, et elle ne ment pas.
- **Une signature sans `error` porte quand même son marqueur**, en commentaire
  sur la première ligne du corps. Sans lui la fonction est invisible au `grep`
  et le décompte ment.

**Un manquement dit où il est.** Fichier, ligne quand le décodeur la connaît, et
chemin complet de la clé — `ability.lookout.effect[0].target`, pas « cible
invalide ». Le chemin se porte dans l'erreur au fur et à mesure de la descente,
il ne se reconstitue pas après coup. Et ce qui est attendu s'énonce avec ce qui
est refusé : la liste des valeurs connues vaut mieux qu'un adjectif.

Un niveau invalide fait échouer son chargement entier plutôt que d'être chargé à
moitié : une pièce manquante laisserait un trou dans la carte, et le flow field
y enverrait les ennemis tourner en rond. Ses manquements sont listés en une
fois — qui met au point un niveau veut la liste, pas un
aller-retour par erreur.

**Deux temps, deux comportements, et c'est délibéré.** Le décodage refuse au
premier écart ; la validation liste tout.

`encoding/json` s'arrête à la première erreur, et on ne cherche pas à contourner
ça par une passe préalable en `map[string]any` qui collecterait les clés
inconnues : deux passes, ce sont deux vérités sur le même fichier, avec leurs
tolérances propres, et le jour où elles divergent un fichier franchit la
première pour échouer à la seconde avec un message qui ne correspond à rien de
ce qu'on vient de lui dire.

La restriction est d'ailleurs ce qui est juste, pas un repli. Une clé inconnue
est presque toujours une faute de frappe isolée — un `rotaton` — et la corriger
fait avancer d'un cran. Les manquements de validation arrivent par grappes : une
pièce absente, un ancrage qui manque, une sortie inatteignable. C'est là que
l'aller-retour coûte, et c'est là que la promesse porte.

En échange, le message du décodeur donne le chemin de la clé fautive et
**ressort tel quel** jusqu'à l'auteur du niveau. Le remplacer par « niveau
invalide » détruirait la seule information utile de la ligne.

### Ce que le décodage d'un fichier partagé exige

Trois règles, qui se tiennent ensemble :

- **`DisallowUnknownFields`.** Sans lui, `rotaton` au lieu de `rotation` se
  charge en silence avec la valeur par défaut, et l'auteur ne comprend pas
  pourquoi sa pièce est de travers.
- **Chaque structure décodée tolère `$comment`**, et pas seulement la racine :
  une pièce se commente là où elle pose question, dans un ancrage ou un objet de
  liste. Un type `Commentable` embarqué dans chacune porte le champ, qui n'est
  jamais lu — le compilateur garantit alors que la convention et le contrôle ne
  divergent pas, et une structure qui l'oublierait ferait échouer le premier
  fichier commenté plutôt que le centième.
- **Une validation des champs obligatoires derrière le décodage.** Refuser les
  clés inconnues attrape la faute de frappe qui *ajoute* une clé, jamais celle
  qui en *supprime* une : un lieu sans graine se décode sans erreur et démarre
  sur la graine zéro, qui est une graine valide.

C'est dans cette validation, et nulle part ailleurs, que « listés en une fois »
devient vrai. C'est donc là qu'il faut résister au `return` sur le premier
manquement : elle accumule et rend tout, y compris quand le premier défaut rend
les suivants prévisibles.

## 7. Écriture de fichiers

`0o600` pour un fichier, `0o750` pour un dossier. Ce que le jeu écrit — niveaux
extraits, sauvegardes, préférences — appartient à qui l'a lancé et n'a aucune
raison d'être lisible par les autres comptes de la machine.

Ce sont les seuils qu'exige `gosec`, et aucun cas n'a justifié de s'en écarter.
Le jour où il y en aura un, c'est un `#nosec` commenté, pas un assouplissement
global.

## 8. Une affirmation de nombre s'adosse à un test, ou perd son quantificateur

« Le seul endroit où », « les quatre recopies », « une par profil », « dix
dessins en moyenne ». Une affirmation de quantité est vérifiable ; une affirmation de
principe ne l'est pas — c'est ce qui rend celle-là exigible et donne un critère
mécanique : chercher ces tournures, regarder lesquelles ont un test en face.

Le danger n'est pas qu'elle devienne fausse, c'est qu'elle **empêche de
chercher**. Une phrase qui affirme l'unicité d'un garde-fou dispense de vérifier
s'il y en a un second, et une relecture ne se déclenche pas quand quelqu'un
ajoute un cas ailleurs.

Deux formes, deux remèdes. Une affirmation d'unicité sur une **catégorie
fermée** — « le seul champ non sérialisé de `Game` » — se vérifie une fois et
tient. Sur une **catégorie ouverte** — les endroits où une recopie est admise,
les causes d'un échec — elle est fausse dès le prochain ajout, et c'est là qu'il
faut un test.

Vaut pour les documents : un chiffre qui **décrit le code** s'adosse à un test
qui lit le document, un chiffre qui **décide** est exercé par le test de
conformité, un chiffre **mesuré** porte sa date et sa mesure rejouable — voir
`-maj-mesures`. Un quantificateur qui n'est aucun des trois n'a rien à faire là.

## 9. Journalisation

`log/slog` uniquement, structuré. Clés en anglais, message en français :

```go
slog.Info("niveau charge", "name", nom, "pieces", nb, "graine", graine)
```

Jamais de `fmt.Println` de débogage laissé derrière soi. Et rien dans la boucle
de mise à jour : un log par image traverse le disque soixante fois par seconde.

## 10. Concurrence

- `context.Context` en premier paramètre de toute fonction faisant de
  l'entrée-sortie, propagé jusqu'au bout. Pas de `context.Background()` hors de
  `main` et des tests.
- La simulation d'équilibrage parallélise des milliers de parties, **avec un
  plafond** : `runtime.NumCPU()`, pas une goroutine par partie.
- Toute goroutine lancée a une condition d'arrêt explicite et est attendue.
- **Le parallélisme ne transparaît jamais dans un résultat.** Les runs simulées
  sont agrégées dans l'ordre de leur graine, jamais dans l'ordre d'arrivée.
- **La boucle de jeu reste mono-goroutine.** Paralléliser la mise à jour des
  entités casserait le déterminisme pour un gain nul à ce volume ; ce qui se
  parallélise, c'est la simulation d'équilibrage hors partie.
- `make race` doit passer.

---


## Le bassin d'entités

Les entités de jeu vivent dans des bassins préalloués, jamais dans des slices
qui croissent. Le motif est le même pour les ennemis, les projectiles, les
gemmes, les caisses, les particules et les cadavres :

```go
type Pool struct {
    enemies []Enemy // capacité fixe, jamais réallouée
    active  int
}
```

Trois règles qui ne se négocient pas — elles sont dans les invariants :

- Suppression par échange avec le dernier actif, puis décrément. Pas d'`append`
  dans la boucle de mise à jour, pas de trou.
- **Aucun pointeur ne sort du bassin.** Après un échange, un `*Enemy` conservé
  ailleurs désigne une autre entité. Une référence qui vit plusieurs images est
  un couple identifiant + génération, que le bassin résout en place courante.
  **L'identifiant n'est pas cette place** : l'échange ramène la dernière entité
  dans le trou, si bien qu'une référence indexée par la place se briserait parce
  qu'une *autre* entité est morte. Le bassin tient donc la redirection dans les
  deux sens, sur des tableaux qui ne se compactent pas.
- Ce qui est partagé par un type d'ennemi — vitesse, PV max, poids de
  séparation — vit dans un `[]EnemyProfile`, et l'entité n'en garde que l'index.

En itérant, prendre l'adresse plutôt que la copie :

```go
for i := range p.enemies[:p.active] {
    e := &p.enemies[i]
}
```

Le tri en profondeur travaille sur une slice d'indices réutilisée avec
`indices = indices[:0]`, jamais sur le bassin : l'échange casse l'ordre.

## Les données ne sont pas du code

Profils d'ennemis, courbes de pression, passifs, pièces, lieux : des tables et
des fichiers. Une nouvelle sorte d'ennemi est une ligne de table, pas une
branche dans une fonction. Si un profil demande du code, c'est qu'il manque un
paramètre au profil.

Même règle pour les manifestes : le moteur ne connaît que des profils et des
cycles, jamais un nom de fichier ni un nombre d'images codés en dur.

## Deux descriptions de la même chose finissent par diverger

Avant d'ajouter une seconde représentation de quelque chose qui en a déjà une,
une question : **qu'est-ce qui garantit qu'elles restent d'accord ?** Quand la
réponse est « la vigilance », c'est non.

La deuxième représentation paraît toujours peu coûteuse, et c'est toujours elle
qui se désynchronise. Trois cas déjà écartés, qui sont le même :

- une pièce en TOML et un lieu partagé en JSON compact — le partage imposerait
  une conversion, donc deux formes du même objet ;
- une passe de décodage en `map[string]any` devant la passe typée, chacune avec
  ses tolérances propres, jusqu'au fichier qui franchit l'une et échoue à
  l'autre ;
- une liste de champs à exclure tenue à côté de l'empreinte d'état, qui se
  périme au premier champ ajouté.

La question laisse passer les cas où la réponse est bonne, et c'est pour cela
qu'elle vaut mieux qu'une interdiction, qui se ferait contourner au premier
d'entre eux : un cache reconstruit à partir de sa source ne peut pas en
diverger, un index dérivé non plus. Ce qui est refusé, c'est la seconde
description qu'on maintient à la main.

**Le cas aggravé est celui où la seconde description est le contrôle lui-même.**
Les trois exemples ci-dessus sont des descriptions parallèles qu'on maintient à
la main : on sait qu'on les tient, même quand on oublie de le faire. Un contrôle
passe pour sûr par nature, puisque vérifier est sa fonction. Quand il teste une
valeur périmée, il ne se contente donc pas d'être inutile — il **protège**
l'écart : l'endroit précis censé le signaler le certifie, et il n'en reste plus
aucun pour le voir.

Une décision sur les destructibles a ainsi survécu dans les données après avoir
été retirée du document, tout en restant au vert. D'où le réflexe, quand une
décision touche des données : chercher tout de suite **qui les contrôle**. Le
contrôle est le dernier endroit où l'ancienne décision reste vraie.

C'est le pendant, sur la représentation, de ce que ce document exige du
comportement : un invariant se vérifie, il ne se surveille pas. Quand une
seconde description est vraiment nécessaire sans être dérivable, alors c'est un
test qui tient l'accord — pas une relecture.

## La valeur zéro ne se partage pas entre l'absence et une valeur légitime

Un défaut de cette famille ne naît pas d'une erreur qu'on pourrait relire, mais
d'une omission qui n'a laissé aucune trace à relire. La valeur zéro est ce qu'on
obtient **sans rien faire** : un champ qu'on n'a pas rempli, une structure
copiée, une tranche agrandie. Quand elle signifie déjà quelque chose de
légitime, l'oubli devient indiscernable du réglage, et il n'y a rien à
l'endroit du défaut qui puisse attirer un regard.

Ce n'est pas la valeur zéro utile qui est en cause — `sync.Mutex` et
`bytes.Buffer` s'en servent bien. Ce qui l'est est plus étroit : **un champ dont
le zéro est à la fois une valeur légitime et ce qu'un oubli produit.**

Les cas de ce dépôt ont le même diagnostic et deux remèdes opposés, ce qui est
la raison pour laquelle le motif est difficile à voir :

| Le champ | Ce que zéro veut dire | Ce qui déménage |
|---|---|---|
| `cout_traversee` | la traversée est gratuite | l'absence, dans un `*int` |
| `max_simultane` | aucun plafond de simultanéité | l'absence, dans un `*int` |
| la génération d'un `Handle` | rien | la validité : les compteurs partent à 1 |

D'où le critère, qui choisit au lieu d'interdire : **le zéro a-t-il une
signification métier ?** S'il en a une, elle ne se retire pas, et c'est
l'absence qui doit se représenter ailleurs. S'il n'en a pas, on lui en donne
une, et le zéro devient l'état que rien de valide ne produit — `Handle{}` ne
désigne alors aucune entité, ce qui est exactement ce qu'un champ oublié doit
valoir.

Cette règle est de la même famille que les deux du chapitre Tests, « un contrôle
privé de son entrée échoue » et « une planche que rien ne fabrique ne relit
rien » : toutes trois portent sur des défauts sans existence textuelle, qu'aucune
relecture ne peut donc trouver. Ce qui les attrape est une forme choisie de
telle sorte que le défaut ne puisse pas s'écrire.

## La projection isométrique ne se recalcule pas

Toute conversion écran/tuile passe par `internal/render`. Un calcul recopié sur
place est un endroit de plus à corriger le jour où la taille de tuile bouge, et
c'est celui qu'on oublie.

## 11. Tests

### Aucun test n'ouvre de fenêtre

Les runners d'intégration sont sans écran, et `internal/game` n'importe pas
Ebitengine, ce qui rend la règle facile à tenir.
Un test qui exigerait un serveur X virtuel est le signe qu'il n'a rien à faire
dans la suite par défaut : le rendu se juge à l'œil.

### Le test qui vend le projet

```
graine -> run simulée sans rendu -> état identique sur toutes les cibles
```

Une run se joue en mémoire, sans fenêtre, sur un nombre de ticks fixé : mêmes
apparitions, mêmes trajectoires, même état final. Sans lui, l'équilibrage n'est
pas comparable d'une version à l'autre et une mort injuste n'est pas
reproductible.

**Sur toutes les cibles, et pas seulement deux fois de suite ici.** Le job des
tests natifs exécute la suite sur windows/amd64, darwin/arm64 et linux/arm64 :
c'est là que se vérifie ce que la virgule fixe achète, et c'est ce qui sépare un
invariant d'une discipline. Un déterminisme qui ne tiendrait que sur la machine
qui l'a écrit ne porterait ni le classement par graine, ni le partage d'un défi.

L'empreinte comparée porte sur l'**état** — positions, points de vie,
générations, état de chaque flux aléatoire, parcourus dans l'ordre des index du
bassin — et jamais sur un résumé. Un compte d'entités vivantes ou un score
passerait au vert alors que deux trajectoires ont divergé puis se sont
recroisées, ce qui est exactement le cas recherché. L'ordre du parcours fait
partie de ce qui doit être stable : l'échange à la suppression le casse, donc
l'empreinte se calcule sur les index, jamais sur un parcours de `map`.

**Elle énumère ce qu'elle inclut, jamais ce qu'elle écarte.** Les champs
cosmétiques en sont exclus — la teinte d'un vêtement ne décide de rien, et un
second test l'exige en jouant deux fois la même graine avec des teintes forcées
différentes. Une liste de champs à ignorer se périmerait au premier champ
ajouté, et le test échouerait pour une raison sans rapport avec ce qu'il garde ;
une liste de ce qui compte laisse un champ nouveau dehors par défaut, ce qui
affaiblit le test au lieu de le casser. Des deux erreurs possibles, c'est celle
qui se rattrape.

Son jumeau, sur le budget : mille itérations à 300 entités sans une allocation.
`testing.AllocsPerRun` le dit en un chiffre, et c'est l'invariant le plus facile
à casser sans s'en apercevoir.

### Fichiers de référence

L'assemblage d'un niveau étant pur, il se teste sans rien : un fichier de niveau
en entrée, une carte attendue en sortie, comparaison stricte. Idem pour le champ
de flux sur une carte figée.

Le jeu de cas couvre délibérément les dispositions pénibles — salle très ouverte,
couloirs étroits, cul-de-sac, obstacle détruit en cours de route.

Mise à jour groupée des attendus derrière `go test ./internal/game -maj-attendus`,
jamais automatique : un attendu régénéré sans être relu ne teste plus rien. Trois
autres artefacts suivent le même motif — `-maj-mesures`, cité plus haut à propos
des quantificateurs, `-maj-notices` et `-maj-schemas`.

### Une planche que rien ne fabrique ne relit rien

Un artefact de relecture ne vaut que par le geste qui le produit et le regard qui
le suit ; c'est la règle de l'attendu régénéré sans être relu, prise par l'autre
bout. Une planche que le code sait dessiner mais qu'aucune cible n'écrit tient
donc lieu de contrôle sans en être un.

Un défaut d'orientation des personnages l'a montré : la planche affichait bien
les huit directions, dont les deux seules où le symptôme était visible, et aucune
commande ne la produisait. Ce que le contrôle automatique attrape — alpha,
silhouette, taille, palette — ne dit rien de ce que l'image représente, et il n'y
a pas de moyen d'y suppléer autrement que de regarder.

D'où la règle : **toute planche de relecture sort d'une cible du `Makefile`**,
dans le même geste que ce qu'elle donne à relire. Ajouter des vues à une planche
que personne ne produit ne corrige rien et donne le sentiment du contraire.

### Un contrôle privé de son entrée échoue, il ne passe pas

Quand un contrôle dépend d'une information venue d'ailleurs — un nom de dossier,
une liste, un manifeste —, son absence est une erreur et jamais un passage en
silence. Sinon on le désarme sans le toucher, en changeant seulement la façon de
l'appeler, et il continue de passer au vert en ne vérifiant plus rien.

Le chargeur de lieux en donne le cas : il confronte le nom du dossier au champ
`identifiant`. Un `fs.FS` monté directement sur le dossier du lieu ne lui laisse
plus de nom à confronter — le contrôle ne devient pas faux, il perd sa prise.
D'où le refus de `Load(".")`, et son propre test.

C'est la même famille que ci-dessus, prise encore ailleurs : un contrôle qui
certifie l'écart au lieu de le signaler.

### Cas limites du jeu

Chaque règle qui peut se contredire a son test. Sur ce jeu, la liste de départ :

- ennemi acculé contre un obstacle bas, avec la poussée de séparation qui le
  pousse dans le décor ;
- caisse détruite pendant qu'un ennemi la traverse — le champ de flux se
  rafraîchit sous ses pieds ;
- aucune cible à portée, l'arme ne doit pas tirer ni consommer sa cadence ;
- le joueur pris entre un Vigile et un mur, seul cas où un corps ennemi l'arrête ;
- sortie atteinte avant l'objectif : la porte reste fermée ;
- bassin plein au moment d'une vague, le spawner ne doit ni allouer ni écraser ;
- deux projectiles atteignant le même ennemi dans le même tick : une seule mort,
  un seul butin, une seule explosion — et le second va chercher derrière ;
- bassin de cadavres plein : le plus ancien cède sa place, sans allocation.

### Un test se vérifie en le faisant échouer

Un test qui n'a jamais échoué ne prouve rien. Écrire l'invariant plutôt que le
cas nominal — pour un effet, qu'appliquer puis annuler rende un état identique à
l'original — et l'éprouver une fois en cassant volontairement ce qu'il garde.

C'est ce qui a révélé qu'une annulation laissait une tranche vide non nulle là
où il y avait `nil`, ce qui aurait fait diverger le rejeu du journal en JSON
sans que rien ne le signale.

### Un test qui bâtit son entrée ne teste pas le chemin qui la produit

Bâtir une entrée à la main est légitime pour **isoler un critère** — atteindre le
quatrième refus d'un validateur sans satisfaire les trois premiers par accident.
Ce qui ne l'est pas, c'est d'affirmer une propriété du **système monté** en
contournant le montage.

Sur un projet antérieur, un mode est resté inerte des mois durant : le test
posait lui-même la clé du mode, une clé que le manifeste livré n'a jamais
portée. Un second test vérifiait la bonne, les deux passaient au vert.

D'où un test de conformité qui charge les manifestes livrés de `assets/`, monte
le jeu et n'injecte rien.

**Et une entrée de test doit être une entrée que le système accepterait.**
Contourner le montage et forger un état invalide sont deux façons de tester ce
qui n'arrivera jamais : un niveau dont les pièces ne se raccordent pas n'est pas
un cas limite, c'est un fichier que le chargeur refuse.

### Un test aléatoire dit ce qu'il a visité, pas seulement qu'il est passé

Une run simulée ne couvre que ce que sa courbe de pression lui présente. Un
profil qui n'apparaît qu'au-delà de la troisième minute ne sera jamais visité par
un test de trente secondes, et son défaut survivra à toute la suite.

Compter les profils réellement apparus, et le dire dans la sortie du test.

### Conformité des manifestes

Un test lit les manifestes de `assets/` et vérifie que chaque cycle déclaré a sa
bande, que la largeur de la bande correspond au nombre d'images annoncé, et que
tout profil référencé par le code existe. C'est ce qui empêche qu'un
renommage de fichier casse le jeu au lancement plutôt qu'à la compilation.

### Ce qu'on ne teste pas

Pas de bouchon simulant Ebitengine, pas de comparaison de rendu pixel à pixel.
Un faux moteur ne testerait que la fidélité du faux.
