# Feuille de route

Les numéros font foi : ce sont eux que portent les
`errors.New("à implémenter : étape N")` du code.

```
grep -rn "à implémenter : étape" internal cmd --include="*.go" --exclude="*_test.go" | wc -l
```

C'est la mesure d'avancement la plus honnête du projet. Elle descend toute
seule, et elle ne ment pas.

**Les numéros ordonnent les dépendances, pas le calendrier.** Ils ne se
renumérotent jamais — ce sont eux que portent les marqueurs du code. Une étape
qui apparaît s'ajoute à la fin, quelle que soit sa place logique.

**Un sous-ensemble d'étape peut être avancé**, quand l'écart entre l'écriture
d'un contrat et sa première mise à l'épreuve devient trop grand. Deux règles
alors : l'avancement nomme son périmètre exact **et ce qui reste à l'étape
d'origine** — c'est cette moitié-là qu'on oublie de vérifier en y arrivant —, et
il porte son motif. Au-delà de deux ou trois, ce ne sont plus des dépendances :
c'est que la numérotation est fausse, et c'est elle qu'il faut refaire,
marqueurs compris.

**Chaque étape franchie est publiée**, qu'elle donne à voir quelque chose ou non,
et qu'elle soit jouable ou non. L'étape N porte la version 0.N.0 — le
`CHANGELOG` dit déjà que le mineur marque une étape et non une rupture d'API ;
ce qui est ajouté ici est l'engagement à taguer, pas seulement à numéroter.

Deux conséquences pour les premières. Les notes d'une version qui ne se joue pas
doivent dire ce qu'elle **ne fait pas** : quelqu'un télécharge une archive dont
le binaire rend « à implémenter : étape N », et sans cette phrase il croit à un
défaut. Et l'étape 1 étant sans rendu par construction, la 0.2.0 sera la
première version qu'on puisse lancer pour voir quelque chose.

La conception complète est dans `docs/conception.md`. Ce fichier n'en est que
l'ordre d'exécution.

---

## 1 — Simulation nue

`internal/game` : bassin d'entités préalloué, champ de flux sur grille, poussée
de séparation par densité, tir automatique.

Rien à l'écran : la boucle tourne en mémoire et se mesure en test. La structure
des entités se fige ici — tableau de structures pleines, index plutôt que
pointeurs, profil partagé hors de l'entité. La reprendre plus tard toucherait
tout le code de jeu.

**Dépend de la lecture d'un lieu à pièce unique, sous-ensemble de l'étape 10
avancé ici.** Le champ de flux a besoin d'obstacles à contourner dès la première
ligne, et son interface n'est pas un format de niveau mais une grille de coûts.
Une carte bâtie en Go la remplirait aussi bien — mais le chargeur serait alors
écrit neuf étapes plus tard contre un contrat que rien n'aurait exercé, et c'est
au moment de l'écrire qu'on découvrirait que la grille attendue n'est pas celle
qu'on sait produire.

Ce qui est avancé se borne à cela : lire un fichier de lieu, le cuire en grille
de coûts. Un lieu d'une seule pièce rend l'assemblage trivial. Les connecteurs,
la validation topologique et la composition de plusieurs pièces restent à
l'étape 10, où ils ont leur place — les numéros ne bougent pas, et le marqueur
de l'étape 10 continue de désigner ce qu'il désignait.

Livré quand 300 poursuivants convergent vers une cible mobile en contournant
des obstacles, à budget d'allocation constant sur mille itérations.

## 2 — Rendu isométrique

`internal/render` : conversion écran/tuile, tri en profondeur par compartiments,
caméra en pixels entiers, tampon interne de 960×540.

**L'agrandissement du tampon vers la fenêtre n'est pas ici, mais à l'étape 15.**
Le tampon est fixe dès cette étape-ci, ce qui est la règle de pixel art ; le
facteur qui l'agrandit, lui, oblige à choisir entre bandes noires, taille de
fenêtre contrainte et redimensionnement par pas, et il se lit sur le facteur
d'échelle du système. C'est un réglage d'affichage, pas une décision de caméra.

