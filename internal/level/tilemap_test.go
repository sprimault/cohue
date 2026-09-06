// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que la carte des formes garde : chaque case en porte une, la grille de
// coûts en descend case par case, et le hors-carte se dit au lieu de rendre une
// forme.

package level

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue/internal/game"
)

// TestLaCarteDesFormesNommeChaqueCase vérifie que la cuisson pose une forme
// partout, et laquelle.
//
// Le même fichier asymétrique que la cuisson en coûts, pour la même raison : un
// assemblage qui inverserait `u` et `v` passerait tous les autres cas et
// échouerait ici.
func TestLaCarteDesFormesNommeChaqueCase(t *testing.T) {
	charge, err := chargeurDeTest(t).Load("essai")
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}
	tuiles := charge.Tiles

	if tuiles.Width() != 8 || tuiles.Height() != 4 {
		t.Fatalf("carte %dx%d, attendu 8x4", tuiles.Width(), tuiles.Height())
	}

	nomDe := func(u, v int) string {
		if i := tuiles.At(u, v); i >= 0 {
			return tuiles.Shapes()[i]
		}
		return ""
	}
	for v := range tuiles.Height() {
		for u := range tuiles.Width() {
			if nomDe(u, v) == "" {
				t.Errorf("(%d,%d) ne porte aucune forme", u, v)
			}
		}
	}
	if got := nomDe(0, 0); got != "mur" {
		t.Errorf("(0,0) porte « %s », attendu « mur » : le nord est un mur plein", got)
	}
	if got := nomDe(3, 2); got != "pilier" {
		t.Errorf("(3,2) porte « %s », attendu « pilier »", got)
	}
	if got := nomDe(2, 3); got == "pilier" {
		t.Error("le pilier est en (2,3) : la carte a été posée en (v,u) au lieu de (u,v)")
	}
}

// TestLaGrilleDescendDeLaCarteDesFormes garde ce qui rend les deux
// incontradictoires.
//
// **C'est la propriété, et non une coïncidence à surveiller.** La grille est
// dérivée de la carte plutôt que remplie en même temps qu'elle : ce qu'on
// dessine et ce que le champ de flux lit ne peuvent donc pas dire deux choses
// d'une même case. Le test rejoue la dérivation depuis le catalogue et confronte
// case par case — il tombe le jour où quelqu'un remet deux remplissages côte à
// côte, ce qui est précisément ce qu'aucune relecture ne rattrape.
func TestLaGrilleDescendDeLaCarteDesFormes(t *testing.T) {
	charge, err := chargeurDeTest(t).Load("essai")
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}

	couts := decorDeTest.Costs()
	vus := map[game.Cost]int{}
	for v := range charge.Tiles.Height() {
		for u := range charge.Tiles.Width() {
			attendu := couts[charge.Tiles.Shapes()[charge.Tiles.At(u, v)]]
			if got := charge.Grid.At(u, v); got != attendu {
				t.Errorf("(%d,%d) coûte %d, la forme posée en dit %d", u, v, got, attendu)
			}
			vus[attendu]++
		}
	}

	// Les deux natures doivent être présentes : une carte uniforme passerait au
	// vert sans rien départager.
	if vus[game.Free] == 0 || vus[game.Blocked] == 0 {
		t.Fatalf("cases vues par coût : %v — il en faut des deux natures", vus)
	}
}

// TestUnThemeSansSolQuiEmploieUneFormeNueEstRefuse garde le refus qui ferme
// réellement le cas.
//
// **Ce n'est ni la forme ni l'absence de sol qui est invalide, c'est leur
// rapprochement** — le même geste que le refus d'un profil qu'une phase autorise
// sans pouvoir le payer. Sans lui, le mécanisme du sol serait correct et un jeu
// de pièces pourrait simplement ne pas l'employer : le trou reviendrait, aussi
// silencieux qu'avant.
//
// Le message est éprouvé autant que le refus. Il s'adresse à un auteur de thème,
// qui doit savoir **laquelle** de ses tuiles exige un sol : « sol manquant » lui
// ferait ouvrir les soixante et une formes du catalogue.
func TestUnThemeSansSolQuiEmploieUneFormeNueEstRefuse(t *testing.T) {
	fsys := fstest.MapFS{
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
		"x/jeu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun",
			"palette": {".": "sol", "O": "pilier"}
		}`)},
		"x/pieces/salle.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "salle", "jeu": "commun",
			"taille": [2, 1], "grille": [".O"]
		}`)},
	}

	_, err := chargeur(fsys).Load("x")
	if err == nil {
		t.Fatal("un thème qui pose un pilier sans déclarer de sol se charge")
	}
	for _, attendu := range []string{"pilier", "0.5×0.5", "sol"} {
		if !strings.Contains(err.Error(), attendu) {
			t.Errorf("le refus ne dit pas « %s » : %v", attendu, err)
		}
	}
}

// TestUnSolQuiNeRemplitPasSaCaseEstRefuse garde la borne du comblement.
//
// Le refus au chargement plutôt qu'une récursion : combler un sol par un autre
// sol n'aurait pas de fin, et la propriété qu'on veut se vérifie en un mot — un
// sol remplit sa case.
func TestUnSolQuiNeRemplitPasSaCaseEstRefuse(t *testing.T) {
	fsys := fstest.MapFS{
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
		"x/jeu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun",
			"sol": "pilier", "palette": {".": "sol"}
		}`)},
		"x/pieces/salle.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "salle", "jeu": "commun",
			"taille": [1, 1], "grille": ["."]
		}`)},
	}

	if _, err := chargeur(fsys).Load("x"); err == nil {
		t.Fatal("un pilier déclaré comme sol du thème est accepté")
	}
}

// TestUneCaseHorsCarteNaPasDeForme garde le repli du balayage.
//
// La fenêtre de la caméra est un rectangle dont les bords tombent au-delà du
// losange du lieu, et ce qui la balaie lit sans tester l'appartenance. Rendre la
// première forme du catalogue plutôt que rien y peindrait une bordure de sol
// autour du lieu.
func TestUneCaseHorsCarteNaPasDeForme(t *testing.T) {
	charge, err := chargeurDeTest(t).Load("essai")
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}

	for _, hors := range [][2]int{{-1, 0}, {0, -1}, {8, 0}, {0, 4}} {
		if i := charge.Tiles.At(hors[0], hors[1]); i >= 0 {
			t.Errorf("(%d,%d) est hors de la carte et rend la forme %d", hors[0], hors[1], i)
		}
	}
}
