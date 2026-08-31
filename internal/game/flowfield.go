// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package game

import "math"

// Unreachable est la distance d'une cellule que la cible n'atteint pas.
//
// Un recoin muré, l'intérieur d'un obstacle, ou toute la carte quand la cible
// elle-même est dans un mur.
const Unreachable uint32 = math.MaxUint32

// deltas est l'ordre des huit voisins, aligné sur la table des orientations :
// le voisin d'indice k se rejoint en suivant `Heading(k)`.
//
// Aligné, et non simplement cohérent : c'est ce qui permet à la direction d'une
// cellule d'être une orientation tabulée plutôt qu'un vecteur normalisé sur
// place, donc d'être exactement unitaire sans un seul calcul.
//
// **L'ordre est fixe et il compte.** Deux chemins de même longueur — l'un droit,
// l'autre en escalier — sont équivalents pour le champ, et c'est le parcours qui
// les départage. Le résultat est déterministe, donc l'invariant tient ; mais
// changer cet ordre change les trajectoires de toute la horde.
var deltas = [Headings][2]int{
	{1, 0}, {1, 1}, {0, 1}, {-1, 1},
	{-1, 0}, {-1, -1}, {0, -1}, {1, -1},
}

// FlowField est le chemin vers la cible, calculé une fois pour toute la horde.
//
// Chaque cellule porte sa distance à la cible et la direction du meilleur
// voisin. Un ennemi ne cherche rien : il lit la cellule sous ses pieds. C'est ce
// qui rend le contournement d'obstacles gratuit — un pilier, un rayonnage, un
// tourniquet se retrouvent dans le champ sans une ligne de code d'évitement.
type FlowField struct {
	grille *CostGrid
	dist   []uint32
	dir    []Vec

	// seaux est la file de Dial : un seau par distance modulo leur nombre.
	//
	// Un tri par seaux et non un tas, parce que les coûts sont trois ou quatre
	// valeurs entières : le parcours pondéré revient alors au temps linéaire
	// d'un parcours en largeur ordinaire, là où un Dijkstra général se paierait
	// pour une variété de coûts qui n'existe pas ici.
	seaux [][]int
	// nbSeaux double `len(seaux)` dans le type des distances, pour que le
	// modulo du parcours ne traverse aucune conversion.
	nbSeaux uint32
}

// NewFlowField prépare un champ pour une grille donnée.
//
// Le nombre de seaux se dérive du plus grand coût franchissable de la grille
// plutôt que d'une constante : une constante serait une seconde description du
// catalogue du manifeste, tenue à la main, et se périmerait au premier coût
// ajouté. Le calcul exclut `Blocked` **par égalité et non par seuil** — un seuil
// se mettrait à mentir sans bruit le jour où un coût légitime en approcherait
// la valeur.
func NewFlowField(g *CostGrid) *FlowField {
	cellules := g.Width() * g.Height()

	maxCout := Free
	for v := range g.Height() {
		for u := range g.Width() {
			if c := g.At(u, v); c != Blocked && c > maxCout {
				maxCout = c
			}
		}
	}

	// Autant de seaux que de distances simultanément vivantes, plus une : avec
	// des coûts d'au moins un pas, une cellule atteinte depuis la distance d
	// entre dans un seau strictement postérieur, donc jamais dans celui qu'on
	// est en train de vider.
	nbSeaux := uint32(maxCout) + 1

	return &FlowField{
		grille:  g,
		dist:    make([]uint32, cellules),
		dir:     make([]Vec, cellules),
		seaux:   make([][]int, nbSeaux),
		nbSeaux: nbSeaux,
	}
}

// Cell rend la cellule qui porte une position du monde.
//
// C'est celle du point d'appui, qui situe déjà l'entité partout ailleurs, tri en
// profondeur compris.
func (f *FlowField) Cell(x, y Fixed) (int, int) { return x.Floor(), y.Floor() }

// Distance rend le coût cumulé du chemin d'une cellule vers la cible, ou
// `Unreachable`.
//
// Ce n'est pas un nombre de cases : une flaque compte pour ce que le manifeste
// lui donne, et c'est ce qui fait contourner ce qui ralentit.
func (f *FlowField) Distance(u, v int) uint32 {
	if !f.grille.InBounds(u, v) {
		return Unreachable
	}
	return f.dist[v*f.grille.Width()+u]
}

// Direction rend l'orientation à suivre depuis une cellule, ou le vecteur nul.
//
// Le vecteur nul est rendu dans trois cas qui n'en font qu'un pour l'appelant :
// la cellule **est** la cible, la cible ne l'atteint pas, ou elle est hors de la
// grille. Aucun ne mérite un traitement propre ici, parce que le mécanisme du
// vecteur dégénéré y répond déjà : `Vec.Direction(index)` donne à l'entité une
// orientation tirée de son identité, donc elle cherche au lieu de rester figée,
// et la projection sur la passabilité l'empêche d'entrer dans un mur. Une
// créature poussée dans un recoin muré par la séparation en ressort ; deux
// créatures posées sur le joueur s'écartent l'une de l'autre.
//
// La sortie n'est pas garantie pour autant : l'orientation tirée de l'identité a
// une chance sur huit d'être la bonne, et elle est stable tant que l'index de
// l'entité ne bouge pas. Une créature enfermée peut donc pousser contre le même
// mur longtemps. C'est le comportement choisi et non un oubli — lui donner une
// recherche de sortie reviendrait à faire délibérer une entité, ce que la
// conception refuse.
func (f *FlowField) Direction(u, v int) Vec {
	if !f.grille.InBounds(u, v) {
		return Vec{}
	}
	return f.dir[v*f.grille.Width()+u]
}

