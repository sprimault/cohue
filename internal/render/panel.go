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

// contenuEmplacement est le côté de ce qu'une case d'emplacement contient.
//
// Il fixe la case, et non l'inverse : `Slot` en dérive son côté en ajoutant la
// marge et le bord. Douze pixels aujourd'hui, ce que fera une icône d'objet à
// l'étape 5 — et c'est alors sa taille réelle qui prendra la place de ce chiffre.
const contenuEmplacement = 12

// toucheAimant est ce que le joueur presse pour déclencher sa charge.
//
// **Les chiffres appartiennent aux emplacements**, et ils les gardent toute la
// partie. C'est ce qui a fait passer le choix des cartes aux flèches : une carte
// mal choisie se rattrape au niveau suivant, un aimant déclenché à vide est perdu
// jusqu'à la prochaine apparition, et le coût n'est pas symétrique.
const toucheAimant = "1"

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
	// Charged dit si le joueur tient un aimant.
	Charged bool
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

	h.emplacement(dst, margeEcran, y+h.Font.Height()+h.Margin(), r.Charged)
}

// emplacement pose la case de l'aimant sous les jauges.
//
// **La case est toujours là, pleine ou vide.** Un emplacement qui n'apparaîtrait
// qu'une fois chargé apprendrait au joueur l'existence de l'objet au moment où il
// le tient déjà, et une case vide est ce qui fait chercher l'aimant dans la
// salle. Ce que la charge change est ce qu'il y a dedans, pas la case.
//
// La touche s'écrit dessous et non le nom, comme la conception l'exige d'un
// emplacement : ce que le joueur cherche sous la case en jouant est ce qu'il doit
// presser, et l'icône dira de quoi il s'agit quand elle existera.
func (h *HUD) emplacement(dst *ebiten.Image, x, y int, chargee bool) {
	cote := h.Slot(dst, x, y, contenuEmplacement, toucheAimant)
	if !chargee {
		return
	}

	// Un aplat centré tient lieu d'icône, dans **la teinte de l'objet au sol** et
	// non dans une couleur du thème : sans elle, rien ne dirait que la case et ce
	// qu'on vient de ramasser sont la même chose. C'est la seule chose que le
	// bandeau emprunte au monde plutôt qu'au manifeste d'interface, et ça cessera
	// quand l'icône de l'aimant existera.
	bord := (cote - contenuEmplacement) / 2
	h.Rect(dst, x+bord, y+bord, contenuEmplacement, contenuEmplacement, teinteAimant)
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
		Charged:    s.monde.Charged(),
	})
}
