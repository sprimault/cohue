// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le chargeur de décor garde : le catalogue livré se lit en entier, une
// image qui dément son manifeste se refuse, et le coin où une forme se pose
// tient à trois nombres qu'aucun œil ne rattraperait.

package sprite

import (
	"maps"
	"testing"
	"testing/fstest"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/level"
)

// decorLivre est le manifeste du décor, tel que le binaire l'embarque.
const decorLivre = "assets/decors/manifeste.json"

// formeDEssai rend une forme d'une tuile pleine, à compléter par le cas.
func formeDEssai() level.Shape {
	return level.Shape{
		Theme:     "commun",
		Size:      [2]int{64, 32},
		Anchor:    [2]int{32, 31},
		Footprint: [2]float64{1, 1},
	}
}

// decorDEssai rend un manifeste d'une seule forme.
func decorDEssai(nom string, forme level.Shape) *level.Decor {
	return &level.Decor{Tile: [2]int{64, 32}, Shapes: map[string]level.Shape{nom: forme}}
}

// TestLeCatalogueLivreSeCharge monte toutes les formes de `assets/` sans rien
// injecter.
//
// C'est la conformité que la doctrine exige, et elle porte plus loin qu'il n'y
// paraît : le manifeste annonce une taille par forme, et le contrôle la
// confronte au fichier. Une forme redessinée sans que son manifeste bouge se
// poserait décalée de la différence, ce qu'aucun test montant sa propre image ne
// pourrait dire.
//
// Il compte ce qu'il a visité et le dit : une boucle qui ne trouverait aucune
// forme passerait au vert en ne vérifiant rien.
func TestLeCatalogueLivreSeCharge(t *testing.T) {
	decor, err := level.LoadDecor(cohue.Assets, decorLivre)
	if err != nil {
		t.Fatalf("manifeste de décor : %v", err)
	}

	jeu, err := LoadTiles(cohue.Assets, cohue.DecorDir, decor)
	if err != nil {
		t.Fatalf("formes livrées : %v", err)
	}

	eleves := 0
	for nom, forme := range decor.Shapes {
		tuile, connue := jeu.Tile(nom)
		if !connue {
			t.Errorf("« %s » est au manifeste et pas au catalogue chargé", nom)
			continue
		}
		if taille := tuile.Image.Bounds().Size(); taille.X != forme.Size[0] {
			t.Errorf("« %s » : image large de %d, le manifeste annonce %d",
				nom, taille.X, forme.Size[0])
		}
		if tuile.Elevation != 0 {
			eleves++
		}
	}

	if len(decor.Shapes) < 60 || eleves < 40 {
		t.Fatalf("%d formes dont %d élevées : trop peu pour que ce test dise quelque chose",
			len(decor.Shapes), eleves)
	}
	t.Logf("%d formes chargées, dont %d qui dépassent du sol", len(decor.Shapes), eleves)
}

// TestUneImageQuiDementSonManifesteEstRefusee garde le manifeste-contrat.
//
// La taille vient du manifeste et jamais du fichier. L'accepter du fichier
// ferait dire au dessin ce que le contrat doit dire, et une forme retouchée se
// poserait de travers sans qu'aucun refus ne le signale — le pire des deux, la
// scène restant plausible.
func TestUneImageQuiDementSonManifesteEstRefusee(t *testing.T) {
	forme := formeDEssai()
	fsys := fstest.MapFS{
		"d/commun/sol.png": &fstest.MapFile{Data: bande(t, forme.Size[0], forme.Size[1]+8)},
	}

	if _, err := LoadTiles(fsys, "d", decorDEssai("sol", forme)); err == nil {
		t.Fatal("image de huit pixels trop haute acceptée")
	}
}

