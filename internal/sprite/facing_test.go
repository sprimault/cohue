// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que la classification des directions garde : les huit axes du monde
// tombent sur les huit bandes de l'écran, les frontières sont où on les attend,
// et un vecteur nul n'invente pas de direction.

package sprite

import (
	"testing"

	"github.com/sprimault/cohue/internal/game"
)

// vers bâtit un vecteur du monde à partir de deux entiers de tuile.
func vers(x, y int) game.Vec {
	return game.Vec{X: game.FromInt(x), Y: game.FromInt(y)}
}

// TestLesHuitAxesDuMondeTombentSurLeursBandes épingle la rotation d'un huitième
// de tour entre les deux repères.
//
// **Elle n'a aucun autre gardien, et elle est fausse de façon plausible.** Le
// monde a deux axes obliques, l'écran huit noms de boussole : une erreur d'un
// cran donne une horde qui marche de travers sans que rien ne tombe, et une
// erreur de signe donne une horde qui marche à reculons. Les deux se lisent à
// l'œil et aucune ne se mesure.
//
// Le sens de chaque ligne : le `+x` du monde descend vers le sud-est de l'écran
// et le `+y` vers le sud-ouest ; leur somme descend donc plein sud et leur
// différence part plein est.
func TestLesHuitAxesDuMondeTombentSurLeursBandes(t *testing.T) {
	for _, c := range []struct {
		x, y   int
		attend string
	}{
		{1, 0, SudEst},
		{1, 1, Sud},
		{0, 1, SudOuest},
		{-1, 1, Ouest},
		{-1, 0, NordOuest},
		{-1, -1, Nord},
		{0, -1, NordEst},
		{1, -1, Est},
	} {
		if got := Facing(vers(c.x, c.y)); got != c.attend {
			t.Errorf("(%d, %d) regarde « %s », attendu « %s »", c.x, c.y, got, c.attend)
		}
	}
}

// TestLaDirectionNeDependPasDeLaLongueur garde ce qu'une classification promet.
//
// Une créature lente et une créature rapide qui vont au même endroit regardent
// le même côté : ce qui décide est l'angle, et un seuil comparé à une longueur
// plutôt qu'à un rapport ferait dépendre l'orientation de la vitesse.
func TestLaDirectionNeDependPasDeLaLongueur(t *testing.T) {
	attendu := Facing(vers(1, 0))
	for _, echelle := range []int{2, 7, 100} {
		if got := Facing(vers(echelle, 0)); got != attendu {
			t.Errorf("à l'échelle %d : « %s », attendu « %s »", echelle, got, attendu)
		}
	}

	// Et sous la tuile, là où la virgule fixe est ce qu'elle a de plus fin : une
	// créature avance de quelques centièmes de tuile par tick, jamais d'une.
	if got := Facing(game.Vec{X: game.One / 64}); got != attendu {
		t.Errorf("à un soixante-quatrième de tuile : « %s », attendu « %s »", got, attendu)
	}
}

// TestLesFrontieresSontAVingtDeuxDegresEtDemi épingle la largeur des secteurs.
//
// **Les huit sont d'égale largeur, et ce n'est pas indifférent.** Des diagonales
// plus étroites que les axes feraient sauter l'orientation d'un cran pendant une
// course en biais, ce qui est le mouvement le plus courant du jeu. Le cas se
// pose des deux côtés de la frontière, faute de quoi le test passerait sur une
// classification qui rendrait toujours la même moitié.
func TestLesFrontieresSontAVingtDeuxDegresEtDemi(t *testing.T) {
	// Un vecteur du monde (dx, dy) donne l'abscisse d'écran dx−dy et l'ordonnée
	// dx+dy ; la frontière entre l'est et le sud-est tombe où la seconde vaut
	// 0,41421 fois la première. On cherche donc de part et d'autre de ce
	// rapport. L'est étant `+x −y`, les deux composantes sont de signes
	// opposés : (700, −300) donne 1000 et 400, soit 0,400 ; (715, −285) donne
	// 1000 et 430, soit 0,430.
	if got := Facing(vers(700, -300)); got != Est {
		t.Errorf("juste sous la frontière : « %s », attendu « %s »", got, Est)
	}
	if got := Facing(vers(715, -285)); got != SudEst {
		t.Errorf("juste au-delà : « %s », attendu « %s »", got, SudEst)
	}
}

// TestUnVecteurNulNaPasDeDirection garde le refus de s'en inventer une.
//
// Rendre le sud ferait pivoter vers l'écran un personnage à l'arrêt sans qu'on
// sache si c'est un choix ou l'absence de réponse. C'est à l'appelant de
// trancher, parce que lui seul sait s'il lui reste une cible à regarder.
func TestUnVecteurNulNaPasDeDirection(t *testing.T) {
	if got := Facing(game.Vec{}); got != "" {
		t.Errorf("le vecteur nul regarde « %s », attendu rien", got)
	}
}
