// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Tilemap est ce que l'assemblage produit : une forme de décor par case. La
// grille de coûts en descend, et rien d'autre ne dit ce qu'une case porte.

package level

import "github.com/sprimault/cohue/internal/game"

// Tilemap dit quelle forme du décor occupe chaque case d'un lieu assemblé.
//
// **C'est elle que la cuisson produit, et la grille de coûts en descend.** La
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

// set pose une forme sur une case, en la nommant si c'est la première fois.
func (t *Tilemap) set(u, v int, forme string) {
	if u < 0 || v < 0 || u >= t.largeur || v >= t.hauteur {
		return
	}
	i, connue := t.index[forme]
	if !connue {
		i = len(t.formes)
		t.formes = append(t.formes, forme)
		t.index[forme] = i
	}
	t.cases[v*t.largeur+u] = i
}

// couts dérive la grille que la simulation lit.
//
// Une forme absente du catalogue vaut un mur, et le chargement la refuse par
// ailleurs : l'assemblage précède la validation, donc il rencontre encore ce
// que celle-ci va rejeter et n'a pas à s'en émouvoir. Le mur est le repli qui ne
// laisse passer personne, ce qu'on préfère à un sol qui s'ouvrirait sur une
// faute de frappe.
//
// Une case qu'aucune pièce ne pose garde en revanche le coût d'une grille
// neuve, celui d'un sol ordinaire. C'est ce que le contrôle de couverture
// décrit et refuse : le trou se traverse et ne se dessine pas, et lui donner un
// mur ici le rendrait invisible au moment même où l'on veut qu'il se voie.
func (t *Tilemap) couts(catalogue map[string]game.Cost) *game.CostGrid {
	grille := game.NewCostGrid(t.largeur, t.hauteur)
	for v := range t.hauteur {
		for u := range t.largeur {
			i := t.cases[v*t.largeur+u]
			if i < 0 {
				continue
			}
			cout, connu := catalogue[t.formes[i]]
			if !connu {
				cout = game.Blocked
			}
			grille.Set(u, v, cout)
		}
	}
	return grille
}