// Rebuild recalcule le champ entier depuis une position.
//
// Les directions sont calculées ici, une fois par rafraîchissement, et non par
// entité et par image : c'est la raison d'être de la structure.
//
// Une cible hors grille ou dans un mur laisse le champ entièrement
// inatteignable, plutôt que d'être ramenée sur la case passable la plus proche.
// Le rapprochement ferait converger toute la horde vers un point où le joueur
// n'est pas, ce qui se diagnostique bien plus mal qu'une horde qui se disperse.
func (f *FlowField) Rebuild(x, y Fixed) {
	for i := range f.dist {
		f.dist[i] = Unreachable
		f.dir[i] = Vec{}
	}
	for i := range f.seaux {
		f.seaux[i] = f.seaux[i][:0]
	}

	u, v := f.Cell(x, y)
	if !f.grille.InBounds(u, v) || !f.grille.Passable(u, v) {
		return
	}

	f.propager(u, v)
	f.orienter()
}

// propager remplit les distances par un parcours à seaux depuis la cible.
//
// Quatre voisins et non huit : à coûts entiers, une diagonale vaudrait un pas
// comme un axe, si bien qu'un détour en diagonale deviendrait gratuit. Mais le
// vrai motif est ailleurs — en huit-connexité, un chemin passe entre deux murs
// perpendiculaires qui se touchent par un angle, et la horde traverse
// visuellement une arête. C'est le genre de défaut qu'on voit tout de suite et
// qu'on met des jours à relier au champ de flux.
func (f *FlowField) propager(u, v int) {
	largeur := f.grille.Width()
	f.dist[v*largeur+u] = 0
	f.seaux[0] = append(f.seaux[0], v*largeur+u)

	restants := 1
	for d := uint32(0); restants > 0; d++ {
		seau := d % f.nbSeaux
		for len(f.seaux[seau]) > 0 {
			fin := len(f.seaux[seau]) - 1
			cellule := f.seaux[seau][fin]
			f.seaux[seau] = f.seaux[seau][:fin]
			restants--

			// Une cellule améliorée après avoir été empilée laisse derrière elle
			// une entrée qui n'est plus la bonne.
			//
			// La sauter n'est pas ce qui rend le résultat juste : sans cette
			// ligne, la cellule serait retraitée avec la distance de son seau,
			// donc pessimiste, et la comparaison des candidats rejetterait tout
			// ce qu'elle propage. Elle évite le retraitement, rien de plus — la
			// retirer ne fait donc échouer aucun test, et c'est normal.
			if f.dist[cellule] != d {
				continue
			}

			cu, cv := cellule%largeur, cellule/largeur
			// Les indices pairs de la table sont les quatre orthogonaux.
			for k := 0; k < Headings; k += 2 {
				vu, vv := cu+deltas[k][0], cv+deltas[k][1]
				if !f.grille.InBounds(vu, vv) {
					continue
				}
				cout := f.grille.At(vu, vv)
				if cout == Blocked {
					continue
				}
				candidate := d + uint32(cout)
				voisin := vv*largeur + vu
				if candidate >= f.dist[voisin] {
					continue
				}
				f.dist[voisin] = candidate
				suivant := candidate % f.nbSeaux
				f.seaux[suivant] = append(f.seaux[suivant], voisin)
				restants++
			}
		}
	}
}

// orienter donne à chaque cellule la direction de son meilleur voisin.
//
// Les huit sont candidats, alors que la propagation n'en a suivi que quatre : la
// distance reste dans l'unité du manifeste, et la direction reste lisse. Une
// diagonale n'est retenue que si les deux orthogonaux qui la bordent sont
// franchissables — sans quoi la horde couperait l'angle des murs que la
// propagation lui a fait contourner.
func (f *FlowField) orienter() {
	largeur, hauteur := f.grille.Width(), f.grille.Height()
	for v := range hauteur {
		for u := range largeur {
			cellule := v*largeur + u
			meilleure := f.dist[cellule]
			if meilleure == Unreachable || meilleure == 0 {
				continue
			}

			choix := -1
			for k := range Headings {
				vu, vv := u+deltas[k][0], v+deltas[k][1]
				if !f.grille.InBounds(vu, vv) || !f.diagonaleOuverte(u, v, k) {
					continue
				}
				if d := f.dist[vv*largeur+vu]; d < meilleure {
					meilleure = d
					choix = k
				}
			}
			if choix >= 0 {
				f.dir[cellule] = Heading(choix)
			}
		}
	}
}

// diagonaleOuverte dit si le voisin k se rejoint sans traverser un angle.
//
// Vrai d'office pour les quatre orthogonaux, dont les indices sont pairs.
func (f *FlowField) diagonaleOuverte(u, v, k int) bool {
	if k%2 == 0 {
		return true
	}
	return f.grille.Passable(u+deltas[k][0], v) && f.grille.Passable(u, v+deltas[k][1])
}
