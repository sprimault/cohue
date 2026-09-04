// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de l'explosion : l'amorce posée à la mort, la mèche qui laisse le
// temps de s'écarter, le choc hors plafond, la horde que la déflagration
// épargne, et les profils qui n'explosent pas.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// areneExplosive monte une salle avec l'arme du joueur active.
//
// **L'arme tire, contrairement aux autres arènes du paquet** : l'amorce se
// branche sur la mort d'une créature, et la seule mort qu'une partie sache
// produire est celle qu'un projectile inflige. La neutraliser reviendrait à
// éprouver l'explosion sans jamais l'obtenir.
func areneExplosive(t *testing.T) (*World, *EnemyProfile) {
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
	w := NewWorld(profils, armes, progressionLivree(t), sansVagues(), g, graineDeTest,
		capacitesDeTest)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)

	p := &profils.Enemies[indexDuProfil(t, profils, "eclateur")]
	if p.BurstRadius == 0 || p.Fuse == 0 || p.BurstDamage == 0 {
		t.Fatalf("la Baudruche n'explose pas : rayon %v, mèche %d, dégâts %d",
			p.BurstRadius, p.Fuse, p.BurstDamage)
	}
	return w, p
}

// abattreUneBaudruche en pose une à portée du joueur et la laisse tomber sous
// son arme, puis rend le souffle amorcé.
func abattreUneBaudruche(t *testing.T, w *World, tuiles Fixed) *Blast {
	t.Helper()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "eclateur"),
		w.playerX+tuiles, w.playerY); !ok {
		t.Fatal("bassin plein")
	}

	for range 10 * TPS {
		w.Step(Vec{})
		if w.Blasts().Len() > 0 {
			return w.Blasts().At(0)
		}
	}
	t.Fatal("aucune explosion amorcée : la Baudruche n'est pas morte, ou sa mort " +
		"n'a rien posé")
	return nil
}

// TestLaBaudrucheAmorceEnMourant vérifie que sa mort pose une explosion qui
// détone plus tard.
//
// **Plus tard, et c'est tout l'intérêt.** Une déflagration instantanée n'aurait
// pas de télégraphe, donc rien à esquiver : ce que la conception lui donne est
// une annonce, et une annonce suppose un délai. Le cas relève donc que la
// créature a bien quitté le bassin — l'explosion ne la garde pas en vie — et que
// la détonation attend sa mèche.
func TestLaBaudrucheAmorceEnMourant(t *testing.T) {
	w, profil := areneExplosive(t)
	b := abattreUneBaudruche(t, w, FromInt(3))

	if n := w.Enemies().Len(); n != 0 {
		t.Errorf("%d créature(s) vivantes : la Baudruche est restée dans le bassin", n)
	}
	if b.Fuse != profil.Fuse {
		t.Errorf("mèche de %d ticks à l'amorce, attendu %d", b.Fuse, profil.Fuse)
	}
	// La perte se compare aux dégâts d'explosion et non à zéro : la Baudruche a
	// approché pendant qu'on l'abattait, donc quelques points de contact
	// ordinaire sont partis avant sa mort. Ce qu'on garde ici est qu'aucune
	// détonation n'a eu lieu, pas que rien ne s'est passé.
	if perdu := w.MaxHealth() - w.Health(); perdu >= profil.BurstDamage {
		t.Errorf("%d points perdus au moment de l'amorce, l'explosion en vaut %d : "+
			"la détonation n'a pas attendu sa mèche", perdu, profil.BurstDamage)
	}
}

// TestLExplosionBlesseHorsPlafond garde la troisième des sources de dégâts que
// la conception distingue.
//
// Trente-cinq points sur un plafond de vingt : si la déflagration passait par
// l'accumulateur de contact, le relevé rendrait au plus le plafond. Le joueur ne
// bouge pas, et il est seul avec l'emprise — ce qu'il perd ne peut venir que
// d'elle.
func TestLExplosionBlesseHorsPlafond(t *testing.T) {
	w, profil := areneExplosive(t)
	if profil.BurstDamage <= w.profils.Player.DamageCap {
		t.Fatalf("explosion à %d sous le plafond de %d : le cas ne sépare rien",
			profil.BurstDamage, w.profils.Player.DamageCap)
	}
	abattreUneBaudruche(t, w, One)

	avant := w.Health()
	for range profil.Fuse + 1 {
		w.Step(Vec{})
	}
	if perdu := avant - w.Health(); perdu != profil.BurstDamage {
		t.Errorf("%d points perdus à la détonation, attendu %d", perdu, profil.BurstDamage)
	}
	if n := w.Blasts().Len(); n != 0 {
		t.Errorf("%d explosion(s) survivent à leur détonation", n)
	}
}

