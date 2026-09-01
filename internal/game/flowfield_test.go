// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du champ de flux : la distance qui compte les coûts et non les cases,
// le contournement de ce qui coûte cher, l'angle de deux murs qu'on ne coupe
// pas, et la marche qui vérifie que toute case atteignable mène à la cible.

package game

import "testing"

// grilleDepuis bâtit une grille de coûts à partir d'un dessin.
//
// `.` est une case ordinaire, `#` un mur, `~` une flaque à trois pas. Le dessin
// vaut mieux qu'une suite d'appels à `Set` : une disposition pénible se relit
// d'un coup d'œil, et le jeu de cas en compte plusieurs.
func grilleDepuis(lignes ...string) *CostGrid {
	g := NewCostGrid(len(lignes[0]), len(lignes))
	for v, ligne := range lignes {
		for u, r := range ligne {
			switch r {
			case '#':
				g.Set(u, v, Blocked)
			case '~':
				g.Set(u, v, 3)
			}
		}
	}
	return g
}

// champSur prépare un champ reconstruit depuis une cellule.
func champSur(g *CostGrid, u, v int) *FlowField {
	f := NewFlowField(g)
	f.Rebuild(FromInt(u), FromInt(v))
	return f
}

// TestDistanceCompteLesCoutsEtNonLesCases éprouve ce qui distingue ce parcours
// d'un BFS ordinaire.
//
// Sans cela, le parcours pondéré serait une superstition : le champ
// contournerait au prix de deux cases ce qui ne coûte rien à traverser, et
// l'écart entre le chemin choisi et le chemin payé ne se verrait nulle part.
func TestDistanceCompteLesCoutsEtNonLesCases(t *testing.T) {
	f := champSur(grilleDepuis(".~.."), 0, 0)

	cas := []struct {
		u    int
		veut uint32
	}{
		{0, 0}, // la cible
		{1, 3}, // la flaque, à son prix
		{2, 4}, // derrière elle
		{3, 5},
	}
	for _, c := range cas {
		if d := f.Distance(c.u, 0); d != c.veut {
			t.Errorf("distance en (%d, 0) : %d, attendu %d", c.u, d, c.veut)
		}
	}
}

// TestLeChampContourneCeQuiCouteCher vérifie que le détour gratuit l'emporte sur
// la traversée coûteuse, ce qui est tout l'intérêt d'une passabilité par coût.
func TestLeChampContourneCeQuiCouteCher(t *testing.T) {
	// Deux chemins de la cible vers (3, 0) : tout droit par deux flaques, à sept,
	// ou par le bas en cinq pas ordinaires. Une seule flaque ne testerait rien —
	// les deux chemins coûteraient quatre, et le champ aurait raison de choisir
	// n'importe lequel.
	f := champSur(grilleDepuis(
		".~~.",
		"....",
	), 0, 0)

	if d := f.Distance(3, 0); d != 5 {
		t.Errorf("distance en (3, 0) : %d, attendu 5 par le détour", d)
	}
	// Et la direction depuis (3, 0) renvoie vers le bas, pas dans les flaques.
	if dir := f.Direction(3, 0); dir.Y <= 0 {
		t.Errorf("direction en (3, 0) : %v, attendu vers le détour", dir)
	}
}

// TestLaCibleNaPasDeDirection est le point où le vecteur nul du champ rencontre
// le mécanisme du vecteur dégénéré.
//
// Une entité posée exactement sur le joueur lit un vecteur nul, donc reçoit une
// orientation tirée de son identité — et deux entités superposées en reçoivent
// deux différentes, donc elles se séparent au tick suivant.
func TestLaCibleNaPasDeDirection(t *testing.T) {
	f := champSur(grilleDepuis("...", "...", "..."), 1, 1)

	if d := f.Distance(1, 1); d != 0 {
		t.Errorf("la cible est à distance %d d'elle-même", d)
	}
	if dir := f.Direction(1, 1); dir != (Vec{}) {
		t.Errorf("la cible porte la direction %v, attendu le vecteur nul", dir)
	}

	premiere := f.Direction(1, 1).Direction(0)
	seconde := f.Direction(1, 1).Direction(1)
	if premiere == seconde {
		t.Error("deux entités superposées sur la cible partent dans la même direction")
	}
	if premiere.Len() != One || seconde.Len() != One {
		t.Errorf("directions de longueurs %d et %d, attendu %d", premiere.Len(), seconde.Len(), One)
	}
}

