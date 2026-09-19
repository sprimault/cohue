# Construction

## En local

```
make build      # binaire dans .tmp/
make run
make apercus    # les vues du rendu, dans .tmp/apercus/
make anime      # l'aperçu animé en WebP, pour ce qui se publie
make test
make race
make fmt
make lint
make vulncheck
make sec
make entetes    # la mention de licence, sur ce que git publie
make sommaire   # le sommaire par question de docs/go.md
make notices    # THIRD-PARTY-NOTICES contre ce que les cibles publiées lient
make cover
make binaries   # windows et wasm, dans dist/
make tools      # installe golangci-lint, govulncheck, gosec
```

`entetes`, `sommaire` et `notices` sont des étapes de l'intégration continue :
les lancer en local, c'est éviter un aller-retour. `notices` en particulier, que
seul un changement de dépendance fait bouger, et qui est rouge sur toute pull
request qui en ajoute une sans régénérer le fichier.

`make apercus` écrit des vues du rendu dans `.tmp/apercus/`. C'est le seul moyen
de relire `internal/render`, qui n'a pas de test et n'en aura pas — importer
Ebitengine initialise GLFW, qui panique sans écran. La planche pilote le rendu du
jeu et non une scène montée à côté, et elle est déterministe : c'est ce qui permet
de comparer une image d'avant et d'après un changement. Elle exige un écran, donc
elle ne tourne pas en intégration continue.

**Une des vues n'est pas une image.** `partie.gif` montre six secondes de jeu à
la troisième minute, jouées par le pilote du test de déterminisme : ce qui ne se
relit qu'en mouvement — une horde qui converge, une salve qui s'élargit, des
gemmes qui partent — n'a aucune place sur une planche. Elle s'achève au dernier
instant où le joueur est vivant, parce qu'un artefact s'arrête sur l'événement
qu'il montre et non sur un compte de pas.

`make anime` y ajoute `partie.webp`, treize fois plus léger à contenu
identique : c'est le format de ce qui se publie, et c'est lui que le `README`
montre depuis `docs/images/`. **La copie s'y fait à la main**, parce qu'une
cible qui écrirait dans le dépôt salirait `git status` à chaque exécution —
on ne publie pas à chaque relecture. **Sans perte, et ce n'est pas
un scrupule** — le mode avec perte est ici le plus lourd, étant fait pour des
dégradés quand le pixel art n'est qu'arêtes franches et aplats. Le GIF reste la
sortie du générateur, qui n'écrit qu'avec la bibliothèque standard.

`makefile.local` porte ce qui est propre au poste — `GOTMPDIR`, `GOCACHE` — et
n'est pas versionné. Passer par le Makefile plutôt que d'appeler `go` à la main :
une commande tapée directement perd ces variables, et l'échec est intermittent.

## Les ressources

```
python -m pip install -r outils/requirements.txt
```

Les versions y sont épinglées pour que le dessin reste stable d'une version de
Pillow à l'autre.

**Ce qui est comparé, ce sont les pixels d'une image, et les octets de tout le
reste.** Le PNG est un conteneur compressé et sa compression n'est pas portable :
à Pillow identique, la wheel Windows charge zlib-ng là où la wheel Linux charge
zlib, si bien que tout le catalogue diffère d'un système à l'autre sans qu'un
pixel ait bougé. Sons et manifestes, eux, sont comparés au bit près — ils le
passent sur les deux systèmes.

Le contrôle n'y perd rien : une retouche manuelle, un script modifié sans
régénération et un rendu de Pillow qui changerait déplacent tous des pixels.

**Régénérer et committer depuis un seul système.** `make ressources-verif`
passera partout, mais une régénération sur l'autre système réécrit les six cents
fichiers à dessin inchangé, et le diff devient illisible.

Le décor est produit par `outils/decor_iso.py` et **versionné** : il ne dépend
d'aucune source tierce.

