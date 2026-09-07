// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du tir : l'arme prête qui reste prête, le plus proche visé quel que
// soit l'ordre du bassin, la créature qui en sort à l'instant de sa mort, et le
// projectile qui disparaît au bout de sa portée.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// champDeTir monte une salle vide, le joueur au centre, avec l'arme livrée.
func champDeTir(t *testing.T) (*World, *Profiles) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("armes livrées : %v", err)
	}

	g := NewCostGrid(32, 32)
	for u := range 32 {
		g.Set(u, 0, Blocked)
		g.Set(u, 31, Blocked)
	}
	for v := range 32 {
		g.Set(0, v, Blocked)
		g.Set(31, v, Blocked)
	}

	w := NewWorld(profils, armes, progressionLivree(t), sansVagues(), g, graineDeTest, capacitesDeTest)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)
	return w, profils
}

// champSansTir monte la même salle, mais avec une arme inerte : les tests qui
// éprouvent la mort ne doivent pas voir le joueur abattre leurs cobayes.
func champSansTir(t *testing.T) (*World, *Profiles) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	g := NewCostGrid(32, 32)
	w := NewWorld(profils, armesInertes(t), progressionLivree(t), sansVagues(), g, graineDeTest, capacitesDeTest)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)
	return w, profils
}

// TestSansCibleLArmeNeConsommeRien éprouve le cas limite que la conception nomme.
//
// Si la cadence se consommait à vide, le joueur qui sort d'un couloir désert
// tirerait sa première salve avec un retard fonction du temps passé sans rien à
// viser. Rien à l'écran ne l'expliquerait, et le comportement de l'arme
// dépendrait du passé récent.
func TestSansCibleLArmeNeConsommeRien(t *testing.T) {
	w, profils := champDeTir(t)

	// Un compte volontairement premier avec la cadence : cent ticks tombaient
	// pile sur un multiple de vingt-cinq, si bien qu'une arme qui se consomme à
	// vide se retrouvait prête au moment de la mesure — et le test passait par
	// coïncidence.
	for range 101 {
		w.Step(Vec{})
	}
	if n := w.Shots().Len(); n != 0 {
		t.Fatalf("%d projectile(s) en vol sans aucune cible", n)
	}

	// Une créature apparaît à portée : le tir doit partir au tick suivant, sans
	// attendre une cadence.
	marcheur := indexDuProfil(t, profils, "marcheur")
	px, py := w.Player()
	if _, ok := w.SpawnEnemy(marcheur, px+FromInt(3), py); !ok {
		t.Fatal("créature refusée")
	}

	w.Step(Vec{})
	if n := w.Shots().Len(); n != 1 {
		t.Errorf("%d projectile(s) au premier tick avec cible, attendu 1 : "+
			"l'arme n'était pas prête", n)
	}
}

// TestLeTirViseLePlusProche fixe le ciblage.
//
// La visée est omnidirectionnelle et le joueur ne choisit pas : c'est ce qui
// donnera son rôle au Secouriste à l'étape 4, dont le seul moyen de se
// débarrasser est d'aller vers lui. Un ciblage qui prendrait n'importe quelle
// créature à portée le désactiverait par avance, et rien ici ne le dirait.
func TestLeTirViseLePlusProche(t *testing.T) {
	// Les deux cibles sont dans des directions perpendiculaires, et non alignées
	// : alignées, un projectile visant la lointaine traverse la proche et la
	// touche au passage, si bien que le test passe quel que soit le ciblage.
	//
	// Les deux ordres de pose, pour la même raison prise ailleurs : un ciblage
	// qui retiendrait la dernière créature rencontrée à portée désignerait la
	// bonne dans l'un des deux cas.
	for _, cas := range []struct {
		quoi         string
		procheDAbord bool
	}{
		{"la proche posée en premier", true},
		{"la proche posée en second", false},
	} {
		t.Run(cas.quoi, func(t *testing.T) {
			w, profils := champDeTir(t)
			px, py := w.Player()
			marcheur := indexDuProfil(t, profils, "marcheur")

			poser := func(dx, dy Fixed) Handle {
				h, ok := w.SpawnEnemy(marcheur, px+dx, py+dy)
				if !ok {
					t.Fatal("créature refusée")
				}
				return h
			}

			var pres, loin Handle
			if cas.procheDAbord {
				pres = poser(FromInt(2), 0)
				loin = poser(0, FromInt(5))
			} else {
				loin = poser(0, FromInt(5))
				pres = poser(FromInt(2), 0)
			}

			for range 10 * TPS {
				w.Step(Vec{})
				switch {
				case !w.Enemies().Alive(pres) && !w.Enemies().Alive(loin):
					t.Fatal("les deux sont mortes dans le même tick, le cas ne dit rien")
				case !w.Enemies().Alive(loin):
					t.Fatal("la créature à cinq tuiles est morte avant celle à deux")
				case !w.Enemies().Alive(pres):
					return
				}
			}
			t.Fatal("aucune des deux n'est morte")
		})
	}
}

