// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de l'arme lourde : ce qu'on tient, ce qu'un déclenchement dépense,
// l'arme jetée à vide, et une déflagration qui n'emporte que la horde.

package game

import "testing"

// TestUneLourdeSePrendParSaCle garde ce que la sonde du montage appelle.
//
// Le second cas est celui qui compte : demander l'arme de base rendrait une arme
// sans charges, donc un emplacement qu'on croirait tenir et qui ne déclencherait
// rien.
func TestUneLourdeSePrendParSaCle(t *testing.T) {
	w, _ := champDeTir(t)

	if !w.GiveHeavy("grenade") {
		t.Fatal("la grenade n'est pas dans la table des armes")
	}
	arme, charges := w.HeldHeavy()
	if arme.Key != "grenade" || charges != arme.Charges {
		t.Errorf("tenue : %q à %d charge(s), attendu grenade à %d",
			arme.Key, charges, arme.Charges)
	}

	if w.GiveHeavy("reglementaire") {
		t.Error("le socle a été pris pour une lourde")
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
	if !w.GiveHeavy("grenade") {
		t.Fatal("la grenade n'est pas dans la table des armes")
	}
	_, avant := w.HeldHeavy()

	w.Trigger()

	if _, apres := w.HeldHeavy(); apres != avant {
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
	if !w.GiveHeavy("grenade") {
		t.Fatal("la grenade n'est pas dans la table des armes")
	}
	px, py := w.Player()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(2), py); !ok {
		t.Fatal("créature refusée")
	}
	_, avant := w.HeldHeavy()

	w.Trigger()

	if _, apres := w.HeldHeavy(); apres != avant-1 {
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
	if !w.GiveHeavy("grenade") {
		t.Fatal("la grenade n'est pas dans la table des armes")
	}
	px, py := w.Player()
	marcheur := indexDuProfil(t, profils, "marcheur")
	_, charges := w.HeldHeavy()

	for range charges {
		// Une cible neuve à chaque fois : la précédente peut être morte de la
		// déflagration, et un déclenchement sans cible ne dépenserait rien.
		if _, ok := w.SpawnEnemy(marcheur, px+FromInt(2), py); !ok {
			t.Fatal("créature refusée")
		}
		w.Trigger()
		w.detoner()
	}

	if arme, reste := w.HeldHeavy(); arme.Key != "" || reste != 0 {
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
