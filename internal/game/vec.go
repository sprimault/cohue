// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Vec est un déplacement du monde en virgule fixe, avec ses opérations et les
// huit orientations tabulées. La normalisation y traite le cas nul, qui n'a pas
// de réponse mathématique et en exige une.

package game

import "math"

// Vec est un déplacement ou une direction dans le monde, en tuiles.
//
// Deux longueurs en virgule fixe, et rien d'autre : ce que la simulation
// manipule ne passe jamais par un flottant, sauf à l'endroit précis où la
// normalisation l'exige et où l'IEEE-754 garantit le même résultat partout.
type Vec struct {
	X, Y Fixed
}

// diag est la composante d'une direction diagonale unitaire : 65536 × √2 ⁄ 2,
// arrondi.
//
// Une diagonale n'est donc pas exactement unitaire — sa norme vaut 65536,07
// pour une tuile de 65536, un millionième de trop. Autant dire exacte ; le
// chiffre est ici pour qui réemploiera la table et se demanderait si ses
// diagonales sont lentes.
const diag Fixed = 46341

// Headings est le nombre d'orientations que le jeu distingue.
const Headings = 8

// headings est la table des huit orientations unitaires du monde, en sens
// trigonométrique depuis l'axe des X.
//
// Tabulées et non calculées, pour une raison qui dépasse la vitesse : `sin` et
// `cos` ne sont pas correctement arrondis par l'IEEE-754, contrairement à
// `sqrt`. Leur dernier bit peut différer d'une architecture à l'autre, et faire
// entrer de la trigonométrie ici rouvrirait la porte que la virgule fixe a
// fermée — sur le cas le plus difficile à diagnostiquer, une divergence qui
// n'apparaît que lorsque deux entités se superposent.
//
// Elle reste privée et s'atteint par `Heading` : exportée, `game.Headings[0] =
// …` compilerait, et un appelant pourrait déplacer le nord de tout le jeu depuis
// n'importe où.
var headings = [Headings]Vec{
	{One, 0}, {diag, diag}, {0, One}, {-diag, diag},
	{-One, 0}, {-diag, -diag}, {0, -One}, {diag, -diag},
}

// Heading rend l'une des huit orientations du monde, l'index étant ramené dans
// la table.
//
// Elle sert au cas nul de `Direction`, et servira à tout ce qui a besoin de
// huit orientations plutôt que d'une infinité — l'orientation d'un sprite, un
// tir en éventail, la gerbe d'un éclat.
//
// L'ordre est celui du **monde**, et non celui des huit noms que le manifeste
// des personnages donne à ses bandes — `S`, `SO`, `O`… — qui sont des
// orientations d'**écran**. La correspondance entre les deux appartient à
// `internal/render`, avec le reste de la projection : la poser ici mélangerait
// les deux repères dans le paquet qui n'en connaît qu'un.
func Heading(index int) Vec { return headings[index&(Headings-1)] }

// Add rend la somme de deux vecteurs.
func (v Vec) Add(w Vec) Vec { return Vec{v.X + w.X, v.Y + w.Y} }

// Sub rend la différence de deux vecteurs.
func (v Vec) Sub(w Vec) Vec { return Vec{v.X - w.X, v.Y - w.Y} }

// Scale multiplie un vecteur par un facteur.
func (v Vec) Scale(f Fixed) Vec { return Vec{v.X.Mul(f), v.Y.Mul(f)} }

// Perp rend le vecteur tourné d'un quart de tour dans le sens trigonométrique.
//
// C'est la dérive latérale d'un flanqueur : ajoutée à l'attirance du champ de
// flux, elle referme le cercle autour du joueur au lieu de foncer dessus.
func (v Vec) Perp() Vec { return Vec{-v.Y, v.X} }

// Len rend la norme du vecteur.
//
// La somme des carrés est exacte, calculée en `int64` : elle ne devient un
// flottant qu'au moment de la racine, dont l'IEEE-754 exige l'arrondi correct.
// C'est la seule opération flottante que la simulation admet, et le résultat est
// arrondi au plus proche plutôt que tronqué — une troncature raccourcit
// toujours, et les diagonales deviendraient plus lentes que les axes.
func (v Vec) Len() Fixed {
	return borner(int64(math.Round(math.Sqrt(float64(v.carres())))))
}

// Direction rend le vecteur ramené à la longueur d'une tuile.
//
// L'index est celui de l'entité dans son bassin, et il ne sert qu'au cas nul.
// Il est demandé même quand il ne servira pas, parce que c'est ce qui rend ce
// cas impossible à oublier : une variante rendant `(Vec, bool)` laisserait
// l'appelant écrire `_` sans que le compilateur ait rien à redire.
//
// Le vecteur nul a une direction, et elle vient de l'entité. La somme des forces
// peut s'annuler — une créature pile sur le joueur, deux créatures exactement
// superposées dans le gradient de densité, ce qu'un anneau d'apparition finit
// toujours par produire. Une direction fixe serait la pire des réponses :
// superposées, deux entités recevraient la même correction et le resteraient.
//
// Deux entités vivantes ont toujours des index distincts, donc deux directions
// différentes, donc elles se séparent au tick suivant. Le pas de trois est
// premier avec huit et porte l'écart entre deux index consécutifs à 135 degrés,
// là où un pas de un n'en donnerait que 45 : deux créatures apparues côte à côte
// ont des index voisins, et se sépareraient lentement.
//
// L'index n'a pas besoin d'être stable dans le temps — la suppression par
// échange le change dès qu'une entité meurt devant. Ce qu'il faut est l'unicité
// à un instant donné, et le cas dure un tick.
func (v Vec) Direction(index int) Vec {
	carres := v.carres()
	if carres == 0 {
		return Heading(index * 3)
	}
	norme := math.Sqrt(float64(carres))
	return Vec{
		borner(int64(math.Round(float64(v.X) * float64(One) / norme))),
		borner(int64(math.Round(float64(v.Y) * float64(One) / norme))),
	}
}

// carres rend la somme des carrés des composantes, exacte.
//
// En `int64` parce que le carré d'une longueur n'entre pas dans un `int32`, et
// sans arrondi intermédiaire : c'est ce qui permet à `Direction` de rendre une
// diagonale unitaire même pour un vecteur d'une seule unité par axe, là où une
// norme arrondie à l'unité aurait donné un vecteur d'une longueur et demie.
func (v Vec) carres() int64 {
	return int64(v.X)*int64(v.X) + int64(v.Y)*int64(v.Y)
}
