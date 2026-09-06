// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que l'aplatissement garde : la forme du dessin et rien de sa couleur, le
// transparent qui le reste, et une taille inchangée.

package sprite

import (
	"image"
	"image/color"
	"testing"
)

// dessinDEssai rend une image dont la moitié gauche est opaque et colorée, la
// droite transparente, avec un pixel à demi opaque au milieu.
func dessinDEssai() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 2))
	img.SetNRGBA(0, 0, color.NRGBA{R: 200, G: 40, B: 10, A: 255})
	img.SetNRGBA(1, 0, color.NRGBA{R: 10, G: 200, B: 40, A: 255})
	img.SetNRGBA(0, 1, color.NRGBA{R: 40, G: 10, B: 200, A: 128})
	return img
}

// TestLaplatissementGardeLaFormeEtPerdLaCouleur épingle ce que l'aplat doit
// être.
//
// **Il n'a aucun autre gardien** : le rendu qui s'en sert n'a pas de test par
// doctrine, et une couleur qui survivrait ne se verrait qu'à l'œil — un contour
// qui prendrait la teinte du sprite au lieu de celle qu'on lui donne se lit
// comme un halo sale, pas comme un défaut.
func TestLaplatissementGardeLaFormeEtPerdLaCouleur(t *testing.T) {
	plat := Flatten(dessinDEssai())

	if taille := plat.Bounds().Size(); taille.X != 4 || taille.Y != 2 {
		t.Fatalf("taille %v, attendu 4x2", taille)
	}
	for _, c := range []struct {
		x, y     int
		r, a     uint32
		pourquoi string
	}{
		{0, 0, 0xffff, 0xffff, "un pixel opaque devient blanc opaque"},
		{1, 0, 0xffff, 0xffff, "sa couleur ne survit pas"},
		{3, 1, 0, 0, "le transparent le reste"},
	} {
		r, _, _, a := plat.At(c.x, c.y).RGBA()
		if r != c.r || a != c.a {
			t.Errorf("(%d,%d) : rouge %d alpha %d, attendu %d et %d — %s",
				c.x, c.y, r, a, c.r, c.a, c.pourquoi)
		}
	}
}

// TestUnAlphaPartielSurvitALaplatissement garde le choix de ne pas binariser.
//
// La forme est celle du dessin et non une idée de sa forme : arrondir ici
// déciderait à la place du générateur, qui fait déjà des contours francs. Un
// jour où un sprite aurait des bords fondus, sa silhouette les aurait aussi —
// ce qui est la réponse juste.
func TestUnAlphaPartielSurvitALaplatissement(t *testing.T) {
	plat := Flatten(dessinDEssai())

	_, _, _, a := plat.At(0, 1).RGBA()
	if a == 0 || a == 0xffff {
		t.Errorf("alpha %d sur un pixel à demi opaque : il a été arrondi", a)
	}
}