// TestLaMecheLaisseLeTempsDeSEcarter vérifie que le télégraphe sert à quelque
// chose.
//
// Une amorce qu'on ne peut pas fuir serait une punition sans réponse, et le
// délai n'aurait aucune raison d'exister. Le joueur part au tick de l'amorce et
// marche droit : à cinq tuiles par seconde pour une mèche d'une demi-seconde, il
// couvre plus que le rayon.
func TestLaMecheLaisseLeTempsDeSEcarter(t *testing.T) {
	w, profil := areneExplosive(t)
	abattreUneBaudruche(t, w, One)

	avant := w.Health()
	for range profil.Fuse + 1 {
		w.Step(Vec{Y: -One})
	}
	if perdu := avant - w.Health(); perdu != 0 {
		t.Errorf("%d points perdus alors que le joueur a fui : la mèche ne laisse "+
			"pas le temps de sortir de l'emprise", perdu)
	}
}

// TestLExplosionEpargneLaHorde garde la décision qui paraîtrait fausse à qui
// reprendrait le code sans la conception.
//
// Une déflagration qui emporte les voisines serait plus imitative du réel et
// retournerait la mécanique : la Baudruche punit le nettoyage à l'aveugle en
// mêlée, et nettoyer autour d'elle récompenserait exactement ce geste.
func TestLExplosionEpargneLaHorde(t *testing.T) {
	w, profil := areneExplosive(t)
	b := abattreUneBaudruche(t, w, FromInt(3))

	// Un Badaud posé au centre même de l'emprise, après l'amorce pour que l'arme
	// du joueur ne l'abatte pas avant la détonation.
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, "marcheur"), b.X, b.Y); !ok {
		t.Fatal("bassin plein")
	}
	temoin := w.Enemies().At(0)
	vie := temoin.Hits

	// **Deux ticks de marge après la mèche**, parce que l'amorce est posée au
	// tick où la créature meurt, c'est-à-dire après le passage des détonations :
	// s'arrêter à `Fuse` relèverait l'état d'avant l'explosion, et le cas
	// passerait sur un code qui emporte la horde — ce qu'une mutation a montré.
	for range profil.Fuse + 2 {
		w.Step(Vec{})
	}

	// La survie plutôt que le compte de touches : l'arme du joueur entame le
	// témoin pendant ce temps, quand l'explosion l'anéantirait. Trois touches ne
	// tombent pas en trente-deux ticks à une salve toutes les vingt-quatre.
	if temoin.Hits <= 0 {
		t.Errorf("le témoin est tombé de %d touches à %d : l'explosion a mordu sur "+
			"la horde", vie, temoin.Hits)
	}
}

// TestUnProfilSansRayonNAmorceRien garde ce qui permet à l'amorce de vivre sur
// la transition de mort commune.
//
// Toutes les créatures y passent : si le rayon nul ne fermait pas le mécanisme,
// chaque Badaud abattu poserait une explosion sans dégâts ni emprise, et le
// bassin se remplirait de déflagrations invisibles.
func TestUnProfilSansRayonNAmorceRien(t *testing.T) {
	w, _ := areneExplosive(t)
	marcheur := indexDuProfil(t, w.profils, "marcheur")
	if w.profils.Enemies[marcheur].BurstRadius != 0 {
		t.Fatal("le Badaud porte un rayon d'explosion, ce cas suppose le contraire")
	}
	if _, ok := w.SpawnEnemy(marcheur, w.playerX+FromInt(3), w.playerY); !ok {
		t.Fatal("bassin plein")
	}

	for range 10 * TPS {
		w.Step(Vec{})
		if n := w.Blasts().Len(); n != 0 {
			t.Fatalf("%d explosion(s) amorcées par un profil sans rayon", n)
		}
	}
}