// TestCelluleInatteignable vérifie ce que le champ répond d'un recoin muré.
//
// La réponse est la même que pour la cible, et c'est voulu : l'appelant n'a
// qu'un cas à traiter, celui du vecteur nul.
func TestCelluleInatteignable(t *testing.T) {
	// Le coin bas-droit est isolé par deux murs.
	f := champSur(grilleDepuis(
		"....",
		"....",
		"..##",
		"..#.",
	), 0, 0)

	if d := f.Distance(3, 3); d != Unreachable {
		t.Errorf("le recoin muré est à distance %d, attendu inatteignable", d)
	}
	if dir := f.Direction(3, 3); dir != (Vec{}) {
		t.Errorf("le recoin muré porte la direction %v, attendu le vecteur nul", dir)
	}
}

// TestCibleDansUnMurLaisseToutInatteignable vérifie le refus de rapprocher la
// cible d'une case passable.
//
// Le rapprochement ferait converger toute la horde vers un point où le joueur
// n'est pas, ce qui se diagnostique bien plus mal qu'une horde qui se disperse.
func TestCibleDansUnMurLaisseToutInatteignable(t *testing.T) {
	f := champSur(grilleDepuis(
		"...",
		".#.",
		"...",
	), 1, 1)

	if d := f.Distance(0, 0); d != Unreachable {
		t.Errorf("distance en (0, 0) : %d, attendu inatteignable", d)
	}
}

// TestLaDirectionNeCoupePasUnAngle éprouve la seule règle qui distingue les huit
// voisins de l'orientation des quatre de la propagation.
//
// En huit-connexité sans cette règle, un chemin passe entre deux murs
// perpendiculaires qui se touchent par un angle, et la horde traverse
// visuellement une arête. C'est le genre de défaut qu'on voit tout de suite et
// qu'on met des jours à relier au champ de flux.
//
// **Il ne fait que la moitié du travail, et l'autre est dans
// `TestLeGlissementNeCoupeAucunAngle`.** Ce qui est gardé ici est l'orientation
// proposée par une cellule ; rien n'oblige une entité à la suivre, et le joueur
// ne la lit même pas. Supprimer celui-là au motif que celui-ci existe laisserait
// le champ libre de désigner l'arête.
func TestLaDirectionNeCoupePasUnAngle(t *testing.T) {
	// Depuis (2, 2), le voisin diagonal (1, 1) est bien plus proche de la cible,
	// mais les deux murs qui le bordent ferment le passage.
	g := grilleDepuis(
		"....",
		"..#.",
		".#..",
		"....",
	)
	f := champSur(g, 0, 0)

	if d := f.Distance(1, 1); d >= f.Distance(2, 2) {
		t.Fatalf("le cas ne teste rien : (1, 1) est à %d et (2, 2) à %d",
			f.Distance(1, 1), f.Distance(2, 2))
	}

	// deltas[5] est la diagonale vers le haut-gauche.
	if dir := f.Direction(2, 2); dir == Heading(5) {
		t.Error("la direction traverse l'angle des deux murs")
	}
}

