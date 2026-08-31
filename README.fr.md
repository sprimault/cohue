# Cohue

English: [README.md](README.md)

Un action-roguelite urbain en vue isométrique, sous pression de horde. Des
salles enchaînées, du tir automatique, une build qui se compose en quinze
minutes.

Apache 2.0 — [`LICENSE`](LICENSE), et [`CREDITS.md`](CREDITS.md) pour les
ressources graphiques.

Le joueur ne contrôle que son déplacement. Il monte de niveau en ramassant les
gemmes, casse des caisses en les traversant, lit la signalétique du décor pour
trouver la sortie, et passe au lieu suivant — parking, supermarché, quartier,
cinéma, station.

L'objectif de conception n'est pas la difficulté, c'est la relance : une touche,
moins d'une seconde, même configuration.

## État

**Le jeu s'ouvre et se traverse, il ne se joue pas encore.** Les étapes 1 et 2
sont livrées : la fenêtre montre un lieu, on s'y déplace au clavier, et la horde
converge en contournant les obstacles.

Ce qui manque pour que ce soit un jeu : la mort et la relance, les portes à
ouvrir, les caisses à casser, les armes à ramasser, les montées de niveau, et
l'enchaînement des lieux. Le décor est fait de rectangles colorés — les images
existent, mais rien ne les charge encore.

Ce qui est acquis : la simulation à trois cents créatures sans allocation par
tick, le rendu isométrique avec sa projection, sa caméra et son tri en
profondeur, le décor généré avec son manifeste, et les personnages générés eux
aussi — un gabarit par famille, huit orientations, et les valeurs de jeu dans le
même manifeste que le rendu.

- [`docs/conception.md`](docs/conception.md) — la conception complète
- [`ROADMAP.md`](ROADMAP.md) — les étapes et ce qui est hors périmètre v1

## Les ressources

Le décor est **généré** par `outils/decor_iso.py`, de façon déterministe : une
forme se corrige dans le script, jamais dans le PNG.

```
make decors      # les lieux, six thèmes
make figurines   # les créatures, six gabarits, variantes de teinte
make objets      # ce qui se ramasse ou se tire
make sons        # les bruitages, par synthèse
```

Décor et personnages sont **générés et versionnés** : ils sortent des mêmes
primitives isométriques, ne dépendent d'aucune source tierce, et les images sont
embarquées dans l'exécutable. Un dépôt fraîchement cloné compile sans rien
installer de plus.

C'est aussi ce qui rend les créatures contribuables : changer les proportions
d'un profil ou ajouter un gabarit se fait dans du code relisible en pull
request, pas dans un PNG.

## Niveaux et partage

Un niveau est une liste de pièces posées, pas une carte : quelques centaines
d'octets, copiables en base64 dans un message.

**Un niveau partagé ne contient aucune image ni aucun son.** Il ne référence que
ce que le binaire fournit — pièces, objets, profils, événements. C'est ce qui
rend la distribution triviale et évite toute question de droits sur ce qui
transite par le jeu.

## Construction

```
make build     # binaire local dans .tmp/
make test
make lint
```

Ebitengine est en Go pur sous Windows : aucun compilateur C pour le
développement quotidien. Les cibles Linux et macOS exigent du natif et passent
par l'intégration continue.

[`docs/construction.md`](docs/construction.md) porte le reste : les cibles
locales, la régénération des ressources et la matrice de compilation.
