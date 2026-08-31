// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le tampon interne et ce qu'on y peint : le sol tel que la simulation le voit,
// ce qui s'y tient, et les touches qui déplacent le joueur. C'est ici que le
// repère de l'écran rencontre celui du monde ; l'ordre de dessin est à côté.

// Package render dessine ce que la simulation a calculé, et ne décide de rien.
//
// La frontière avec `internal/game` va dans un seul sens : ce paquet lit l'état
// du monde, et rien de ce qu'il produit n'y revient. C'est ce qui lui permet de
// calculer en flottants — interpolations, lissage de caméra — sans menacer le
// déterminisme de la run, et ce qui interdit à une notion d'écran de redescendre
// dans la simulation.
//
// **Ce paquet n'a pas de fichier de test, et n'en aura pas.** Importer
// Ebitengine initialise GLFW, qui panique sans `DISPLAY` : sur un runner sans
// écran, n'importe quel test du paquet tombe avant d'avoir commencé, y compris
// un test qui n'ouvrirait aucune fenêtre. La suite par défaut reste donc
// exécutable partout, et le rendu se juge à l'œil — c'est ce que la doctrine de
// test énonçait déjà, et l'absence de tests ici en est la conséquence et non un
// oubli. Ce qui doit être vérifié mécaniquement vit du côté simulation, qui
// n'importe pas Ebitengine.
package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/game"
)

// Les dimensions du tampon interne, agrandi en entier vers la fenêtre.
//
// Un tampon de 480×270 ne montrerait que sept tuiles de large, bien trop serré
// pour voir la horde arriver ; 960×540 en donne une quinzaine et se multiplie
// par deux pour du 1080p, donc pixels carrés garantis.
const (
	Largeur = 960
	Hauteur = 540
)

// Les teintes du rendu provisoire, qui tiendront jusqu'à ce que l'atlas entre.
//
// Elles ne cherchent pas à ressembler à un lieu : ce sont trois états de la
// grille de coûts, plus le joueur, choisis pour se distinguer et pour rien
// d'autre. La palette fermée du jeu vaut pour les images du décor, qui sortent
// des générateurs, et non pour ces aplats qui disparaîtront avec eux.
var (
	fond      = color.RGBA{R: 24, G: 24, B: 28, A: 255}
	solBloque = color.RGBA{R: 58, G: 58, B: 66, A: 255}
	solLibre  = color.RGBA{R: 96, G: 98, B: 104, A: 255}
	solLent   = color.RGBA{R: 74, G: 96, B: 120, A: 255}

	// Le joueur en clair et les créatures en sombre : le chapitre de la
	// lisibilité veut que le personnage reste distinguable à cent ennemis à
	// l'écran, ce qui se joue d'abord sur la valeur et non sur la teinte.
	teinteJoueur = color.RGBA{R: 236, G: 214, B: 120, A: 255}
	teinteHorde  = color.RGBA{R: 150, G: 78, B: 74, A: 255}
	teinteTir    = color.RGBA{R: 226, G: 232, B: 238, A: 255}
)

// Screen est le jeu tel qu'Ebitengine le voit.
type Screen struct {
	monde *game.World
	carte *game.CostGrid
	cam   *camera

	scene *scene

	// Les trois formes blanches que le dessin teinte au blit : la face d'une
	// case, la silhouette d'un personnage, et le point d'un projectile.
	face     *ebiten.Image
	figurine *ebiten.Image
	eclat    *ebiten.Image
	// demiTuile est l'abscisse du sommet dans l'image d'une face, ce que le
	// manifeste appellera son ancrage quand les images viendront de lui.
	demiTuile int

	// op est réutilisée d'un blit à l'autre, et remise à zéro à chaque fois :
	// une case visible en produit un millier par image.
	op ebiten.DrawImageOptions
}

// NewScreen monte le rendu sur une partie et le lieu qu'elle joue.
//
// La taille de tuile est celle du manifeste de décor et jamais une constante :
// le manifeste la porte, le chargeur en exige le rapport de deux pour un, et
// c'est de lui que la projection la tient.
func NewScreen(monde *game.World, carte *game.CostGrid, tuile [2]int) *Screen {
	s := &Screen{
		monde:     monde,
		carte:     carte,
		cam:       nouvelleCamera(tuile, carte),
		scene:     nouvelleScene(carte, monde.Enemies().Cap(), monde.Shots().Cap()),
		face:      face(tuile),
		figurine:  aplat(tuile[0]/4, tuile[0]*3/4),
		eclat:     aplat(tuile[1]/8, tuile[1]/8),
		demiTuile: tuile[0] / 2,
	}
	s.cam.suivre(monde.Player())
	return s
}

