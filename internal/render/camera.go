// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La caméra : ce qu'elle suit, ce qui l'arrête aux bords du lieu, et le pixel
// entier où elle pose ce qu'on lui donne. C'est par elle que tout dessin passe.

package render

import (
	"math"

	"github.com/sprimault/cohue/internal/game"
)

// camera compose la projection avec le cadrage du lieu.
//
// Son décalage est entier, et c'est ce qui compte : un déplacement en sous-pixel
// fait scintiller le pixel art, ce que le style du jeu interdit. Elle arrondit
// donc avant de rendre quoi que ce soit, y compris pour une entité dont la
// position n'a jamais de raison de tomber sur un pixel.
//
// **Le vide dans les angles n'est pas un défaut de cadrage.** Un lieu est un
// losange et le tampon un rectangle : dès que la caméra bute sur un bord, les
// angles de l'écran tombent hors du losange. Resserrer les bornes pour l'effacer
// décentre le joueur sans y parvenir, puisque le vide ne vient pas des bornes
// mais de la forme. C'est un fond de lieu qui le traitera, pas la caméra.
type camera struct {
	proj projection
	// L'étendue du lieu en pixels, prise sur les quatre sommets du losange.
	minX, minY, maxX, maxY float64
	// Ce qu'il faut ajouter à un pixel du lieu pour obtenir un pixel du tampon.
	dx, dy int
}

// nouvelleCamera monte la caméra sur une taille de tuile et l'étendue d'un lieu.
//
// Les bornes sortent de la projection des quatre sommets plutôt que des formules
// récrites ici : le sommet nord donne le haut, l'ouest la gauche, l'est la
// droite, et le sud le bas.
func nouvelleCamera(tuile [2]int, carte *game.CostGrid) *camera {
	p := nouvelleProjection(tuile)
	largeur, hauteur := game.FromInt(carte.Width()), game.FromInt(carte.Height())

	_, minY := p.pixel(0, 0)
	minX, _ := p.pixel(0, hauteur)
	maxX, _ := p.pixel(largeur, 0)
	_, maxY := p.pixel(largeur, hauteur)

	return &camera{proj: p, minX: minX, minY: minY, maxX: maxX, maxY: maxY}
}

// suivre recentre la caméra sur un point du monde.
func (c *camera) suivre(x, y game.Fixed) {
	ex, ey := c.proj.pixel(x, y)
	c.dx = cadrer(Largeur, ex, c.minX, c.maxX)
	c.dy = cadrer(Hauteur, ey, c.minY, c.maxY)
}

// ecran rend le pixel du tampon où dessiner un point du monde.
//
// C'est la seule conversion que le dessin appelle : elle cadre et elle arrondit,
// là où `projection.pixel` ne fait ni l'un ni l'autre.
func (c *camera) ecran(x, y game.Fixed) (int, int) {
	ex, ey := c.proj.pixel(x, y)
	return c.dx + arrondi(ex), c.dy + arrondi(ey)
}

// casesVisibles rend l'intervalle de cases que le tampon peut montrer, bornes
// comprises.
//
// La fenêtre est un rectangle à l'écran, donc un losange dans le monde : ce qui
// est rendu est son englobant en cases, soit environ le double de ce qui sera
// réellement peint. Le balayer coûte un millier de cases quelle que soit la
// taille du lieu, là où balayer la carte coûterait la sienne.
//
// Une case de marge de chaque côté, parce qu'une face est posée depuis son
// sommet et s'étend vers le bas : celle dont le sommet vient de sortir de
// l'écran y a encore son corps.
func (c *camera) casesVisibles() (u0, v0, u1, v1 int) {
	gauche, haut := -float64(c.dx), -float64(c.dy)
	droite, bas := gauche+Largeur, haut+Hauteur

	u0, v0 = math.MaxInt, math.MaxInt
	u1, v1 = math.MinInt, math.MinInt
	for _, coin := range [4][2]float64{{gauche, haut}, {droite, haut}, {gauche, bas}, {droite, bas}} {
		x, y := c.proj.tuile(coin[0], coin[1])
		u0, u1 = min(u0, x.Floor()), max(u1, x.Floor())
		v0, v1 = min(v0, y.Floor()), max(v1, y.Floor())
	}
	return u0 - 1, v0 - 1, u1 + 1, v1 + 1
}

// cadrer rend le décalage d'un axe : la cible au milieu du tampon, ramenée dans
// ce que le lieu laisse voir.
//
// Quand le lieu est plus étroit que le tampon, l'intervalle admissible est vide
// et il n'y a plus rien à border : le lieu se centre alors une fois pour toutes,
// et le joueur s'y déplace sans que l'image bouge. C'est la seconde moitié de ce
// que veut la conception — se bloquer aux limites plutôt que découvrir du vide,
// et centrer ce qui tient tout entier.
func cadrer(tampon int, cible, borneBasse, borneHaute float64) int {
	mini, maxi := float64(tampon)-borneHaute, -borneBasse
	if mini > maxi {
		return arrondi((float64(tampon) - borneBasse - borneHaute) / 2)
	}
	return arrondi(math.Min(math.Max(float64(tampon)/2-cible, mini), maxi))
}

// arrondi ramène un pixel flottant au plus proche pixel entier.
func arrondi(pixels float64) int { return int(math.Round(pixels)) }
