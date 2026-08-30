// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package level

import (
	"testing"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
)

// TestManifesteLivreDonneLeCatalogue vérifie que le décor tel qu'il est publié
// se dérive sans un seul coût écrit à la main.
//
// Les autres tests du chargeur fournissent leur propre catalogue, ce qui est le
// bon découpage pour eux : ils éprouvent l'assemblage, pas le manifeste. Mais
// aucun ne dirait rien si le générateur cessait d'écrire `cout_traversee`, ou
// si une forme entrait avec un rôle contradictoire. Sur un projet antérieur, un
// mode est resté inerte des mois parce que le test posait lui-même la clé que
// le manifeste livré ne portait pas.
func TestManifesteLivreDonneLeCatalogue(t *testing.T) {
	_, couts, err := LoadDecor(cohue.Assets, "assets/decors/manifeste.json")
	if err != nil {
		t.Fatalf("catalogue de coûts : %v", err)
	}

	// Trois formes, trois natures : ce qui se marche, ce qui ralentit, ce qui
	// arrête. Une seule d'entre elles suffirait à passer si le catalogue était
	// devenu binaire.
	attendus := map[string]game.Cost{"sol": game.Free, "flaque": 2, "mur": game.Blocked}
	for nom, veut := range attendus {
		if a, connu := couts[nom]; !connu {
			t.Errorf("« %s » absent du catalogue", nom)
		} else if a != veut {
			t.Errorf("« %s » coûte %d, attendu %d", nom, a, veut)
		}
	}
}

// TestLieuLivre monte le lieu publié sur le catalogue publié, sans rien injecter.
//
// C'est le seul test qui exerce la chaîne entière — `go:embed`, le manifeste du
// décor, la palette du jeu de pièces, la cuisson — et donc le seul qui tombe
// quand deux maillons cessent d'être d'accord. Un renommage de forme dans le
// générateur casse ici, pas au lancement du jeu.
func TestLieuLivre(t *testing.T) {
	_, couts, err := LoadDecor(cohue.Assets, "assets/decors/manifeste.json")
	if err != nil {
		t.Fatalf("catalogue de coûts : %v", err)
	}
	grille, err := NewLoader(cohue.Assets, couts).Load("assets/lieux/place")
	if err != nil {
		t.Fatalf("chargement du lieu : %v", err)
	}
	if grille.Width() != 32 || grille.Height() != 32 {
		t.Fatalf("grille %dx%d, attendu 32x32", grille.Width(), grille.Height())
	}

	// Quatre cases relevées sur le dessin de `carrefour.json`, en `[u, v]`.
	cases := []struct {
		u, v int
		quoi string
		veut game.Cost
	}{
		{0, 0, "mur du bord", game.Blocked},
		{16, 16, "sol au centre", game.Free},
		{14, 12, "flaque", 2},
		{2, 8, "muret de la chicane", game.Blocked},
	}
	for _, c := range cases {
		if a := grille.At(c.u, c.v); a != c.veut {
			t.Errorf("%s en (%d, %d) : %d, attendu %d", c.quoi, c.u, c.v, a, c.veut)
		}
	}
}
