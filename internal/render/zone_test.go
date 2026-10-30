// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que la désignation à la souris garde : un rectangle borne ce qu'il
// contient, une case absente ne se clique pas, et la rangée de cases ne sort
// jamais de la bande que l'interface s'est réservée.

package render

import "testing"

// TestUneZoneExclutSesBordsOpposes garde la convention qui empêche deux cases
// voisines de se disputer la colonne qui les sépare.
func TestUneZoneExclutSesBordsOpposes(t *testing.T) {
	z := zone{x: 10, y: 20, largeur: 4, hauteur: 6}

	for _, c := range []struct {
		nom    string
		x, y   int
		dedans bool
	}{
		{"coin haut-gauche", 10, 20, true},
		{"dernier pixel", 13, 25, true},
		{"un pixel à gauche", 9, 20, false},
		{"un pixel au-dessus", 10, 19, false},
		{"le bord droit", 14, 20, false},
		{"le bord bas", 10, 26, false},
	} {
		if got := z.contient(c.x, c.y); got != c.dedans {
			t.Errorf("%s : (%d, %d) rend %v, attendu %v", c.nom, c.x, c.y, got, c.dedans)
		}
	}
}

// TestUneZoneVideNeSurvolePas éprouve ce qui fait qu'un clic ne déclenche rien
// là où la partie ne montre rien.
//
// **La valeur zéro porte l'absence**, et c'est ce qui se vérifie ici : sans ce
// refus, une case vide aurait un rectangle d'un pixel au coin de l'image, et un
// clic à cet endroit déclencherait une arme que le joueur ne tient pas.
func TestUneZoneVideNeSurvolePas(t *testing.T) {
	souris := pointeur{x: 0, y: 0, dedans: true}

	if souris.survole(zone{}) {
		t.Error("une zone de taille nulle se laisse survoler")
	}
	if souris.survole(zone{largeur: 4}) {
		t.Error("une zone sans hauteur se laisse survoler")
	}
	if !souris.survole(zone{largeur: 1, hauteur: 1}) {
		t.Error("une zone d'un pixel ne se laisse pas survoler")
	}
}

// TestLePointeurHorsFenetreNeSurvoleRien garde ce que le moteur demande de son
// second résultat.
//
// Sorti de la fenêtre, il garde la dernière position plutôt que de la remettre
// à zéro — qui désignerait un coin, donc la première case du bandeau. S'y fier
// sans lire le drapeau ferait cliquer l'aimant en sortant de l'écran.
func TestLePointeurHorsFenetreNeSurvoleRien(t *testing.T) {
	z := zone{x: 5, y: 5, largeur: 10, hauteur: 10}

	if dehors := (pointeur{x: 7, y: 7}); dehors.survole(z) {
		t.Error("un pointeur hors fenêtre survole une zone qui le contient")
	}
	if dedans := (pointeur{x: 7, y: 7, dedans: true}); !dedans.survole(z) {
		t.Error("un pointeur dans la fenêtre ne survole pas la zone qui le contient")
	}
}

// TestUneCaseVideNaPasDeRectangle éprouve la moitié que la lecture de la partie
// décide : l'aimant garde sa case, les autres n'existent que tenues.
func TestUneCaseVideNaPasDeRectangle(t *testing.T) {
	ecran, _ := monter(t)

	cases := ecran.hud.casesDuBandeau(ecran.Scene(), Readings{})
	if cases[caseAimant].vide() {
		t.Error("la case de l'aimant disparaît alors qu'elle reste toujours posée")
	}
	if !cases[caseFiole].vide() {
		t.Error("la case des fioles existe sans fiole tenue")
	}
	for place := range cases[1 : nombreDeCases-1] {
		if !cases[place+1].vide() {
			t.Errorf("l'emplacement %d existe sans arme tenue", place)
		}
	}

	var pleines Readings
	pleines.Vials.Charges = 1
	for place := range pleines.Heavies {
		pleines.Heavies[place].Charges = 1
	}
	for i, z := range ecran.hud.casesDuBandeau(ecran.Scene(), pleines) {
		if z.vide() {
			t.Errorf("la case %d manque alors que la partie la montre", i)
		}
	}
}

// TestLesCasesTiennentDansLeBandeau garde ce qu'aucune des deux formules ne dit
// toute seule.
//
// La hauteur du bandeau et l'ordonnée de la rangée se dérivent l'une de l'autre,
// mais rien n'oblige la seconde à laisser la place d'une case sous elle : une
// jauge ajoutée ou une case agrandie pousserait la rangée hors de la bande, et
// le bandeau se mettrait à mordre sur le jeu sans qu'aucun contrôle ne le dise.
func TestLesCasesTiennentDansLeBandeau(t *testing.T) {
	ecran, _ := monter(t)
	h := ecran.hud

	_, ay, _, _ := aireDans(ecran.Scene())
	haut := ay - hauteurBandeau(h)

	var pleines Readings
	pleines.Vials.Charges = 1
	for place := range pleines.Heavies {
		pleines.Heavies[place].Charges = 1
	}

	for i, z := range h.casesDuBandeau(ecran.Scene(), pleines) {
		if z.y < haut {
			t.Errorf("la case %d commence à %d, au-dessus du bandeau qui part de %d",
				i, z.y, haut)
		}
		if bas := z.y + z.hauteur; bas > ay {
			t.Errorf("la case %d descend à %d, sous l'aire de jeu qui part de %d",
				i, bas, ay)
		}
	}
}
