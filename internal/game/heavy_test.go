// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de l'arme lourde : ce qu'une caisse laisse, ce qu'on ramasse en
// marchant dessus, l'échange quand les deux emplacements sont pleins, ce qu'un
// déclenchement dépense et la déflagration qui n'emporte que la horde.

package game

import "testing"

// tomber pose une arme lourde au sol sous le joueur, par le bassin.
//
// **Le hasard est contourné, jamais le ramassage** : ce que ces cas éprouvent
// est ce qui suit la chute, et passer par une caisse ferait dépendre chacun d'un
// tirage. La chute elle-même est gardée par
// `TestUneCaisseLaisseParfoisUneArme`.
func tomber(t *testing.T, w *World, cle string) {
	t.Helper()
	px, py := w.Player()
	if _, ok := w.armesAuSol.Spawn(Drop{X: px, Y: py, Weapon: rangDeLArme(t, w, cle)}); !ok {
		t.Fatal("bassin des armes au sol plein")
	}
}

// TestUneArmeSeRamasseEnMarchantDessus garde ce que la conception veut du geste.
//
// **Aucun menu, aucune touche** tant qu'une place est libre : on passe dessus
// pour prendre, on contourne pour laisser.
func TestUneArmeSeRamasseEnMarchantDessus(t *testing.T) {
	w, _ := champDeTir(t)
	tomber(t, w, "grenade")

	w.ramasserUneArme()

	arme, charges := w.HeldHeavy(0)
	if arme.Key != "grenade" || charges != arme.Charges {
		t.Errorf("tenue : %q à %d charge(s), attendu grenade pleine", arme.Key, charges)
	}
	if w.armesAuSol.Len() != 0 {
		t.Error("l'arme est restée au sol après avoir été prise")
	}
}

// TestLesDeuxEmplacementsSeRemplissentPuisSArretent garde la borne.
//
// **Deux et pas trois** : la conception en fait une règle, le joueur ayant une
// décision — laquelle garder — et non une gestion. La troisième arme reste au
// sol, ce qui est la condition de l'échange.
func TestLesDeuxEmplacementsSeRemplissentPuisSArretent(t *testing.T) {
	w, _ := champDeTir(t)

	for range Slots {
		tomber(t, w, "grenade")
		w.ramasserUneArme()
	}
	for place := range Slots {
		if _, charges := w.HeldHeavy(place); charges == 0 {
			t.Errorf("emplacement %d vide après deux ramassages", place)
		}
	}

	tomber(t, w, "grenade")
	w.ramasserUneArme()
	if w.armesAuSol.Len() != 1 {
		t.Error("la troisième arme a été prise alors que les deux places sont tenues")
	}
}

// TestLaToucheEchangeQuandLesDeuxPlacesSontTenues garde la règle unique.
//
// **La touche d'un emplacement, pressée sur une arme au sol, y met cette arme** —
// vide ou plein. Les trois autres lectures possibles — remplacer la plus
// ancienne, la moins chargée ou la première — font perdre une arme sans que le
// joueur sache laquelle, et aucun aperçu au sol ne peut le lui dire à l'avance.
func TestLaToucheEchangeQuandLesDeuxPlacesSontTenues(t *testing.T) {
	w, _ := champDeTir(t)
	for range Slots {
		tomber(t, w, "grenade")
		w.ramasserUneArme()
	}

	// Une arme entamée dans la seconde place : c'est ce qui rend l'échange
	// visible, la neuve arrivant pleine.
	w.lourdes[1].Charges = 1
	tomber(t, w, "grenade")

	if !w.TakeDrop(1) {
		t.Fatal("l'échange a été refusé alors qu'une arme est sous les pieds")
	}
	arme, charges := w.HeldHeavy(1)
	if charges != arme.Charges {
		t.Errorf("la place échangée porte %d charge(s), attendu une arme pleine", charges)
	}
	if w.armesAuSol.Len() != 0 {
		t.Error("l'arme échangée est restée au sol")
	}
}

// TestLEchangeSansArmeSousLesPiedsNeFaitRien garde ce qui laisse la touche
// déclencher.
//
// C'est ce qui permet à la même touche de faire deux choses : sans arme au sol,
// elle rend faux et l'appelant déclenche ce que l'emplacement tient.
func TestLEchangeSansArmeSousLesPiedsNeFaitRien(t *testing.T) {
	w, _ := champDeTir(t)
	tomber(t, w, "grenade")
	w.ramasserUneArme()

	if w.TakeDrop(0) {
		t.Error("un échange a eu lieu sans arme au sol")
	}
}

// TestUneCaisseLaisseParfoisUneArme garde le premier lecteur du flux « butin ».
//
// **`Loot` attendait le sien depuis sa déclaration**, sa godoc annonçant « au
// futur, et rien ne l'alimente encore » : seul le témoin de l'empreinte le
// gardait numéroté. C'est le même moment que les figurants pour le flux
// cosmétique.
//
// Le compte n'est pas vérifié — une chance sur trois n'a pas de fréquence exacte
// sur un échantillon —, seulement qu'il tombe des armes et qu'il n'en tombe pas à
// chaque fois. Sans la seconde moitié, un code qui en lâcherait toujours passerait.
func TestUneCaisseLaisseParfoisUneArme(t *testing.T) {
	w, _ := champDeTir(t)
	px, py := w.Player()

	tombees, essais := 0, 60
	for range essais {
		avant := w.armesAuSol.Len()
		w.lacherUneArme(px, py)
		if w.armesAuSol.Len() > avant {
			tombees++
			w.armesAuSol.RemoveAt(0)
		}
	}

	if tombees == 0 {
		t.Errorf("aucune arme sur %d caisses, la chance déclarée est de une sur %d",
			essais, w.progression.HeavyOdds)
	}
	if tombees == essais {
		t.Error("chaque caisse a laissé une arme : le tirage ne départage rien")
	}
}

