// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que la pose d'un volume en entité garde : le dessin ne bouge pas d'un
// pixel, et l'ordre change — ce qui s'abrite derrière lui disparaît derrière lui,
// ce qui passe devant reste entier.

package render

import (
	"testing"

	loz "github.com/sprimault/lozengine"
	"github.com/sprimault/lozengine/raster"
	"github.com/sprimault/lozengine/render"
	"github.com/sprimault/lozengine/scene"
)

const (
	// coteDEssai est le côté de la scène d'essai, en cases : assez pour que la
	// vue tienne dedans et que la case du milieu ait ses quatre voisines.
	coteDEssai = 32
	// tuileDEssai est la tuile de cohue, celle du manifeste livré.
	tuileLargeurDEssai, tuileHauteurDEssai = 64, 32
	// muretDEssai est la taille de l'image du muret au manifeste, et son
	// élévation de vingt-deux pixels en fait un obstacle bas.
	muretLargeurDEssai, muretHauteurDEssai = 64, 54
	// spriteDEssai est le côté d'une figurine de cohue.
	spriteDEssai = 64
	// ecartDEssai est ce qui sépare la créature d'essai du centre de la case,
	// vers le nord puis vers le sud, en cases.
	//
	// **Sept dixièmes, et non une case entière.** C'est la position d'une
	// créature qui s'appuie contre le muret, et c'était celle du défaut : une
	// case entière met les deux à égalité de profondeur, où le départage se
	// décide ailleurs et où ce test ne dirait plus rien de l'ordre.
	ecartDEssai = 0.7
)

// teinteDecorDEssai et teinteCreatureDEssai sont deux teintes franches
// qu'aucun fond ne porte, pour qu'un compte de pixels dise lequel est devant.
var (
	teinteDecorDEssai    = render.Color{R: 255, A: 255}
	teinteCreatureDEssai = render.Color{G: 255, A: 255}
)

// planchesDEssai rend un atlas portant deux images pleines, l'une pour le décor
// et l'autre pour la créature.
//
// **Des planches et non des aplats** : `render.Flat` peint un rectangle sans
// lire de source, et le relevé de recouvrement du moteur ne compte alors rien.
func planchesDEssai(t *testing.T) (*raster.Atlas, uint16, uint16) {
	t.Helper()

	remplir := func(s raster.Sheet, c render.Color) raster.Sheet {
		for i := range s.Pixels {
			s.Pixels[i] = c
		}
		return s
	}

	atlas := &raster.Atlas{}
	decor, err := atlas.Add(remplir(raster.NewSheet(spriteDEssai, spriteDEssai), teinteDecorDEssai))
	if err != nil {
		t.Fatalf("planche du décor : %v", err)
	}
	creature, err := atlas.Add(remplir(raster.NewSheet(spriteDEssai, spriteDEssai), teinteCreatureDEssai))
	if err != nil {
		t.Fatalf("planche de la créature : %v", err)
	}
	return atlas, decor, creature
}

// imageDEssai rend un aspect qui lit une planche entière, à l'étage où cohue
// range ce qui se dresse sur le sol.
//
// Aucune ancre : l'image se pose alors par son pied centré sur le point visé, ce
// que les deux voies de pose font d'elles-mêmes. Le dessin du muret livré en
// porte une, mais elle ne change pas l'ordre — le moteur le dit sur
// `Aspect.AnchorX` — et ce qui s'éprouve ici est l'ordre.
func imageDEssai(planche, largeur, hauteur uint16) scene.Aspect {
	return scene.Aspect{
		Sheet:  planche,
		Width:  largeur,
		Height: hauteur,
		Tint:   render.White(),
		Floor:  etageDebout,
	}
}

// sceneDEssai monte une scène vide à la projection de cohue, centrée sur la case
// du milieu.
func sceneDEssai(t *testing.T) (*scene.Scene, int, int) {
	t.Helper()

	sc, err := scene.New(coteDEssai, coteDEssai,
		projection([2]int{tuileLargeurDEssai, tuileHauteurDEssai}),
		scene.Options{Entities: 8})
	if err != nil {
		t.Fatalf("montage de la scène : %v", err)
	}
	u, v := coteDEssai/2, coteDEssai/2
	sc.Focus(float64(u)+0.5, float64(v)+0.5, Width, Height)
	return sc, u, v
}

// imageDe rend une image de la scène, et le compte de pixels de chaque teinte.
func imageDe(t *testing.T, sc *scene.Scene, atlas *raster.Atlas) ([]render.Color, int, int) {
	t.Helper()

	moteur, err := loz.NewEngine(sc, loz.Options{Width: Width, Height: Height, Atlas: atlas})
	if err != nil {
		t.Fatalf("montage du moteur : %v", err)
	}
	moteur.Step(nil, loz.Input{})

	tampon := moteur.Frame()
	var decor, creature int
	for _, p := range tampon.Pixels {
		switch p {
		case teinteDecorDEssai:
			decor++
		case teinteCreatureDEssai:
			creature++
		}
	}
	return tampon.Pixels, decor, creature
}

