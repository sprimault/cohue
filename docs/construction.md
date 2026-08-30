# Construction

## En local

```
make build      # binaire dans .tmp/
make run
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

`makefile.local` porte ce qui est propre au poste — `GOTMPDIR`, `GOCACHE` — et
n'est pas versionné. Passer par le Makefile plutôt que d'appeler `go` à la main :
une commande tapée directement perd ces variables, et l'échec est intermittent.

## Les ressources

```
python -m pip install -r outils/requirements.txt
```

Les versions y sont épinglées, et le fichier dit pourquoi : les images sont
comparées au bit près, or c'est la compression de Pillow qui fixe ces octets à
dessin identique. Une version différente fait échouer `make ressources-verif`
sur la totalité du catalogue sans qu'un seul pixel ait bougé.

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