// TestLaCadenceEspaceLesTirs vérifie que l'arme ne tire pas à chaque tick une
// fois qu'elle a une cible.
func TestLaCadenceEspaceLesTirs(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	// Un Vigile : douze touches, il survivra à la mesure.
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "bloqueur"), px+FromInt(3), py); !ok {
		t.Fatal("créature refusée")
	}

	tirs := 0
	precedent := 0
	for range 3 * 24 {
		w.Step(Vec{})
		if n := w.Shots().Len(); n > precedent {
			tirs++
		}
		precedent = w.Shots().Len()
	}
	if tirs != 3 {
		t.Errorf("%d tir(s) en trois cadences, attendu 3", tirs)
	}
}

// TestLeTirTueEtLaCreatureQuitteLeBassin éprouve la mort comme état.
//
// La résistance du Quidam vaut trois touches de l'arme de base : trois tirs, et
// il n'est plus dans le bassin. Aucun drapeau, aucune liste de morts en attente.
// salveDe joue un tick de tir sur une cible unique et rend les projectiles nés.
//
// Le nombre est posé sur l'arme de la partie plutôt que gagné par une carte : ce
// qu'on isole ici est la géométrie de la salve, et le chemin qui l'accroît est
// gardé ailleurs, par `TestChoisirAppliqueLePalierDeProjectiles`.
func salveDe(t *testing.T, nombre int) []Projectile {
	t.Helper()
	w, profils := champDeTir(t)
	w.arme.Projectiles = nombre

	px, py := w.Player()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(3), py); !ok {
		t.Fatal("créature refusée")
	}
	w.Step(Vec{})

	if vol := len(w.tirs.Active()); vol != nombre {
		t.Fatalf("%d projectile(s) en vol, attendu %d", vol, nombre)
	}
	salve := make([]Projectile, nombre)
	for i := range nombre {
		salve[i] = *w.tirs.At(i)
	}
	return salve
}

// TestLeFrontEcarteSansTournerLaCourse garde ce qui sépare le front de
// l'éventail.
//
// **Un front n'est pas un cône** : les projectiles se décalent perpendiculairement
// à la visée et gardent tous la même course. C'est ce qui laisse à l'axe
// `eventail` quelque chose à apporter, et à la recette « trois projectiles plus
// éventail » deux termes à combiner — si le nombre écartait déjà les courses, la
// synergie n'aurait rien à faire.
func TestLeFrontEcarteSansTournerLaCourse(t *testing.T) {
	salve := salveDe(t, 3)

	for i := 1; i < len(salve); i++ {
		if salve[i].Step != salve[0].Step {
			t.Errorf("projectile %d : pas %v, attendu %v", i, salve[i].Step, salve[0].Step)
		}
	}
	if salve[0].X == salve[1].X && salve[0].Y == salve[1].Y {
		t.Error("deux projectiles au même point : la salve est confondue")
	}
}

// TestLeFrontGardeSaLargeurQuandLeNombreMonte garde la décision de structure.
//
// **C'est la largeur qui se règle, et l'écartement qui s'en dérive.** Éprouvé
// contre l'implémentation qu'on écarterait naturellement — un écartement fixe
// entre voisins —, qui rendrait ici deux tuiles à trois projectiles et six à
// sept, au lieu d'une seule à chaque fois. On découvrirait alors au sixième
// palier que les extrêmes tirent à plusieurs tuiles de la visée, et l'on
// corrigerait l'écartement pour tous les paliers d'un coup.
func TestLeFrontGardeSaLargeurQuandLeNombreMonte(t *testing.T) {
	// Sept est le dernier palier de l'axe, celui où un écartement fixe aurait
	// coûté le plus cher.
	for _, nombre := range []int{2, 3, 7} {
		salve := salveDe(t, nombre)
		premier, dernier := salve[0], salve[len(salve)-1]
		largeur := Vec{X: dernier.X - premier.X, Y: dernier.Y - premier.Y}.Len()

		// Quelques unités de virgule fixe : la division par le nombre d'écarts
		// ne tombe pas juste, et une tolérance nulle mesurerait l'arrondi.
		if ecart := largeur - One; ecart.Abs() > 16 {
			t.Errorf("%d projectiles : front de %v, attendu une tuile", nombre, largeur)
		}
	}
}

