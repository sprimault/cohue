// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La grille de densité : le comptage des créatures par cellule, et la pente qui
// s'en déduit. Elle répond à une question factuelle — combien, et dans quel sens
// cela décroît ; la force que cette pente exerce se décide dans la boucle.

package game

// DensityGrid compte les entités par cellule, pour que la horde ne s'empile pas
// sur un pixel.
//
// C'est l'alternative bon marché au voisinage : chaque entité incrémente sa
// cellule, et la pente du comptage sert de répulsion. Deux passes sur les
// entités, aucune requête de voisinage, et le coût ne dépend pas du nombre de
// voisins d'une créature — c'est ce qui tient à deux cents monstres dans un
// couloir.
//
// **Seuls les ennemis y entrent, jamais le joueur.** Si le joueur comptait, la
// horde s'écarterait de lui : elle éviterait ce qu'elle est censée poursuivre,
// en contradiction directe avec le champ de flux. Le cas du Vigile, dont le
// corps arrête le joueur, n'est pas une exception à cette règle mais un autre
// mécanisme — une collision, pas une répulsion douce — et les confondre
// donnerait envie d'ajouter le joueur ici pour uniformiser.
//
// La grille ne dit rien de la force que cette pente exerce : elle répond à une
// question factuelle, combien d'entités par cellule et dans quel sens cela
// décroît. Convertir une pente en poussée est une décision de jeu, elle vit dans
// la boucle avec les autres réglages qu'on cherchera en même temps.
type DensityGrid struct {
	largeur, hauteur int
	comptes          []uint16
}

// NewDensityGrid prépare une grille aux dimensions de la carte.
//
// Une cellule par tuile : c'est la maille du champ de flux, et une entité y lit
// sa densité au même endroit qu'elle lit sa direction.
func NewDensityGrid(largeur, hauteur int) *DensityGrid {
	return &DensityGrid{
		largeur: largeur,
		hauteur: hauteur,
		comptes: make([]uint16, largeur*hauteur),
	}
}

// Clear remet tous les comptes à zéro, au début d'un tick.
//
// Le coût est celui de la carte et non celui de la horde, ce qui paraît un
// mauvais échange et n'en est pas un : `clear` sur une tranche compile en un
// effacement de bloc, et la carte entière tient en quelques dizaines de
// kilo-octets. Suivre les cellules touchées pour n'effacer que celles-là
// coûterait une liste à tenir, donc un état de plus à garder juste.
func (d *DensityGrid) Clear() { clear(d.comptes) }

// Add compte une entité dans une cellule.
//
// La cellule, et non la position : la conversion se fait une fois par entité
// dans la boucle, qui la donne aussi au champ de flux. La refaire ici la
// dédoublerait, et deux conventions de correspondance finiraient par diverger.
func (d *DensityGrid) Add(u, v int) {
	if !d.inBounds(u, v) {
		return
	}
	d.comptes[v*d.largeur+u]++
}

// At rend le nombre d'entités comptées dans une cellule.
func (d *DensityGrid) At(u, v int) int {
	if !d.inBounds(u, v) {
		return 0
	}
	return int(d.comptes[v*d.largeur+u])
}

// Gradient rend la pente du comptage, orientée vers la foule.
//
// L'appelant la **soustrait** de l'attirance du champ de flux : c'est ce qui
// écarte une créature de ses voisines tout en la laissant descendre vers le
// joueur.
//
// L'échelle est explicite et vaut une tuile par entité d'écart. Ce n'est pas un
// réglage d'équilibrage déguisé mais l'unité la plus simple à relire : le poids
// de séparation du profil, qui multiplie ce vecteur, reste un rapport, et le
// facteur qui donne sa force à l'ensemble vit dans la boucle.
//
// Hors des bords, l'extérieur compte pour vide — une créature collée au mur de
// la carte est donc poussée vers l'intérieur, ce qui est le comportement voulu.
//
// Une foule parfaitement uniforme rend un gradient nul : sans pente, pas de
// répulsion. C'est correct et non un défaut à corriger — dans une masse de
// densité égale, aucune direction n'est moins encombrée qu'une autre, et c'est
// la poussée des bords qui la desserre.
func (d *DensityGrid) Gradient(u, v int) Vec {
	return Vec{
		FromInt(d.At(u+1, v) - d.At(u-1, v)),
		FromInt(d.At(u, v+1) - d.At(u, v-1)),
	}
}

// inBounds dit si une cellule est dans la grille.
func (d *DensityGrid) inBounds(u, v int) bool {
	return u >= 0 && v >= 0 && u < d.largeur && v < d.hauteur
}