Aucun asset final : des rectangles colorés. C'est le premier moment où la boucle
se juge à l'œil plutôt qu'en test.

Livré quand l'étape 1 est visible et jouable au clavier, sans scintillement de
sous-pixel.

## 3 — La boucle mort → relance

Écran de mort, relance sur une touche en moins d'une seconde, même
configuration. Une arme, un profil d'ennemi, une courbe de pression.

**Jalon éliminatoire.** Il tranche une seule question : le déplacement et le tir
sont-ils agréables ? Si la réponse est non, aucun sprite ne le sauvera et on ne
continue pas — c'est le seul jalon qui puisse arrêter le projet, et il arrive
tôt pour cette raison. Un oui, en revanche, ne prouve rien sur l'envie de
refaire : c'est l'étape 8 qui le dira.

Son critère se mesure plutôt qu'il ne se ressent : **si la bascule de puissance
n'est pas ressentie avant la minute 9**, la courbe est trop lente — voir le
chapitre 2 de la conception.

## 4 — Les profils d'ennemis

Les comportements comme données, pas comme code : marcheur, sprinteur,
flanqueur, cracheur, bloqueur, éclateur, soigneur. Un `EnemyProfile` par ligne
de table.

Le spawner achète des ennemis dans un budget de pression par seconde plutôt que
de poser des compteurs fixes : c'est ce qui gardera les niveaux tiers jouables.

Les dégâts au contact sont continus, avec un plafond par seconde quel que soit le
nombre d'ennemis collés. Sans lui, un encerclement tue instantanément et la mort
devient illisible.

La résistance d'un profil s'exprime en touches de l'arme de base au premier
niveau, jamais en points absolus : l'arme grossit toute la run, un chiffre
absolu ne voudrait rien dire. Un multiplicateur adossé à la courbe de pression
la fait monter au fil du temps.

**Y compris une porte et une caisse, sous-ensembles des étapes 8 et 7 avancés
ici** — non comme contenu mais comme sonde. Cinq étapes séparent le jalon
éliminatoire du jalon décisif, ce qui est long sans retour : un objectif
d'ouverture et une caisse à casser suffisent à sentir la tension « rester ou
partir » bien avant que tout soit écrit. Restent à leurs étapes : le temps mort
à la porte, le choix de branche, le score, le recyclage de la traîne, et tout ce
que l'étape 7 dit des ressources.

## 5 — Les assets

Chargement des manifestes de `assets/`, atlas, cycles d'animation avec cadence
et bouclage, orientation du sprite sur la direction de visée.

Le moteur ne connaît que des profils et des cycles, jamais un nom de fichier ni
un nombre d'images codé en dur.

## 6 — Armes, niveaux et synergies

Table d'armes et de passifs, trois choix par montée de niveau, recettes de
fusion. Armement de base infini ; armes lourdes à charges, déclenchement
conditionnel et lisible.

## 7 — Ressources et caisses

Caisse cassée en la traversant, avec délai de contact et ralentissement.
Blocante dans le champ de flux, rafraîchissement local à sa destruction.
Consommables limités à deux ou trois emplacements, sans menu.

## 8 — L'enchaînement de salles, points et score

Objectif d'ouverture de porte, temps mort à la porte, choix de branche.
Recyclage de la traîne d'ennemis, champ de flux calculé sur une fenêtre.

Score d'un lieu : points d'ennemis plus bonus de temps restant. Les deux
s'opposent — farmer rapporte, partir vite rapporte aussi — et c'est cette
tension qui donne du poids au choix de la porte. Affichage en fin de lieu, pas
pendant l'action.

**Jalon décisif.** C'est ici, et pas à l'étape 3, qu'on sait si l'on a envie de
relancer : un enchaînement complet dit ce que le seul déplacement ne pouvait pas
dire. Celui-là ne peut que valider — c'est l'étape 3 qui pouvait arrêter.

## 9 — La signalétique

Panneaux orientés au chargement depuis le chemin réel vers la sortie, en relais
de carrefour en carrefour. Boussole en filet de sécurité après quarante secondes
sans progression.