// restantDe rend la résistance d'une créature, ou zéro si elle est morte.
//
// Par le `Handle` et non par la place : celle-ci change à chaque suppression par
// échange, et c'est précisément ce que les cas de perforation produisent.
func restantDe(w *World, h Handle) int {
	place, vivante := w.Enemies().Slot(h)
	if !vivante {
		return 0
	}
	return w.Enemies().At(place).Hits
}

// TestLaPerforationTraverseEtPoursuit garde ce que l'axe apporte, et ce qu'il
// n'apporte pas sans palier.
//
// Deux cas, et le second est ce qui rend le premier concluant : sans perforation
// le tir s'arrête sur la première créature, si bien qu'un test qui ne montrerait
// que le cas perforant passerait aussi sur un code qui traverse toujours.
func TestLaPerforationTraverseEtPoursuit(t *testing.T) {
	for _, cas := range []struct {
		quoi     string
		pierce   int
		derriere bool
	}{
		{"sans palier, le tir s'arrête sur la première", 0, false},
		{"avec un palier, il traverse et touche la suivante", 1, true},
	} {
		t.Run(cas.quoi, func(t *testing.T) {
			w, profils := champDeTir(t)
			px, py := w.Player()
			marcheur := indexDuProfil(t, profils, "marcheur")

			devant, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py)
			if !ok {
				t.Fatal("créature refusée")
			}
			derriere, ok := w.SpawnEnemy(marcheur, px+FromInt(2)+One/4, py)
			if !ok {
				t.Fatal("créature refusée")
			}
			plein := restantDe(w, derriere)

			// Le segment couvre les deux : ce qu'on éprouve est le sort du
			// projectile après la première, jamais la portée de son pas.
			tir := Projectile{
				X: px + FromInt(2) - One/16, Y: py,
				Step:      Vec{X: One/2 + One/4},
				Remaining: FromInt(6),
				Hits:      1,
				Pierce:    cas.pierce,
			}
			consomme := w.toucher(Vec{tir.X, tir.Y}, &tir)

			if restantDe(w, devant) >= plein {
				t.Error("la première est intacte : le cas ne teste rien")
			}
			if consomme == cas.derriere {
				t.Errorf("projectile consommé = %t avec %d perforation(s)",
					consomme, cas.pierce)
			}
			// Une seule passe : ce que la perforation change est le sort du
			// projectile, et la seconde créature est touchée au segment suivant.
			if reste := tir.Pierce; cas.derriere && reste != 0 {
				t.Errorf("perforations restantes : %d, attendu 0", reste)
			}
		})
	}
}

// TestUnTirNeRetouchePasCeQuIlVientDeTraverser garde la parade à la double
// touche.
//
// **Le cas n'existait pas avant la perforation** : un projectile mourait sur ce
// qu'il frappait, donc il ne repassait jamais. Celui qui survit reste dans le
// segment d'une créature qui a survécu elle aussi — trois touches contre une —,
// et la reprendrait à chaque tick sans la référence qu'il garde.
func TestUnTirNeRetouchePasCeQuIlVientDeTraverser(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()

	cible, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(2), py)
	if !ok {
		t.Fatal("créature refusée")
	}
	plein := restantDe(w, cible)

	tir := Projectile{
		X: px + FromInt(2) - One/16, Y: py,
		Step:      Vec{X: One / 16},
		Remaining: FromInt(6),
		Hits:      1,
		Pierce:    3,
	}
	// Quatre passes sur place : sans la référence, chacune reprendrait la même
	// créature, et le pas est trop court pour la faire sortir du segment.
	for range 4 {
		w.toucher(Vec{tir.X, tir.Y}, &tir)
		tir.X += tir.Step.X
	}

	if perdu := plein - restantDe(w, cible); perdu != 1 {
		t.Errorf("%d touche(s) encaissée(s), attendu 1", perdu)
	}
}

