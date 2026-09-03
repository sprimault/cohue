# Cibles en anglais, commentaires en français : convention des autres projets.

BINAIRE   ?= cohue
VERSION   ?= dev
LDFLAGS    = -s -w -X main.version=$(VERSION)
SORTIE    ?= .tmp

# Les distributions Linux et macOS ne posent que `python3`, l'installateur
# Windows que `python`. Une variable plutôt qu'un nom en dur, surchargée dans
# makefile.local : sur un poste Windows, `python3` est un raccourci vers le
# Microsoft Store dont l'échec ne ressemble pas à une commande absente.
PYTHON    ?= python3

# makefile.local porte ce qui est propre au poste — GOTMPDIR, GOCACHE — et n'est
# pas versionné. Inclus ici et non en tête : une affectation immédiate qui y
# référencerait une variable définie plus haut trouverait une chaîne vide.
#
# Passer par le Makefile plutôt que d'appeler go à la main : une commande tapée
# directement perd ces variables, et l'échec est intermittent.
-include makefile.local

.PHONY: build run test race fmt lint vulncheck sec notices cover binary binaries clean tools ressources ressources-verif decors figurines objets sons interface controle entetes sommaire apercus

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(SORTIE)/$(BINAIRE) ./cmd/cohue

run:
	go run ./cmd/cohue

# internal/render n'a pas de test et n'en aura pas : le rendu se juge à l'œil, et
# une planche que rien ne fabrique ne relit rien. Elle ouvre une fenêtre le temps
# d'obtenir un contexte graphique, donc elle ne tourne pas en intégration
# continue — c'est une planche, pas un contrôle.
apercus:
	go run ./cmd/preview

# Aucun test ne doit ouvrir de fenêtre : les runners sont sans écran, et un test
# qui exigerait xvfb n'a rien à faire dans la suite par défaut.
#
# PKG et RUN restreignent la portée sans sortir du Makefile : une commande go
# tapée directement perd les réglages de makefile.local, et l'oubli ne se voit
# pas dans la sortie.
PKG ?= ./...
RUN ?=

# Le compte des mécaniques inertes est affiché avec la suite, jamais séparément :
# un test qui se saute passe inaperçu là où un test rouge fait mal, et six mois
# plus tard il reste des skips que personne ne regarde. C'est le pendant du
# compte de stubs de ROADMAP.md — il descend tout seul et ne ment pas.
#
# Le compte vient du test, jamais d'un motif recopié ici. La première version
# posait un grep dont le motif divergeait d'un espace de celui du fichier : elle
# affichait zéro, ce qui est pire qu'aucun compteur. Aucun filtre non plus sur la
# sortie — il dépendrait du texte du message et casserait à la première
# reformulation, en silence.
test:
	go test $(if $(RUN),-run '$(RUN)') $(PKG)

race:
	CGO_ENABLED=1 go test -race ./...

# Doublon assumé avec les formatters de golangci-lint : gofmt porte sur tout
# l'arbre, sans exclusion ni configuration, et reste vrai le jour où quelqu'un
# touche à la liste des formatters. Coût nul, et surtout une cible à lancer
# avant de pousser, ce qui manquait.
fmt:
	@restants=$$(gofmt -l .); \
	if [ -n "$$restants" ]; then \
	  echo "fichiers non formatés :"; \
	  echo "$$restants"; \
	  exit 1; \
	fi

lint:
	golangci-lint run

vulncheck:
	govulncheck ./...

# .tmp porte les programmes jetables, qui ne sont pas versionnés : gosec les
# analyse en local et pas en intégration continue, où le dossier n'existe pas.
# Sans exclusion, la cible échoue ici et passe là-bas, ce qui finit par la faire
# ignorer — et par masquer un vrai signalement.
sec:
	gosec -exclude-dir=.tmp ./...

# Les licences des dépendances liées au binaire. Le contrôle confronte le fichier
# à ce que `go list -deps` rapporte pour les cibles publiées, dans les deux sens
# et versions comprises : une notice manquante est un manquement légal, une
# notice de trop désigne un composant que le binaire n'embarque pas.
notices:
	$(PYTHON) outils/notices.py

