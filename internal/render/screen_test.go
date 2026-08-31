// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la fenêtre, sans en ouvrir une : le tampon interne garde ses
// dimensions quelle que soit la taille demandée, et le rapport au 1080p reste
// entier.

package render

import "testing"

// TestLeTamponNeSuitPasLaFenetre vérifie que la résolution interne est fixe.
//
// C'est ce qui distingue ce jeu d'un rendu qui s'adapte : la surface visible ne
// doit dépendre ni de la taille de la fenêtre ni de l'écran, sans quoi un joueur
// en plein écran verrait arriver la horde bien plus tôt qu'un autre — un
// avantage que rien n'équilibre.
func TestLeTamponNeSuitPasLaFenetre(t *testing.T) {
	var s Screen
	for _, fenetre := range [][2]int{{640, 480}, {1920, 1080}, {3440, 1440}, {1, 1}} {
		l, h := s.Layout(fenetre[0], fenetre[1])
		if l != Largeur || h != Hauteur {
			t.Errorf("fenêtre de %dx%d : tampon de %dx%d, attendu %dx%d",
				fenetre[0], fenetre[1], l, h, Largeur, Hauteur)
		}
	}
}

// TestLeTamponSAgranditEnEntier garde le chiffre sur lequel la résolution a été
// choisie.
//
// 960×540 se multiplie par deux pour du 1080p, donc pixels carrés garantis. Un
// tampon dont le rapport ne serait pas entier ferait scintiller le pixel art à
// l'agrandissement, ce que la conception interdit par ailleurs à la caméra.
func TestLeTamponSAgranditEnEntier(t *testing.T) {
	const largeur1080p, hauteur1080p = 1920, 1080
	if largeur1080p%Largeur != 0 || hauteur1080p%Hauteur != 0 {
		t.Fatalf("tampon de %dx%d : le rapport au 1080p n'est pas entier", Largeur, Hauteur)
	}
	if largeur1080p/Largeur != hauteur1080p/Hauteur {
		t.Errorf("facteurs différents en largeur et en hauteur : %d et %d",
			largeur1080p/Largeur, hauteur1080p/Hauteur)
	}
}