// TestLeRicochetRepartVersUneAutreCible garde les trois décisions de l'axe.
func TestLeRicochetRepartVersUneAutreCible(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")

	frappee, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py)
	if !ok {
		t.Fatal("créature refusée")
	}
	// Hors du segment, sinon elle serait touchée par la course plutôt que par le
	// rebond, et le test passerait sans rien départager.
	voisine, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py+One)
	if !ok {
		t.Fatal("créature refusée")
	}
	plein := restantDe(w, voisine)

	tir := Projectile{
		X: px + FromInt(2) - One/16, Y: py,
		Step:      Vec{X: One / 8},
		Remaining: FromInt(6),
		Hits:      1,
		Bounces:   1,
	}
	if w.toucher(Vec{tir.X, tir.Y}, &tir) {
		t.Fatal("le projectile est consommé : il n'a pas rebondi")
	}
	if restantDe(w, frappee) == 0 {
		t.Fatal("la première n'a rien encaissé : le cas ne teste rien")
	}
	if tir.Bounces != 0 {
		t.Errorf("rebonds restants : %d, attendu 0", tir.Bounces)
	}
	if tir.Step.Y <= 0 {
		t.Errorf("le pas ne s'est pas tourné vers la voisine : %v", tir.Step)
	}
	if plein == 0 {
		t.Fatal("la voisine part morte")
	}
}

// TestLeRebondNeRechargePasLaPortee garde ce que l'axe ne donne pas.
//
// Une cible au-delà de la portée restante n'est pas un rebond possible : le
// projectile meurt là où il a frappé. Sans cette borne il repartirait
// indéfiniment, chaque créature en amenant une autre.
func TestLeRebondNeRechargePasLaPortee(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")

	if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py); !ok {
		t.Fatal("créature refusée")
	}
	if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py+FromInt(5)); !ok {
		t.Fatal("créature refusée")
	}

	tir := Projectile{
		X: px + FromInt(2) - One/16, Y: py,
		Step:      Vec{X: One / 8},
		Remaining: FromInt(1), // moins que l'écart entre les deux créatures
		Hits:      1,
		Bounces:   1,
	}
	if !w.toucher(Vec{tir.X, tir.Y}, &tir) {
		t.Error("le projectile a rebondi vers une cible hors de sa portée restante")
	}
}

// TestLaPerforationPasseAvantLeRicochet garde l'ordre des deux axes.
//
// **L'ordre inverse rendrait le perforant inutile chez qui a les deux** : le
// projectile repartirait vers une autre cible au lieu de traverser, donc il ne
// traverserait jamais. Éprouvé par ce que le projectile devient — il poursuit sa
// course sans avoir dépensé son rebond.
func TestLaPerforationPasseAvantLeRicochet(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")

	if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py); !ok {
		t.Fatal("créature refusée")
	}
	if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py+One); !ok {
		t.Fatal("créature refusée")
	}

	tir := Projectile{
		X: px + FromInt(2) - One/16, Y: py,
		Step:      Vec{X: One / 8},
		Remaining: FromInt(6),
		Hits:      1,
		Pierce:    1,
		Bounces:   1,
	}
	avant := tir.Step
	if w.toucher(Vec{tir.X, tir.Y}, &tir) {
		t.Fatal("le projectile est consommé : ni traversée ni rebond")
	}
	if tir.Pierce != 0 {
		t.Errorf("perforations restantes : %d, attendu 0", tir.Pierce)
	}
	if tir.Bounces != 1 {
		t.Errorf("rebonds dépensés avant la perforation : %d restant(s)", tir.Bounces)
	}
	if tir.Step != avant {
		t.Errorf("la course a été redirigée : %v, attendu %v", tir.Step, avant)
	}
}

func TestLeTirTueEtLaCreatureQuitteLeBassin(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")
	if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py); !ok {
		t.Fatal("créature refusée")
	}
	depart := profils.Enemies[marcheur].Hits

	for range 10 * TPS {
		w.Step(Vec{})
		if w.Enemies().Len() == 0 {
			break
		}
	}
	if n := w.Enemies().Len(); n != 0 {
		t.Fatalf("%d créature(s) vivante(s) après dix secondes de tir sur %d touches",
			n, depart)
	}
}