// TestUnDeclenchementSansCibleNeDepenseRien éprouve le cas limite du tir à vide,
// transposé à ce qui se déclenche.
//
// **La charge ne part pas dans le vide**, pour la raison qui fait que la cadence
// ne se consomme pas sans cible : une arme à trois charges dont une disparaît
// sans rien produire se lit comme un défaut, et le joueur n'a aucun moyen de
// relier la perte à son geste.
func TestUnDeclenchementSansCibleNeDepenseRien(t *testing.T) {
	w, _ := champDeTir(t)
	tomber(t, w, "grenade")
	w.ramasserUneArme()
	_, avant := w.HeldHeavy(0)

	w.Trigger(0)

	if _, apres := w.HeldHeavy(0); apres != avant {
		t.Errorf("%d charge(s) après un déclenchement sans cible, attendu %d", apres, avant)
	}
	if w.souffles.Len() != 0 {
		t.Error("une déflagration a été posée sans cible")
	}
}

// TestUnDeclenchementPoseUneDeflagrationEtDepense ferme le chemin de la touche à
// l'explosion.
func TestUnDeclenchementPoseUneDeflagrationEtDepense(t *testing.T) {
	w, profils := champDeTir(t)
	tomber(t, w, "grenade")
	w.ramasserUneArme()
	px, py := w.Player()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(2), py); !ok {
		t.Fatal("créature refusée")
	}
	_, avant := w.HeldHeavy(0)

	w.Trigger(0)

	if _, apres := w.HeldHeavy(0); apres != avant-1 {
		t.Errorf("%d charge(s) après un déclenchement, attendu %d", apres, avant-1)
	}
	if w.souffles.Len() != 1 {
		t.Fatalf("%d déflagration(s) posée(s), attendu 1", w.souffles.Len())
	}
	if b := w.souffles.At(0); b.Source != BlastWeapon {
		t.Errorf("source %d, attendu celle d'une arme", b.Source)
	}
}

// TestUneLourdeVideEstJetee garde ce que la conception exige d'elle.
//
// **L'arme quitte l'emplacement au lieu d'y rester inerte** : le joueur voit sa
// place se libérer plutôt qu'un compteur à zéro, et rien ne l'invite à presser
// une touche qui ne fera plus rien.
func TestUneLourdeVideEstJetee(t *testing.T) {
	w, profils := champDeTir(t)
	tomber(t, w, "grenade")
	w.ramasserUneArme()
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")
	_, charges := w.HeldHeavy(0)

	for range charges {
		// Une cible neuve à chaque fois : la précédente peut être morte de la
		// déflagration, et un déclenchement sans cible ne dépenserait rien.
		if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py); !ok {
			t.Fatal("créature refusée")
		}
		w.Trigger(0)
		w.detoner()
	}

	if arme, reste := w.HeldHeavy(0); arme.Key != "" || reste != 0 {
		t.Errorf("l'arme vide est restée : %q à %d charge(s)", arme.Key, reste)
	}
}

// TestUneDeflagrationDArmeNEmporteQueLaHorde garde la décision de jeu.
//
// **Le joueur ne dirige pas son lancer**, donc le blesser serait une sanction
// sans recours — il ne contrôle que son déplacement. C'est ce qui la sépare de
// celle d'une Baudruche, qu'on voit venir et qu'on esquive, et la question se
// rouvrira le jour où une lourde deviendra dirigeable.
func TestUneDeflagrationDArmeNEmporteQueLaHorde(t *testing.T) {
	w, profils := champDeTir(t)
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")

	cible, ok := w.SpawnEnemy(marcheur, px+One/2, py)
	if !ok {
		t.Fatal("créature refusée")
	}
	plein := restantDe(w, cible)
	vie := w.Health()

	// Sous les pieds du joueur : si l'explosion le touchait, ce serait ici.
	w.souffles.Spawn(Blast{X: px, Y: py, Source: BlastWeapon, Index: rangDeLArme(t, w, "grenade")})
	w.detoner()

	if restantDe(w, cible) >= plein {
		t.Error("la créature dans le rayon est intacte : le cas ne teste rien")
	}
	if w.Health() != vie {
		t.Errorf("le joueur a encaissé %d point(s) de sa propre grenade", vie-w.Health())
	}
}

// rangDeLArme rend la place d'une arme dans la table livrée.
func rangDeLArme(t *testing.T, w *World, cle string) int {
	t.Helper()
	for i := range w.armes.All {
		if w.armes.All[i].Key == cle {
			return i
		}
	}
	t.Fatalf("arme « %s » absente de la table", cle)
	return 0
}
