// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les primitives de l'interface : un cadre, une jauge, un bandeau, une case
// d'emplacement. Toutes sont des rectangles unis, posés à partir des réglages du
// manifeste — aucune teinte ni épaisseur n'est écrite ici.

package render

import (
	"io/fs"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"
)

// HUD pose l'interface : le texte et ce qui l'entoure.
//
// **Rien de ce qu'il pose n'est une image.** Une jauge dont la longueur suit la
// vie ne peut pas en être une : ce qui varie est une proportion recalculée à
// chaque image, et étirer un dessin casserait le pixel entier que tout le reste
// tient. Un cadre uni ne demande pas davantage — le découpage en neuf morceaux
// n'a d'intérêt que si le cadre porte un motif, et aucun n'a été décidé.
//
// **Il ne tient plus ni pixel blanc ni options de dessin.** Un aplat du moteur
// peint sa teinte sans lire de planche, donc le blanc d'un pixel qu'on étirait
// n'a plus d'objet — et avec lui disparaît `Image`, qui n'existait que parce que
// `Rect` laissait sa teinte dans des options partagées.
type HUD struct {
	// Font pose le texte. Il est exporté parce que la mesure d'une chaîne
	// décide de la taille de ce qui l'entoure, et que l'appelant compose.
	Font  *Font
	theme *Theme
}

// LoadHUD lit le manifeste d'interface : la police et les réglages.
func LoadHUD(fsys fs.FS, chemin string, atlas *raster.Atlas) (*HUD, error) {
	police, err := LoadFont(fsys, chemin, atlas)
	if err != nil {
		return nil, err
	}
	theme, err := LoadTheme(fsys, chemin)
	if err != nil {
		return nil, err
	}
	return &HUD{Font: police, theme: theme}, nil
}

// Margin rend la marge entre un bord et ce qu'il entoure.
func (h *HUD) Margin() int { return h.theme.Margin }

// Border rend l'épaisseur d'un bord de cadre.
func (h *HUD) Border() int { return h.theme.Border }

// Color rend une teinte du thème, pour que l'appelant compose avec les mêmes.
func (h *HUD) Color(nom string) render.Color { return h.theme.Color(nom) }

// Rect pose un rectangle plein.
//
// Une taille nulle ne pose rien : l'aplat rend alors la valeur zéro, que la
// scène reconnaît comme ce qu'elle n'a pas à dessiner.
func (h *HUD) Rect(sc *scene.Scene, x, y, largeur, hauteur int, teinte render.Color) {
	sc.PlaceScreen(pixelDEcran(x), pixelDEcran(y), aplat(largeur, hauteur, teinte))
}

// Sprite pose un dessin de l'interface à son coin, tel qu'il est dessiné.
//
// **Elle retire l'ancre, et c'est ce qui la fait tenir dans sa case.** Tout
// aspect que `sprite` fabrique en porte une, parce qu'elle convertit une pose de
// monde — où le moteur centre le pied de l'image sur le point visé — en coin
// haut-gauche. `Scene.PlaceScreen` vise déjà ce coin et ajoute l'ancre
// par-dessus : une icône de vingt pixels s'en trouvait poussée de dix vers la
// droite et de vingt vers le bas, soit hors de l'emplacement qu'elle garnit.
//
// C'est le même piège que la police portait, et il vient du même endroit : un
// aspect ne dit pas par quelle voie il sera posé, donc c'est la voie qui doit
// savoir ce qu'elle attend.
func (h *HUD) Sprite(sc *scene.Scene, a scene.Aspect, x, y int) {
	if a.Width == 0 || a.Height == 0 {
		return
	}
	a.Tint = render.White()
	a.AnchorX, a.AnchorY = 0, 0
	a.Layer = coucheInterface
	sc.PlaceScreen(pixelDEcran(x), pixelDEcran(y), a)
}

// Frame pose un cadre : un fond translucide et un bord d'un pixel.
//
// Le fond laisse deviner ce qu'il couvre, et ce n'est pas de l'ornement : une
// carte de choix se pose au milieu d'une partie, et masquer la horde qui
// approche retirerait au joueur ce qu'il doit lire pendant qu'il choisit.
func (h *HUD) Frame(sc *scene.Scene, x, y, largeur, hauteur int) {
	h.Rect(sc, x, y, largeur, hauteur, h.theme.Color("cadre_fond"))
	h.Outline(sc, x, y, largeur, hauteur, h.theme.Color("cadre_bord"))
}