// TestLeChampNalloueRien garde le budget : le champ est reconstruit toutes les
// cinq ou six images, donc dans la boucle de mise à jour.
//
// **Il reconstruit depuis une cible fixe, et c'est ce qui le limite.** Les seaux
// sont réutilisés avec `[:0]`, si bien qu'une capacité acquise suffit tant que la
// distribution des distances ne change pas — ce qu'une cible immobile garantit,
// et ce qu'un joueur qui traverse le lieu défait. La capacité venant désormais du
// montage, la distinction ne décide plus de rien ici ; elle décidait avant, et
// c'est `TestLesSeauxNeGrandissentPasAuFilDesReconstructions` qui la garde.
//
// `TestLaBoucleNalloueRien` couvre le même budget à l'échelle du tick, et ne
// remplace pas celui-ci : il mesure un assemblage, si bien qu'un échec y dit
// qu'une allocation a eu lieu sans dire d'où. Celui-ci isole le champ, et c'est
// à quoi sert de le garder quand l'autre passe déjà.
func TestLeChampNalloueRien(t *testing.T) {
	g := grilleDepuis(
		"................",
		".#####..####....",
		".....#..#..#....",
		"..~~.#..#..#....",
		"..~~.#..####....",
		".....#..........",
		"..############..",
		"................",
	)
	f := NewFlowField(g)

	moyenne := testing.AllocsPerRun(100, func() {
		f.Rebuild(FromInt(15), FromInt(7))
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par reconstruction, attendu aucune", moyenne)
	}
}

// TestLesSeauxNeGrandissentPasAuFilDesReconstructions garde ce que la cible fixe
// ne peut pas garder : la capacité des seaux vient du montage, elle ne se
// constitue pas reconstruction après reconstruction.
//
// Son jumeau `TestLeChampNalloueRien` mesure le régime établi, et il y est aveugle
// par construction — `AllocsPerRun` préchauffe, donc il joue lui-même les tours
// pendant lesquels une capacité manquante se constituerait, puis mesure ceux où
// elle ne manque plus. Supprimer l'un des deux au motif que l'autre couvre les
// allocations du champ laisserait la moitié de la propriété sans garde.
//
// D'où les deux écarts avec lui : le champ est **neuf**, et la cible **change à
// chaque pas**. C'est cette seconde condition qui compte — les seaux étant
// réutilisés avec `[:0]`, une capacité acquise suffit tant que la distribution
// des distances ne change pas, et seule une cible qui se déplace la renouvelle.
// C'est ce que fait un joueur qui traverse le lieu.
//
// **Il lit les capacités plutôt qu'il ne compte les allocations.**
// `runtime.ReadMemStats` mesure tout le processus : les finaliseurs et les
// goroutines laissés par les tests précédents tombent dans la fenêtre, et le
// verdict dépend alors de ce qui entoure le test au lieu de ce qu'il garde. La
// capacité d'un seau, elle, est la propriété elle-même.
//
// **Le relevé part du montage et non de la première reconstruction**, faute de
// quoi le test se désarme : la première cible fait croître les seaux jusqu'à son
// pire cas, les quinze suivantes n'en demandent pas davantage, et la préallocation
// peut disparaître sans que rien ne bouge entre les deux mesures.
//
// La cible mobile reste nécessaire pour l'autre moitié : une croissance qu'une
// seule distribution de distances ne provoquerait pas.
func TestLesSeauxNeGrandissentPasAuFilDesReconstructions(t *testing.T) {
	g := grilleDepuis(
		"................",
		".#####..####....",
		".....#..#..#....",
		"..~~.#..#..#....",
		"..~~.#..####....",
		".....#..........",
		"..############..",
		"................",
	)
	f := NewFlowField(g)

	cibles := [][2]int{
		{15, 7}, {0, 0}, {8, 5}, {1, 7}, {14, 0}, {6, 0}, {0, 4}, {12, 7},
		{2, 0}, {15, 0}, {9, 7}, {0, 7}, {13, 5}, {4, 7}, {11, 0}, {7, 5},
	}

	auMontage := make([]int, len(f.seaux))
	for i, seau := range f.seaux {
		auMontage[i] = cap(seau)
	}

	for _, c := range cibles {
		f.Rebuild(FromInt(c[0]), FromInt(c[1]))
	}

	for i, seau := range f.seaux {
		if cap(seau) != auMontage[i] {
			t.Errorf("seau %d : capacité de %d au montage, %d après %d reconstructions",
				i, auMontage[i], cap(seau), len(cibles))
		}
	}
}

// TestToutCeQuiEstAtteignablePorteUneDirection est l'invariant que la boucle
// suppose : suivre les directions depuis n'importe quelle case mène à la cible.
//
// Il est vérifié en marchant, et non en relisant les vecteurs un à un : un
// champ dont une seule cellule pointerait à l'envers ferait tourner en rond,
// et aucune inspection cellule par cellule ne le dirait.
func TestToutCeQuiEstAtteignablePorteUneDirection(t *testing.T) {
	g := grilleDepuis(
		"................",
		".#####..####....",
		".....#..#..#....",
		"..~~.#..#..#....",
		"..~~.#..####....",
		".....#..........",
		"..############..",
		"................",
	)
	const cibleU, cibleV = 15, 7
	f := champSur(g, cibleU, cibleV)

	for v := range g.Height() {
		for u := range g.Width() {
			if f.Distance(u, v) == Unreachable {
				continue
			}
			pas, cu, cv := 0, u, v
			for (cu != cibleU || cv != cibleV) && pas < g.Width()*g.Height() {
				d := f.Direction(cu, cv)
				if d == (Vec{}) {
					t.Fatalf("(%d, %d) est atteignable et n'a pas de direction", cu, cv)
				}
				cu, cv = pasVers(cu, cv, d)
				pas++
			}
			if cu != cibleU || cv != cibleV {
				t.Fatalf("depuis (%d, %d), les directions ne mènent pas à la cible", u, v)
			}
		}
	}
}

// pasVers avance d'une case dans la direction donnée.
func pasVers(u, v int, d Vec) (int, int) {
	return u + signe(d.X), v + signe(d.Y)
}

// signe rend -1, 0 ou 1 selon le signe d'une longueur.
func signe(f Fixed) int {
	switch {
	case f > 0:
		return 1
	case f < 0:
		return -1
	}
	return 0
}
