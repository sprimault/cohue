// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du chargeur : la cuisson en grille de coûts, la clé inconnue refusée,
// le commentaire admis en tout point, les manquements listés en une fois, le
// dossier renommé sans son identifiant, et le lieu qu'on tente d'ouvrir par un
// fichier.

package level

import (
	"errors"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/manifest"
)

// couts est le catalogue que le manifeste du décor fournira ; les tests en
// donnent la part dont ils ont besoin.
var couts = map[string]game.Cost{
	"sol":    game.Free,
	"mur":    game.Blocked,
	"pilier": game.Blocked,
}

// chargeurDeTest monte un chargeur sur les fichiers de testdata.
//
// La racine est `testdata` et non le dossier du lieu : c'est le dossier qui
// nomme le lieu, et le monter directement priverait le chargeur du nom qu'il
// doit confronter à l'identifiant.
func chargeurDeTest(t *testing.T) *Loader {
	t.Helper()
	return NewLoader(os.DirFS("testdata"), couts)
}

// TestCuissonEnGrilleDeCouts vérifie qu'un lieu devient la carte qu'il dessine.
//
// La grille du fichier est volontairement asymétrique : mur plein sur la
// première ligne, ouverture pleine sur la dernière. Un chargeur qui inverserait
// `u` et `v` passerait tous les autres tests et échouerait ici.
func TestCuissonEnGrilleDeCouts(t *testing.T) {
	grille, err := chargeurDeTest(t).Load("essai")
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
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0, "rotaton": 1}]
		}`)},
	}
	_, err := NewLoader(fsys, couts).Load("x")
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
	if _, err := chargeurDeTest(t).Load("essai"); err != nil {
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
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
		"x/jeu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun", "palette": {".": "sol"}
		}`)},
		// Trois défauts d'un coup : deux lignes au lieu de trois, une ligne trop
		// courte, un caractère absent de la palette.
		"x/pieces/salle.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "salle", "jeu": "commun",
			"taille": [4, 3], "grille": ["....", "..#"]
		}`)},
	}
	_, err := NewLoader(fsys, couts).Load("x")
	if err == nil {
		t.Fatal("un lieu invalide s'est chargé")
	}
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("erreur de type %T, attendu *manifest.Invalid", err)
	}
	if len(invalide.Missing) < 4 {
		t.Errorf("%d manquement(s) listé(s), attendu au moins 4 :\n%v",
			len(invalide.Missing), invalide.Missing)
	}
}

// TestDossierRenommeSansIdentifiant vérifie que le nom du dossier et
// l'identifiant doivent s'accorder.
//
// Le dossier nomme le lieu, l'identifiant aussi : deux descriptions de la même
// chose, donc l'une des deux finira par mentir. Le cas concret est celui de qui
// duplique un lieu pour en faire une variante — il renomme le dossier et oublie
// l'identifiant, et sans ce refus la copie se charge en se croyant l'original.
func TestDossierRenommeSansIdentifiant(t *testing.T) {
	fsys := fstest.MapFS{
		"variante/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
		"variante/jeu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun", "palette": {".": "sol"}
		}`)},
		"variante/pieces/salle.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "salle", "jeu": "commun",
			"taille": [1, 1], "grille": ["."]
		}`)},
	}
	_, err := NewLoader(fsys, couts).Load("variante")
	if err == nil {
		t.Fatal("un lieu dont le dossier et l'identifiant divergent s'est chargé")
	}
	// Les deux noms dans le message : sans eux, l'auteur sait qu'il y a
	// désaccord sans savoir lequel des deux il voulait garder.
	for _, attendu := range []string{"x", "variante"} {
		if !strings.Contains(err.Error(), attendu) {
			t.Errorf("« %s » absent du refus : %v", attendu, err)
		}
	}
}

// TestLieuChargeParUnFichier vérifie qu'un chemin sans dossier est refusé.
//
// Un `fs.FS` monté sur le dossier du lieu lui-même priverait le chargeur du nom
// qu'il doit confronter à l'identifiant, et le contrôle deviendrait muet sans
// que rien ne le dise.
func TestLieuChargeParUnFichier(t *testing.T) {
	fsys := fstest.MapFS{
		"lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
	}
	if _, err := NewLoader(fsys, couts).Load("."); err == nil {
		t.Fatal("un lieu chargé depuis la racine du système de fichiers")
	}
}

// TestFormatNonPrisEnCharge vérifie qu'une version inconnue est refusée plutôt
// que lue de travers.
func TestFormatNonPrisEnCharge(t *testing.T) {
	fsys := fstest.MapFS{
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 99, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
	}
	_, err := NewLoader(fsys, couts).Load("x")
	if !errors.Is(err, manifest.ErrUnsupportedFormat) {
		t.Errorf("format 99 accepté, ou refusé pour une autre raison : %v", err)
	}
}

// TestLieuSansPiece vérifie qu'un lieu vide est refusé.
//
// Il se chargerait sinon en une grille de zéro par zéro, où le joueur
// apparaîtrait hors de la carte.
func TestLieuSansPiece(t *testing.T) {
	fsys := fstest.MapFS{
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun", "pieces": []
		}`)},
	}
	if _, err := NewLoader(fsys, couts).Load("x"); !errors.Is(err, ErrEmptyLevel) {
		t.Errorf("lieu vide accepté : %v", err)
	}
}

