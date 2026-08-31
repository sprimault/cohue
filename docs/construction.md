# Construction

## En local

```
make build      # binaire dans .tmp/
make run
make apercus    # les vues du rendu, dans .tmp/apercus/
make test
make race
make fmt
make lint
make vulncheck
make sec
make entetes    # la mention de licence, sur ce que git publie
make cover
make tools      # installe golangci-lint, govulncheck, gosec
```

`make apercus` écrit des vues du rendu dans `.tmp/apercus/`. C'est le seul moyen
de relire `internal/render`, qui n'a pas de test et n'en aura pas — importer
Ebitengine initialise GLFW, qui panique sans écran. La planche pilote le rendu du
jeu et non une scène montée à côté, et elle est déterministe : c'est ce qui permet
de comparer une image d'avant et d'après un changement. Elle exige un écran, donc
elle ne tourne pas en intégration continue.

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

Les personnages sortent de `outils/figurines.py`, sur le même principe : des
volumes isométriques composés, avec un gabarit par famille — bipède, quadrupède,
rampant, bulbe, colosse, gonflé — et des variantes de teinte par profil.

Régénérer le décor après une modification du script :

```
python outils/decor_iso.py --sortie assets/decors
```

La sortie est déterministe. Un diff sur `assets/` après régénération sans
changement de script signale un problème — version de Pillow, ou retouche
manuelle d'un PNG.

## Les cibles publiées

Ebitengine ne se compile en croisé que vers Windows et WebAssembly ; Linux et
macOS exigent du natif. D'où la matrice multi-runners de `release.yml`.

`js/wasm` se compile mais ne se publie pas : l'archive serait inutilisable seule.
La cible reste dans la matrice parce qu'elle est ce qui empêche une dépendance
d'introduire du cgo sans qu'on le voie.

## Les assets dans le binaire

Embarqués par `go:embed` depuis `assets/`. L'archive publiée ne contient que
l'exécutable et ses mentions de licence : rien à installer à côté, et déplacer
le fichier ne casse rien.

## Les ressources et les contributions

Décor et personnages sortent des mêmes primitives isométriques, et les deux sont
versionnés. **Une forme se corrige dans le script, jamais dans le PNG** : une
retouche manuelle serait écrasée à la prochaine génération sans que personne ne
le voie.

C'est aussi ce qui rend les créatures contribuables : `outils/figurines.py` est
du code relisible en pull request, là où un PNG ne l'est pas. Changer les
proportions d'un profil, ajouter un gabarit ou une variante de teinte se fait
dans ce fichier, puis :

```
make ressources        # les quatre générateurs, puis le contrôle
make ressources-verif  # régénère à côté et exige l'identique
```

`make ressources-verif` doit passer si rien n'a changé dans les scripts. Un écart
signale une retouche manuelle d'un PNG ou une version de Pillow différente.

Pour mettre au point une forme, garder la boucle courte : chaque générateur reste
autonome et accepte ses propres options.

```
python outils/decor_iso.py --theme parking --sortie assets/decors
python outils/decor_iso.py voiture camion --sortie assets/decors
python outils/figurines.py --apercu --sortie .tmp/apercus
python outils/objets.py --sortie assets/objets
python outils/ressources.py --controle        # sans rien régénérer
python outils/ressources.py --controle --pentes
```
