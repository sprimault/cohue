// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package level

import (
	"errors"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue/internal/game"
)

// couts est le catalogue que le manifeste du décor fournira ; les tests en
// donnent la part dont ils ont besoin.
var couts = map[string]game.Cost{
	"sol":    game.Free,
	"mur":    game.Blocked,
	"pilier": game.Blocked,
}

// chargeurDeTest monte un chargeur sur les fichiers de testdata.
func chargeurDeTest(t *testing.T) *Loader {
	t.Helper()
	return NewLoader(os.DirFS("testdata/simple"), couts)
}

// TestCuissonEnGrilleDeCouts vérifie qu'un lieu devient la carte qu'il dessine.
//
// La grille du fichier est volontairement asymétrique : mur plein sur la
// première ligne, ouverture pleine sur la dernière. Un chargeur qui inverserait
// `u` et `v` passerait tous les autres tests et échouerait ici.
func TestCuissonEnGrilleDeCouts(t *testing.T) {
	grille, err := chargeurDeTest(t).Load("lieu.json")
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}
	if grille.Width() != 8 || grille.Height() != 4 {
		t.Fatalf("grille %dx%d, attendu 8x4", grille.Width(), grille.Height())
	}

	// v = 0 est le bord nord : mur sur toute sa longueur.
	for u := 0; u < 8; u++ {
		if grille.Passable(u, 0) {
			t.Errorf("(%d,0) franchissable, le nord est un mur plein", u)
		}
	}
	// v = 3 est le bord sud : ouvert sur toute sa longueur.
	for u := 0; u < 8; u++ {
		if !grille.Passable(u, 3) {
			t.Errorf("(%d,3) bloqué, le sud est une ouverture pleine", u)
		}
	}
	// Le pilier est en (3,2), et nulle part ailleurs.
	if grille.Passable(3, 2) {
		t.Error("le pilier de (3,2) ne bloque pas")
	}
	if !grille.Passable(2, 3) {
		t.Error("(2,3) bloqué : le pilier a été posé en (v,u) au lieu de (u,v)")
	}
}

// TestCleInconnueRefusee vérifie qu'une faute de frappe ne se charge pas en
// silence, et que le message dit laquelle.
func TestCleInconnueRefusee(t *testing.T) {
	fsys := fstest.MapFS{
		"lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0, "rotaton": 1}]
		}`)},
	}
	_, err := NewLoader(fsys, couts).Load("lieu.json")
	if err == nil {
		t.Fatal("une clé inconnue s'est chargée en silence")
	}
	if !strings.Contains(err.Error(), "rotaton") {
		t.Errorf("le message ne nomme pas la clé fautive : %v", err)
	}
}

// TestCommentaireAccepteEnTousPoints vérifie que `$comment` passe partout où un
// auteur en met un, et pas seulement en tête de fichier.
//
// C'est ce qui rend le JSON tenable pour une pièce écrite à la main : le fichier
// de test en porte un dans son en-tête et un autre au fond d'un ancrage.
func TestCommentaireAccepteEnTousPoints(t *testing.T) {
	if _, err := chargeurDeTest(t).Load("lieu.json"); err != nil {
		t.Fatalf("un commentaire a fait échouer le chargement : %v", err)
	}
}

// TestManquementsListesEnUneFois vérifie que la validation rend tout, et non le
// premier défaut.
//
// C'est la seule promesse de ce genre que le code puisse tenir, et c'est ici
// qu'elle compte : qui met au point une pièce veut la liste, pas un
// aller-retour par erreur.
func TestManquementsListesEnUneFois(t *testing.T) {
	fsys := fstest.MapFS{
		"lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
		"commun.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun", "palette": {".": "sol"}
		}`)},
		// Trois défauts d'un coup : deux lignes au lieu de trois, une ligne trop
		// courte, un caractère absent de la palette.
		"salle.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "salle", "jeu": "commun",
			"taille": [4, 3], "grille": ["....", "..#"]
		}`)},
	}
	_, err := NewLoader(fsys, couts).Load("lieu.json")
	if err == nil {
		t.Fatal("un lieu invalide s'est chargé")
	}
	var invalide *Invalide
	if !errors.As(err, &invalide) {
		t.Fatalf("erreur de type %T, attendu *Invalide", err)
	}
	if len(invalide.Manques) < 4 {
		t.Errorf("%d manquement(s) listé(s), attendu au moins 4 :\n%v",
			len(invalide.Manques), invalide.Manques)
	}
}

// TestFormatNonPrisEnCharge vérifie qu'une version inconnue est refusée plutôt
// que lue de travers.
func TestFormatNonPrisEnCharge(t *testing.T) {
	fsys := fstest.MapFS{
		"lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 99, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
	}
	_, err := NewLoader(fsys, couts).Load("lieu.json")
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Errorf("format 99 accepté, ou refusé pour une autre raison : %v", err)
	}
}

// TestLieuSansPiece vérifie qu'un lieu vide est refusé.
//
// Il se chargerait sinon en une grille de zéro par zéro, où le joueur
// apparaîtrait hors de la carte.
func TestLieuSansPiece(t *testing.T) {
	fsys := fstest.MapFS{
		"lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun", "pieces": []
		}`)},
	}
	if _, err := NewLoader(fsys, couts).Load("lieu.json"); !errors.Is(err, ErrEmptyLevel) {
		t.Errorf("lieu vide accepté : %v", err)
	}
}

// TestPieceInconnue vérifie qu'un lieu citant une pièce absente est refusé.
func TestPieceInconnue(t *testing.T) {
	fsys := fstest.MapFS{
		"lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "fantome", "u": 0, "v": 0}]
		}`)},
		"commun.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun", "palette": {".": "sol"}
		}`)},
	}
	if _, err := NewLoader(fsys, couts).Load("lieu.json"); !errors.Is(err, ErrUnknownRoom) {
		t.Errorf("pièce inconnue acceptée : %v", err)
	}
}

// TestTuileHorsCatalogueBloque vérifie qu'une forme que le manifeste ne connaît
// pas devient un mur.
//
// Le chargeur ne code aucun nom de tuile en dur : il lit un catalogue. Une forme
// qu'il n'y trouve pas ne peut pas être supposée franchissable — un trou dans la
// carte se traverse en silence, un mur se voit.
func TestTuileHorsCatalogueBloque(t *testing.T) {
	fsys := fstest.MapFS{
		"lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
		"commun.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun", "palette": {"?": "forme_inventee"}
		}`)},
		"salle.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "salle", "jeu": "commun",
			"taille": [1, 1], "grille": ["?"]
		}`)},
	}
	grille, err := NewLoader(fsys, couts).Load("lieu.json")
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}
	if grille.Passable(0, 0) {
		t.Error("une forme hors catalogue se dit franchissable")
	}
}
