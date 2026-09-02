// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de l'aimant : le manifeste livré, l'apparition à distance, la charge
// unique, la ruée des gemmes et ce qu'elle sauve de l'effacement.

package game

import "testing"

// champDAimant monte une salle vide sur les réglages d'aimant qu'on lui donne.
//
// La collecte est complétée comme ailleurs, sauf ce que le cas règle lui-même :
// ces cas parlent d'aimants, et une gemme qui s'effacerait au milieu brouillerait
// ce qu'ils mesurent.
func champDAimant(t *testing.T, regler func(*Progression)) (*World, *Profiles) {
	t.Helper()
	seuils := collecte(&Progression{FirstThreshold: 1000, Floor: 100000, GemValue: 1})
	regler(seuils)
	return champDeProgression(t, seuils)
}

// TestManifesteLivreDonneLAimant charge les réglages publiés sans rien injecter.
func TestManifesteLivreDonneLAimant(t *testing.T) {
	p := progressionLivree(t)

	// 30 000 ms à 60 ticks par seconde. La valeur est écrite en clair : un test
	// qui refait la conversion du code passe même quand les deux sont faux.
	if p.MagnetPeriod != 1800 {
		t.Errorf("période : %d ticks, attendu 1800", p.MagnetPeriod)
	}
	if p.MagnetMinRange != FromInt(6) {
		t.Errorf("distance minimale : %d, attendu %d", p.MagnetMinRange, FromInt(6))
	}
	// 12 tuiles par seconde converties au pas de simulation, comme toutes les
	// vitesses du jeu.
	if p.PullSpeed != 13107 {
		t.Errorf("vitesse d'attraction : %d, attendu 13107", p.PullSpeed)
	}
}

// TestLAimantApparaitLoinDuJoueur garde ce qui en fait un trajet.
//
// Un aimant posé sous les pieds se ramasserait sans qu'on ait bougé, donc sans
// décision — et la conception en fait précisément une décision. La contrainte est
// d'être loin, non d'être hors du champ : on doit le voir pour vouloir y aller.
func TestLAimantApparaitLoinDuJoueur(t *testing.T) {
	w, _ := champDAimant(t, func(p *Progression) {
		p.MagnetPeriod = 1
		p.MagnetMinRange = FromInt(6)
	})

	// La période se compte depuis le montage : rien n'apparaît au premier tick,
	// ce qui serait un cadeau avant que quoi que ce soit se soit passé.
	w.Step(Vec{})
	if w.Magnets().Len() != 0 {
		t.Fatal("un aimant est apparu au premier tick")
	}

	w.Step(Vec{})
	if w.Magnets().Len() != 1 {
		t.Fatalf("%d aimant(s) au sol, attendu 1", w.Magnets().Len())
	}

	a := w.Magnets().At(0)
	px, py := w.Player()
	mini := int64(FromInt(6))
	if d := (Vec{X: a.X - px, Y: a.Y - py}).carres(); d < mini*mini {
		t.Errorf("aimant posé à moins de six tuiles du joueur")
	}
}

// TestUnSeulAimantALaFois garde la règle qui produit la tension.
//
// Un aimant qu'on voit sans pouvoir le prendre, parce qu'on en tient déjà un, est
// une raison de dépenser celui qu'on garde. Les laisser s'accumuler retirerait
// cette raison, et l'objet cesserait d'être un événement.
func TestUnSeulAimantALaFois(t *testing.T) {
	w, _ := champDAimant(t, func(p *Progression) {
		p.MagnetPeriod = 1
		p.MagnetMinRange = FromInt(6)
	})

	for range 20 {
		w.Step(Vec{})
	}
	if w.Magnets().Len() != 1 {
		t.Errorf("%d aimant(s) au sol après vingt ticks, attendu 1", w.Magnets().Len())
	}
}