// TestUnMortCesseDEtreUneCibleSansQuitterLeBassin garde la règle que la
// conception fixe : l'entité morte reste en place jusqu'à la fin du tick, pour
// que les index tiennent, mais un projectile traité plus tard l'ignore.
//
// Le test force une résistance à zéro, ce que le système ne produit qu'au milieu
// d'un tick — c'est le seul moyen d'observer de l'extérieur un état qui, en
// marche normale, ne survit pas à la passe de nettoyage.
//
// `TestLeNettoyageNeLaisseAucunMort` vérifie qu'aucun mort ne reste après un
// `Step`, ce qui est vrai que la garde existe ou non : il ne garde donc pas
// celle-ci. Elle ne se voit qu'en observant le milieu du tick, et c'est ce que
// fait celui-ci.
//
// Ce qu'il garde est le **projectile** : qu'il ne soit pas absorbé par ce qui
// est déjà mort. `TestDeuxProjectilesNeDonnentQuUneVolee` garde l'autre
// conséquence du même filtre, au sol — le butin qui ne repart pas une seconde
// fois. Retirer l'un des deux laisse la moitié de la règle sans épreuve.
func TestUnMortCesseDEtreUneCibleSansQuitterLeBassin(t *testing.T) {
	w, profils := champSansTir(t)
	px, py := w.Player()

	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(2), py); !ok {
		t.Fatal("créature refusée")
	}
	w.Enemies().At(0).Hits = 0

	// Le projectile est posé un pas en arrière de la créature et avance d'un pas :
	// il arrive donc sur elle au moment où `toucher` s'exécute. La poser à un
	// rayon d'elle ne suffirait pas — la créature se déplace plus tôt dans le
	// tick, et le contact se manquerait pour une raison géométrique, sans rapport
	// avec ce que le test annonce garder.
	e := w.Enemies().At(0)
	if _, ok := w.Shots().Spawn(Projectile{
		X: e.X - One/8, Y: e.Y,
		Step:      Vec{One / 8, 0},
		Remaining: FromInt(4),
		Hits:      1,
	}); !ok {
		t.Fatal("projectile refusé")
	}

	w.Step(Vec{})

	if n := w.Shots().Len(); n != 1 {
		t.Errorf("%d projectile(s) en vol, attendu 1 : le tir a été absorbé par une "+
			"créature qui n'était plus une cible", n)
	}
	if n := w.Enemies().Len(); n != 0 {
		t.Errorf("%d créature(s) restante(s) : le nettoyage de fin de tick n'a pas eu lieu", n)
	}
}

// TestLeNettoyageNeLaisseAucunMort vérifie que la passe de fin de tick réexamine
// la place qu'elle libère.
//
// C'est le seul endroit du paquet où la place libérée doit être réexaminée : une
// passe de mise à jour ferait avancer deux fois l'entité remontée, alors que
// celle-ci ne fait que filtrer. La sauter laisserait un mort jusqu'au tick
// suivant, et deux morts adjacents suffisent à le montrer.
//
// Il ne dit rien de ce qui se passe **pendant** le tick : un mort encore en
// place cesse d'être une cible, ce qu'aucune observation de fin de tick ne peut
// voir. C'est `TestUnMortCesseDEtreUneCibleSansQuitterLeBassin` qui le garde.
func TestLeNettoyageNeLaisseAucunMort(t *testing.T) {
	w, profils := champSansTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")

	for i := range 3 {
		if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2+i), py); !ok {
			t.Fatal("créature refusée")
		}
	}
	// Les deux dernières places : la première retirée fait remonter la seconde
	// à l'endroit qu'on vient de vider.
	w.Enemies().At(1).Hits = 0
	w.Enemies().At(2).Hits = 0

	w.Step(Vec{})

	if n := w.Enemies().Len(); n != 1 {
		t.Errorf("%d créature(s) vivante(s), attendu 1", n)
	}
	for i := range w.Enemies().Active() {
		if h := w.Enemies().At(i).Hits; h <= 0 {
			t.Errorf("une créature à %d touche(s) a survécu au nettoyage", h)
		}
	}
}

