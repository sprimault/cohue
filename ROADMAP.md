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

Livré quand 300 poursuivants convergent vers une cible mobile en contournant
des obstacles, à budget d'allocation constant sur mille itérations.

## 2 — Rendu isométrique

`internal/render` : conversion écran/tuile, tri en profondeur par compartiments,
caméra en pixels entiers, tampon interne de 960×540 agrandi en entier.

Aucun asset final : des rectangles colorés. C'est le premier moment où la boucle
se juge à l'œil plutôt qu'en test.

Livré quand l'étape 1 est visible et jouable au clavier, sans scintillement de
sous-pixel.

## 3 — La boucle mort → relance

Écran de mort, relance sur une touche en moins d'une seconde, même
configuration. Une arme, un profil d'ennemi, une courbe de pression.

**Jalon décisif.** Si le prototype ne donne pas envie d'enchaîner cinq parties,
aucun sprite ne le sauvera, et il vaut mieux le savoir ici qu'après trois ans.

## 4 — Les profils d'ennemis

Les six comportements comme données, pas comme code : marcheur, sprinteur,
flanqueur, cracheur, bloqueur, éclateur. Un `EnemyProfile` par ligne de table.

Le spawner achète des ennemis dans un budget de pression par seconde plutôt que
de poser des compteurs fixes : c'est ce qui gardera les niveaux tiers jouables.

Les dégâts au contact sont continus, avec un plafond par seconde quel que soit le
nombre d'ennemis collés. Sans lui, un encerclement tue instantanément et la mort
devient illisible.

La résistance d'un profil s'exprime en touches de l'arme de base au premier
niveau, jamais en points absolus : l'arme grossit toute la run, un chiffre
absolu ne voudrait rien dire. Un multiplicateur adossé à la courbe de pression
la fait monter au fil du temps.

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

---

## Hors périmètre v1

Multijoueur, boutique, classements en ligne, portage mobile. Le mode web reste
compilable — c'est ce qui empêche une dépendance d'introduire du cgo sans qu'on
le voie — mais n'est pas publié.
