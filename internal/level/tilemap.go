// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Tilemap est ce que l'assemblage produit : une forme de décor par case. La
// grille de coûts en descend, et rien d'autre ne dit ce qu'une case porte.

package level

import "github.com/sprimault/cohue/internal/game"

// Tilemap dit quelle forme du décor occupe chaque case d'un lieu assemblé.
//
// **C'est elle que l'assemblage produit, et la grille de coûts en descend.** La
// conception énonce l'ordre — assemblage en une seule tilemap, puis dérivation
// de la passabilité — et il vaut mieux que l'inverse : deux tables construites
// côte à côte à partir des mêmes pièces seraient deux descriptions, et le jour
// où elles divergeraient le rendu montrerait un sol là où la simulation lit un
// mur. Dérivée, la grille ne peut pas contredire le dessin.
//
// Les cases portent un index et non un nom : ce qui la lit soixante fois par
// seconde résout une image, et une table par nom lui coûterait une empreinte de
// hachage par case visible.
type Tilemap struct {
	// formes sont les noms employés par le lieu, sans doublon.
	//
	// Elles ne sont pas triées, et leur ordre est celui de la première
	// rencontre : rien ne le lit, puisque les cases citent des index et que
	// l'appelant ne résout que les noms qu'on lui donne. Un tri n'apporterait
	// que le sentiment d'une stabilité dont personne n'a besoin.
	formes []string
	// sol est la forme peinte sous une case dont la forme ne remplit pas son
	// losange, vide quand le thème n'en déclare pas. Elle ne figure dans aucune
	// case : ce qu'elle comble n'est pas ce qu'une case porte.
	sol string
	// couches porte ce qui se pose sur le terrain, de la plus basse à la plus
	// haute : moins un là où une couche ne met rien.
	//
	// **Le terrain se remplit par couches, et une case dit donc plusieurs
	// choses.** Les mêler en une seule faisait disparaître ce qui était
	// dessous : un bus posé sur une chaussée la remplaçait, et le comblement,
	// ne sachant plus ce que la case était, peignait le sol du thème — un carré
	// de béton clair sous chaque véhicule garé. Une flaque sur cette même
	// chaussée aurait fait exactement pareil.
	//
	// Leur nombre n'est pas borné parce que rien ne le demande : un revêtement,
	// ce qui le marque, ce qui s'y tient font trois, et le jour où il en faudra
	// une quatrième elle ne coûtera qu'une ligne de plus au fichier. Les index
	// sont ceux de la même table que `cases` : ce sont les mêmes tuiles, et
	// deux tables donneraient deux résolutions pour un seul dessin.
	couches [][]int
	// index rend la place d'un nom déjà rencontré, le temps de l'assemblage.
	index            map[string]int
	largeur, hauteur int
	// cases porte l'index de la forme de chaque case, rangée par rangée.
	cases []int
}

// newTilemap rend une carte de la taille voulue, dont aucune case n'est encore
// posée.
//
// Les cases partent à moins un plutôt qu'à zéro : zéro serait la première forme
// rencontrée, si bien qu'une case qu'aucune pièce ne pose se lirait comme une
// case ordinaire. Le contrôle de couverture refuse ce cas, mais il tourne après
// l'assemblage — et une valeur zéro légitime ne saurait plus le montrer.
func newTilemap(largeur, hauteur int) *Tilemap {
	cases := make([]int, largeur*hauteur)
	for i := range cases {
		cases[i] = -1
	}
	return &Tilemap{index: map[string]int{}, largeur: largeur, hauteur: hauteur, cases: cases}
}

// vide rend une couche dont aucune case ne porte rien.
func vide(n int) []int {
	couche := make([]int, n)
	for i := range couche {
		couche[i] = -1
	}
	return couche
}

// Shapes rend les noms des formes employées, dans l'ordre où les index les
// désignent.
func (t *Tilemap) Shapes() []string { return t.formes }

// Ground rend la forme à peindre sous une case que la sienne ne remplit pas, ou
// une chaîne vide quand le thème n'en déclare pas.
//
// Vide n'est pas un cas à traiter par prudence : le chargement exige un sol dès
// qu'une palette emploie une forme non couvrante, si bien qu'une carte sans sol
// n'a rien à combler.
func (t *Tilemap) Ground() string { return t.sol }

// Width rend la largeur en tuiles.
func (t *Tilemap) Width() int { return t.largeur }

// Height rend la hauteur en tuiles.
func (t *Tilemap) Height() int { return t.hauteur }

// At rend l'index de la forme d'une case, et moins un hors de la carte.
//
// Hors bornes plutôt qu'une panique, pour la raison qui vaut déjà à `CostGrid`
// de rendre un mur : ce qui la balaie parcourt une fenêtre rectangulaire dont
// les bords tombent au-delà du losange du lieu, et tester l'appartenance avant
// chaque lecture alourdirait le seul balayage qui ait lieu à chaque image.
func (t *Tilemap) At(u, v int) int {
	if u < 0 || v < 0 || u >= t.largeur || v >= t.hauteur {
		return -1
	}
	return t.cases[v*t.largeur+u]
}

// set pose une forme sur le terrain d'une case, en la nommant si c'est la
// première fois.
func (t *Tilemap) set(u, v int, forme string) {
	if u < 0 || v < 0 || u >= t.largeur || v >= t.hauteur {
		return
	}
	t.cases[v*t.largeur+u] = t.nommer(forme)
}

