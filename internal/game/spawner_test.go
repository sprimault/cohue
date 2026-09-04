// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du spawner : le budget qui achète à l'heure, la meute qui n'arrive ni
// seule ni rognée, l'anneau qui pose hors de portée, et les trois façons de ne
// rien poser — dont deux perdent le budget et une le reporte.

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
		Capacities{Enemies: capacite, Shots: 8, EnemyShots: 8, Blasts: 8, Gems: 8})
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

// meuteDeMolosses monte une salle où le Molosse est le seul profil achetable, et
// vérifie que son groupe vaut bien trois.
//
// Le contrôle du groupe n'est pas une précaution : les trois cas qui suivent
// perdraient leur sujet si le manifeste ramenait le Molosse à un, et ils
// passeraient tous les trois sans rien garder.
func meuteDeMolosses(t *testing.T, pression, capacite int) (*World, *EnemyProfile) {
	t.Helper()
	w, profils := salleOuverte(t, nil, capacite)
	sprinteur := indexDuProfil(t, profils, "sprinteur")
	molosse := &profils.Enemies[sprinteur]
	if molosse.Group != 3 {
		t.Fatalf("le Molosse arrive par %d, ces cas en attendent trois", molosse.Group)
	}
	w.scenario = vagueUnique(pression, sprinteur)
	return w, molosse
}

// TestLaMeuteApparaitEntiere vérifie que le Molosse n'arrive jamais seul.
//
// Trois qui chargent en décalé sont ce qui oblige à cesser de reculer en ligne
// droite ; un chien isolé se contourne, et la conception range donc la taille de
// groupe dans le profil plutôt que dans une exception du spawner.
//
// Le cas relève au premier tick qui pose quelque chose, et non après une durée
// fixe : un compte de ticks tomberait entre deux achats et rendrait six.
func TestLaMeuteApparaitEntiere(t *testing.T) {
	w, molosse := meuteDeMolosses(t, 12, 16)

	for range 5 * TPS {
		w.Step(Vec{})
		if w.Enemies().Len() > 0 {
			break
		}
	}
	if n := w.Enemies().Len(); n != molosse.Group {
		t.Errorf("%d créature(s) à la première apparition, la meute en compte %d",
			n, molosse.Group)
	}
}

// TestLaMeuteArriveDUnSeulCote vérifie qu'une seule position est tirée pour
// toute la meute.
//
// Un tirage par membre les ferait naître aux quatre coins de l'anneau : trois
// créatures du même profil, et plus une meute. Elles partent donc du même point
// et la séparation les écarte ensuite, ce qui borne l'écart d'un tick à un pas
// de course — très en deçà des tuiles que deux directions indépendantes
// mettraient entre elles sur un anneau de dix-neuf.
func TestLaMeuteArriveDUnSeulCote(t *testing.T) {
	w, _ := meuteDeMolosses(t, 12, 16)

	for range 5 * TPS {
		w.Step(Vec{})
		if w.Enemies().Len() > 0 {
			break
		}
	}

	premier := w.Enemies().At(0)
	for i := range w.Enemies().Active() {
		e := w.Enemies().At(i)
		ecart := (Vec{X: e.X - premier.X, Y: e.Y - premier.Y}).Len()
		if ecart > One/2 {
			t.Errorf("créature %d à %v de la première : la meute a été tirée membre "+
				"par membre", i, ecart)
		}
	}
}

// TestUneMeuteSePaieEntiere vérifie que le prix d'un achat est celui de la
// meute, le coût du manifeste étant unitaire.
//
// **Le premier relevé est celui qui discrimine.** Six de pression par seconde
// paient un Molosse en une seconde et la meute en deux : un spawner qui
// facturerait le prix unitaire aurait déjà posé trois chiens au premier relevé,
// et les deux implémentations rendraient la même chose au second.
func TestUneMeuteSePaieEntiere(t *testing.T) {
	w, molosse := meuteDeMolosses(t, 6, 16)
	if molosse.PressureCost != 4 {
		t.Fatalf("le Molosse coûte %d, ce cas en attend quatre", molosse.PressureCost)
	}

	for range TPS {
		w.Step(Vec{})
	}
	if n := w.Enemies().Len(); n != 0 {
		t.Errorf("%d créature(s) après une seconde : la meute a été payée au prix "+
			"d'un seul chien", n)
	}

	for range 3 * TPS / 2 {
		w.Step(Vec{})
	}
	if n := w.Enemies().Len(); n != molosse.Group {
		t.Errorf("%d créature(s) après deux secondes et demie, attendu %d",
			n, molosse.Group)
	}
}

// TestUneMeuteNeSeRognePasSurLaPlaceRestante vérifie qu'un bassin presque plein
// ne fait pas apparaître un Molosse seul.
//
// C'est l'endroit exact où l'exception s'écrirait — il reste une place, le
// budget est là, et poser un chien plutôt que rien paraît une bonne affaire. Le
// bassin s'arrête donc à un multiple de la meute, et le budget refusé est perdu
// comme celui du plafond d'effectif.
func TestUneMeuteNeSeRognePasSurLaPlaceRestante(t *testing.T) {
	const capacite = 16
	w, molosse := meuteDeMolosses(t, 120, capacite)

	for range 60 * TPS {
		w.Step(Vec{})
	}

	veut := capacite - capacite%molosse.Group
	if n := w.Enemies().Len(); n != veut {
		t.Errorf("%d créature(s) dans un bassin de %d, attendu %d : la dernière "+
			"meute a été rognée", n, capacite, veut)
	}
}

// TestUnPlafondDeSimultaneiteSeCompteEnMeutes vérifie qu'un plafond ne se
// franchit pas par le bas d'une meute.
//
// Aucun profil livré ne porte les deux à la fois — le Secouriste plafonne mais
// arrive seul, le Molosse arrive à trois sans plafond —, si bien que rien
// n'exercerait la ligne qui les croise. Le plafond est donc posé ici, sur une
// table que le manifeste produirait : sept au-dessus de trois est une écriture
// que le chargeur accepte.
//
// **Sept, et non six.** Un plafond multiple de la meute laisse les deux
// arithmétiques tomber d'accord, et le cas passerait sous celle qui compare le
// nombre de vivants avant l'achat plutôt qu'après.
func TestUnPlafondDeSimultaneiteSeCompteEnMeutes(t *testing.T) {
	w, molosse := meuteDeMolosses(t, 120, 32)
	molosse.MaxAlive = 7

	for range 10 * TPS {
		w.Step(Vec{})
	}

	veut := molosse.MaxAlive - molosse.MaxAlive%molosse.Group
	if n := w.Enemies().Len(); n != veut {
		t.Errorf("%d Molosse(s) vivants pour un plafond de %d, attendu %d : la "+
			"dernière meute a franchi le plafond", n, molosse.MaxAlive, veut)
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
	w.scenario.Phases[0].Cheapest = profils.Enemies[marcheur].PackCost()

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