// Update avance la simulation d'un pas, puis recadre.
//
// Un pas par appel et rien qui lise l'horloge : Ebitengine appelle cette méthode
// à cadence fixe et rattrape un retard en l'appelant plusieurs fois d'affilée,
// ce qui est exactement ce que la simulation attend d'un appelant.
func (s *Screen) Update() error {
	s.monde.Step(voulu())
	s.cam.suivre(s.monde.Player())
	return nil
}

// Draw peint le tampon interne : le sol, puis ce qui s'y tient, en profondeur.
func (s *Screen) Draw(ecran *ebiten.Image) {
	ecran.Fill(fond)
	s.peindreSol(ecran)
	s.peindreEntites(ecran)
}

// Layout fixe la taille du tampon interne, quelle que soit celle de la fenêtre.
//
// Ebitengine agrandit ensuite vers la fenêtre, et le facteur n'est pas
// nécessairement entier : une fenêtre large de 1400 pixels montre le tampon
// agrandi de 1,45 fois, donc des pixels de tailles inégales. Ce qui reste à
// trancher n'appartient ni à la projection ni à la caméra, qui travaillent
// toutes deux dans le tampon : c'est le facteur d'échelle du système qui décide,
// il se lit par `LayoutF`, et aucun des deux ne le connaît.
func (s *Screen) Layout(largeurFenetre, hauteurFenetre int) (int, int) {
	return Largeur, Hauteur
}

// peindreSol pose la face de chaque case visible, teintée par son coût.
//
// Ce que ce sol montre n'est pas le décor mais la grille de coûts, c'est-à-dire
// ce que le champ de flux lit : franchissable, coûteux, ou mur. Tant qu'aucun
// atlas n'est chargé, c'est l'information la plus utile qu'une case puisse
// porter — un lieu se relit alors comme la simulation le voit, et un écart entre
// les deux se verrait ici avant de se deviner ailleurs.
func (s *Screen) peindreSol(ecran *ebiten.Image) {
	u0, v0, u1, v1 := s.cam.casesVisibles()
	for v := v0; v <= v1; v++ {
		for u := u0; u <= u1; u++ {
			if !s.carte.InBounds(u, v) {
				continue
			}
			x, y := s.cam.ecran(game.FromInt(u), game.FromInt(v))
			s.op.GeoM.Reset()
			s.op.GeoM.Translate(float64(x-s.demiTuile), float64(y))
			s.op.ColorScale.Reset()
			s.op.ColorScale.ScaleWithColor(teinte(s.carte.At(u, v)))
			ecran.DrawImage(s.face, &s.op)
		}
	}
}

// peindreEntites pose ce qui se tient sur le sol, du plus lointain au plus
// proche.
//
// La position est relue dans le bassin plutôt que portée par la séquence : le
// tri range des rangs, pas des coordonnées, et une copie faite au moment du tri
// aurait une image de retard le jour où quelque chose bougera entre les deux.
func (s *Screen) peindreEntites(ecran *ebiten.Image) {
	for _, e := range s.scene.ranger(s.monde) {
		switch e.sorte {
		case sorteEnnemi:
			c := s.monde.Enemies().At(e.place)
			s.silhouette(ecran, s.figurine, c.X, c.Y, teinteHorde)
		case sorteTir:
			p := s.monde.Shots().At(e.place)
			s.silhouette(ecran, s.eclat, p.X, p.Y, teinteTir)
		case sorteJoueur:
			x, y := s.monde.Player()
			s.silhouette(ecran, s.figurine, x, y, teinteJoueur)
		}
	}
}

// silhouette pose une forme teintée, son pied sur le point où le monde la situe.
//
// L'appui est au milieu du bas, ce que sera l'ancrage d'un sprite de personnage
// quand le manifeste en fournira : c'est le point qui touche le sol, et le seul
// qui puisse coïncider avec une position du monde.
func (s *Screen) silhouette(ecran, forme *ebiten.Image, x, y game.Fixed, teinte color.RGBA) {
	ex, ey := s.cam.ecran(x, y)
	taille := forme.Bounds()
	s.op.GeoM.Reset()
	s.op.GeoM.Translate(float64(ex-taille.Dx()/2), float64(ey-taille.Dy()))
	s.op.ColorScale.Reset()
	s.op.ColorScale.ScaleWithColor(teinte)
	ecran.DrawImage(forme, &s.op)
}

