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
// Les seaux sont réutilisés avec `[:0]` et leur capacité se stabilise au premier
// parcours, que `AllocsPerRun` joue en préchauffage.
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
