// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de l'épargne : le profil cher qui finit par sortir, la masse que le
// tirage pondéré produit, le convoité abandonné quand il se ferme, et le budget
// qui lui survit.

package game

import "testing"

// TestUnProfilCherFinitParSortir garde ce que l'épargne apporte, et rien
// d'autre.
//
// **C'est le défaut que ce lot ferme.** Le spawner achetait dès qu'un profil
// était payable : le budget se vidait au prix minimal de la phase et n'atteignait
// jamais les prix élevés. Un seul profil apparaissait dans toute une run — mesuré
// sur huit minutes du lieu livré —, et les six autres étaient écrits sans jamais
// arriver.
//
// Le cas met un Quidam à trois et un Vigile à douze dans la même phase : sans
// épargne, le Vigile ne sort jamais, quelle que soit la durée.
func TestUnProfilCherFinitParSortir(t *testing.T) {
	w, profils := salleOuverte(t, nil, 64)
	marcheur := indexDuProfil(t, profils, "marcheur")
	bloqueur := indexDuProfil(t, profils, "bloqueur")
	if profils.Enemies[bloqueur].PressureCost <= profils.Enemies[marcheur].PressureCost {
		t.Fatal("le Vigile ne coûte pas plus qu'un Quidam : le cas ne sépare rien")
	}
	w.scenario = vagueUnique(6, marcheur, bloqueur)

	for range 60 * TPS {
		w.Step(Vec{})
		if w.vivants[bloqueur] > 0 {
			return
		}
	}
	t.Errorf("aucun Vigile en soixante secondes : le profil le moins cher a "+
		"épuisé le budget avant que le sien soit atteint (%d Quidams posés)",
		w.vivants[marcheur])
}

// TestLeTirageProduitDeLaMasse vérifie que la pondération donne à chaque profil
// la même part de budget.
//
// **C'est ce qui fait une horde de masse avec des exceptions.** Un tirage
// uniforme sortirait autant de Vigiles que de Quidams en nombre, donc quatre fois
// plus de budget dépensé par les gros : une horde chère et clairsemée. Pondéré
// par l'inverse du prix, le nombre suit l'inverse du prix.
//
// La marge est large — un facteur deux sur un rapport attendu de quatre — parce
// que ce cas garde la forme du mécanisme, pas la précision d'un tirage : sur
// quelques centaines d'achats, l'écart type a son mot à dire.
func TestLeTirageProduitDeLaMasse(t *testing.T) {
	w, profils := salleOuverte(t, nil, 300)
	marcheur := indexDuProfil(t, profils, "marcheur")
	bloqueur := indexDuProfil(t, profils, "bloqueur")
	w.scenario = vagueUnique(120, marcheur, bloqueur)

	for range 120 * TPS {
		w.Step(Vec{})
	}
	quidams, vigiles := w.vivants[marcheur], w.vivants[bloqueur]
	if vigiles == 0 {
		t.Fatalf("aucun Vigile sur %d Quidams", quidams)
	}

	// Quatre pour un, aux mêmes prix : douze contre trois.
	rapport := float64(quidams) / float64(vigiles)
	if rapport < 2 || rapport > 8 {
		t.Errorf("%d Quidams pour %d Vigiles, soit %.1f pour un : attendu autour "+
			"de quatre, qui est le rapport de leurs prix", quidams, vigiles, rapport)
	}
	t.Logf("%d Quidams pour %d Vigiles (%.1f pour un)", quidams, vigiles, rapport)
}

// TestLeConvoiteSAbandonneQuandIlSeFerme garde ce qui empêche un plafond de
// bloquer toute la horde.
//
// Le Secouriste ne vit qu'à un exemplaire. Convoité alors qu'un premier est déjà
// vivant, il ne sera jamais payé — et épargner pour lui arrêterait tout achat
// jusqu'à sa mort. Le cas le rend seul achetable, laisse le spawner l'épargner,
// puis ferme son plafond : la horde doit continuer d'arriver.
//
// **Le bassin est large à dessein.** Avec soixante-quatre places, il se remplit
// avant la fin du cas et plus rien n'apparaît — pour une raison qui n'est pas
// celle qu'on éprouve. C'est ce que le premier jet mesurait.
func TestLeConvoiteSAbandonneQuandIlSeFerme(t *testing.T) {
	w, profils := salleOuverte(t, nil, 300)
	marcheur := indexDuProfil(t, profils, "marcheur")
	soigneur := indexDuProfil(t, profils, "soigneur")
	if profils.Enemies[soigneur].MaxAlive != 1 {
		t.Fatalf("le Secouriste plafonne à %d, ce cas en attend un",
			profils.Enemies[soigneur].MaxAlive)
	}
	w.scenario = vagueUnique(9, marcheur, soigneur)

	for range 30 * TPS {
		w.Step(Vec{})
	}
	if w.vivants[soigneur] != 1 {
		t.Fatalf("%d Secouriste(s) vivants, ce cas en attend un pour fermer son "+
			"plafond", w.vivants[soigneur])
	}

	avant := w.ennemis.Len()
	for range 10 * TPS {
		w.Step(Vec{})
	}
	if w.ennemis.Len() <= avant {
		t.Error("plus rien n'apparaît une fois le plafond du Secouriste atteint : " +
			"le spawner épargne pour un profil qu'il ne pourra jamais payer")
	}
}

// TestLeBudgetSurvitAuConvoiteAbandonne vérifie que ce qui a été épargné n'est
// pas perdu.
//
// Le budget ne se perd que par les plafonds, comme avant ce lot : abandonner un
// convoité devenu impayable ne doit pas coûter ce qu'on avait mis de côté, sans
// quoi un plafond atteint au mauvais moment effacerait plusieurs secondes de
// pression.
func TestLeBudgetSurvitAuConvoiteAbandonne(t *testing.T) {
	w, profils := salleOuverte(t, nil, 64)
	marcheur := indexDuProfil(t, profils, "marcheur")
	w.scenario = vagueUnique(3, marcheur)

	// Un convoité posé à la main sur un profil que la phase n'autorise pas : il
	// sera abandonné au premier tick, et le budget doit rester.
	w.budget = FromInt(2)
	w.convoite = indexDuProfil(t, profils, "bloqueur") + 1
	w.Step(Vec{})

	if w.convoite-1 == indexDuProfil(t, profils, "bloqueur") {
		t.Error("le convoité hors phase a survécu au tick")
	}
	if w.budget < FromInt(2) {
		t.Errorf("budget %v après abandon, attendu au moins les deux mis de côté",
			w.budget)
	}
}