// TestLeProjectileMeurtAuBoutDeSaPortee vérifie la seconde cause de suppression.
//
// Même chemin que celle du projectile qui touche : deux causes, une seule
// suppression, sinon l'une des deux oublie un jour de libérer sa place — et le
// bassin se remplit de tirs qui n'existent plus.
func TestLeProjectileMeurtAuBoutDeSaPortee(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	// Une cible juste à portée, qu'on retire aussitôt le tir parti : le
	// projectile poursuit dans le vide.
	h, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(5), py)
	if !ok {
		t.Fatal("créature refusée")
	}

	w.Step(Vec{})
	if w.Shots().Len() != 1 {
		t.Fatalf("%d projectile(s), attendu 1", w.Shots().Len())
	}
	w.Enemies().Remove(h)

	// La portée vaut six tuiles et le projectile en parcourt un cinquième par
	// tick : trente ticks suffisent largement.
	for range 60 {
		w.Step(Vec{})
	}
	if n := w.Shots().Len(); n != 0 {
		t.Errorf("%d projectile(s) encore en vol après leur portée", n)
	}
}

// TestLeTirNalloueRien garde le budget sur la boucle complète, tir compris.
func TestLeTirNalloueRien(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	// Des Vigiles, qui encaissent : la mesure ne doit pas vider le bassin.
	bloqueur := indexDuProfil(t, profils, "bloqueur")
	for i := range 16 {
		if _, ok := w.SpawnEnemy(bloqueur, px+FromInt(2+i%3), py+FromInt(i%4)); !ok {
			t.Fatal("créature refusée")
		}
	}
	for range 3 * flowPeriod {
		w.Step(Vec{})
	}

	moyenne := testing.AllocsPerRun(500, func() {
		w.Step(Vec{})
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par tick avec tir, attendu aucune", moyenne)
	}
}

// TestLeTirTouchePremierCeQuIlRencontre garde l'ordre le long du segment.
//
// Deux créatures sur la trajectoire, la plus proche doit tomber. Retenir la
// première du bassin ferait mourir celle de derrière — un projectile qui tue à
// travers un corps, ce qui se verra dès que la horde sera dense.
//
// **Le projectile est forgé et la passe appelée directement, parce qu'un tir
// ordinaire ne discrimine pas.** Il avance de 0,2 tuile par tick pour un rayon
// de 0,125 : deux cibles espacées d'une tuile ne sont jamais candidates dans le
// même tick, si bien qu'un test joué en entier passerait quel que soit l'ordre
// retenu. Il faut donc un segment qui les couvre toutes les deux, et c'est un
// cas que la cadence ne produit pas d'elle-même.
//
// `TestLeTirNEnjambePasSaCible` garde l'autre moitié : que la passe soit
// appelée, et qu'une cible enjambée soit atteinte par le chemin complet.
func TestLeTirTouchePremierCeQuIlRencontre(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")

	// La lointaine d'abord : sans tri le long du segment, c'est elle que le
	// parcours du bassin trouverait en premier.
	loin, ok := w.SpawnEnemy(marcheur, px+FromInt(2)+One/4, py)
	if !ok {
		t.Fatal("créature refusée")
	}
	pres, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py)
	if !ok {
		t.Fatal("créature refusée")
	}

	restant := func(h Handle) int {
		place, vivant := w.Enemies().Slot(h)
		if !vivant {
			return 0
		}
		return w.Enemies().At(place).Hits
	}
	depart := restant(loin)

	// Un segment qui part juste avant la plus proche et couvre les deux.
	tir := Projectile{
		X: px + FromInt(2) - One/16, Y: py,
		Step:      Vec{X: One/2 + One/4},
		Remaining: FromInt(6),
		Hits:      1,
	}
	if !w.toucher(Vec{tir.X, tir.Y}, &tir) {
		t.Fatal("le segment ne touche personne : le cas ne teste rien")
	}

	if restant(loin) != depart {
		t.Errorf("la créature de derrière a encaissé : le tir traverse celle de devant")
	}
	if restant(pres) >= depart {
		t.Errorf("la créature de devant est intacte : le tir ne l'a pas rencontrée")
	}
}

