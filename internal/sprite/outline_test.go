// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le bord garde : le pourtour de la forme, son intérieur laissé libre,
// une taille inchangée, et des cases qui s'ignorent l'une l'autre.

package sprite

import (
	"image"
	"testing"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/render"
)

// carreDEssai rend une planche de six pixels de côté portant un carré opaque de
// quatre, laissant une marge d'un pixel tout autour.
func carreDEssai() raster.Sheet {
	return pave(6, 6, image.Rect(1, 1, 5, 5))
}

// TestLeBordCerneLaFormeSansLaRemplir épingle ce qu'un bord doit être.
//
// **Il n'a aucun autre gardien** : un bord qui remplirait sa forme donnerait
// exactement l'aplat qu'on cherche à remplacer — le défaut passerait alors pour
// l'ancien comportement plutôt que pour une régression.
func TestLeBordCerneLaFormeSansLaRemplir(t *testing.T) {
	bord := Outline(carreDEssai())

	if bord.Width != 6 || bord.Height != 6 {
		t.Fatalf("taille %dx%d, attendu 6x6", bord.Width, bord.Height)
	}
	for _, c := range []struct {
		x, y     int
		a        uint8
		pourquoi string
	}{
		{1, 1, 255, "un coin de la forme touche le vide par trois voisins"},
		{2, 1, 255, "le haut de la forme est un bord"},
		{1, 3, 255, "son flanc aussi"},
		{2, 2, 0, "l'intérieur reste libre, c'est tout l'objet"},
		{3, 3, 0, "et il l'est sur toute son étendue"},
		{0, 0, 0, "rien ne déborde de la forme"},
	} {
		if a := pixel(bord, c.x, c.y).A; a != c.a {
			t.Errorf("(%d,%d) : alpha %d, attendu %d — %s", c.x, c.y, a, c.a, c.pourquoi)
		}
	}
}

// TestUnBordSeVoitEnBlanc garde la teinte neutre, celle que la pose multiplie.
//
// C'est la même raison qu'à l'aplat : le blanc est l'élément neutre, et une
// couleur retenue ici obligerait à une image par teinte.
func TestUnBordSeVoitEnBlanc(t *testing.T) {
	if c := pixel(Outline(carreDEssai()), 1, 1); c != render.White() {
		t.Errorf("bord %v, attendu du blanc plein", c)
	}
}

// TestUneFormeColleeAuCadreGardeSonBord vaut pour le choix du bord intérieur :
// un pixel du cadre n'a pas de voisin au-delà, et il est donc lui-même un bord.
//
// Le bord extérieur, lui, aurait eu à déborder de l'image pour dire la même
// chose, et ce qu'il aurait posé hors du cadre serait perdu.
func TestUneFormeColleeAuCadreGardeSonBord(t *testing.T) {
	bord := Outline(pave(3, 3, image.Rect(0, 0, 3, 3)))

	if a := pixel(bord, 0, 0).A; a != 255 {
		t.Errorf("coin du cadre : alpha %d, attendu opaque", a)
	}
	if a := pixel(bord, 1, 1).A; a != 0 {
		t.Errorf("centre : alpha %d, attendu libre", a)
	}
}

// TestUnAlphaPartielAppartientAuBord garde l'accord de seuil avec l'aplat et le
// masque : un pixel entre dans la forme dès que son alpha n'est pas nul.
func TestUnAlphaPartielAppartientAuBord(t *testing.T) {
	p := raster.NewSheet(3, 1)
	p.Pixels[1] = render.Premultiply(10, 10, 10, 1)

	if a := pixel(Outline(p), 1, 0).A; a != 255 {
		t.Errorf("pixel à peine opaque : alpha %d, attendu retenu comme bord", a)
	}
}

// TestLesCasesDUneBandeSignorent garde ce que la planche unique a changé.
//
// Les images d'un cycle se touchent désormais dans la même planche, là où une
// sous-image les isolait : relever le bord sur la planche entière ferait perdre
// à chaque case son bord droit, parce que la case suivante commence là. Le
// défaut tomberait sur le liseré d'un personnage en marche, et seulement sur
// deux de ses huit directions.
func TestLesCasesDUneBandeSignorent(t *testing.T) {
	bande := pave(6, 3, image.Rect(0, 0, 6, 3))
	if err := bande.CutGrid(3, 3); err != nil {
		t.Fatal(err)
	}

	bord := Outline(bande)
	for _, c := range []struct {
		x, y     int
		a        uint8
		pourquoi string
	}{
		{2, 1, 255, "le bord droit de la première case en est un, sa voisine ne le comble pas"},
		{3, 1, 255, "le bord gauche de la seconde aussi"},
		{1, 1, 0, "le centre d'une case reste libre"},
		{4, 1, 0, "celui de la seconde également"},
	} {
		if a := pixel(bord, c.x, c.y).A; a != c.a {
			t.Errorf("(%d,%d) : alpha %d, attendu %d — %s", c.x, c.y, a, c.a, c.pourquoi)
		}
	}
}
