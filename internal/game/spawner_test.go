// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du spawner : le budget qui achète à l'heure, l'anneau qui pose hors de
// portée, et les trois façons de ne rien poser — dont deux perdent le budget et
// une le reporte.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// salleOuverte monte un monde vide sur une salle assez large pour porter
// l'anneau d'apparition.
//
// Quarante-huit cases : le rayon publié vaut dix-neuf tuiles, et le joueur posé
// au centre doit avoir de la place dans les quatre directions. Une salle plus
// étroite ferait échouer l'anneau, ce qui est le sujet d'un autre cas et non le
// décor de tous.
//
// L'arme est inerte : ces cas comptent des apparitions, et un joueur qui abat ce
// qu'on vient de poser compterait autre chose.
func salleOuverte(t *testing.T, scenario *Scenario, capacite int) (*World, *Profiles) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	g := NewCostGrid(48, 48)
	w := NewWorld(profils, armesInertes(t), progressionLivree(t), scenario, g, graineDeTest,
		capacite, 8, 8)
	w.Place(FromInt(24)+One/2, FromInt(24)+One/2)
	return w, profils
}

// vagueUnique rend un scénario d'une phase, sans pointe.
func vagueUnique(pression int, profils ...int) *Scenario {
	return &Scenario{Phases: []Phase{{
		Pressure: parTick(float64(pression)),
		Profiles: profils,
	}}}
}

// TestLeBudgetAchèteALHeure vérifie que la pression s'accumule au lieu d'acheter
// tout de suite ou jamais.
//
// **Le chiffre est le sujet du cas.** Trois de pression par seconde et un Badaud
// à trois : la première créature coûte une seconde de budget, et rien avant.
// Un spawner qui poserait à chaque tick passerait le second relevé, un spawner
// qui n'accumulerait pas passerait le premier ; il faut les deux.
func TestLeBudgetAchèteALHeure(t *testing.T) {
	w, profils := salleOuverte(t, nil, 16)
	w.scenario = vagueUnique(3, indexDuProfil(t, profils, "marcheur"))

	for range TPS - 1 {
		w.Step(Vec{})
	}
	if n := w.Enemies().Len(); n != 0 {
		t.Errorf("%d créature(s) avant la seconde, la première coûte trois de budget", n)
	}

	for range 2 {
		w.Step(Vec{})
	}
	if n := w.Enemies().Len(); n != 1 {
		t.Errorf("%d créature(s) après la seconde, attendu une", n)
	}
}

// TestLApparitionSePoseSurLAnneau vérifie que la créature naît à la distance
// publiée, et non sur le joueur ni au hasard de la salle.
//
// **C'est la seule chose qui tient la règle « jamais dans le champ de vision ».**
// La simulation ne connaît pas l'écran ; ce qu'elle peut garantir est la
// distance, et le manifeste dit d'où vient le chiffre.
//
// Un seul tick, et une pression qui paie un Badaud à chaque tick : la mesure
// doit se prendre à l'apparition, une créature s'approchant de trois tuiles par
// seconde dès qu'on la laisse vivre.
func TestLApparitionSePoseSurLAnneau(t *testing.T) {
	w, profils := salleOuverte(t, nil, 16)
	w.scenario = vagueUnique(3*TPS, indexDuProfil(t, profils, "marcheur"))

	w.Step(Vec{})
	if w.Enemies().Len() == 0 {
		t.Fatal("aucune créature posée")
	}

	rayon := w.progression.SpawnRadius
	for i := range w.Enemies().Active() {
		e := w.Enemies().At(i)
		distance := (Vec{X: e.X - w.playerX, Y: e.Y - w.playerY}).Len()
		if distance < rayon-One/4 || distance > rayon+One/4 {
			t.Errorf("créature %d à %v du joueur, l'anneau est à %v", i, distance, rayon)
		}
	}
}

