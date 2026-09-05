// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le manifeste d'interface doit refuser, et ce que la police livrée doit
// porter. C'est le versant vérifiable de la police : le dessin, lui, se juge à
// l'œil sur la planche.

package ui

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
)

// manifesteValide est le manifeste minimal dont chaque cas d'échec part.
//
// Deux glyphes suffisent : ce qui est éprouvé ici est la cohérence des
// déclarations entre elles, pas la fonte.
const manifesteValide = `{
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
    "icone": {
      "fichiers": ["icone_16.png"]
    }
  }
}`

// TestLeManifesteDInterfaceRefuseCeQuiEstIncoherent éprouve les refus un par un.
//
// Chaque cas retire une cohérence et une seule, en partant du manifeste valide :
// c'est ce qui permet d'affirmer que le refus vient de ce qu'on a changé. Un
// manifeste faux sur deux points serait refusé pour le premier, et le second ne
// serait jamais atteint.
//
// **Ce que l'attendu porte est le chemin de la clé fautive, et c'est délibéré.**
// Ces messages ont d'abord nommé le groupe sans la feuille — « planche non
// nommée » pour `police.fichier`, un mot qui n'existe nulle part dans le
// manifeste : l'auteur savait qu'une chose manquait et pas laquelle. Un attendu
// qui ne porterait que l'explication laisserait ce défaut revenir.
func TestLeManifesteDInterfaceRefuseCeQuiEstIncoherent(t *testing.T) {
	cas := []struct {
		nom     string
		avant   string
		apres   string
		attendu string
	}{
		{"format inconnu", `"version_format": 1`, `"version_format": 2`,
			"version de format"},
		{"planche non nommée", `"fichier": "police.png"`, `"fichier": ""`,
			"police.fichier : absent ou vide"},
		{"cellule nulle", `"cellule": [11, 9]`, `"cellule": [0, 9]`,
			"police.cellule : 0x9"},
		{"table vide", `"glyphes": "AB"`, `"glyphes": ""`,
			"police.glyphes : table vide"},
		{"une avance de trop", `"avances": [7, 7]`, `"avances": [7, 7, 7]`,
			"police.avances : 3 pour 2 glyphes"},
		{"ligne de base hors cellule", `"ligne_de_base": 8`, `"ligne_de_base": 9`,
			"police.ligne_de_base : 9"},
		{"glyphe en double", `"glyphes": "AB"`, `"glyphes": "AA"`,
			"police.glyphes : U+0041"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			texte := strings.Replace(manifesteValide, c.avant, c.apres, 1)
			if texte == manifesteValide {
				t.Fatalf("le cas ne change rien : %q est absent du manifeste", c.avant)
			}
			fsys := fstest.MapFS{"m.json": {Data: []byte(texte)}}
			_, err := LoadFont(fsys, "m.json")
			if err == nil {
				t.Fatalf("manifeste accepté alors qu'il porte : %s", c.nom)
			}
			if !strings.Contains(err.Error(), c.attendu) {
				t.Errorf("message %q, attendu contenant %q", err, c.attendu)
			}
		})
	}
}

// TestLeManifesteValideSeCharge est le témoin des refus de
// `TestLeManifesteDInterfaceRefuseCeQuiEstIncoherent`.
//
// Sans lui, un `LoadFont` qui refuserait tout les ferait tous passer : ils
// vérifient qu'un manifeste est refusé, jamais qu'un autre est accepté. C'est la
// paire dont chaque moitié garde ce que l'autre ne garde pas.
func TestLeManifesteValideSeCharge(t *testing.T) {
	fsys := fstest.MapFS{"m.json": {Data: []byte(manifesteValide)}}
	police, err := LoadFont(fsys, "m.json")
	if err != nil {
		t.Fatalf("manifeste valide refusé : %v", err)
	}
	if _, ok := police.Place('A'); !ok {
		t.Error("le glyphe A n'a pas de place")
	}
	if n := police.Advance("AB"); n != 14 {
		t.Errorf("avance de « AB » : %d, attendu 14", n)
	}
}

// TestUnGlypheAbsentOccupeUneCellule garde un choix qu'on croirait faux.
//
// Le geste naturel serait d'ignorer un caractère que la table ne porte pas, ce
// qui rendrait un texte plausible auquel il manque une lettre — donc un défaut
// qui se lit comme du texte. Une cellule pleine laisse un trou, et un trou se
// voit là où il se produit.
func TestUnGlypheAbsentOccupeUneCellule(t *testing.T) {
	fsys := fstest.MapFS{"m.json": {Data: []byte(manifesteValide)}}
	police, err := LoadFont(fsys, "m.json")
	if err != nil {
		t.Fatalf("manifeste valide refusé : %v", err)
	}
	if n := police.Advance("AZB"); n != 14+police.Cell[0] {
		t.Errorf("avance de « AZB » : %d, attendu %d", n, 14+police.Cell[0])
	}
}

// TestLaPoliceLivreeSeCharge lit le manifeste publié, sans rien injecter.
//
// `TestLeManifesteDInterfaceRefuseCeQuiEstIncoherent` bâtit son entrée pour
// isoler un refus ; celui-ci passe par le chemin qui produit la vraie, et c'est
// le seul qui tombe si le générateur change ce qu'il écrit. Ce qu'il garde en propre : la table couvre
// l'ASCII imprimable et les capitales accentuées, que le français impose et
// qu'une fonte de jeu anglophone porte rarement.
func TestLaPoliceLivreeSeCharge(t *testing.T) {
	police, err := LoadFont(cohue.Assets, cohue.InterfaceManifest)
	if err != nil {
		t.Fatalf("chargement de la police livrée : %v", err)
	}
	for r := rune(' '); r <= '~'; r++ {
		if _, ok := police.Place(r); !ok {
			t.Errorf("le glyphe U+%04X manque à la police livrée", r)
		}
	}

	// L'espace insécable s'écrit par son code et non en littéral : posée telle
	// quelle, elle se confond avec une espace ordinaire dans le source, et un
	// éditeur la normalise sans que personne ne le voie. C'est elle qui sépare
	// les milliers, donc elle manquerait au premier score à quatre chiffres.
	insecable := rune(0x00A0)
	for _, r := range []rune{'À', 'É', 'È', 'Ê', 'Ç', 'Î', 'Ô', 'Ù', 'Ÿ', 'Œ',
		'Æ', '«', '»', '°', '—', insecable} {
		if _, ok := police.Place(r); !ok {
			t.Errorf("le glyphe U+%04X manque à la police livrée", r)
		}
	}

	if police.Baseline >= police.Cell[1] {
		t.Errorf("ligne de base %d dans une cellule haute de %d",
			police.Baseline, police.Cell[1])
	}
}
