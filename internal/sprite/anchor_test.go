// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package sprite

import "testing"

// TestLAncrePoseLePiedCentre garde la conversion des deux conventions du dépôt
// vers le pied centré d'un moteur de rendu.
//
// Les trois cas sont des valeurs du catalogue, et ils sont choisis pour ce qu'ils
// séparent : un sol prend la convention additive sur une largeur paire, un banc
// la même sur une largeur **impaire** — 61 pixels, où le centrage entier tombe
// juste au lieu de pencher —, et un personnage la convention soustractive.
//
// **Les deux familles ne rendent pas la même ancre, et ce n'est pas un défaut.**
// L'appui d'un personnage est sa dernière rangée peinte, d'où un y de 1 ; un
// ancrage de décor est le sommet bas du losange de l'emprise, d'où les 32 pixels
// d'un sol. Attendre une valeur commune reviendrait à croire que les deux
// ancrages désignent la même chose.
func TestLAncrePoseLePiedCentre(t *testing.T) {
	cas := []struct {
		nom      string
		taille   [2]int
		decalage [2]int // la convention additive ; vide pour un appui
		appui    [2]int // la convention soustractive ; vide pour un décalage
		attend   [2]int
	}{
		{
			nom:      "sol, largeur paire",
			taille:   [2]int{64, 32},
			decalage: [2]int{-32, 0},
			attend:   [2]int{0, 32},
		},
		{
			nom:      "banc, largeur impaire",
			taille:   [2]int{61, 42},
			decalage: [2]int{-15, -3},
			attend:   [2]int{15, 39},
		},
		{
			nom:    "personnage, par son appui",
			taille: [2]int{64, 64},
			appui:  [2]int{32, 63},
			attend: [2]int{0, 1},
		},
	}

	for _, c := range cas {
		var x, y int
		if c.appui == [2]int{} {
			x, y = AnchorFromOffset(c.decalage, c.taille)
		} else {
			x, y = AnchorFromFooting(c.appui, c.taille)
		}
		if got := [2]int{x, y}; got != c.attend {
			t.Errorf("%s : ancre %v, attendue %v", c.nom, got, c.attend)
		}
	}
}

// TestLesDeuxConventionsSOpposent garde que l'une est l'autre au signe près.
//
// C'est la seule propriété qui les relie, et c'est elle qu'un report de valeur
// d'une famille à l'autre casserait sans qu'aucune image ne paraisse absurde :
// une forme décalée du double de son ancrage reste une forme posée quelque part.
func TestLesDeuxConventionsSOpposent(t *testing.T) {
	taille := [2]int{64, 64}
	appui := [2]int{32, 63}
	oppose := [2]int{-appui[0], -appui[1]}

	ax, ay := AnchorFromFooting(appui, taille)
	ox, oy := AnchorFromOffset(oppose, taille)
	if ax != ox || ay != oy {
		t.Errorf("appui %v rend (%d, %d), decalage %v rend (%d, %d)",
			appui, ax, ay, oppose, ox, oy)
	}

	// Et la même valeur prise dans la mauvaise convention s'en écarte du double
	// de l'appui : c'est ce que le lecteur doit pouvoir constater plutôt que
	// croire.
	mx, my := AnchorFromOffset(appui, taille)
	if ecart := [2]int{mx - ax, my - ay}; ecart != [2]int{2 * appui[0], 2 * appui[1]} {
		t.Errorf("ecart entre les deux conventions %v, attendu %v",
			ecart, [2]int{2 * appui[0], 2 * appui[1]})
	}
}

// TestLAncreDUnDecalageNulCentreLePied garde ce que la valeur zéro désigne.
//
// Un décalage nul veut dire « le coin de l'image est au point visé », ce qui
// n'est pas le pied centré : l'ancre vaut alors la demi-largeur et la hauteur.
// C'est le cas qu'on croit neutre et qui ne l'est pas, et il arriverait à qui
// composerait un aspect sans passer par ici.
func TestLAncreDUnDecalageNulCentreLePied(t *testing.T) {
	x, y := AnchorFromOffset([2]int{}, [2]int{64, 32})
	if x != 32 || y != 32 {
		t.Errorf("ancre d'un decalage nul (%d, %d), attendue (32, 32)", x, y)
	}
}