// TestLAnneauBoucheReporteSonBudget vérifie qu'un abri ne fait pas tomber la
// pression à zéro.
//
// Sans report, un couloir étroit devient un endroit où l'on ne risque rien, ce
// que la conception cherche précisément à éviter. Le cas mure la salle, laisse
// le budget s'accumuler pour rien, puis l'ouvre : ce qui apparaît alors est ce
// qui avait été mis de côté.
func TestLAnneauBoucheReporteSonBudget(t *testing.T) {
	w, profils := salleOuverte(t, nil, 16)
	w.scenario = vagueUnique(3, indexDuProfil(t, profils, "marcheur"))
	murer(w)

	for range 2 * TPS {
		w.Step(Vec{})
	}
	if n := w.Enemies().Len(); n != 0 {
		t.Fatalf("%d créature(s) dans une salle murée", n)
	}

	degager(w)
	w.Step(Vec{})
	if n := w.Enemies().Len(); n == 0 {
		t.Error("rien n'apparaît à la réouverture : le budget a été perdu au lieu d'être reporté")
	}
}

// TestLeReportEstBorne vérifie que se terrer coûte du temps plutôt que de
// produire une punition différée.
//
// **Sans borne, le report est un mur d'ennemis à retardement** — exactement ce
// que la règle « jamais dans le champ de vision » interdit, avec en prime
// l'impression que le jeu triche. Dix secondes murées valent trente de budget ;
// la borne publiée en garde trois secondes, soit trois Badauds.
func TestLeReportEstBorne(t *testing.T) {
	w, profils := salleOuverte(t, nil, 64)
	w.scenario = vagueUnique(3, indexDuProfil(t, profils, "marcheur"))
	murer(w)

	for range 10 * TPS {
		w.Step(Vec{})
	}
	degager(w)
	w.Step(Vec{})

	if n := w.Enemies().Len(); n > 3 {
		t.Errorf("%d créature(s) d'un coup, la borne en autorise trois", n)
	}
}

// TestLePlafondDEffectifPerdSonBudget vérifie que le budget refusé par le
// plafond ne revient pas plus tard.
//
// **C'est l'autre moitié du report, et elle va dans l'autre sens.** L'anneau
// bouché est une contrariété passagère dont la pression ne doit pas pâtir ; le
// plafond d'effectif est une limite tenue, et lui reporter son budget rendrait
// chaque mort suivie d'une bouffée que rien n'explique.
func TestLePlafondDEffectifPerdSonBudget(t *testing.T) {
	w, profils := salleOuverte(t, nil, 2)
	w.scenario = vagueUnique(3, indexDuProfil(t, profils, "marcheur"))

	for range 60 * TPS {
		w.Step(Vec{})
	}
	if n := w.Enemies().Len(); n != 2 {
		t.Fatalf("%d créature(s) pour un bassin de deux", n)
	}

	w.Enemies().RemoveAt(0)
	w.Enemies().RemoveAt(0)
	w.Step(Vec{})
	if n := w.Enemies().Len(); n != 0 {
		t.Errorf("%d créature(s) au tick suivant la place libérée : soixante secondes "+
			"de budget avaient été gardées", n)
	}
}

// TestLePlafondDeSimultaneiteTientMalgreLeBudget vérifie qu'un profil rare le
// reste quand le budget pourrait en payer vingt.
//
// Le Secouriste ne vaut rien seul et double la difficulté au milieu de vingt
// Badauds : sa rareté ne peut pas se régler par son prix, trop bas il
// déséquilibre et trop haut sa mécanique ne s'apprend jamais.
func TestLePlafondDeSimultaneiteTientMalgreLeBudget(t *testing.T) {
	w, profils := salleOuverte(t, nil, 32)
	soigneur := indexDuProfil(t, profils, "soigneur")
	if profils.Enemies[soigneur].MaxAlive != 1 {
		t.Fatalf("le Secouriste plafonne à %d, ce cas en attend un",
			profils.Enemies[soigneur].MaxAlive)
	}
	w.scenario = vagueUnique(120, soigneur)

	for range 10 * TPS {
		w.Step(Vec{})
	}
	if n := w.Enemies().Len(); n != 1 {
		t.Errorf("%d Secouriste(s) vivants, son profil en plafonne un", n)
	}
}

