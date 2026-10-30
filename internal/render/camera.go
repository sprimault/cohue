// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La caméra : ce qu'elle suit, et ce qui l'arrête aux bords du lieu. Ce qu'elle
// pose est la vue de la scène, qui place et trie elle-même.

package render

import (
	"math"

	"github.com/sprimault/lozengine/scene"

	"github.com/sprimault/cohue/internal/game"
)

// margeDeCadrage est ce qu'on garde à l'intérieur du losange, en pixels.
//
// **La borne exacte est inatteignable, et c'est pourquoi on ne la vise pas.** Le
// cadrage se calcule en pixels — c'est le seul repère où la limite de l'image
// est un rectangle — mais `Scene.Focus` ne prend qu'un point du monde, si bien
// que le retour par `Matrix.Unproject` admet le quantum que sa godoc annonce,
// « sous un vingtième de cellule » : environ 1,6 pixel à la tuile 64×32. Viser
// la limite laisserait donc la vue en sortir de ce quantum, et deux pixels de
// vide au bord se voient sur du pixel art.
//
// Deux pixels couvrent le quantum avec de la marge, et c'est `docs/go.md` —
// « une valeur posée sur la borne du problème est fausse en permanence ».
const margeDeCadrage = 2

// camera pose la vue de la scène sur ce qu'elle suit, sans la laisser sortir du
// lieu.
//
// **Elle ne convertit plus rien.** La projection vit dans la matrice de la
// scène, le pixel entier est garanti par l'origine de sa vue, et le tri lui
// appartient : ce qui reste ici est le cadrage, que le moteur renvoie
// explicitement au jeu — lui seul sait où sa carte s'arrête.
//
// **Le vide dans les angles n'est pas un défaut de cadrage.** Un lieu est un
// losange et le tampon un rectangle : dès que la caméra bute sur un bord, les
// angles de l'écran tombent hors du losange. Resserrer les bornes pour l'effacer
// décentre le joueur sans y parvenir, puisque le vide ne vient pas des bornes
// mais de la forme. C'est un fond de lieu qui le traitera, pas la caméra.
type camera struct {
	scene *scene.Scene
	// L'étendue du lieu en pixels absolus, prise sur les quatre sommets du
	// losange.
	minX, minY, maxX, maxY float64
}

// nouvelleCamera monte la caméra sur une scène et l'étendue d'un lieu.
//
// Les bornes sortent de la projection des quatre sommets plutôt que de formules
// récrites ici. **Et elles s'en sortent par un minimum et un maximum, sans dire
// quel sommet tombe où** : la matrice décide du sens de l'axe des abscisses, et
// nommer le sommet qui donne la gauche est une seconde description d'elle. Celle
// qui était écrite ici disait l'inverse de ce que `Dimetric` rend, si bien que
// l'intervalle admissible était vide sur toute sa largeur : chaque vue retombait
// sur la branche du lieu trop étroit, se centrait en abscisse, et le joueur
// sortait du champ dès qu'il s'éloignait de la diagonale.
func nouvelleCamera(sc *scene.Scene, carte *game.CostGrid) *camera {
	largeur, hauteur := float64(carte.Width()), float64(carte.Height())

	c := &camera{
		scene: sc,
		minX:  math.Inf(1), minY: math.Inf(1),
		maxX: math.Inf(-1), maxY: math.Inf(-1),
	}
	for _, sommet := range [4][2]float64{
		{0, 0}, {largeur, 0}, {0, hauteur}, {largeur, hauteur},
	} {
		x, y := pixelDe(sc, sommet[0], sommet[1])
		c.minX, c.maxX = math.Min(c.minX, x), math.Max(c.maxX, x)
		c.minY, c.maxY = math.Min(c.minY, y), math.Max(c.maxY, y)
	}
	return c
}

