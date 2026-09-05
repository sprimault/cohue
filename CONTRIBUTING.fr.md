# Contribuer

English: [CONTRIBUTING.md](CONTRIBUTING.md)

## Comment ce projet est écrit

Le code est écrit en binôme avec un assistant, sous les règles de ce dépôt. Ce
n'est pas une note de bas de page : la méthode est le second objet du projet, et
la façon dont les règles sont posées en découle.

- **Les règles précèdent le code**, elles n'en sont pas déduites.
  [`docs/conception.md`](docs/conception.md) fait foi, et un désaccord entre le
  document et le code est un défaut du code.
- **Une décision porte ce qu'elle écarte.** Chaque arbitrage garde le motif des
  options rejetées, pour qu'on puisse le rouvrir sans le rejouer.
- **Ce qui se mesure se mesure.** Plusieurs décisions ont été corrigées par une
  mesure qui contredisait un raisonnement ; c'est la mesure qui tranche.
- **Les messages de commit ne racontent pas la fabrication.** Ils disent ce qui
  change et pourquoi — le reste est dans le diff.

Rien de cela ne s'applique différemment à une contribution extérieure : mêmes
contrôles, mêmes règles de style, même bilingue.

## Avant d'écrire du code

Ouvrir une issue d'abord, pour tout ce qui dépasse une correction. La conception
est écrite dans [`docs/conception.md`](docs/conception.md) ; la changer est une
discussion, pas un correctif.

## Ce qui se discute avant d'être écrit

Une pull request qui touche l'un de ces chemins sans discussion préalable sera
renvoyée à une issue, quelle que soit sa qualité — non par principe, mais parce
que ce sont les endroits où une modification en fait basculer d'autres.

- **[`docs/conception.md`](docs/conception.md)** fait foi. Le code s'y conforme,
  donc en changer une ligne change ce que le code doit faire.
- **Le format des niveaux et des pièces** est un contrat public : un niveau
  partagé aujourd'hui doit se charger demain. Un champ ajouté est optionnel, un
  champ retiré ou renommé casse tout ce qui circule.
- **`assets/` et `outils/`** vont ensemble. Le décor est généré : une forme se
  corrige dans `outils/decor_iso.py`, jamais dans le PNG — une retouche manuelle
  serait écrasée à la prochaine génération sans que personne ne le voie.
- **`.github/workflows/`** décide de ce qui est vérifié. La protection de
  branche exige les contrôles par leur nom, pas par leur contenu : un workflow
  modifié peut rendre vert un contrôle qui ne vérifie plus rien.

Le reste — code, tests, documentation d'accompagnement — se propose directement.

## Ce sur quoi une contribution est jugée

Les conventions de code et la doctrine de test sont dans
[`docs/go.md`](docs/go.md). Ce qui suit en est le résumé exigible.

- `make fmt && make lint && make test && make race && make vulncheck && make sec` passent.
- Toute déclaration a sa documentation. Les commentaires disent *pourquoi*, ils
  ne paraphrasent jamais la ligne suivante.
- Pas de bannière, pas d'emoji décoratif, ni dans le code, ni dans les logs, ni
  dans les messages de commit.
- **Aucune allocation dans la boucle de mise à jour.** Bassins préalloués,
  slices réutilisées, suppression par échange. Aucun pointeur ne sort d'un
  bassin.
- **Le déterminisme est préservé.** Rien ne lit l'horloge ni l'entropie système :
  l'aléatoire vient du générateur alimenté par la graine de la partie. Un
  changement qui fait diverger un rejeu est un défaut même si tous les tests
  passent.
- **Les données restent des données.** Un nouveau profil d'ennemi est une ligne
  de table, pas une branche dans une fonction.
- Une dépendance ajoutée entre dans `THIRD-PARTY-NOTICES`. Une ressource
  graphique ou sonore ajoutée entre dans [`CREDITS.md`](CREDITS.md), **dans le
  même commit**.
- Rien dans `internal/game` n'importe le rendu. Les runners sont sans écran ; un
  test qui exige une fenêtre n'a pas sa place dans la suite par défaut.

## Livraison

**Un lot, une branche, un commit.** La branche part de `master` à jour et se
nomme `<type>/<sujet>`, où le type est le préfixe conventionnel de son commit :
`feat/`, `fix/`, `docs/`, `chore/`, `test/`, `refactor/`. Ne pas enchaîner deux
lots sur la même branche — chacun doit rester relisible et annulable seul.

Elle retourne dans `master` **par une pull request**, jamais par une fusion
locale : c'est la PR qui laisse la trace de ce qui a été livré, et sa fusion qui
supprime la branche des deux côtés.

**Vérifier avant de pousser, pas après :**

```
make fmt && make lint && make test && make race && make vulncheck && make sec
```