// TestUneFormeIntrouvableSeDit garde le refus d'une image absente.
//
// Le chemin vient du thème que le manifeste donne et du nom de la forme : une
// forme déclarée dans le mauvais thème ne se trouve pas, et c'est le genre
// d'écart qu'un générateur produit en déplaçant une forme d'un thème à l'autre.
func TestUneFormeIntrouvableSeDit(t *testing.T) {
	forme := formeDEssai()
	fsys := fstest.MapFS{
		"d/quartier/sol.png": &fstest.MapFile{Data: bande(t, forme.Size[0], forme.Size[1])},
	}

	if _, err := LoadTiles(fsys, "d", decorDEssai("sol", forme)); err == nil {
		t.Fatal("forme déclarée dans « commun » et rangée dans « quartier » acceptée")
	}
}

// TestLeCoinCentreLEmpriseSurSaCase épingle l'arithmétique de pose.
//
// **Elle n'a aucun autre gardien.** Le rendu qui s'en sert n'a pas de test par
// doctrine, et un décalage d'un pixel ou d'une demi-tuile donne une scène qui
// paraît juste : les formes s'alignent entre elles, puisqu'elles se décalent
// toutes pareil, et rien ne dit que le décor a glissé sous les créatures.
//
// Les trois cas sont les trois natures que le décor livré emploie : un sol
// d'une tuile pleine, un mur qui monte, et une forme plus petite que sa case. Le
// dernier est le seul qui distingue le centrage du sommet bas — les deux
// premiers donnent la même réponse.
func TestLeCoinCentreLEmpriseSurSaCase(t *testing.T) {
	tuile := [2]int{64, 32}
	cas := []struct {
		nom    string
		forme  level.Shape
		attend [2]int
	}{
		{
			nom:    "sol",
			forme:  level.Shape{Size: [2]int{64, 32}, Anchor: [2]int{32, 31}, Footprint: [2]float64{1, 1}},
			attend: [2]int{-32, 0},
		},
		{
			// Le losange occupe les trente-deux pixels du bas, les soixante-
			// quatre du haut étant l'élévation : le coin remonte d'autant.
			nom:    "mur",
			forme:  level.Shape{Size: [2]int{64, 96}, Anchor: [2]int{32, 95}, Elevation: 64, Footprint: [2]float64{1, 1}},
			attend: [2]int{-32, -64},
		},
		{
			// Une demi-tuile centrée : son losange de seize pixels tient au
			// milieu de celui de la case, à huit pixels de chacun de ses bords.
			nom:    "pilier",
			forme:  level.Shape{Size: [2]int{32, 80}, Anchor: [2]int{16, 79}, Elevation: 64, Footprint: [2]float64{0.5, 0.5}},
			attend: [2]int{-16, -56},
		},
	}

	for _, c := range cas {
		if got := coin(c.forme, tuile); got != c.attend {
			t.Errorf("%s : coin %v, attendu %v", c.nom, got, c.attend)
		}
	}
}

// TestUnAncrageHorsDeLImageEstRefuse garde le point qu'on pose sur la case.
//
// Un ancrage hors du dessin décale la forme de la différence, et la scène reste
// plausible : c'est le défaut qu'on cherche des heures dans la projection, où il
// n'est pas.
func TestUnAncrageHorsDeLImageEstRefuse(t *testing.T) {
	forme := formeDEssai()
	forme.Anchor = [2]int{32, 32} // la dernière rangée d'une image de 32 est la 31e.
	fsys := fstest.MapFS{
		"d/commun/sol.png": &fstest.MapFile{Data: bande(t, forme.Size[0], forme.Size[1])},
	}

	if _, err := LoadTiles(fsys, "d", decorDEssai("sol", forme)); err == nil {
		t.Fatal("ancrage d'une rangée hors de l'image accepté")
	}
}