// TestLaPhaseEnVigueurEstLaDerniereCommencee vérifie qu'un palier vaut jusqu'à
// ce que le suivant le remplace.
//
// La dernière phase n'a donc pas de fin à écrire, ce qui évite au format un
// champ que tout auteur oublierait sur la dernière ligne de sa courbe.
func TestLaPhaseEnVigueurEstLaDerniereCommencee(t *testing.T) {
	s := &Scenario{Phases: []Phase{
		{Start: 0, Pressure: 1},
		{Start: 10 * TPS, Pressure: 2},
		{Start: 20 * TPS, Pressure: 3},
	}}

	for _, c := range []struct {
		tick Tick
		veut Fixed
	}{
		{0, 1},
		{10*TPS - 1, 1},
		{10 * TPS, 2},
		{20 * TPS, 3},
		{600 * TPS, 3},
	} {
		if a := s.phase(c.tick).Pressure; a != c.veut {
			t.Errorf("au tick %d : pression %v, attendu %v", c.tick, a, c.veut)
		}
	}
}

// TestLaPointeMultipliePuisRelache vérifie que la pointe borne sa fenêtre des
// deux côtés.
//
// Une pointe qui ne relâcherait pas serait un palier de plus, écrit là où on
// croit lire un événement.
func TestLaPointeMultipliePuisRelache(t *testing.T) {
	p := Phase{Pressure: 100, Peak: Peak{At: 10 * TPS, Until: 12 * TPS, Pressure: 300}}

	for _, c := range []struct {
		tick Tick
		veut Fixed
	}{
		{10*TPS - 1, 100},
		{10 * TPS, 300},
		{12*TPS - 1, 300},
		{12 * TPS, 100},
	} {
		if a := p.budget(c.tick); a != c.veut {
			t.Errorf("au tick %d : budget %v, attendu %v", c.tick, a, c.veut)
		}
	}
}

// murer rend toute la salle infranchissable, sauf la case du joueur.
//
// La case du joueur reste ouverte parce que le déplacement et le champ de flux
// la lisent ; ce que le cas veut boucher est l'anneau, pas le monde entier.
func murer(w *World) {
	for v := range w.grille.Height() {
		for u := range w.grille.Width() {
			w.grille.Set(u, v, Blocked)
		}
	}
	w.grille.Set(w.playerX.Floor(), w.playerY.Floor(), Free)
}

// degager rouvre la salle entière.
func degager(w *World) {
	for v := range w.grille.Height() {
		for u := range w.grille.Width() {
			w.grille.Set(u, v, Free)
		}
	}
}

// TestUneBorneNeTuePasUnePhase garde le cas où la borne de report tombe sous le
// prix d'une créature.
//
// **Il est né du réglage de la courbe et d'aucune relecture.** Une pression d'un
// par seconde et un report de trois secondes plafonnent le budget au prix exact
// d'un Badaud, que l'arrondi de la conversion par tick place un millième en
// dessous : le budget monte, bute sur la borne, et la phase n'achète jamais rien.
// Rien ne le signalait — pas un refus au chargement, pas une erreur, une salle
// simplement vide.
//
// Ce que le cas garde n'est donc pas la borne mais son plancher : elle limite
// l'accumulation, elle ne l'arrête pas.
func TestUneBorneNeTuePasUnePhase(t *testing.T) {
	w, profils := salleOuverte(t, nil, 16)
	marcheur := indexDuProfil(t, profils, "marcheur")
	w.scenario = vagueUnique(1, marcheur)
	w.scenario.Phases[0].Cheapest = FromInt(profils.Enemies[marcheur].PressureCost)

	// Trois secondes de report pour trois de budget par seconde : le plafond et
	// le prix se touchent, et c'est là que le cas se joue.
	if plafond := parTick(1) * Fixed(w.progression.CarryOver); plafond >= FromInt(3) {
		t.Fatalf("le plafond vaut %v pour un prix de %v : le cas n'est plus sur l'arête",
			plafond, FromInt(3))
	}

	for range 5 * TPS {
		w.Step(Vec{})
	}
	if n := w.Enemies().Len(); n == 0 {
		t.Error("aucune créature en cinq secondes : la borne empêche tout achat")
	}
}