# Décor et personnages sortent des mêmes primitives isométriques et sont
# versionnés : une forme se corrige dans le script, jamais dans le PNG. Les
# deux cibles doivent laisser le dépôt propre si rien n'a changé.
ressources:
	$(PYTHON) outils/ressources.py

# Cibles unitaires : mettre au point une forme sans tout régénérer. Chaque
# générateur reste autonome et accepte ses propres options — un thème, une forme
# précise, un aperçu. Voir son --help.
decors:
	$(PYTHON) outils/decor_iso.py --sortie assets/decors

figurines:
	$(PYTHON) outils/figurines.py --sortie assets/personnages
	$(PYTHON) outils/figurines.py --apercu

objets:
	$(PYTHON) outils/objets.py --sortie assets/objets

sons:
	$(PYTHON) outils/sons.py --sortie assets/sons

interface:
	$(PYTHON) outils/interface.py --sortie assets/interface

controle:
	$(PYTHON) outils/ressources.py --controle

# La règle et la liste close des dispensés sont dans docs/go.md. Le périmètre
# vient de git ls-files : ce qui n'est pas publié n'a pas à porter de mention.
entetes:
	$(PYTHON) outils/entetes.py

# Le sommaire de docs/go.md indexe par la question, pas par les titres — donc il
# n'est pas une copie de la structure. Ses ancres, elles, en sont une : une ancre
# morte ne produit aucune erreur en Markdown, elle ne fait rien, et un sommaire
# que rien ne vérifie égare au lieu de guider.
sommaire:
	$(PYTHON) outils/sommaire.py

# Ce que passe l'intégration continue : régénère à côté et exige l'identique.
# Un écart signale une retouche manuelle d'un PNG, ce qui serait écrasé au
# prochain lancement sans que personne ne le voie.
ressources-verif:
	$(PYTHON) outils/ressources.py --verifier

cover:
	go test -coverprofile=$(SORTIE)/couverture.out ./...
	go tool cover -html=$(SORTIE)/couverture.out

# Une seule cible paramétrée, pas cinq recettes : la matrice complète vit dans
# le workflow, qui est le seul endroit où elle peut être exécutée. En local on
# ne produit que ce qui se croise sans compilateur C — Windows et wasm.
binary:
	CGO_ENABLED=$(CGO) GOOS=$(OS) GOARCH=$(ARCH) go build -trimpath \
	  -ldflags "$(LDFLAGS)" -o dist/$(BINAIRE)_$(OS)_$(ARCH)$(EXT) ./cmd/cohue

binaries:
	$(MAKE) binary OS=windows ARCH=amd64 CGO=0 EXT=.exe
	$(MAKE) binary OS=js ARCH=wasm CGO=0
	@echo "Les cibles linux et darwin exigent une compilation native : voir release.yml"

clean:
	rm -rf $(SORTIE) dist

# Les versions sont épinglées ici et nulle part ailleurs : le workflow appelle
# make tools plutôt que de réécrire ses go install, et l'action qui lance le lint
# lit GOLANGCI_VERSION par print-%, faute de quoi les deux définitions divergent
# sans que rien ne le signale. Avec @latest des deux côtés, le poste prend du
# retard dès qu'une version sort : le contrôle passe en local et échoue en
# intégration continue, sur du code que personne n'a touché.
#
# Ce sont les règles de ces deux outils qui bougent, et c'est ce qui les rend
# épinglables. govulncheck est le cas contraire, plus bas.
GOLANGCI_VERSION ?= v2.13.1
GOSEC_VERSION    ?= v2.29.0

# print-<VARIABLE> écrit la valeur d'une variable et rien d'autre, pour que
# l'intégration continue lise l'épinglage plutôt que de le recopier. Sans elle,
# ci.yml porterait une seconde version de golangci-lint, que seule la vigilance
# tiendrait d'accord avec celle-ci.
print-%:
	@echo $($*)

tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	# @latest délibérément : govulncheck ne signale pas des règles mais des
	# avis, et son intérêt est de connaître les derniers. L'épingler figerait
	# ce qu'il sait lire des avis publiés depuis.
	go install golang.org/x/vuln/cmd/govulncheck@latest