// poser met une forme sur une case, à la couche donnée, en ouvrant les couches
// manquantes.
//
// Les couches s'ouvrent à la demande plutôt qu'au dimensionnement : un lieu
// dont une seule pièce en emploie deux n'a pas à payer la seconde sur toute son
// étendue, et rien ne dit leur nombre avant d'avoir lu les pièces.
func (t *Tilemap) poser(u, v, couche int, forme string) {
	if u < 0 || v < 0 || u >= t.largeur || v >= t.hauteur {
		return
	}
	for len(t.couches) <= couche {
		t.couches = append(t.couches, vide(t.largeur*t.hauteur))
	}
	t.couches[couche][v*t.largeur+u] = t.nommer(forme)
}

// Layers rend le nombre de couches posées sur le terrain.
func (t *Tilemap) Layers() int { return len(t.couches) }

// LayerAt rend l'index de ce qu'une couche met sur une case, et moins un quand
// elle n'y met rien.
func (t *Tilemap) LayerAt(couche, u, v int) int {
	if couche < 0 || couche >= len(t.couches) ||
		u < 0 || v < 0 || u >= t.largeur || v >= t.hauteur {
		return -1
	}
	return t.couches[couche][v*t.largeur+u]
}

// nommer rend l'index d'une forme, en l'ajoutant si c'est la première fois.
//
// Sols et formes partagent la même table : ce sont les mêmes tuiles, et deux
// tables donneraient deux index pour un seul dessin — donc deux résolutions
// pour la même image au montage du rendu.
func (t *Tilemap) nommer(forme string) int {
	i, connue := t.index[forme]
	if !connue {
		i = len(t.formes)
		t.formes = append(t.formes, forme)
		t.index[forme] = i
	}
	return i
}

// couts dérive la grille que la simulation lit.
//
// **Une forme peint son bloc et non sa seule case.** Une gondole de deux tuiles
// ferme les deux qu'elle couvre, sans quoi la moitié de son dessin se
// traverserait — le défaut le plus coûteux qui soit, puisque la scène reste
// plausible et que seul le joueur qui marche dans le décor l'apprend. Le compte
// vient de `Shape.Block`, que le rendu appelle aussi pour centrer le dessin :
// une seule règle, donc aucun écart possible entre ce qui se voit et ce qui
// arrête.
//
// **Le plus cher l'emporte, et c'est ce qui retire tout effet à l'ordre des
// poses.** Deux blocs qui se recouvrent — un bus le long d'une façade, une
// enseigne au-dessus d'un mur — écriraient sinon selon l'ordre où les pièces
// ont été lues, que rien n'annonce et dont aucun auteur n'aurait idée. Le
// maximum est commutatif, si bien que la grille ne dépend plus que de ce qui
// est posé ; et `Blocked` valant le plus grand coût possible, un mur gagne
// contre tout sans que cette fonction ait à le savoir.
//
// Une forme absente du catalogue vaut un mur d'une case, et le chargement la
// refuse par ailleurs : l'assemblage précède la validation, donc il rencontre
// encore ce que celle-ci va rejeter et n'a pas à s'en émouvoir. Le mur est le
// repli qui ne laisse passer personne, ce qu'on préfère à un sol qui s'ouvrirait
// sur une faute de frappe.
//
// Une case qu'aucune pièce ne pose garde en revanche le coût d'une grille
// neuve, celui d'un sol ordinaire. C'est ce que le contrôle de couverture
// décrit et refuse : le trou se traverse et ne se dessine pas, et lui donner un
// mur ici le rendrait invisible au moment même où l'on veut qu'il se voie.
//
// Ce qu'un bloc pousse hors de la grille est perdu sans conséquence : au-delà
// des bords, `CostGrid.At` rend déjà un mur, si bien qu'il n'y a rien à y
// fermer.
func (t *Tilemap) couts(catalogue map[string]Footing) *game.CostGrid {
	grille := game.NewCostGrid(t.largeur, t.hauteur)
	for v := range t.hauteur {
		for u := range t.largeur {
			t.peindre(grille, catalogue, u, v, t.cases[v*t.largeur+u])
			for couche := range t.couches {
				t.peindre(grille, catalogue, u, v, t.couches[couche][v*t.largeur+u])
			}
		}
	}
	return grille
}

// peindre écrit le coût d'une forme sur le bloc qu'elle ferme, sans jamais le
// faire baisser.
//
// **Les couches passent par le même chemin que le terrain**, et l'arbitrage par
// le maximum leur suffit : une flaque posée sur un sol libre le renchérit, un
// véhicule sur une chaussée la ferme, et rien de ce qui est dessous ne rouvre
// ce qui est dessus.
func (t *Tilemap) peindre(grille *game.CostGrid, catalogue map[string]Footing, u, v, i int) {
	if i < 0 {
		return
	}
	assise, connue := catalogue[t.formes[i]]
	if !connue {
		assise = Footing{Cost: game.Blocked, Block: [2]int{1, 1}}
	}
	for dv := range assise.Block[1] {
		for du := range assise.Block[0] {
			if assise.Cost > grille.At(u+du, v+dv) {
				grille.Set(u+du, v+dv, assise.Cost)
			}
		}
	}
}