// TestPieceInconnue vérifie qu'un lieu citant une pièce absente est refusé.
func TestPieceInconnue(t *testing.T) {
	fsys := fstest.MapFS{
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "fantome", "u": 0, "v": 0}]
		}`)},
		"x/jeu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun", "palette": {".": "sol"}
		}`)},
	}
	if _, err := NewLoader(fsys, couts).Load("x"); !errors.Is(err, ErrUnknownRoom) {
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
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
		"x/jeu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun", "palette": {"?": "forme_inventee"}
		}`)},
		"x/pieces/salle.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "salle", "jeu": "commun",
			"taille": [1, 1], "grille": ["?"]
		}`)},
	}
	grille, err := NewLoader(fsys, couts).Load("x")
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}
	if grille.Passable(0, 0) {
		t.Error("une forme hors catalogue se dit franchissable")
	}
}

// TestJeuDePiecesQuiDementLeLieu garde ce que le nom fixe a cessé de vérifier.
//
// **Ce contrôle est né du renommage.** Le jeu de pièces s'appelait autrefois du
// nom de son identifiant, si bien que le chemin le vérifiait au passage : un
// fichier mal nommé ne se chargeait pas. Le nom étant désormais fixe, plus rien
// ne rapprochait les deux, et un `jeu.json` déposé dans le mauvais lieu se
// serait chargé en silence — avec sa palette, donc en changeant le sens de tous
// les caractères des pièces.
func TestJeuDePiecesQuiDementLeLieu(t *testing.T) {
	fsys := fstest.MapFS{
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [{"id": "salle", "u": 0, "v": 0}]
		}`)},
		"x/jeu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "autre", "palette": {".": "sol"}
		}`)},
		"x/pieces/salle.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "salle", "jeu": "commun",
			"taille": [1, 1], "grille": ["."]
		}`)},
	}
	_, err := NewLoader(fsys, couts).Load("x")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("jeu de pièces étranger au lieu accepté : %v", err)
	}
}

// lieuDePoses monte un lieu d'une seule sorte de pièce, posée où on le demande.
//
// Les trois cas de couverture ne diffèrent que par des positions : leur donner
// chacun son système de fichiers aurait mis trois fois le même carré de quatre
// sur trois sous les yeux du lecteur, qui aurait à le comparer pour trouver ce
// qui change.
func lieuDePoses(poses string) fstest.MapFS {
	return fstest.MapFS{
		"x/lieu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "x", "jeu_pieces": "commun",
			"pieces": [` + poses + `]
		}`)},
		"x/jeu.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "commun", "palette": {".": "sol"}
		}`)},
		"x/pieces/salle.json": &fstest.MapFile{Data: []byte(`{
			"version_format": 1, "identifiant": "salle", "jeu": "commun",
			"taille": [4, 3], "grille": ["....", "....", "...."]
		}`)},
	}
}

// TestCaseQueNullePieceNePose refuse le lieu qui laisse un trou.
//
// **Le trou est franchissable et invisible**, ce qui en fait le pire défaut que
// ce format puisse produire : une grille neuve vaut le coût d'un sol ordinaire,
// si bien qu'une case oubliée se marche comme les autres et ne se dessine pas.
// Personne ne la voit avant qu'une créature y flotte.
func TestCaseQueNullePieceNePose(t *testing.T) {
	fsys := lieuDePoses(`{"id": "salle", "u": 0, "v": 0}, {"id": "salle", "u": 8, "v": 0}`)

	_, err := NewLoader(fsys, couts).Load("x")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("un lieu troué s'est chargé : %v", err)
	}
	// Douze cases, et la première en (4, 0) : un message qui dirait seulement
	// « trou » laisserait à chercher de quel côté.
	if !strings.Contains(invalide.Error(), "(4, 0)") {
		t.Errorf("le message ne situe pas le trou : %v", invalide)
	}
}

// TestCasePoseeDeuxFois refuse le lieu dont deux pièces se recouvrent.
//
// Le recouvrement se résout aujourd'hui par l'ordre des poses, la dernière
// écrivant par-dessus la première. Rien n'annonce cet ordre, et le refuser le
// rend sans effet — ce qui vaut mieux que de l'écrire dans un document que
// l'auteur d'un éditeur ne lira pas.
func TestCasePoseeDeuxFois(t *testing.T) {
	fsys := lieuDePoses(`{"id": "salle", "u": 0, "v": 0}, {"id": "salle", "u": 2, "v": 0}`)

	_, err := NewLoader(fsys, couts).Load("x")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("deux pièces superposées se sont chargées : %v", err)
	}
	if !strings.Contains(invalide.Error(), "(2, 0)") {
		t.Errorf("le message ne situe pas le recouvrement : %v", invalide)
	}
}

// TestPoseAvantLOrigine refuse la pièce posée en amont du coin du lieu.
//
// La cuisson laisse tomber ce qui sort de la grille : une pièce posée en `u` de
// moins un perdrait sa première colonne sans un mot, et le lieu s'ouvrirait avec
// un mur en moins là où son auteur en avait dessiné un.
func TestPoseAvantLOrigine(t *testing.T) {
	fsys := lieuDePoses(`{"id": "salle", "u": -1, "v": 0}`)

	_, err := NewLoader(fsys, couts).Load("x")
	var invalide *manifest.Invalid
	if !errors.As(err, &invalide) {
		t.Fatalf("une pièce posée avant l'origine s'est chargée : %v", err)
	}
	if !strings.Contains(invalide.Error(), "(-1, 0)") {
		t.Errorf("le message ne dit pas où la pièce est posée : %v", invalide)
	}
}