**La matière des revêtements vient de `outils/textures/`**, des images carrées
jointives que le générateur applique sur la face supérieure d'un volume. Elles
sont des sources et non des ressources du jeu — elles vivent donc auprès des
outils, que `go:embed` n'embarque pas —, et le catalogue reste comparable à sa
régénération. Une matière se corrige en remplaçant sa texture, un tracé dans le
script.

**Quelques formes sont dessinées en entier**, et `DESSINES` les énumère dans
`outils/decor_iso.py` : ce qu'aucun volume ne saurait rendre. Le script en écrit l'entrée de manifeste — taille, ancrage,
élévation et couverture se mesurent sur le dessin livré — mais plus l'image, et
`make ressources-verif` ne les compare pas. Elles se remplacent en déposant un
nouveau PNG, jamais en régénérant, et le dossier porte son `LICENSE.txt` comme
celui d'un profil.

**Les personnages sont dessinés**, et `DESSINES` les énumère tous dans
`outils/figurines.py`. Leurs bandes sont des dessins tenus à la main, avec un
`LICENSE.txt` dans leur dossier, comme la police. Le générateur n'écrit plus que
leur entrée de manifeste, `make ressources-verif` ne les compare pas, et le
contrôle des images les voit comme les autres. Un dessin se remplace en
déposant ses bandes, jamais en régénérant.

Régénérer le décor après une modification du script :

```
python outils/decor_iso.py --sortie assets/decors
```

La sortie est déterministe. Un diff sur `assets/` après régénération sans
changement de script signale un problème — version de Pillow, ou retouche
manuelle d'un PNG.

## La police et l'interface

`outils/interface.py` produit deux dessins : la planche de glyphes et l'icône de
fenêtre, en trois tailles parmi lesquelles le système choisit. L'icône se dessine
sur une grille de cases plutôt qu'en pixels, si bien que les trois tailles
portent le même dessin sans qu'aucune soit l'agrandissement d'une autre ; sa
palette est la sienne et ne descend pas du rendu, dont les aplats sont
provisoires. **Les jauges, cadres et cases n'en sont pas** : ce sont des
rectangles unis dont la
taille dépend de leur contenu — une jauge suit la vie, une carte son texte, une
case son icône. Une image fixe devrait être étirée, ce qui casserait le pixel
entier, et le découpage en neuf morceaux n'a d'intérêt que si le cadre porte un
motif, biseau ou rivets, qu'aucune décision n'a demandé. Le rendu les dessine
donc, à partir des **réglages** que le manifeste porte : épaisseur du bord,
marge, hauteur de jauge, teintes. Aucune dimension d'élément n'y figure — elle
serait une seconde description de ce que le contenu impose.

La police, elle, est **tierce**, et c'est la première ressource du dépôt à
l'être — dessiner quarante glyphes lisibles et accentués est un métier, et le
temps qu'il coûterait ne va pas au jalon de l'étape 3.

Trois exigences la contraignent :

- **bitmap et non vectorielle.** Un rendu vectoriel dans le tampon interne
  réintroduit de l'anticrénelage, et le pixel art tombe avec lui.
- **les accents français**, qui ne sont pas acquis sur une fonte de jeu
  anglophone, et une cellule d'au moins sept pixels de haut : en deçà, l'accent
  n'a pas de ligne libre au-dessus de la capitale et vient se coller à la lettre.
  C'est une propriété de la fonte, vraie à n'importe quel agrandissement — donc
  indépendante de ce que l'étape 15 décidera de l'affichage.
- **un manifeste**, comme tout ce que le moteur lit. Taille de cellule, avance,
  hauteur de ligne et table des glyphes s'y déclarent ; sans lui les métriques
  deviennent des constantes du rendu, et changer de fonte redevient un changement
  de code — c'est-à-dire exactement ce qu'un manifeste existe pour éviter.
  `outils/interface.py` l'écrit en cuisant la planche, parce que ces mesures sont
  dérivées du dessin et non décidées : le fichier porte « ne pas modifier à la
  main », et une retouche serait écrasée à la prochaine régénération.

