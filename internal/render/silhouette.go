// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'aplatissement d'un dessin en sa seule forme, et les deux emplois qu'on en
// fait : le contour qui détache en permanence, la silhouette qui révèle quand
// quelque chose recouvre.

package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/sprite"
)

// aplatir rend la forme d'une image, convertie en texture.
//
// Le calcul vit dans `internal/sprite`, où un test peut le garder : ce paquet
// n'en a pas, et une découpe fausse d'un pixel ne se verrait qu'à l'œil, sur un
// contour qui aurait un trou.
func aplatir(src image.Image) *ebiten.Image {
	return ebiten.NewImageFromImage(sprite.Flatten(src))
}

// trace est ce qu'un dessin laisse à la séquence : le coin où il s'est posé, sa
// forme, et son aplat quand il est de ceux qu'on révèle.
//
// Elle remplace la boîte que les fonctions de dessin rendaient. Une boîte
// suffisait tant que le recouvrement se jugeait dessus ; du jour où il se juge
// au pixel, ce qu'il faut transporter est la forme, et la porter à côté de la
// boîte aurait laissé deux descriptions du même dessin.
type trace struct {
	x, y   int
	masque *sprite.Mask
	// forme est l'aplat blanc, non nul pour les deux seules sortes que la
	// conception révèle : le joueur et le projectile de la horde.
	forme *ebiten.Image
}

// revele est une chose posée qu'on redessinera si quelque chose la recouvre.
//
// Elle porte l'aplat et son coin plutôt que l'entité d'où elle vient : au moment
// de révéler, la séquence est finie et rien ne garantit qu'une place de bassin
// désigne encore la même chose. Ce qu'il faut alors est une image et deux
// nombres, pas une référence.
type revele struct {
	forme  *ebiten.Image
	masque *sprite.Mask
	x, y   int
	teinte color.RGBA
	// couvert dit qu'au moins un pixel de la forme a disparu sous ce qui a été
	// posé après elle. C'est la condition, et la seule, pour la redessiner.
	couvert bool
}

// recouvertPar dit si un dessin posé après celui-ci en cache un pixel.
//
// **La boîte ne suffit pas, et c'est tout le sujet.** Une bande de personnage
// fait soixante-quatre pixels de côté quand le joueur en occupe dix-neuf sur
// quarante-cinq : quatre croisements de boîtes sur cinq ne touchent aucun pixel
// de lui. Le test portait sur les boîtes, si bien qu'à cent ennemis le
// personnage était remplacé par son aplat blanc quatre images sur cinq — ce qui
// lui retirait l'orientation que sa bande venait de lui donner.
//
// **Les deux côtés doivent porter un masque.** Serrer la boîte du seul révélé
// laisse celle de ce qui recouvre, et le faux déclenchement revient par là.
func (r revele) recouvertPar(t trace) bool {
	return r.masque.Overlaps(r.x, r.y, t.masque, t.x, t.y)
}

// Les huit décalages d'un contour d'un pixel.
//
// **Huit et non quatre**, parce qu'un sprite isométrique n'a pas d'arête droite :
// ses bords descendent en escalier, et un contour en croix y laisse des trous
// dans les diagonales — exactement là où la silhouette d'un personnage est le
// plus mince.
var contourAutour = [8][2]int{
	{-1, -1}, {0, -1}, {1, -1},
	{-1, 0}, {1, 0},
	{-1, 1}, {0, 1}, {1, 1},
}

// contour cerne une forme d'un liseré, sous l'image qu'il détache.
//
// **Il détache le personnage de la horde, et c'est mesuré.** Le pixel le plus
// clair d'un profil monte à 162, celui d'un figurant à 175 ; un liseré à plein
// blanc gagne donc plus de quatre-vingts de luminance sur tout ce qu'il peut
// toucher. C'est l'ordre de grandeur qui avait fait retenir le liseré du tir de
// la Buse, et c'est le même remède au même problème — reconnaître un objet dans
// une masse qui partage sa gamme.
//
// **Ce que la valeur ne pouvait plus faire.** Le chapitre 2 sépare le joueur de
// la horde par la valeur, ce qui supposait un sol sombre ; sur le décor livré, à
// 162 de luminance, l'éclaircir le rapproche du sol et l'assombrir le renvoie
// dans la foule. Aucune teinte ne satisfait les deux, et c'est pourquoi la
// réponse est une forme.
//
// Reste un fond qu'il ne bat que de vingt-quatre : le carrelage du supermarché,
// à 231. Ce n'est pas traité — un joueur seul sur du sol clair n'est pas le cas
// à risque, et ce que le chapitre 2 protège est le personnage sous un
// empilement.
func (s *Screen) contour(ecran, forme *ebiten.Image, x, y int, teinte color.RGBA) {
	for _, pas := range contourAutour {
		s.op.GeoM.Reset()
		s.op.GeoM.Translate(float64(x+pas[0]), float64(y+pas[1]))
		s.op.ColorScale.Reset()
		s.op.ColorScale.ScaleWithColor(teinte)
		ecran.DrawImage(forme, &s.op)
	}
}
