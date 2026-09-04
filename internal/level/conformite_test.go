// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le montage du décor et du lieu publiés, sans rien injecter : `go:embed`, le
// manifeste, la palette, la cuisson, case par case. Le montage d'une partie
// entière est éprouvé dans `internal/session`.

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
// Il exerce la chaîne du lieu — `go:embed`, le manifeste du décor, la palette du
// jeu de pièces, la cuisson — et tombe quand deux de ces maillons cessent d'être
// d'accord. Un renommage de forme dans le générateur casse ici, pas au lancement
// du jeu.
//
// Il s'arrête à la grille, et `TestPartieLivreeSeMonte` reprend au-delà : celui-là
// monte profils, armes et joueur, mais ne regarde aucune case. Le supprimer au
// motif que l'autre monte davantage laisserait la cuisson sans épreuve.
func TestLieuLivre(t *testing.T) {
	_, couts, err := LoadDecor(cohue.Assets, "assets/decors/manifeste.json")
	if err != nil {
		t.Fatalf("catalogue de coûts : %v", err)
	}
	profils, err := game.LoadProfiles(cohue.Assets, "assets/personnages/manifeste.json")
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	// La progression livrée, et non une borne d'essai : ce cas monte ce que le
	// jeu monte, et le refus d'un profil trop cher pour son plafond fait partie
	// de ce qu'il doit vérifier sur le lieu réel.
	progression, err := game.LoadProgression(cohue.Assets, "assets/progression/manifeste.json")
	if err != nil {
		t.Fatalf("progression livrée : %v", err)
	}
	charge, err := NewLoader(cohue.Assets, couts, profils, progression.CarryOver).
		Load("assets/campagnes/demonstration/place")
	if err != nil {
		t.Fatalf("chargement du lieu : %v", err)
	}
	grille, scenario := charge.Grid, charge.Scenario

	// **Le scénario du lieu livré est le seul que rien d'autre n'exigerait.** Le
	// format admet une salle sans horde — une boutique, un passage —, si bien
	// qu'un `vagues` effacé par mégarde se chargerait sans un mot et rendrait une
	// partie où rien n'arrive. C'est ici que ça se voit, et nulle part ailleurs.
	if len(scenario.Phases) == 0 {
		t.Error("le lieu livré n'a aucune phase : la partie se jouerait sans horde")
	}

	// **Les sept profils sont convoqués, et ça ne se vérifie que d'ici.** Une run
	// de référence ne peut pas le dire : ses instants tombent dans la première
	// minute, et son pilote — une direction constante — meurt avant le troisième
	// palier, après quoi le bassin sature et plus rien n'apparaît. Ce qui garantit
	// qu'un profil peut apparaître est le refus au chargement d'une phase qui ne
	// peut pas le payer ; ce qui garantit qu'il est **écrit** est ce cas-ci.
	//
	// Sans lui, un profil retiré du lieu par mégarde ne se verrait qu'en jouant
	// dix minutes, et seulement si l'on savait déjà lequel chercher.
	convoques := map[int]bool{}
	for _, phase := range scenario.Phases {
		for _, rang := range phase.Profiles {
			convoques[rang] = true
		}
	}
	if len(convoques) != len(profils.Enemies) {
		t.Errorf("le lieu livré convoque %d profils sur %d : l'étape 4 les rend "+
			"tous jouables, et ce lieu est le seul endroit où ça se voit",
			len(convoques), len(profils.Enemies))
	}
	t.Logf("profils convoqués : %d sur %d", len(convoques), len(profils.Enemies))
	if grille.Width() != 98 || grille.Height() != 98 {
		t.Fatalf("grille %dx%d, attendu 98x98", grille.Width(), grille.Height())
	}

	// Une case par pièce posée, relevée sur son dessin et décalée de son
	// origine, en `[u, v]`.
	//
	// **C'est la pose qui est éprouvée ici, pas le dessin.** Une seule case
	// suffirait à dire que la cuisson lit une grille ; il en faut une par pièce
	// pour dire qu'elle les pose au bon endroit, et un lieu qui n'en portait
	// qu'une n'a jamais rien demandé à ce code.
	cases := []struct {
		u, v int
		quoi string
		veut game.Cost
	}{
		{0, 0, "enceinte du nord", game.Blocked},
		{0, 50, "enceinte de l'ouest", game.Blocked},
		{3, 9, "muret de la chicane de « carrefour »", game.Blocked},
		{49, 13, "muret long d'« avenue »", game.Blocked},
		{41, 54, "flaque d'« esplanade »", 2},
		{49, 49, "sol au centre du lieu", game.Free},
	}
	for _, c := range cases {
		if a := grille.At(c.u, c.v); a != c.veut {
			t.Errorf("%s en (%d, %d) : %d, attendu %d", c.quoi, c.u, c.v, a, c.veut)
		}
	}
}
