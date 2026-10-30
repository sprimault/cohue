// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que l'aplatissement garde : la forme du dessin et rien de sa couleur, le
// transparent qui le reste, et une taille inchangée.

package sprite

import (
	"image"
	"testing"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/render"
)

// pave rend une planche transparente où les rectangles donnés sont opaques.
func pave(largeur, hauteur uint16, pleins ...image.Rectangle) raster.Sheet {
	p := raster.NewSheet(largeur, hauteur)
	for _, plein := range pleins {
		for y := plein.Min.Y; y < plein.Max.Y; y++ {
			for x := plein.Min.X; x < plein.Max.X; x++ {
				p.Pixels[y*int(largeur)+x] = render.Premultiply(200, 200, 200, 255)
			}
		}
	}
	return p
}

// pixel rend la couleur d'un pixel d'une planche.
func pixel(p raster.Sheet, x, y int) render.Color {
	return p.Pixels[y*int(p.Width)+x]
}

// dessinDEssai rend une planche dont la moitié gauche est opaque et colorée, la
// droite transparente, avec un pixel à demi opaque au milieu.
func dessinDEssai() raster.Sheet {
	p := raster.NewSheet(4, 2)
	p.Pixels[0] = render.Premultiply(200, 40, 10, 255)
	p.Pixels[1] = render.Premultiply(10, 200, 40, 255)
	p.Pixels[4] = render.Premultiply(40, 10, 200, 128)
	return p
}

// TestLaplatissementGardeLaFormeEtPerdLaCouleur épingle ce que l'aplat doit
// être.
//
// **Il n'a aucun autre gardien** : une couleur qui survivrait ne se verrait qu'à
// l'œil — un contour qui prendrait la teinte du sprite au lieu de celle qu'on
// lui donne se lit comme un halo sale, pas comme un défaut.
func TestLaplatissementGardeLaFormeEtPerdLaCouleur(t *testing.T) {
	plat := Flatten(dessinDEssai())

	if plat.Width != 4 || plat.Height != 2 {
		t.Fatalf("taille %dx%d, attendu 4x2", plat.Width, plat.Height)
	}
	for _, c := range []struct {
		x, y     int
		r, a     uint8
		pourquoi string
	}{
		{0, 0, 255, 255, "un pixel opaque devient blanc opaque"},
		{1, 0, 255, 255, "sa couleur ne survit pas"},
		{3, 1, 0, 0, "le transparent le reste"},
	} {
		got := pixel(plat, c.x, c.y)
		if got.R != c.r || got.A != c.a {
			t.Errorf("(%d,%d) : rouge %d alpha %d, attendu %d et %d — %s",
				c.x, c.y, got.R, got.A, c.r, c.a, c.pourquoi)
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
	if a := pixel(Flatten(dessinDEssai()), 0, 1).A; a == 0 || a == 255 {
		t.Errorf("alpha %d sur un pixel à demi opaque : il a été arrondi", a)
	}
}

// TestLaplatissementGardeLaDecoupe garde l'accord des deux planches.
//
// Un aspect prend le même rectangle dans la forme que dans le dessin : une
// dérivée sans découpe obligerait à la recalculer à côté, et c'est cette
// seconde description qui finirait par désigner une autre case.
func TestLaplatissementGardeLaDecoupe(t *testing.T) {
	dessin := dessinDEssai()
	if err := dessin.CutGrid(2, 2); err != nil {
		t.Fatal(err)
	}

	plat := Flatten(dessin)
	if len(plat.Rectangles) != len(dessin.Rectangles) {
		t.Fatalf("%d case(s) dans l'aplat, %d dans le dessin",
			len(plat.Rectangles), len(dessin.Rectangles))
	}
	for i, r := range dessin.Rectangles {
		if plat.Rectangles[i] != r {
			t.Errorf("case %d : %v, attendue %v", i, plat.Rectangles[i], r)
		}
	}
}