Entre deux fontes qui les respectent, le départage se fait sur le É. Beaucoup
réservent la ligne d'accent aux minuscules : le é est correct, et la capitale
accentuée bute sur le bord de sa cellule. Les libellés s'écrivant en casse
mixte — la conception dit pourquoi —, cela ne bloque rien ; mais avoir à changer
de fonte le jour d'un titre en capitales coûterait plus cher que de le vérifier
maintenant.

Le texte se pose en coordonnées entières, au même titre que la caméra : une fonte
bitmap posée à une position fractionnaire perd ce pour quoi on l'a choisie.

La ligne de [`../CREDITS.md`](../CREDITS.md) s'écrit avec le choix de la fonte,
sa règle étant de tenir dès l'introduction de la ressource.

## Les cibles publiées

Ebitengine ne se compile en croisé que vers Windows et WebAssembly ; Linux et
macOS exigent du natif. D'où la matrice multi-runners de `release.yml`.

`js/wasm` se compile mais ne se publie pas : l'archive serait inutilisable seule.
La cible reste dans la matrice parce qu'elle est ce qui empêche une dépendance
d'introduire du cgo sans qu'on le voie.

**Les deux cibles Linux sont épinglées sur une image nommée**, jamais
`ubuntu-latest`. Le binaire y est lié à la glibc, faute de construction
statique : **la version du runner est donc le plancher exigé chez le joueur**, et
une image qui migre le relève sans prévenir — l'exécutable cesse de démarrer sur
les distributions un peu anciennes, ce qu'on apprend par un rapport de joueur.

Le choix est asymétrique, et c'est ce qui le tranche : compiler sur une glibc
ancienne ne coûte rien, le binaire tournant aussi sur les systèmes récents, quand
compiler sur une récente exclut les anciens. Le jeu n'emploie aucune API système
nouvelle, donc la dernière image n'apporte rien. Ceux qu'elle écarterait sont les
distributions au long support, c'est-à-dire le public qui ne s'occupe pas de son
système.

Ce que l'épinglage coûte : une image finit par être retirée, et ce jour-là le job
échoue franchement. C'est préférable à un binaire qui cesse silencieusement de
démarrer. La même règle vaut déjà pour `macos-15-intel`, dernière image x86_64
des runners, dont la ligne porte sa date de péremption.

## Les assets dans le binaire

Embarqués par `go:embed` depuis `assets/`. L'archive publiée ne contient que
l'exécutable et ses mentions de licence : rien à installer à côté, et déplacer
le fichier ne casse rien.

## Les ressources et les contributions

Le décor, les objets et les sons sortent de générateurs, et sont versionnés.
**Une forme se corrige dans le script, jamais dans le PNG** : une retouche
manuelle serait écrasée à la prochaine génération sans que personne ne le voie.

C'est aussi ce qui les rend contribuables : un générateur est du code relisible
en pull request, là où un PNG ne l'est pas. Changer une forme se fait dans son
script, puis :

```
make ressources        # tous les générateurs, puis le contrôle
make ressources-verif  # régénère à côté et exige l'identique
```

`make ressources-verif` doit passer si rien n'a changé dans les scripts. Un écart
signale une retouche manuelle d'un PNG ou une version de Pillow différente.

Pour mettre au point une forme, garder la boucle courte : chaque générateur reste
autonome et accepte ses propres options.

```
python outils/decor_iso.py --theme parking --sortie assets/decors
python outils/decor_iso.py voiture camion --sortie assets/decors
python outils/objets.py --sortie assets/objets
python outils/interface.py --sortie assets/interface
python outils/ressources.py --controle        # sans rien régénérer
python outils/ressources.py --controle --pentes
```
