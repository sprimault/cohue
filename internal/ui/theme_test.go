// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que les réglages d'apparence doivent refuser, et ce que le thème livré doit
// porter.

package ui

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
)

// reglagesValides est le manifeste minimal dont chaque cas d'échec part.
//
// La police y est réduite à deux glyphes : ce qui est éprouvé ici est la section
// des réglages, et un manifeste complet y ajouterait du bruit sans rien garder.
const reglagesValides = `{
  "version_format": 1,
  "interface": {
    "police": {
      "fichier": "police.png",
      "source": "PixelOperator8.ttf",
      "cellule": [11, 9],
      "ligne_de_base": 8,
      "taille_native": 8,
      "glyphes": "AB",
      "avances": [7, 7]
    },
    "reglages": {
      "bord_px": 1,
      "marge_px": 4,
      "hauteur_jauge_px": 6,
      "teintes": {
        "cadre_fond": [26, 28, 34, 235],
        "cadre_bord": [92, 96, 106, 255],
        "cadre_choisi": [236, 196, 96, 255],
        "bandeau_fond": [16, 17, 21, 200],
        "jauge_fond": [40, 42, 48, 255],
        "jauge_vie": [176, 62, 58, 255],
        "vignette_danger": [28, 10, 9, 40],
        "jauge_experience": [86, 132, 186, 255],
        "texte": [232, 234, 238, 255],
        "texte_attenue": [150, 154, 162, 255],
        "texte_valeur": [236, 196, 96, 255],
        "texte_contour": [0, 0, 0, 255]
      }
    }
  }
}`

// TestLesReglagesRefusentCeQuiEstIncoherent éprouve les refus un par un.
//
// Le cas de la teinte inconnue est celui qui compte le plus : sans lui, une clé
// mal orthographiée s'ajouterait au manifeste sans que rien ne la lise, et la
// couleur qu'on croirait avoir réglée resterait celle d'avant. C'est le contrôle
// qui certifie l'écart au lieu de le signaler, pris à sa source.
func TestLesReglagesRefusentCeQuiEstIncoherent(t *testing.T) {
	cas := []struct {
		nom     string
		avant   string
		apres   string
		attendu string
	}{
		{"bord nul", `"bord_px": 1`, `"bord_px": 0`, "reglages.bord_px : 0"},
		{"jauge sans épaisseur", `"hauteur_jauge_px": 6`, `"hauteur_jauge_px": 0`,
			"reglages.hauteur_jauge_px : 0"},
		{"teinte absente", `"texte_valeur": [236, 196, 96, 255],`, ``,
			"reglages.teintes.texte_valeur : absente"},
		{"teinte inconnue", `"texte_contour": [0, 0, 0, 255]`,
			`"texte_contour": [0, 0, 0, 255], "texte_gras": [1, 2, 3, 4]`,
			"reglages.teintes.texte_gras : inconnue"},
		// La vignette est la seule teinte dont l'alpha soit assez bas pour qu'une
		// écriture à plat se voie : sur une teinte opaque, aucune composante ne
		// peut dépasser 255, et le cas ne se poserait pas.
		{"teinte non prémultipliée", `"vignette_danger": [28, 10, 9, 40]`,
			`"vignette_danger": [176, 62, 58, 40]`,
			"reglages.teintes.vignette_danger : rouge à 176 au-dessus de l'alpha 40"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			texte := strings.Replace(reglagesValides, c.avant, c.apres, 1)
			if texte == reglagesValides {
				t.Fatalf("le cas ne change rien : %q est absent du manifeste", c.avant)
			}
			fsys := fstest.MapFS{"m.json": {Data: []byte(texte)}}
			if _, err := LoadTheme(fsys, "m.json"); err == nil {
				t.Fatalf("réglages acceptés alors qu'ils portent : %s", c.nom)
			} else if !strings.Contains(err.Error(), c.attendu) {
				t.Errorf("message %q, attendu contenant %q", err, c.attendu)
			}
		})
	}
}

// TestLesReglagesValidesSeChargent est le témoin des refus ci-dessus.
//
// Sans lui, un `LoadTheme` qui refuserait tout les ferait tous passer.
func TestLesReglagesValidesSeChargent(t *testing.T) {
	fsys := fstest.MapFS{"m.json": {Data: []byte(reglagesValides)}}
	theme, err := LoadTheme(fsys, "m.json")
	if err != nil {
		t.Fatalf("réglages valides refusés : %v", err)
	}
	if theme.Border != 1 || theme.Margin != 4 || theme.GaugeHeight != 6 {
		t.Errorf("bord %d, marge %d, jauge %d",
			theme.Border, theme.Margin, theme.GaugeHeight)
	}
	if c := theme.Color("jauge_vie"); c.R != 176 {
		t.Errorf("teinte de la vie : %v", c)
	}
}

// TestLeThemeLivreSeCharge lit le manifeste publié, sans rien injecter.
//
// Les cas ci-dessus bâtissent leur entrée pour isoler un refus ; celui-ci passe
// par le chemin qui produit la vraie, et c'est le seul qui tombe si le
// générateur cesse d'écrire une teinte que la table exige.
func TestLeThemeLivreSeCharge(t *testing.T) {
	theme, err := LoadTheme(cohue.Assets, cohue.InterfaceManifest)
	if err != nil {
		t.Fatalf("chargement du thème livré : %v", err)
	}
	for _, nom := range teintesRequises {
		if c := theme.Color(nom); c.A == 0 {
			t.Errorf("la teinte « %s » est entièrement transparente", nom)
		}
	}
}