// TestLaHauteurDuSolSuitCeQuOnMarche garde la hauteur où se pose un marquage.
//
// Les quatre cas sont ceux du catalogue livré, et ils tiennent en une phrase :
// une forme qui remplit sa case en est la surface, une forme qui la traverse ne
// l'est pas. Le rail et la porte ouverte sont les deux contre-exemples, et la
// porte est le plus parlant — quarante-huit pixels de hauteur pour une case
// qu'on franchit à plat.
func TestLaHauteurDuSolSuitCeQuOnMarche(t *testing.T) {
	cas := []struct {
		nom       string
		emprise   [2]float64
		elevation int
		attend    int
	}{
		{nom: "sol", emprise: [2]float64{1, 1}, elevation: 0, attend: 0},
		{nom: "trottoir", emprise: [2]float64{1, 1}, elevation: 6, attend: 6},
		{nom: "quai", emprise: [2]float64{1, 1}, elevation: 10, attend: 10},
		{nom: "rail", emprise: [2]float64{2, 0.1}, elevation: 4, attend: 0},
		{nom: "porte_ouverte", emprise: [2]float64{1, 0.2}, elevation: 48, attend: 0},
	}

	for _, c := range cas {
		forme := formeDEssai()
		forme.Footprint = c.emprise
		forme.Elevation = c.elevation
		if got := hauteurSol(forme); got != c.attend {
			t.Errorf("%s : hauteur de sol %d, attendu %d", c.nom, got, c.attend)
		}
	}
}

// TestLeCatalogueLivreNaQueDeuxSolsSurelevees épingle ce que le correctif du
// télégraphe suppose du décor.
//
// Le marquage d'une explosion monte à `GroundHeight`, et cette hauteur ne vaut
// que pour une forme couvrante. Le jour où le générateur en ajoute une
// troisième, ou change l'élévation de l'une des deux, c'est ici qu'on l'apprend
// plutôt que sur une capture d'écran.
func TestLeCatalogueLivreNaQueDeuxSolsSurelevees(t *testing.T) {
	decor, err := level.LoadDecor(cohue.Assets, decorLivre)
	if err != nil {
		t.Fatal(err)
	}
	jeu, err := LoadTiles(cohue.Assets, "assets/decors", decor)
	if err != nil {
		t.Fatal(err)
	}

	hauteurs := map[string]int{}
	for nom, forme := range decor.Shapes {
		if forme.Category != "sol" {
			continue
		}
		tuile, connue := jeu.Tile(nom)
		if !connue {
			t.Fatalf("%s : absente du catalogue chargé", nom)
		}
		if tuile.GroundHeight != 0 {
			hauteurs[nom] = tuile.GroundHeight
		}
	}

	attendu := map[string]int{"trottoir": 6, "quai": 10}
	if !maps.Equal(hauteurs, attendu) {
		t.Errorf("sols surélevés %v, attendu %v", hauteurs, attendu)
	}
}

// TestSeulesLesFormesHautesCachentUnPersonnage épingle ce qui déclenche une
// silhouette.
//
// **Le seuil est celui du générateur, vingt-quatre pixels**, et il vaut d'être
// gardé ici parce qu'il est la frontière entre un décor de bordure et un
// obstacle à contourner. Un trottoir de six pixels recouvre la case d'à côté
// sans cacher qui que ce soit ; le révéler faisait clignoter le personnage en
// blanc à chaque bordure.
func TestSeulesLesFormesHautesCachentUnPersonnage(t *testing.T) {
	decor, err := level.LoadDecor(cohue.Assets, decorLivre)
	if err != nil {
		t.Fatal(err)
	}
	jeu, err := LoadTiles(cohue.Assets, "assets/decors", decor)
	if err != nil {
		t.Fatal(err)
	}

	plafond, plancher := 0, 1<<30
	for nom, forme := range decor.Shapes {
		tuile, connue := jeu.Tile(nom)
		if !connue {
			t.Fatalf("%s : absente du catalogue chargé", nom)
		}
		if tuile.Masking != forme.Masking {
			t.Errorf("%s : masquant %v au catalogue, %v au manifeste",
				nom, tuile.Masking, forme.Masking)
		}
		if tuile.Masking {
			plancher = min(plancher, forme.Elevation)
		} else {
			plafond = max(plafond, forme.Elevation)
		}
	}

	t.Logf("la plus haute forme non masquante : %d px ; la plus basse masquante : %d px",
		plafond, plancher)
	if plafond >= plancher {
		t.Errorf("une forme de %d px ne cache pas quand une de %d px cache : le seuil "+
			"n'en est plus un", plafond, plancher)
	}
	if plancher > 24+1 {
		t.Errorf("la plus basse forme masquante est à %d px, le générateur annonce 24",
			plancher)
	}
}