// TestLAimantSeRamasseEtSeGarde éprouve la charge.
//
// Il ne se déclenche pas au contact : c'est ce qui en fait une décision plutôt
// qu'un cadeau, et toute la tension de l'objet en dépend.
func TestLAimantSeRamasseEtSeGarde(t *testing.T) {
	w, _ := champDAimant(t, func(p *Progression) {
		p.MagnetPeriod = 1
		p.MagnetMinRange = 0
	})

	// L'aimant est déposé sous le joueur : ce que ce cas mesure est le ramassage
	// et la charge, pas la distance, qui a son propre test.
	px, py := w.Player()
	w.Magnets().Spawn(Magnet{X: px, Y: py})
	w.Step(Vec{})

	if w.Magnets().Len() != 0 {
		t.Error("l'aimant est resté au sol alors que le joueur était dessus")
	}
	if !w.Charged() {
		t.Error("le joueur n'a pas gardé la charge")
	}
}

// TestLaRueeRameneLesGemmes éprouve ce que l'aimant fait, et ce qu'il coûte.
//
// La gemme est semée hors de portée : sans cela, elle serait ramassée en marchant
// et le cas ne dirait rien de l'attraction.
func TestLaRueeRameneLesGemmes(t *testing.T) {
	w, profils := champDAimant(t, func(p *Progression) {
		p.MagnetPeriod = 100000
		p.PullSpeed = One / 4
	})
	w.charge = true

	px, py := w.Player()
	e := Enemy{Profile: indexDuProfil(t, profils, "marcheur"), X: px + FromInt(5), Y: py}
	w.lacher(&e)

	w.Attract()
	if w.Charged() {
		t.Error("la charge n'a pas été dépensée")
	}

	// Cinq tuiles à un quart de tuile par tick : une vingtaine de ticks, et la
	// passe de ramassage la prend à l'arrivée.
	for range 30 {
		w.Step(Vec{})
	}
	if w.Experience() != 1 {
		t.Errorf("expérience : %d, attendu 1 : la gemme devait rejoindre le joueur",
			w.Experience())
	}
}

// TestUneGemmeAttireeNeSEteintPas garde ce qui empêche l'aimant d'échouer.
//
// **C'est le recours contre l'effacement, il ne peut pas en être la victime.**
// Une gemme qui continuerait de vieillir en vol disparaîtrait avant d'arriver
// quand on déclenche sur un tas ancien — c'est-à-dire dans le cas exact où l'on
// dépense sa charge.
func TestUneGemmeAttireeNeSEteintPas(t *testing.T) {
	w, profils := champDAimant(t, func(p *Progression) {
		p.MagnetPeriod = 100000
		p.GemLife = 4
		// Assez lente pour que la gemme soit encore en vol bien après l'âge où
		// elle aurait dû s'éteindre.
		p.PullSpeed = One / 16
	})
	w.charge = true

	px, py := w.Player()
	e := Enemy{Profile: indexDuProfil(t, profils, "marcheur"), X: px + FromInt(5), Y: py}
	w.lacher(&e)
	w.Attract()

	for range 10 {
		w.Step(Vec{})
	}
	if w.Gems().Len() != 1 {
		t.Fatalf("%d gemme(s) après dix ticks, attendu 1 encore en vol", w.Gems().Len())
	}
	if w.Experience() != 0 {
		t.Fatalf("expérience : %d, la gemme est arrivée trop tôt et le cas ne dit "+
			"plus rien de l'effacement", w.Experience())
	}
}

// TestSansChargeLaRueeNaPasLieu garde le silence de l'appel à vide.
//
// L'appelant est un clavier : une touche pressée sans charge ne doit ni
// consommer, ni avertir, ni saisir les gemmes. L'emplacement vide à l'écran
// suffit à le dire.
func TestSansChargeLaRueeNaPasLieu(t *testing.T) {
	w, profils := champDAimant(t, func(p *Progression) {
		p.MagnetPeriod = 100000
		p.PullSpeed = One / 4
	})

	px, py := w.Player()
	e := Enemy{Profile: indexDuProfil(t, profils, "marcheur"), X: px + FromInt(5), Y: py}
	w.lacher(&e)

	w.Attract()
	for range 30 {
		w.Step(Vec{})
	}
	if w.Experience() != 0 {
		t.Errorf("expérience : %d, attendu 0 sans charge", w.Experience())
	}
}