// Outline repose le bord d'un cadre dans une autre teinte.
//
// Elle existe pour la désignation d'une carte : c'est le bord qui la porte, un
// fond assombri masquant la horde que le joueur doit continuer de lire.
func (h *HUD) Outline(sc *scene.Scene, x, y, largeur, hauteur int, teinte render.Color) {
	bord := h.theme.Border
	h.Rect(sc, x, y, largeur, bord, teinte)
	h.Rect(sc, x, y+hauteur-bord, largeur, bord, teinte)
	h.Rect(sc, x, y, bord, hauteur, teinte)
	h.Rect(sc, x+largeur-bord, y, bord, hauteur, teinte)
}

// Band pose un bandeau du fond du thème, sur le rectangle donné.
//
// C'est ce qui manquait à la ligne de titre posée à même le décor : un texte sur
// le monde a besoin de son propre fond ou d'un contour, faute de quoi il
// disparaît là où le sol est clair.
//
// **Son rectangle lui est donné et ne se devine plus.** Elle couvrait la largeur
// du tampon, ce qui était la même chose tant que le tampon valait l'aire de jeu ;
// depuis que la surface suit la fenêtre, un bandeau large comme elle déborderait
// du bloc que l'aire et l'interface forment ensemble.
func (h *HUD) Band(sc *scene.Scene, x, y, largeur, hauteur int) {
	h.Rect(sc, x, y, largeur, hauteur, h.theme.Color("bandeau_fond"))
}

// Gauge pose une jauge remplie d'une fraction, dans la teinte donnée.
//
// La fraction est bornée : une vie négative ou un dépassement de maximum sont
// des états que le jeu peut traverser un instant, et une jauge qui déborderait
// de son cadre le dirait moins bien qu'une jauge pleine.
//
// **Le remplissage est arrondi vers le bas, et jamais à zéro tant qu'il reste
// quelque chose.** Sinon une vie de deux points sur cent rendrait une jauge
// vide, indiscernable de la mort, au moment précis où le joueur doit savoir
// qu'il lui reste un souffle.
func (h *HUD) Gauge(sc *scene.Scene, x, y, largeur int, part float64, teinte render.Color) {
	h.Rect(sc, x, y, largeur, h.theme.GaugeHeight, h.theme.Color("jauge_fond"))

	part = min(max(part, 0), 1)
	rempli := int(float64(largeur) * part)
	if rempli == 0 && part > 0 {
		rempli = 1
	}
	h.Rect(sc, x, y, rempli, h.theme.GaugeHeight, teinte)
}

// SlotSide rend le côté qu'une case aurait, sans rien poser.
//
// **Séparée de `Slot` parce que placer précède peindre** : poser deux cases côte
// à côte demande de connaître la largeur de la première avant de poser la
// seconde, et appeler `Slot` pour l'apprendre la poserait au mauvais endroit.
func (h *HUD) SlotSide(contenu int) int {
	return contenu + 2*(h.theme.Margin+h.theme.Border)
}

// Slot pose une case d'emplacement et rend son côté.
//
// **Le côté n'est pas un réglage, il se calcule** : c'est la taille du contenu,
// plus la marge et le bord de chaque côté. Le déclarer au manifeste en aurait
// fait une seconde description de ce que l'icône impose, fausse au premier
// changement de taille d'icône.
//
// La touche s'écrit sous la case plutôt que le nom de l'objet : l'icône dit déjà
// de quoi il s'agit, et ce que le joueur cherche est ce qu'il doit presser.
func (h *HUD) Slot(sc *scene.Scene, x, y, contenu int, touche string) int {
	cote := h.SlotSide(contenu)
	h.Frame(sc, x, y, cote, cote)

	largeur := h.Font.Advance(touche)
	h.Font.Draw(sc, touche, x+(cote-largeur)/2, y+cote+2, h.theme.Color("texte_attenue"))
	return cote
}