// TestUnObstacleBasSePeintAuMemePixelQuUneCellule garde ce qui rend la bascule
// sans conséquence sur le dessin.
//
// **C'est la moitié du changement qui ne doit rien changer.** Poser un obstacle
// bas en entité ne vise qu'à corriger son ordre : si le point donné à la scène
// n'était pas celui depuis lequel le moteur projette une cellule, tout le
// mobilier d'un lieu se décalerait d'un quart de tuile en diagonale, et l'écart
// se chercherait dans le manifeste, où il n'est pas.
//
// Les deux images sont comparées en entier plutôt que par le compte de pixels de
// la teinte : un décalage d'une case en garderait le compte intact.
func TestUnObstacleBasSePeintAuMemePixelQuUneCellule(t *testing.T) {
	atlas, planche, _ := planchesDEssai(t)
	muret := imageDEssai(planche, muretLargeurDEssai, muretHauteurDEssai)

	parCellule, u, v := sceneDEssai(t)
	if !parCellule.PlaceCell(u, v, strateTerrain, muret) {
		t.Fatal("la cellule n'a pas été posée")
	}
	attendu, decor, _ := imageDe(t, parCellule, atlas)
	if decor == 0 {
		t.Fatal("la cellule ne peint rien : le cas ne pose rien")
	}

	parEntite, u, v := sceneDEssai(t)
	if _, pose := parEntite.Place(volumeDebout{u: u, v: v}.entite(muret)); !pose {
		t.Fatal("l'entité n'a pas été posée")
	}
	obtenu, _, _ := imageDe(t, parEntite, atlas)

	for i := range attendu {
		if attendu[i] != obtenu[i] {
			t.Fatalf("les deux images divergent au pixel %d sur %d : %v en cellule, %v en entité",
				i, len(attendu), attendu[i], obtenu[i])
		}
	}
}

// TestUnObstacleBasCacheLaCaseDuNordEtLaisseCelleDuSud garde le défaut que ce
// lot ferme, et son revers.
//
// **Le nord est le cas du défaut** : une créature qui s'abrite derrière un muret
// se peignait par-dessus la maçonnerie, et se lisait comme debout dessus. Une
// cellule se trie à la profondeur de son coin le plus lointain, ce qui la met
// derrière tout ce que la case voisine contient.
//
// **Le sud est ce qui empêche de corriger de travers.** Avancer le décor jusqu'à
// recouvrir ce qui passe devant lui ferait disparaître une créature au premier
// plan, et l'image resterait plausible : le nord seul passerait au vert avec
// cette faute.
func TestUnObstacleBasCacheLaCaseDuNordEtLaisseCelleDuSud(t *testing.T) {
	for _, cas := range []struct {
		nom    string
		ecart  float64
		cachee bool
	}{
		{"au nord, derrière le muret", ecartDEssai, true},
		{"au sud, devant le muret", -ecartDEssai, false},
	} {
		t.Run(cas.nom, func(t *testing.T) {
			atlas, plancheDecor, plancheCreature := planchesDEssai(t)
			sc, u, v := sceneDEssai(t)

			if _, pose := sc.Place(volumeDebout{u: u, v: v}.entite(
				imageDEssai(plancheDecor, muretLargeurDEssai, muretHauteurDEssai),
			)); !pose {
				t.Fatal("l'obstacle bas n'a pas été posé")
			}
			if _, pose := sc.Place(scene.Entity{
				X:      float64(u) + 0.5,
				Y:      float64(v) + 0.5 - cas.ecart,
				Aspect: imageDEssai(plancheCreature, spriteDEssai, spriteDEssai),
			}); !pose {
				t.Fatal("la créature n'a pas été posée")
			}

			_, decor, creature := imageDe(t, sc, atlas)
			if decor == 0 {
				t.Fatal("l'obstacle ne peint rien : le cas ne pose rien")
			}

			entiere := spriteDEssai * spriteDEssai
			switch {
			case cas.cachee && creature >= entiere:
				t.Errorf("la créature derrière l'obstacle ressort à %d pixels sur %d :"+
					" elle se peint par-dessus la maçonnerie", creature, entiere)
			case !cas.cachee && creature != entiere:
				t.Errorf("la créature devant l'obstacle n'en montre que %d pixels sur %d :"+
					" le décor passe devant ce qui est devant lui", creature, entiere)
			}
		})
	}
}