**La liste est fixe et se passe entière**, jamais réduite à ce qui touche au
changement qu'on vient d'écrire. Composer sa liste revient à ne vérifier que ce
qu'on a déjà en tête, et le défaut est ailleurs par construction : s'il avait été
là où l'on regardait, on l'aurait vu en écrivant. **Celui qui trouve est celui
qu'on n'avait pas de raison de lancer.**

Deux contrôles s'y ajoutent dès qu'un changement touche `assets/`, `outils/` ou
la forme d'un fichier :

```
make ressources-verif && make entetes
```

Un troisième dès qu'une dépendance entre ou sort, parce que les notices
accompagnent chaque archive publiée :

```
make notices
```

Un quatrième dès qu'une section de `docs/go.md` est ajoutée, retirée ou
renommée : son sommaire indexe par la question qu'on se pose et non par les
titres, si bien qu'une ancre morte n'y produit aucune erreur — elle ne fait
rien, et un sommaire que rien ne vérifie égare au lieu de guider.

```
make sommaire
```

`govulncheck` interroge sa base d'avis **en direct** : un job vert le matin peut
être rouge l'après-midi sur exactement le même code. Ne pas se reposer sur
l'intégration continue seule, qui valide une fois la branche déjà poussée.

**La section du `CHANGELOG` part avec le lot**, pas au moment du tag : elle est
relue en pull request, donc au moment où elle compte. La publication en tire le
nom et les notes de la version, et une section absente l'arrête.

**La documentation part avec le changement.** Avant de commiter, vérifier ce que
le changement rend faux ailleurs : l'état annoncé dans le README, une décision de
[`docs/conception.md`](docs/conception.md), une étape de
[`ROADMAP.md`](ROADMAP.md). Cas propre à ce projet : une décision de conception
qui change rend le document faux, et ce document fait foi — le code ne prend
jamais de l'avance sur lui.

**Un message dit ce qui change et pourquoi**, en quelques lignes. Le défaut est
le titre seul : un corps n'existe que s'il porte quelque chose que le titre ne
dit pas et que le diff ne montre pas.

## Corriger une vulnérabilité sans en créer une autre

Ne pas adopter une version publiée **le jour même**, même corrective. Chercher
la plus ancienne qui suffit :

```
go list -m -versions <module>
```

Une version parue dans l'heure est le profil type d'une compromission de compte
mainteneur.

Un épinglage s'explique : un `require` figé plus bas que le dernier disponible
porte un commentaire de fin de ligne disant pourquoi, et **quand le retirer**.

## Trois numéros, à ne pas confondre

| Numéro | Où | Ce qu'il suit |
|---|---|---|
| version du dépôt | tag git | le binaire |
| `version_format` | chaque niveau et chaque pièce | le format de fichier |
| `empreinte_jeu_pieces` | chaque niveau | l'état réel du jeu de pièces |

Les deux derniers ne suivent pas SemVer. `version_format` est un entier :
ajouter un champ optionnel ne l'incrémente pas, tout le reste l'incrémente, et
un incrément oblige à écrire la migration des niveaux existants.

`empreinte_jeu_pieces` n'est pas une version mais une somme de contrôle : elle
change dès qu'une pièce livrée bouge. Sans elle, un niveau se chargerait en
silence avec une géométrie différente de celle qu'a construite son auteur.

Le dépôt suit SemVer avec la clause du zéro, définie dans
[`CHANGELOG.md`](CHANGELOG.md) : **en `0.x`, rien n'est imposé.**
Le mineur marque une étape de [`ROADMAP.md`](ROADMAP.md), pas une rupture d'API ;
tout le reste s'accumule en correctif. Conséquence directe : **le numéro ne
prévient de rien**, et ce sont les notes de version qui doivent dire ce qu'un
auteur de niveau doit reprendre.

## Langue

**Les identifiants sont en anglais** — répertoires, fichiers, paquets, types,
fonctions, champs : `Enemy`, `FlowField`, `Tile`. **La documentation est en
français** : godoc, commentaires, messages d'erreur et journaux. L'API se lit en
anglais parce que c'est du code ; le raisonnement se lit en français parce que
c'est de la pensée.

Messages de commit en français d'abord, anglais ensuite, dans un seul texte
séparé par `***`. Jamais `---` : `git am` le traite comme un séparateur de patch
et tronque tout ce qui suit.

Les contributions en anglais sont bienvenues et ne sont pas soumises à la règle
bilingue.

## Niveaux partagés

Un niveau est de la donnée, jamais du code : il ne référence que le vocabulaire
fourni par le binaire — pièces, objets, profils, événements.

**Aucun fichier binaire, sous aucune extension.** Ni image, ni son, ni exécutable.
C'est une règle mécanique et non un jugement : elle supprime toute question de
provenance et de droits sur ce qui transite par le jeu.
