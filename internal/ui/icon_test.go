// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que les icônes livrées doivent porter, et le refus d'une fenêtre sans
// icône. Le dessin, lui, se juge dans une barre des tâches.

package ui

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
)

// TestLesIconesLivreesSeChargent monte les icônes du binaire sans rien injecter.
//
// **Le manifeste livré et non un manifeste d'essai**, parce que ce qui est
// éprouvé est la chaîne entière : la liste que le générateur écrit, les fichiers
// qu'il produit, et `go:embed` qui les emporte. Un fichier renommé dans le
// générateur sans que le manifeste suive casse ici, et non au lancement du jeu.
func TestLesIconesLivreesSeChargent(t *testing.T) {
	icones, err := LoadIcons(cohue.Assets, cohue.InterfaceManifest)
	if err != nil {
		t.Fatalf("icônes livrées : %v", err)
	}
	if len(icones) == 0 {
		t.Fatal("aucune icône livrée")
	}

	// Carrées, parce qu'un gestionnaire de fenêtres qui reçoit un rectangle
	// l'étire : l'icône ne serait pas refusée, elle serait déformée.
	for _, icone := range icones {
		b := icone.Bounds()
		if b.Dx() != b.Dy() {
			t.Errorf("icône de %dx%d, une icône est carrée", b.Dx(), b.Dy())
		}
	}
}

// TestUnManifesteSansIconeEstRefuse garde le refus plutôt que la tolérance.
//
// Une tranche vide rendue sans erreur donnerait la fenêtre avec l'icône par
// défaut du système : rien ne planterait, et personne ne saurait dire à quel
// moment le jeu a cessé d'avoir la sienne. C'est le seul endroit qui puisse le
// dire, puisque c'est le seul qui sache qu'il en attendait une.
func TestUnManifesteSansIconeEstRefuse(t *testing.T) {
	sansIcone := strings.Replace(manifesteValide,
		`"fichiers": ["icone_16.png"]`, `"fichiers": []`, 1)
	if sansIcone == manifesteValide {
		t.Fatal("le manifeste d'essai ne déclare aucune icône à retirer")
	}

	_, err := LoadIcons(fstest.MapFS{
		"manifeste.json": &fstest.MapFile{Data: []byte(sansIcone)},
	}, "manifeste.json")
	if err == nil {
		t.Fatal("un manifeste sans icône est accepté")
	}
	if !strings.Contains(err.Error(), "icone") {
		t.Errorf("le refus ne nomme pas ce qui manque : %v", err)
	}
}
