// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le chargeur de feuilles garde : les bandes livrées se rangent dans
// l'atlas, une taille qui ment se refuse, et un cycle absent se dit au lieu de
// se deviner.

package sprite

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"testing/fstest"

	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/scene"

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
//
// **Et il compte les planches**, parce que c'est la borne du moteur qui décide
// de la découpe : un atlas en porte mille vingt-quatre, et une planche par image
// d'animation au lieu d'une par fichier la dépasserait.
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

	atlas := &raster.Atlas{}
	bandes, images := 0, 0
	for _, f := range figures {
		feuille, err := Load(cohue.Assets, racineLivree, f, atlas)
		if err != nil {
			t.Fatalf("feuille de %s : %v", f.Key, err)
		}
		for cycle, c := range f.Cycles {
			for _, direction := range f.Directions {
				for variante := range f.Variants {
					bandes++
					images += c.Frames
					if _, ok := feuille.Pose(cycle, direction, variante, c.Frames-1); !ok {
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
	if len(atlas.Sheets) != bandes {
		t.Errorf("%d planches pour %d bandes : la découpe n'est plus de la donnée",
			len(atlas.Sheets), bandes)
	}
	t.Logf("%d profils, %d bandes, %d images, %d planches",
		len(figures), bandes, images, len(atlas.Sheets))
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

	if _, err := Load(fsys, "p", f, &raster.Atlas{}); err == nil {
		t.Fatal("bande de quatre images acceptée pour un cycle qui en annonce trois")
	}
}

// TestUnCycleAbsentSeDitAuLieuDeSeDeviner garde le second retour de `Pose`.
//
// Tous les profils n'ont pas les mêmes cycles : le Molosse n'a ni repos ni
// attaque. Rendre un aspect vide plutôt qu'un refus ferait disparaître une
// créature à l'arrêt sans qu'aucune erreur ne le dise — la valeur zéro d'un
// aspect ne dessinant rien —, et le rendu ne saurait pas qu'il doit se replier
// sur une pose existante.
func TestUnCycleAbsentSeDitAuLieuDeSeDeviner(t *testing.T) {
	f := figureDEssai()
	fsys := fstest.MapFS{
		"p/marcheur/marche_S.png": &fstest.MapFile{Data: bande(t, 3*f.Side, f.Side)},
	}

	feuille, err := Load(fsys, "p", f, &raster.Atlas{})
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}
	if _, ok := feuille.Pose("repos", "S", 0, 0); ok {
		t.Error("un cycle que le manifeste ne déclare pas a rendu une image")
	}
	if _, ok := feuille.Pose("marche", "S", 0, 3); ok {
		t.Error("la quatrième image d'un cycle qui en porte trois a été rendue")
	}
	if _, ok := feuille.Pose("marche", "S", 0, 2); !ok {
		t.Error("la troisième image d'un cycle qui en porte trois manque")
	}
}

// TestLesImagesDUneBandeSeSuiventDansLaPlanche garde la découpe.
//
// Les trois images d'un cycle vivent dans une planche et s'y distinguent par
// leur seule abscisse de source. Une découpe qui les ferait coïncider rendrait
// une animation figée, ce qu'aucune erreur ne dirait : les trois aspects
// resteraient valides.
func TestLesImagesDUneBandeSeSuiventDansLaPlanche(t *testing.T) {
	f := figureDEssai()
	fsys := fstest.MapFS{
		"p/marcheur/marche_S.png": &fstest.MapFile{Data: bande(t, 3*f.Side, f.Side)},
	}
	feuille, err := Load(fsys, "p", f, &raster.Atlas{})
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}

	for image := range 3 {
		pose, ok := feuille.Pose("marche", "S", 0, image)
		if !ok {
			t.Fatalf("image %d manquante", image)
		}
		if got, veut := int(pose.Aspect.SourceX), image*f.Side; got != veut {
			t.Errorf("image %d : source en %d, attendue en %d", image, got, veut)
		}
		if pose.Aspect.SourceY != 0 {
			t.Errorf("image %d : source en ordonnée %d, une bande est sur une ligne",
				image, pose.Aspect.SourceY)
		}
		if int(pose.Aspect.Width) != f.Side || int(pose.Aspect.Height) != f.Side {
			t.Errorf("image %d : %dx%d, attendu %dx%d",
				image, pose.Aspect.Width, pose.Aspect.Height, f.Side, f.Side)
		}
	}
}

// TestLancrageDUnePoseVientDeLappui garde la conversion des deux conventions.
//
// Le dépôt retranche un point d'appui du pixel visé, le moteur centre le pied de
// l'image : `AnchorFromFooting` est le passage entre les deux, et c'est ici
// qu'il reçoit son premier appelant. Reporter l'appui tel quel poserait la
// créature à cent vingt-six pixels de là sans qu'aucune valeur ne le trahisse.
func TestLancrageDUnePoseVientDeLappui(t *testing.T) {
	f := figureDEssai()
	fsys := fstest.MapFS{
		"p/marcheur/marche_S.png": &fstest.MapFile{Data: bande(t, 3*f.Side, f.Side)},
	}
	feuille, err := Load(fsys, "p", f, &raster.Atlas{})
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}

	pose, ok := feuille.Pose("marche", "S", 0, 0)
	if !ok {
		t.Fatal("première image manquante")
	}
	x, y := AnchorFromFooting(f.Anchor, [2]int{f.Side, f.Side})
	if int(pose.Aspect.AnchorX) != x || int(pose.Aspect.AnchorY) != y {
		t.Errorf("ancre (%d,%d), attendue (%d,%d)",
			pose.Aspect.AnchorX, pose.Aspect.AnchorY, x, y)
	}
}

// TestUnAspectRenduSeDessine garde la teinte neutre.
//
// La valeur zéro d'une couleur est transparente : un aspect livré sans teinte ne
// dessinerait rien, et le défaut se chercherait dans le tri ou dans la position.
// Le blanc est l'élément neutre de la multiplication, donc l'absence de teinte
// plutôt que l'absence de couleur.
func TestUnAspectRenduSeDessine(t *testing.T) {
	f := figureDEssai()
	fsys := fstest.MapFS{
		"p/marcheur/marche_S.png": &fstest.MapFile{Data: bande(t, 3*f.Side, f.Side)},
	}
	feuille, err := Load(fsys, "p", f, &raster.Atlas{})
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}

	pose, _ := feuille.Pose("marche", "S", 0, 0)
	if pose.Aspect.Tint.A == 0 {
		t.Error("teinte transparente : l'aspect ne dessinerait rien")
	}
}

