// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le bandeau de la partie : la vie, l'expérience et le temps écoulé, composés à
// partir des primitives. Aucune dimension d'élément n'est écrite ici — une
// jauge suit son contenu, un libellé se pose à la mesure du texte.

package render

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/game"
)

// margeEcran est la distance entre un bord du tampon et ce qui s'y accroche.
//
// La marge du thème sépare un contenu de son cadre ; au bord de l'écran il n'y a
// pas de cadre, et la reprendre telle quelle collerait la jauge à l'arête. Douze
// pixels sont un retrait de mise en page, la seule valeur de ce fichier qui ne
// se dérive de rien.
const margeEcran = 12

// largeurJauge est la longueur d'une jauge du bandeau, en pixels.
//
// Plus longue que la vie n'a de points, et c'est ce qui la fixe : le
// remplissage est arrondi vers le bas, si bien qu'une jauge de quatre-vingts
// pixels pour cent points de vie laisserait un point sur cinq ne rien déplacer.
// Cent quarante-huit en donnent une et demie par point.
const largeurJauge = 148

// Readings est ce que le bandeau montre d'une partie.
//
// Des nombres et non le monde : la planche de relecture compose le bandeau sans
// partie derrière, et une signature qui exigerait un `game.World` l'aurait
// obligée à en monter une pour juger une mise en page.
type Readings struct {
	// Health et MaxHealth sont la vie restante et celle du profil.
	Health, MaxHealth int
	// Level est le niveau atteint, le premier valant un.
	Level int
	// Experience et Threshold sont les gemmes acquises vers le niveau suivant,
	// et ce qu'il coûte.
	Experience, Threshold int
	// Elapsed est l'âge de la partie.
	Elapsed game.Tick
}

// Panel pose les trois lectures : la vie, l'expérience et le temps écoulé.
//
// **Les libellés sont posés à plat, sans contour.** Le contour est fait pour un
// chiffre de dégâts isolé, à qui il ajoute une épaisseur ; sur du texte à cette
// taille il ferme les contre-formes — un zéro devient un carré plein, « 62 / 100 »
// se lit « 620/1000 ». La relecture le montre en une image.
//
// Ce que le contour aurait couvert reste donc ouvert : un libellé clair sur un
// sol clair. Le rendu ne peint aujourd'hui que trois gris et un bleu, si bien
// que le cas ne peut pas être jugé avant les sprites — c'est la même échéance
// que celle déjà notée sur la vue du texte. Les jauges, elles, portent leur
// propre fond et ne dépendent pas de ce qu'elles couvrent.
func (h *HUD) Panel(dst *ebiten.Image, r Readings) {
	x, y := margeEcran, margeEcran
	ecart := 2 * h.Margin()

	h.Gauge(dst, x, y, largeurJauge, part(r.Health, r.MaxHealth), h.Color("jauge_vie"))
	h.libelle(dst, fmt.Sprintf("%d / %d", r.Health, r.MaxHealth),
		x+largeurJauge+ecart, y, h.Color("texte"))

	y += h.Font.Height()
	h.Gauge(dst, x, y, largeurJauge, part(r.Experience, r.Threshold),
		h.Color("jauge_experience"))
	h.libelle(dst, fmt.Sprintf("Niveau %d", r.Level),
		x+largeurJauge+ecart, y, h.Color("texte_attenue"))

	// Le minuteur s'aligne sur le bord droit par mesure, et non à distance fixe :
	// il s'allonge d'un caractère au passage de la dixième minute.
	temps := minuteur(r.Elapsed)
	h.libelle(dst, temps, Width-margeEcran-h.Font.Advance(temps), margeEcran,
		h.Color("texte"))
}

// libelle pose un texte du bandeau, aligné sur la jauge qu'il commente.
//
// Le pixel de décalage vers le haut n'est pas un ajustement à l'œil : la jauge
// est haute de six pixels et la ligne de texte de neuf, et les centrer l'une sur
// l'autre demande la moitié de leur différence.
func (h *HUD) libelle(dst *ebiten.Image, texte string, x, y int, teinte color.RGBA) {
	h.Font.Draw(dst, texte, x, y-(h.Font.Height()-h.theme.GaugeHeight)/2, teinte)
}

// part rend la fraction d'une jauge, zéro quand son maximum n'en est pas un.
//
// Le garde-fou n'est pas de la défensive : sans lui, un maximum nul produirait
// un NaN, et le bornage de `Gauge` ne l'arrête pas — toute comparaison avec un
// NaN étant fausse, la jauge afficherait n'importe quelle longueur au lieu de
// rien.
func part(valeur, maximum int) float64 {
	if maximum <= 0 {
		return 0
	}
	return float64(valeur) / float64(maximum)
}

// minuteur rend l'âge d'une partie en minutes et secondes.
//
// **C'est le cas où `Tick.Seconds` est légitime.** Sa godoc interdit de décider
// avec — un seuil, une cadence, un plancher se comptent en ticks —, pas de
// montrer avec. Sans cette phrase, l'avertissement se relit comme une
// interdiction générale, et quelqu'un réinventerait une conversion à côté.
func minuteur(ecoule game.Tick) string {
	secondes := int(ecoule.Seconds())
	return fmt.Sprintf("%02d:%02d", secondes/60, secondes%60)
}

// peindreBandeau pose le bandeau de la partie en cours.
//
// Après les entités et avant le voile de mort : par-dessus le monde, parce
// qu'une créature qui passerait devant la jauge de vie la rendrait illisible au
// pire moment ; sous le voile, parce que le niveau atteint et le temps tenu font
// partie de ce que le joueur relit en mourant.
func (s *Screen) peindreBandeau(ecran *ebiten.Image) {
	if s.hud == nil {
		return
	}
	s.hud.Panel(ecran, Readings{
		Health:     s.monde.Health(),
		MaxHealth:  s.monde.MaxHealth(),
		Level:      s.monde.Level(),
		Experience: s.monde.Experience(),
		Threshold:  s.monde.Threshold(),
		Elapsed:    s.monde.Tick(),
	})
}