// TestLeTirNEnjambePasSaCible garde ce que le point d'arrivée ne peut pas voir.
//
// Un projectile avance de 0,2 tuile par tick et une créature en fait 0,125 de
// rayon : une cible qui tombe entre deux positions échantillonnées n'est jamais
// dans le rayon au moment du test, et le tir la traverse sans effet. Le défaut
// est resté invisible tant qu'aucune créature n'approchait à moins d'un
// demi-tuile.
//
// **Le cas est posé et la passe appelée directement.** Un tir joué en entier ne
// discrimine pas : la cible bouge d'un tick à l'autre, et une créature enjambée
// une fois se retrouve à portée du point d'arrivée au suivant — le test
// passerait alors avec ou sans le correctif, pour une raison qui n'est pas la
// sienne. Ce qui est éprouvé ici est un segment et une position, sans rien qui
// bouge entre les deux.
func TestLeTirNEnjambePasSaCible(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")

	// À un cinquième de tuile : le pas la dépasse, et le point d'arrivée la
	// laisse hors du rayon de 0,125.
	cible, ok := w.SpawnEnemy(marcheur, px+One/5, py)
	if !ok {
		t.Fatal("créature refusée")
	}
	depart := profils.Enemies[marcheur].Hits

	tir := Projectile{X: px, Y: py, Step: Vec{X: One * 2 / 5},
		Remaining: FromInt(6), Hits: 1}

	// Le cas doit être celui qu'on croit : la cible est hors de portée du seul
	// point d'arrivée, sinon le test ne dirait rien du segment.
	fin := Vec{tir.X + tir.Step.X, tir.Y}
	rayon := profils.Enemies[marcheur].Radius
	e := w.Enemies().At(0)
	if (Vec{e.X - fin.X, e.Y - fin.Y}).carres() <= int64(rayon)*int64(rayon) {
		t.Fatal("la cible est à portée du point d'arrivée : le cas ne teste pas le segment")
	}

	if !w.toucher(Vec{tir.X, tir.Y}, &tir) {
		t.Fatal("le projectile enjambe sa cible")
	}
	place, vivant := w.Enemies().Slot(cible)
	if !vivant || w.Enemies().At(place).Hits >= depart {
		t.Error("la cible n'a rien encaissé")
	}
}

// TestLeTirAtteintUneCibleCollee garde ce sur quoi repose l'exception du Vigile.
//
// La conception tient que son corps solide ne peut pas devenir un piège **parce
// que** le joueur acculé lui tire nécessairement dessus : c'est le plus proche,
// et la visée est omnidirectionnelle.
//
// **Six autres tests tombent sur une portée minimale ajoutée au ciblage, et
// aucun ne garde cette propriété** : leurs cibles poursuivent le joueur, donc
// elles finissent collées avant d'encaisser leur dernière touche. Ils
// rencontrent le cas, ils ne le posent pas — celui qui figerait sa cible ou
// raccourcirait sa boucle cesserait de l'attraper sans que rien ne le dise.
//
// **L'écart nul est le cas qui compte**, et pas seulement le contact : c'est là
// que la visée anticipée normalise un vecteur nul, et la seule distance qu'un
// carré nul admet. Les écarts intermédiaires l'encadrent plutôt qu'ils ne le
// répètent — une portée minimale se poserait quelque part entre eux.
//
// Il ne dit rien de ce qu'on **voit** : l'éclair d'impact s'allume, et le rendu
// le recouvre du sprite du joueur, qui passe devant ce qui partage sa
// profondeur. Une partie jouée l'a fait prendre pour un tir qui ne part plus.
func TestLeTirAtteintUneCibleCollee(t *testing.T) {
	for _, ecart := range []Fixed{0, FromFloat(0.1), FromFloat(0.25)} {
		w, profils := champDeTir(t)
		px, py := w.Player()
		marcheur := indexDuProfil(t, profils, "marcheur")

		cible, ok := w.SpawnEnemy(marcheur, px+ecart, py)
		if !ok {
			t.Fatal("créature refusée")
		}

		morte := false
		for range 5 * TPS {
			w.Step(Vec{})
			if !w.Enemies().Alive(cible) {
				morte = true
				break
			}
		}
		if !morte {
			t.Errorf("collée à %v tuile du joueur, la cible n'est pas tombée en cinq secondes",
				ecart.Float())
		}
	}
}