// teinte dit de quelle couleur une case se peint, selon ce qu'elle coûte.
func teinte(cout game.Cost) color.RGBA {
	switch {
	case cout == game.Blocked:
		return solBloque
	case cout > game.Free:
		return solLent
	}
	return solLibre
}

// voulu lit les touches et rend la direction demandée, dans le repère du monde.
//
// Les touches sont en repère d'écran et le monde ne connaît que ses deux axes
// obliques : « haut » vaut donc (-1, -1), et « gauche » (-1, +1). C'est la
// conversion dont ce paquet a la charge, prise par son autre bout, et c'est
// pourquoi elle vit ici plutôt que dans la boucle de jeu.
//
// La somme des touches enfoncées donne les huit directions sans table de
// combinaisons — haut et gauche font (-2, 0), que la simulation ramène à
// l'unité. Elle ne normalise pas non plus : c'est `Vec.Direction` qui le fait,
// avec l'arrondi que le déterminisme exige, et le refaire ici en flottants
// donnerait deux réponses pour une question.
//
// `ebiten.Key` désigne une touche par sa **place** sur un clavier américain :
// `KeyW` est la touche marquée Z sur un clavier français, et le carré ZQSD tombe
// donc juste sans qu'on ait à le nommer.
func voulu() game.Vec {
	var v game.Vec
	if enfonce(ebiten.KeyW, ebiten.KeyArrowUp) {
		v = v.Add(game.Vec{X: -game.One, Y: -game.One})
	}
	if enfonce(ebiten.KeyS, ebiten.KeyArrowDown) {
		v = v.Add(game.Vec{X: game.One, Y: game.One})
	}
	if enfonce(ebiten.KeyA, ebiten.KeyArrowLeft) {
		v = v.Add(game.Vec{X: -game.One, Y: game.One})
	}
	if enfonce(ebiten.KeyD, ebiten.KeyArrowRight) {
		v = v.Add(game.Vec{X: game.One, Y: -game.One})
	}
	return v
}

// enfonce dit si l'une des touches est pressée.
func enfonce(touches ...ebiten.Key) bool {
	for _, t := range touches {
		if ebiten.IsKeyPressed(t) {
			return true
		}
	}
	return false
}

// face peint la face supérieure d'une case, en blanc, pour être teintée au blit.
//
// La forme reprend le test de `outils/primitives_iso.py` — bords sur des droites
// de pente 1/2, avec la même marge d'un demi-pixel — pour que les faces se
// jointent comme le feront les images du décor. Ce n'est pas une seconde
// description qu'il faudrait tenir d'accord avec lui : les deux ne s'affichent
// jamais ensemble, et celle-ci s'en va quand l'atlas entre.
func face(tuile [2]int) *ebiten.Image {
	largeur, hauteur := tuile[0], tuile[1]
	demi := float64(largeur) / 2
	marge := 0.5 / float64(largeur)

	pixels := make([]byte, 4*largeur*hauteur)
	for y := range hauteur {
		for x := range largeur {
			px, py := float64(x)+0.5-demi, float64(y)+0.5
			u := px/float64(largeur) + py/demi
			v := -px/float64(largeur) + py/demi
			if u < -marge || u > 1+marge || v < -marge || v > 1+marge {
				continue
			}
			i := 4 * (y*largeur + x)
			pixels[i], pixels[i+1], pixels[i+2], pixels[i+3] = 255, 255, 255, 255
		}
	}

	img := ebiten.NewImage(largeur, hauteur)
	img.WritePixels(pixels)
	return img
}

// aplat rend un rectangle blanc plein, à teinter au blit.
//
// Un quart de tuile de large et trois quarts de haut pour la silhouette du
// joueur : un personnage tient dans une image de la largeur d'une tuile et s'y
// dresse presque entier, si bien que ces proportions donnent l'échelle sans
// prétendre au sprite. Elles se dérivent de la tuile pour ne pas mentir si elle
// change.
func aplat(largeur, hauteur int) *ebiten.Image {
	img := ebiten.NewImage(largeur, hauteur)
	img.Fill(color.White)
	return img
}
