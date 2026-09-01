// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les primitives de l'interface : un cadre, une jauge, un bandeau, une case
// d'emplacement. Toutes sont des rectangles unis, dessinés à partir des réglages
// du manifeste — aucune teinte ni épaisseur n'est écrite ici.

package render

import (
	"image/color"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/ui"
)

// HUD dessine l'interface : le texte et ce qui l'entoure.
//
// **Rien de ce qu'il pose n'est une image.** Une jauge dont la longueur suit la
// vie ne peut pas en être une : ce qui varie est une proportion recalculée à
// chaque image, et étirer un dessin casserait le pixel entier que tout le reste
// tient. Un cadre uni ne demande pas davantage — le découpage en neuf morceaux
// n'a d'intérêt que si le cadre porte un motif, et aucun n'a été décidé.
type HUD struct {
	// Font pose le texte. Il est exporté parce que la mesure d'une chaîne
	// décide de la taille de ce qui l'entoure, et que l'appelant compose.
	Font  *Font
	theme *ui.Theme

	// pixel est le blanc d'un pixel que tout aplat étire. Un `ebiten.Image` par
	// rectangle ferait une allocation par élément et par image.
	pixel *ebiten.Image
	op    ebiten.DrawImageOptions
}

// LoadHUD lit le manifeste d'interface : la police et les réglages.
func LoadHUD(fsys fs.FS, chemin string) (*HUD, error) {
	police, err := LoadFont(fsys, chemin)
	if err != nil {
		return nil, err
	}
	theme, err := ui.LoadTheme(fsys, chemin)
	if err != nil {
		return nil, err
	}

	pixel := ebiten.NewImage(1, 1)
	pixel.Fill(color.White)
	return &HUD{Font: police, theme: theme, pixel: pixel}, nil
}

// Margin rend la marge entre un bord et ce qu'il entoure.
func (h *HUD) Margin() int { return h.theme.Margin }

// Border rend l'épaisseur d'un bord de cadre.
func (h *HUD) Border() int { return h.theme.Border }

// Color rend une teinte du thème, pour que l'appelant compose avec les mêmes.
func (h *HUD) Color(nom string) color.RGBA { return h.theme.Color(nom) }

// Rect peint un rectangle plein.
func (h *HUD) Rect(dst *ebiten.Image, x, y, largeur, hauteur int, teinte color.RGBA) {
	if largeur <= 0 || hauteur <= 0 {
		return
	}
	h.op = ebiten.DrawImageOptions{}
	h.op.GeoM.Scale(float64(largeur), float64(hauteur))
	h.op.GeoM.Translate(float64(x), float64(y))
	h.op.ColorScale.ScaleWithColor(teinte)
	dst.DrawImage(h.pixel, &h.op)
}

// Frame peint un cadre : un fond translucide et un bord d'un pixel.
//
// Le fond laisse deviner ce qu'il couvre, et ce n'est pas de l'ornement : une
// carte de choix se pose au milieu d'une partie, et masquer la horde qui
// approche retirerait au joueur ce qu'il doit lire pendant qu'il choisit.
func (h *HUD) Frame(dst *ebiten.Image, x, y, largeur, hauteur int) {
	h.Rect(dst, x, y, largeur, hauteur, h.theme.Color("cadre_fond"))

	bord, teinte := h.theme.Border, h.theme.Color("cadre_bord")
	h.Rect(dst, x, y, largeur, bord, teinte)
	h.Rect(dst, x, y+hauteur-bord, largeur, bord, teinte)
	h.Rect(dst, x, y, bord, hauteur, teinte)
	h.Rect(dst, x+largeur-bord, y, bord, hauteur, teinte)
}

// Band peint un bandeau sur toute la largeur du tampon.
//
// C'est ce qui manquait à la ligne de titre posée à même le décor : un texte sur
// le monde a besoin de son propre fond ou d'un contour, faute de quoi il
// disparaît là où le sol est clair.
func (h *HUD) Band(dst *ebiten.Image, y, hauteur int) {
	h.Rect(dst, 0, y, Width, hauteur, h.theme.Color("bandeau_fond"))
}

// Gauge peint une jauge remplie d'une fraction, dans la teinte donnée.
//
// La fraction est bornée : une vie négative ou un dépassement de maximum sont
// des états que le jeu peut traverser un instant, et une jauge qui déborderait
// de son cadre le dirait moins bien qu'une jauge pleine.
//
// **Le remplissage est arrondi vers le bas, et jamais à zéro tant qu'il reste
// quelque chose.** Sinon une vie de deux points sur cent rendrait une jauge
// vide, indiscernable de la mort, au moment précis où le joueur doit savoir
// qu'il lui reste un souffle.
func (h *HUD) Gauge(dst *ebiten.Image, x, y, largeur int, part float64, teinte color.RGBA) {
	h.Rect(dst, x, y, largeur, h.theme.GaugeHeight, h.theme.Color("jauge_fond"))

	part = min(max(part, 0), 1)
	rempli := int(float64(largeur) * part)
	if rempli == 0 && part > 0 {
		rempli = 1
	}
	h.Rect(dst, x, y, rempli, h.theme.GaugeHeight, teinte)
}

// Slot peint une case d'emplacement et rend son côté.
//
// **Le côté n'est pas un réglage, il se calcule** : c'est la taille du contenu,
// plus la marge et le bord de chaque côté. Le déclarer au manifeste en aurait
// fait une seconde description de ce que l'icône impose, fausse au premier
// changement de taille d'icône.
//
// La touche s'écrit sous la case plutôt que le nom de l'objet : l'icône dit déjà
// de quoi il s'agit, et ce que le joueur cherche est ce qu'il doit presser.
func (h *HUD) Slot(dst *ebiten.Image, x, y, contenu int, touche string) int {
	cote := contenu + 2*(h.theme.Margin+h.theme.Border)
	h.Frame(dst, x, y, cote, cote)

	largeur := h.Font.Advance(touche)
	h.Font.Draw(dst, touche, x+(cote-largeur)/2, y+cote+2, h.theme.Color("texte_attenue"))
	return cote
}
