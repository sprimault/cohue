// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le chargeur de feuilles garde : les bandes livrées se découpent, une
// taille qui ment se refuse, et un cycle absent se dit au lieu de se deviner.

package sprite

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
)

// racineLivree est le dossier des figurines dans les ressources embarquées.
const racineLivree = "assets/personnages"

// manifesteLivre est le manifeste des personnages, tel que le binaire l'embarque.
const manifesteLivre = "assets/personnages/manifeste.json"

// bande fabrique une bande de la taille demandée, opaque.
func bande(t *testing.T, largeur, hauteur int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, largeur, hauteur))
	for y := range hauteur {
		for x := range largeur {
			img.Set(x, y, color.NRGBA{R: uint8(x), G: uint8(y), B: 200, A: 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatalf("encodage de la bande : %v", err)
	}
	return b.Bytes()
}

// figureDEssai rend une figure minimale, à compléter par le cas.
func figureDEssai() game.Figure {
	return game.Figure{
		Key:        "marcheur",
		Side:       8,
		Anchor:     [2]int{4, 7},
		Variants:   1,
		Directions: []string{"S"},
		Cycles:     map[string]game.Cycle{"marche": {Frames: 3, Duration: 6, Loop: true}},
	}
}

// TestLesFeuillesLivreesSeChargent monte toutes les bandes de `assets/` sans
// rien injecter.
//
// C'est la conformité que la doctrine exige : un test qui bâtit son entrée ne
// dit rien du système monté, et un mode resté inerte des mois sur un projet
// antérieur venait exactement de là — le test posait lui-même la clé que le
// manifeste livré n'a jamais portée.
//
// Il compte ce qu'il a visité et le dit : une boucle qui ne trouverait aucun
// profil passerait au vert en ne vérifiant rien, ce qui est le motif du contrôle
// privé de son entrée.
func TestLesFeuillesLivreesSeChargent(t *testing.T) {
	profils, err := game.LoadProfiles(cohue.Assets, manifesteLivre)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	figures := []game.Figure{profils.Player.Figure}
	for _, p := range profils.Enemies {
		figures = append(figures, p.Figure)
	}
	for _, p := range profils.Ambient {
		figures = append(figures, p.Figure)
	}

	bandes := 0
	for _, f := range figures {
		feuille, err := Load(cohue.Assets, racineLivree, f)
		if err != nil {
			t.Fatalf("feuille de %s : %v", f.Key, err)
		}
		for cycle, c := range f.Cycles {
			for _, direction := range f.Directions {
				for variante := range f.Variants {
					bandes++
					if _, ok := feuille.Frame(cycle, direction, variante, c.Frames-1); !ok {
						t.Errorf("%s : dernière image de %s_%s manquante en v%d",
							f.Key, cycle, direction, variante)
					}
				}
			}
		}
	}

	if len(figures) < 9 || bandes < 400 {
		t.Fatalf("%d profils et %d bandes visités : trop peu pour que ce test dise quelque chose",
			len(figures), bandes)
	}
	t.Logf("%d profils, %d bandes découpées", len(figures), bandes)
}

// TestUneBandeQuiMentSurSaTailleEstRefusee garde le manifeste-contrat.
//
// Une bande de 320 pixels est indéchiffrable seule — cinq images de 64 ou quatre
// de 80 —, et c'est le manifeste qui tranche. Accepter la taille du fichier
// reviendrait à faire dire au dessin ce que le contrat doit dire, et un
// renommage de fichier casserait le jeu au lancement plutôt qu'au chargement.
func TestUneBandeQuiMentSurSaTailleEstRefusee(t *testing.T) {
	f := figureDEssai()
	// Quatre images là où le manifeste en annonce trois.
	fsys := fstest.MapFS{
		"p/marcheur/marche_S.png": &fstest.MapFile{Data: bande(t, 4*f.Side, f.Side)},
	}

	if _, err := Load(fsys, "p", f); err == nil {
		t.Fatal("bande de quatre images acceptée pour un cycle qui en annonce trois")
	}
}

// TestUnCycleAbsentSeDitAuLieuDeSeDeviner garde le second retour de `Frame`.
//
// Tous les profils n'ont pas les mêmes cycles : le Molosse n'a ni repos ni
// attaque. Rendre une image vide plutôt qu'un refus ferait disparaître une
// créature à l'arrêt sans qu'aucune erreur ne le dise, et le rendu ne saurait
// pas qu'il doit se replier sur une pose existante.
func TestUnCycleAbsentSeDitAuLieuDeSeDeviner(t *testing.T) {
	f := figureDEssai()
	fsys := fstest.MapFS{
		"p/marcheur/marche_S.png": &fstest.MapFile{Data: bande(t, 3*f.Side, f.Side)},
	}

	feuille, err := Load(fsys, "p", f)
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}
	if _, ok := feuille.Frame("repos", "S", 0, 0); ok {
		t.Error("un cycle que le manifeste ne déclare pas a rendu une image")
	}
	if _, ok := feuille.Frame("marche", "S", 0, 3); ok {
		t.Error("la quatrième image d'un cycle qui en porte trois a été rendue")
	}
	if _, ok := feuille.Frame("marche", "S", 0, 2); !ok {
		t.Error("la troisième image d'un cycle qui en porte trois manque")
	}
}

// TestLeSousDossierNexisteQuAuDelaDUneTeinte garde la règle des variantes.
//
// Un profil à teinte unique n'a pas de `v0` : le chargeur n'a rien à savoir des
// variantes qui n'existent pas. Chercher `v0` partout ferait échouer huit profils
// sur neuf, et l'ajouter au générateur pour satisfaire le chargeur reviendrait à
// faire décider le lecteur de ce que le catalogue contient.
func TestLeSousDossierNexisteQuAuDelaDUneTeinte(t *testing.T) {
	for _, cas := range []struct {
		quoi      string
		variantes int
		chemin    string
	}{
		{"teinte unique, à plat", 1, "p/marcheur/marche_S.png"},
		{"deux teintes, en sous-dossier", 2, "p/marcheur/v1/marche_S.png"},
	} {
		t.Run(cas.quoi, func(t *testing.T) {
			f := figureDEssai()
			f.Variants = cas.variantes

			fsys := fstest.MapFS{}
			for _, chemin := range []string{"p/marcheur/marche_S.png", "p/marcheur/v0/marche_S.png",
				"p/marcheur/v1/marche_S.png"} {
				fsys[chemin] = &fstest.MapFile{Data: bande(t, 3*f.Side, f.Side)}
			}
			if _, err := Load(fsys, "p", f); err != nil {
				t.Fatalf("chargement : %v", err)
			}

			// Le fichier attendu est retiré : ce qui doit échouer est le chemin
			// que le chargeur construit, et non un fichier qui manquerait aussi.
			delete(fsys, cas.chemin)
			if _, err := Load(fsys, "p", f); err == nil {
				t.Errorf("chargement réussi alors que %s manque", cas.chemin)
			}
		})
	}
}

// TestUneFigureSansCoteEstRefusee garde ce que `internal/game` ne contrôle plus.
//
// Le manifeste exige ces champs sans vérifier leur contenu, `internal/game` les
// portant sans les lire. Le refus vit donc ici, avec le seul paquet qui sache ce
// qu'il en attend — et sans lui, un `cote` à zéro produirait des images vides
// plutôt qu'un message.
func TestUneFigureSansCoteEstRefusee(t *testing.T) {
	for _, cas := range []struct {
		quoi    string
		abimer  func(*game.Figure)
		attendu string
	}{
		{"côté nul", func(f *game.Figure) { f.Side = 0 }, "cote"},
		{"aucune direction", func(f *game.Figure) { f.Directions = nil }, "directions"},
		{"aucun cycle", func(f *game.Figure) { f.Cycles = nil }, "cycles"},
		{"aucune variante", func(f *game.Figure) { f.Variants = 0 }, "variantes"},
		{"appui hors de la case", func(f *game.Figure) { f.Anchor = [2]int{4, 99} }, "appui"},
		{"cycle sans image", func(f *game.Figure) {
			f.Cycles = map[string]game.Cycle{"marche": {Frames: 0, Duration: 6}}
		}, "images"},
	} {
		t.Run(cas.quoi, func(t *testing.T) {
			f := figureDEssai()
			cas.abimer(&f)
			if _, err := Load(fstest.MapFS{}, "p", f); err == nil {
				t.Fatalf("figure acceptée alors que %s est fautif", cas.attendu)
			}
		})
	}
}
