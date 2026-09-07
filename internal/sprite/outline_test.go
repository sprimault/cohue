// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le bord garde : le pourtour de la forme, son intérieur laissé libre, et
// une taille inchangée.

package sprite

import (
	"image"
	"image/color"
	"testing"
)

// carreDEssai rend une image de six pixels de côté portant un carré opaque de
// quatre, laissant une marge d'un pixel tout autour.
func carreDEssai() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, 6, 6))
	for y := 1; y < 5; y++ {
		for x := 1; x < 5; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 90, G: 90, B: 90, A: 255})
		}
	}
	return img
}

// TestLeBordCerneLaFormeSansLaRemplir épingle ce qu'un bord doit être.
//
// **Il n'a aucun autre gardien** : le rendu qui s'en sert n'a pas de test par
// doctrine, et un bord qui remplirait sa forme donnerait exactement l'aplat
// qu'on cherche à remplacer — le défaut passerait alors pour l'ancien
// comportement plutôt que pour une régression.
func TestLeBordCerneLaFormeSansLaRemplir(t *testing.T) {
	bord := Outline(carreDEssai())

	if taille := bord.Bounds().Size(); taille.X != 6 || taille.Y != 6 {
		t.Fatalf("taille %v, attendu 6x6", taille)
	}
	for _, c := range []struct {
		x, y     int
		a        uint32
		pourquoi string
	}{
		{1, 1, 0xffff, "un coin de la forme touche le vide par trois voisins"},
		{2, 1, 0xffff, "le haut de la forme est un bord"},
		{1, 3, 0xffff, "son flanc aussi"},
		{2, 2, 0, "l'intérieur reste libre, c'est tout l'objet"},
		{3, 3, 0, "et il l'est sur toute son étendue"},
		{0, 0, 0, "rien ne déborde de la forme"},
	} {
		_, _, _, a := bord.At(c.x, c.y).RGBA()
		if a != c.a {
			t.Errorf("(%d,%d) : alpha %d, attendu %d — %s", c.x, c.y, a, c.a, c.pourquoi)
		}
	}
}

// TestUnBordSeVoitEnBlanc garde la teinte neutre, celle que la pose multiplie.
//
// C'est la même raison qu'à l'aplat : le blanc est l'élément neutre, et une
// couleur retenue ici obligerait à une image par teinte.
func TestUnBordSeVoitEnBlanc(t *testing.T) {
	r, v, b, _ := Outline(carreDEssai()).At(1, 1).RGBA()
	if r != 0xffff || v != 0xffff || b != 0xffff {
		t.Errorf("bord (%d,%d,%d), attendu du blanc plein", r, v, b)
	}
}

// TestUneFormeColleeAuCadreGardeSonBord vaut pour le choix du bord intérieur :
// un pixel du cadre n'a pas de voisin au-delà, et il est donc lui-même un bord.
//
// Le bord extérieur, lui, aurait eu à déborder de l'image pour dire la même
// chose, et ce qu'il aurait posé hors du cadre serait perdu.
func TestUneFormeColleeAuCadreGardeSonBord(t *testing.T) {
	plein := image.NewNRGBA(image.Rect(0, 0, 3, 3))
	for y := range 3 {
		for x := range 3 {
			plein.SetNRGBA(x, y, color.NRGBA{R: 20, G: 20, B: 20, A: 255})
		}
	}
	bord := Outline(plein)

	if _, _, _, a := bord.At(0, 0).RGBA(); a != 0xffff {
		t.Errorf("coin du cadre : alpha %d, attendu opaque", a)
	}
	if _, _, _, a := bord.At(1, 1).RGBA(); a != 0 {
		t.Errorf("centre : alpha %d, attendu libre", a)
	}
}

// TestUnAlphaPartielAppartientAuBord garde l'accord de seuil avec l'aplat et le
// masque : un pixel entre dans la forme dès que son alpha n'est pas nul.
func TestUnAlphaPartielAppartientAuBord(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	img.SetNRGBA(1, 0, color.NRGBA{R: 10, G: 10, B: 10, A: 1})

	if _, _, _, a := Outline(img).At(1, 0).RGBA(); a != 0xffff {
		t.Errorf("pixel à peine opaque : alpha %d, attendu retenu comme bord", a)
	}
}