## 10 — Le format de pièces

Lecture, validation, assemblage en une seule tilemap à la cuisson. Passabilité
et hauteurs dérivées des tuiles, jamais saisies.

Les lieux livrés sont bâtis en pièces, comme un niveau tiers : même chemin de
code, une seule chose à déboguer.

## 11 — L'éditeur

Vue de dessus, pose de pièces, aimantation par connecteurs, jauges de contrôle
(aire jouable, culs-de-sac, boucles), lancement de partie sur place et retour au
même endroit.

Pas de mode tuiles dans cette étape : le noyau se réduit à la pose de pièces.

## 12 — Le partage

Niveau encodé en base64 compressé, dossier scanné au démarrage, validation au
chargement avec message explicite. Empreinte du jeu de pièces vérifiée.

Aucun asset importable : un niveau partagé ne référence que ce que le binaire fournit.

## 13 — Campagne et méta-progression

Graphe de lieux, déblocages qui ajoutent des options plutôt que de la puissance,
socle d'améliorations permanentes plafonné bas.

## 14 — Mode tuiles

Peinture d'une pièce avec les tuiles du jeu. Différé jusqu'ici pour de bonnes
raisons : le champ existe dans le format depuis l'étape 10, mais rien ne l'écrit.

## 15 — Les écrans de réglages

Sensibilité, volumes par catégorie de mixage, remappage des touches, difficulté,
et l'agrandissement du tampon interne vers la fenêtre — en entier, par `LayoutF`
et le facteur d'échelle du système, avec ce qu'il faut décider autour : plein
écran, bandes noires ou fenêtre contrainte.

Ici et pas plus tôt parce que leur contenu dépend de ce qu'il y aura à régler, et
que la moitié n'existe pas avant : les catégories de mixage sont posées dès le
manifeste des sons, la difficulté n'est un curseur qu'une fois la courbe de
pression écrite, et le remappage suppose la liste complète des commandes.

Ce qui se décidait tôt l'a été à l'étape 3 et vit dans la conception : ce qu'une
pause fige, quand on écrit sur le disque, ce qui persiste d'une run à l'autre.
Cette étape ne porte que le contenu, et elle existe pour qu'il ne soit pas une
dette sans domicile.

## 16 — Le catalogue de lieux

L'écran de démarrage énumère les lieux et laisse en choisir un. Après l'étape 12
parce qu'un catalogue sans lieux téléchargeables n'a rien à énumérer : la
dépendance est réelle, pas seulement chronologique.

**Traversé une fois par session, jamais par run.** Le chapitre 2 proscrit la
sélection de campagne parmi les quatre frictions de la relance, et fixe « même
configuration » : la mort rejoue le lieu en cours sans rien redemander. Un écran
au premier lancement ne coûte rien, y repasser après chaque mort coûte les
relances.

**La clé d'une entrée est le couple source et identifiant**, et un lieu
téléchargé ne masque pas le lieu livré du même nom. Les deux s'affichent avec
leur provenance : si le téléchargé masquait l'autre, le supprimer ferait changer
« son » lieu de contenu sans que rien n'ait touché à l'original.

**L'énumération est triée explicitement**, par source puis par identifiant, et
non rendue dans l'ordre de parcours du système de fichiers — qui n'est stable ni
d'une plateforme à l'autre, ni toujours d'un lancement au suivant. Une liste qui
change d'ordre toute seule donne l'impression d'un jeu instable.

**Un lieu invalide s'affiche barré, avec son motif.** Ni disparu en silence, où
son auteur ne comprendrait pas ce qui lui arrive, ni bloquant pour la liste
entière, où un fichier de trop empêcherait de jouer. C'est ce que la validation
sait déjà dire, porté à l'écran.

---

## Hors périmètre v1

Multijoueur, boutique, classements en ligne, portage mobile. Le mode web reste
compilable — c'est ce qui empêche une dépendance d'introduire du cgo sans qu'on
le voie — mais n'est pas publié.
