// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La forme d'un dessin réduite à un bit par pixel, et le recouvrement de deux
// formes posées à l'écran.

package sprite

import (
	"image"
	"image/color"
)

// Mask est la forme d'un dessin : un bit par pixel, levé là où le dessin est
// opaque.
//
// **Il existe parce qu'une boîte n'est pas une forme.** Une bande de personnage
// fait soixante-quatre pixels de côté et le joueur en occupe dix-neuf sur
// quarante-cinq : quatre recouvrements de boîtes sur cinq ne touchent aucun
// pixel de lui. Ce qui décide de révéler une silhouette doit donc porter sur les
// pixels, et un bit par pixel se teste sans lire de texture — c'est-à-dire sans
// rien demander au processeur graphique au milieu d'une image.
type Mask struct {
	largeur, hauteur int
	// serree est l'étendue des pixels opaques, relative au coin de l'image.
	// Vide quand rien ne l'est.
	//
	// Elle sert de premier tri : deux formes dont les étendues ne se croisent
	// pas ne se recouvrent pas, et c'est le cas courant.
	serree image.Rectangle
	bits   []uint64
}

// NewMask relève la forme d'un dessin.
//
// **Le seuil est celui de `Flatten`, et il ne peut pas en différer** : un pixel
// entre dans la forme dès que son alpha n'est pas nul. Un masque plus sévère que
// l'aplat révélerait des silhouettes dont le liseré déborde de ce qu'on a testé,
// et l'écart ne se verrait qu'à l'œil sur un bord fondu.
//
// Rend nil pour un dessin absent, ce que `Overlaps` accepte : toutes les poses
// n'ont pas d'image, et l'appelant n'a pas à le savoir deux fois.
func NewMask(src image.Image) *Mask {
	if src == nil {
		return nil
	}
	boite := src.Bounds()
	m := &Mask{
		largeur: boite.Dx(),
		hauteur: boite.Dy(),
		bits:    make([]uint64, (boite.Dx()*boite.Dy()+63)/64),
	}

	x0, y0 := m.largeur, m.hauteur
	x1, y1 := 0, 0
	for y := range m.hauteur {
		for x := range m.largeur {
			if color.NRGBAModel.Convert(src.At(boite.Min.X+x, boite.Min.Y+y)).(color.NRGBA).A == 0 {
				continue
			}
			i := y*m.largeur + x
			m.bits[i>>6] |= 1 << uint(i&63)
			x0, y0 = min(x0, x), min(y0, y)
			x1, y1 = max(x1, x+1), max(y1, y+1)
		}
	}
	if x0 < x1 {
		m.serree = image.Rect(x0, y0, x1, y1)
	}
	return m
}

// Bounds rend l'étendue des pixels opaques, relative au coin du dessin.
func (m *Mask) Bounds() image.Rectangle {
	if m == nil {
		return image.Rectangle{}
	}
	return m.serree
}

// opaque dit si le pixel est dans la forme. Hors du dessin, il n'y est pas.
func (m *Mask) opaque(x, y int) bool {
	if x < 0 || y < 0 || x >= m.largeur || y >= m.hauteur {
		return false
	}
	i := y*m.largeur + x
	return m.bits[i>>6]&(1<<uint(i&63)) != 0
}

// Overlaps dit si deux formes posées à l'écran partagent au moins un pixel.
//
// Les coins sont ceux des dessins, en pixels d'écran, tels qu'on les a passés au
// dessin lui-même : c'est la seule façon de ne pas avoir à retrancher les
// ancrages une seconde fois ici.
//
// **L'étendue serrée est testée d'abord**, et c'est elle qui rend l'appel bon
// marché : deux boîtes de bande qui se croisent sans que les personnages se
// touchent sortent en quatre comparaisons.
func (m *Mask) Overlaps(x, y int, o *Mask, ox, oy int) bool {
	if m == nil || o == nil {
		return false
	}
	croix := m.serree.Add(image.Pt(x, y)).Intersect(o.serree.Add(image.Pt(ox, oy)))
	if croix.Empty() {
		return false
	}
	for py := croix.Min.Y; py < croix.Max.Y; py++ {
		for px := croix.Min.X; px < croix.Max.X; px++ {
			if m.opaque(px-x, py-y) && o.opaque(px-ox, py-oy) {
				return true
			}
		}
	}
	return false
}