// TestLaSilhouetteAjouteDeuxPlanchesParBande garde ce que la borne de l'atlas
// impose.
//
// L'aplat et le bord d'une bande sont deux planches, pas deux par image : à six
// images par cycle, dériver image par image multiplierait par sept ce que le
// catalogue occupe et franchirait le plafond. Et les cases suivent, pour qu'un
// aspect prenne le même rectangle dans la forme que dans le dessin.
func TestLaSilhouetteAjouteDeuxPlanchesParBande(t *testing.T) {
	f := figureDEssai()
	fsys := fstest.MapFS{
		"p/marcheur/marche_S.png": &fstest.MapFile{Data: bande(t, 3*f.Side, f.Side)},
	}
	atlas := &raster.Atlas{}
	feuille, err := Load(fsys, "p", f, atlas)
	if err != nil {
		t.Fatalf("chargement : %v", err)
	}
	if len(atlas.Sheets) != 1 {
		t.Fatalf("%d planche(s) pour une bande", len(atlas.Sheets))
	}

	if _, ok := feuille.Shape("marche", "S", 0, 0); ok {
		t.Error("un aplat existe avant qu'on l'ait demandé")
	}
	if err := feuille.Silhouette(atlas); err != nil {
		t.Fatalf("silhouette : %v", err)
	}
	if len(atlas.Sheets) != 3 {
		t.Errorf("%d planches après la silhouette, attendu 3", len(atlas.Sheets))
	}

	pose, _ := feuille.Pose("marche", "S", 0, 2)
	forme, ok := feuille.Shape("marche", "S", 0, 2)
	if !ok {
		t.Fatal("aplat de la troisième image manquant")
	}
	bord, ok := feuille.Edge("marche", "S", 0, 2)
	if !ok {
		t.Fatal("bord de la troisième image manquant")
	}
	for _, c := range []struct {
		quoi   string
		aspect scene.Aspect
	}{
		{"aplat", forme},
		{"bord", bord},
	} {
		if c.aspect.SourceX != pose.Aspect.SourceX || c.aspect.Width != pose.Aspect.Width {
			t.Errorf("%s : source %d largeur %d, attendu %d et %d",
				c.quoi, c.aspect.SourceX, c.aspect.Width,
				pose.Aspect.SourceX, pose.Aspect.Width)
		}
		if c.aspect.Sheet == pose.Aspect.Sheet {
			t.Errorf("%s : même planche que le dessin, le dérivé n'en est pas un", c.quoi)
		}
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
			if _, err := Load(fsys, "p", f, &raster.Atlas{}); err != nil {
				t.Fatalf("chargement : %v", err)
			}

			// Le fichier attendu est retiré : ce qui doit échouer est le chemin
			// que le chargeur construit, et non un fichier qui manquerait aussi.
			delete(fsys, cas.chemin)
			if _, err := Load(fsys, "p", f, &raster.Atlas{}); err == nil {
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
			if _, err := Load(fstest.MapFS{}, "p", f, &raster.Atlas{}); err == nil {
				t.Fatalf("figure acceptée alors que %s est fautif", cas.attendu)
			}
		})
	}
}