// suivre recentre la vue sur un point du monde, borné à ce que le lieu montre.
//
// Le point revient au monde avant d'être posé, parce que c'est la seule entrée
// que la scène offre : le cadrage se décide en pixels, la vue se pose en
// cellules.
//
// **La taille vient de la vue et non des constantes.** L'aire de monde est la
// surface moins la bande du bandeau, et la surface suit la fenêtre : elle ne
// vaut donc la résolution interne à aucun moment, et borner le cadrage sur
// 960×540 laisserait la caméra dépasser le losange du lieu de la différence.
// Elle se relit à chaque pas plutôt que d'être retenue, la boucle du moteur la
// réimposant — les deux derniers paramètres de `Focus` sont inertes sous `Run`,
// et seule la planche, qui tient sa propre boucle, les voit.
func (c *camera) suivre(x, y game.Fixed) {
	_, _, largeur, hauteur := c.scene.View()

	ex, ey := pixelDe(c.scene, x.Float(), y.Float())
	vise := cadrer(int(largeur), ex, c.minX, c.maxX)
	visey := cadrer(int(hauteur), ey, c.minY, c.maxY)

	mx, my := c.scene.Projection().Unproject(arrondi32(vise), arrondi32(visey))
	c.scene.Focus(mx, my, largeur, hauteur)
}

// pixelDe rend le pixel absolu où tombe un point du monde, dans le plan du sol.
func pixelDe(sc *scene.Scene, x, y float64) (float64, float64) {
	ex, ey, _, _ := sc.Projection().Project(x, y, 0)
	return float64(ex), float64(ey)
}

// casesVisibles rend l'intervalle de cases que le tampon peut montrer, bornes
// comprises.
//
// **Elle survit au portage, et la nappe est son seul lecteur.** Le décor ne se
// relève plus par image — la scène le garde posé —, mais le marquage d'une zone
// naît et meurt, donc il faut encore savoir quelles cases l'image couvre. Ce qui
// a disparu est le relevé du décor, pas le cadrage en cases.
//
// La fenêtre est un rectangle à l'écran, donc un losange dans le monde : ce qui
// est rendu est son englobant en cases, soit environ le double de ce qui sera
// réellement peint.
//
// Une case de marge de chaque côté, parce qu'une face est posée depuis son
// sommet et s'étend vers le bas : celle dont le sommet vient de sortir de
// l'écran y a encore son corps.
func (c *camera) casesVisibles() (u0, v0, u1, v1 int) {
	x, y, largeur, hauteur := c.scene.View()
	m := c.scene.Projection()

	u0, v0 = math.MaxInt, math.MaxInt
	u1, v1 = math.MinInt, math.MinInt
	for _, coin := range [4][2]int32{
		{x, y}, {x + largeur, y}, {x, y + hauteur}, {x + largeur, y + hauteur},
	} {
		mx, my := m.Unproject(coin[0], coin[1])
		u, v := int(math.Floor(mx)), int(math.Floor(my))
		u0, u1 = min(u0, u), max(u1, u)
		v0, v1 = min(v0, v), max(v1, v)
	}
	return u0 - 1, v0 - 1, u1 + 1, v1 + 1
}

// cadrer rend le point d'écran sur lequel centrer la vue : la cible, ramenée
// dans ce que le lieu laisse voir.
//
// Quand le lieu est plus étroit que le tampon, l'intervalle admissible est vide
// et il n'y a plus rien à border : le lieu se centre alors une fois pour toutes,
// et le joueur s'y déplace sans que l'image bouge. C'est la seconde moitié de ce
// que veut la conception — se bloquer aux limites plutôt que découvrir du vide,
// et centrer ce qui tient tout entier.
func cadrer(tampon int, cible, borneBasse, borneHaute float64) float64 {
	demi := float64(tampon) / 2
	mini, maxi := borneBasse+demi+margeDeCadrage, borneHaute-demi-margeDeCadrage
	if mini > maxi {
		return (borneBasse + borneHaute) / 2
	}
	return math.Min(math.Max(cible, mini), maxi)
}

// arrondi32 ramène un pixel flottant au plus proche pixel entier du moteur.
func arrondi32(pixels float64) int32 {
	// #nosec G115 -- un pixel absolu est borné par l'étendue d'un lieu, que la
	// clé de tri du moteur compte sur trente-deux bits
	return int32(math.Round(pixels))
}