// TestLeTirNeCoupePasUnAngle garde ce qu'une seule partie jouée a montré.
//
// **Le défaut demandait de raser un angle**, ce qui explique qu'il ait fallu
// jouer pour le voir : le pas d'un projectile vaut deux dixièmes de tuile, et il
// faut qu'il entre dans une case bloquée et en ressorte à l'intérieur du même
// pas. Aucun des deux bouts n'y tombe alors, et le tir passait au travers.
//
// Le cas pose la case bloquée en diagonale du départ, sur l'angle où quatre
// cases se rencontrent : c'est la seule géométrie où le pas courbe la règle.
func TestLeTirNeCoupePasUnAngle(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	g := NewCostGrid(8, 8)
	g.Set(1, 1, Blocked)
	w := NewWorld(profils, armesInertes(t), progressionLivree(t), sansVagues(), g,
		graineDeTest, Capacities{Enemies: 4, Shots: 4, EnemyShots: 4, Blasts: 4, Gems: 4})

	// De la case (2,1) vers la case (1,2), en passant par l'angle du coin (2,2).
	// Le départ et l'arrivée sont franchissables ; seule (1,1) ne l'est pas.
	depart := Vec{X: FromFloat(2.02), Y: FromFloat(1.98)}
	if _, ok := w.tirs.Spawn(Projectile{
		X: depart.X, Y: depart.Y,
		Step:      Vec{X: FromFloat(-0.04), Y: FromFloat(0.04)},
		Remaining: FromInt(6),
		Hits:      1,
	}); !ok {
		t.Fatal("projectile refusé")
	}

	w.deplacerTirs()
	if w.tirs.Len() != 0 {
		p := w.tirs.At(0)
		t.Errorf("le tir a franchi l'angle et se trouve en (%v, %v)", p.X, p.Y)
	}
}

// TestLeTirViseOuLaCibleSera garde ce qu'une partie jouée a fait remarquer.
//
// **La cible traverse, et c'est le cas que la visée manquait.** Le projectile met
// une demi-seconde à parcourir six tuiles ; un Quidam en couvre une et demie dans
// ce temps, pour un rayon d'un huitième. Viser sa position rendait le tir sans
// effet sur tout ce qui ne venait pas droit dessus, et le joueur n'a aucun moyen
// de corriger une visée qu'il ne tient pas.
//
// La créature est avancée à la main : ce que le cas éprouve est la prédiction, et
// une cible que le champ de flux ramènerait vers le joueur serait précisément
// celle que l'ancienne visée touchait déjà.
func TestLeTirViseOuLaCibleSera(t *testing.T) {
	w, profils := champDeTir(t)
	marcheur := indexDuProfil(t, profils, "marcheur")
	if _, ok := w.SpawnEnemy(marcheur, FromInt(20)+One/2, FromInt(16)+One/2); !ok {
		t.Fatal("créature refusée")
	}

	// Quatre tuiles à l'est du joueur, traversant plein sud à sa vitesse.
	e := w.ennemis.At(0)
	e.Step = Vec{Y: profils.Enemies[marcheur].Speed}
	resistance := e.Hits

	w.tirer()
	if w.tirs.Len() == 0 {
		t.Fatal("aucun tir parti")
	}
	for range 2 * TPS {
		cible := w.ennemis.At(0)
		cible.X, cible.Y = cible.X+cible.Step.X, cible.Y+cible.Step.Y
		w.deplacerTirs()
		if w.tirs.Len() == 0 {
			break
		}
	}

	if w.ennemis.At(0).Hits >= resistance {
		t.Error("le tir a manqué une cible qui traverse")
	}
}

// TestLEclairSAllumeAuCoupEtSEteintSeul garde le seul retour qu'une touche
// produise.
//
// **Une créature encaisse trois touches et le joueur ne voyait que la dernière.**
// Rien ne distinguait un tir qui rate d'un tir qui entame, si bien qu'en jouant
// on déduisait les ratés au lieu de les lire. Le cas garde les deux moitiés du
// mécanisme : l'éclair s'allume au coup, et il s'éteint sans que personne
// n'ait à le remettre à zéro.
func TestLEclairSAllumeAuCoupEtSEteintSeul(t *testing.T) {
	w, profils := champDeTir(t)
	// Le Vigile encaisse douze touches : la créature survit assez pour qu'on
	// voie son éclair s'éteindre.
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "bloqueur"),
		FromInt(19)+One/2, FromInt(16)+One/2); !ok {
		t.Fatal("créature refusée")
	}

	touche := 0
	for range 4 * TPS {
		w.Step(Vec{})
		if w.ennemis.At(0).Flash > 0 {
			break
		}
		touche++
	}
	if w.ennemis.At(0).Flash == 0 {
		t.Fatal("aucun éclair après quatre secondes de tir")
	}

	// La cadence vaut vingt-quatre ticks et l'éclair quatre : aucun second coup
	// ne peut le rallumer dans cet intervalle.
	for range eclairImpact {
		w.Step(Vec{})
	}
	if reste := w.ennemis.At(0).Flash; reste != 0 {
		t.Errorf("éclair encore à %d après sa durée : il ne s'éteint pas", reste)
	}
}
