// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du soin : la voisine remise debout, le soigneur qui ne se soigne pas,
// le plafond de la résistance, le mort qu'on ne ressuscite pas, et la horde
// ordinaire qui ne soigne personne.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// areneDeSoin monte une salle vide et rend le profil du Secouriste.
func areneDeSoin(t *testing.T) (*World, *EnemyProfile) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	g := NewCostGrid(32, 32)
	w := NewWorld(profils, armesInertes(t), progressionLivree(t), sansVagues(), g,
		graineDeTest, capacitesDeTest)
	w.Place(FromInt(16)+One/2, FromInt(16)+One/2)

	p := &profils.Enemies[indexDuProfil(t, profils, "soigneur")]
	if p.HealRange == 0 || p.HealCooldown == 0 || p.HealHits == 0 {
		t.Fatalf("le Secouriste ne soigne pas : portée %v, cadence %d, touches %d",
			p.HealRange, p.HealCooldown, p.HealHits)
	}
	return w, p
}

// poser place une créature à l'écart du joueur et rend son entité.
//
// À dix tuiles : l'arme est inerte, mais la horde converge, et ces cas comptent
// des touches que le contact ne doit pas troubler.
func poser(t *testing.T, w *World, cle string, dx, dy Fixed) *Enemy {
	t.Helper()
	avant := w.Enemies().Len()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, w.profils, cle),
		w.playerX+dx, w.playerY+dy); !ok {
		t.Fatal("bassin plein")
	}
	return w.Enemies().At(avant)
}

// TestLeSecouristeRemetUneVoisineDebout garde ce que la conception lui donne :
// tant qu'il vit, le reste ne meurt pas.
//
// Le cas entame une voisine à la main plutôt que de la faire tirer dessus —
// l'arme est inerte pour que rien n'abatte les cobayes.
//
// **Il s'arrête sur le soin, pas après une cadence.** Les deux éclairs durent un
// quart de seconde quand la cadence en dure une et demie : un relevé pris à
// `HealCooldown` les trouve éteints depuis longtemps, et ne garde alors que la
// moitié de ce que ce cas promet. C'est ce que le premier jet mesurait.
func TestLeSecouristeRemetUneVoisineDebout(t *testing.T) {
	w, profil := areneDeSoin(t)
	blessee := poser(t, w, "marcheur", FromInt(10), 0)
	poser(t, w, "soigneur", FromInt(10)+One, 0)

	blessee.Hits = 1
	for range profil.HealCooldown + 1 {
		w.Step(Vec{})
		if blessee.Hits > 1 {
			break
		}
	}

	if blessee.Hits != 1+profil.HealHits {
		t.Errorf("résistance %d après un soin, attendu %d",
			blessee.Hits, 1+profil.HealHits)
	}
	if blessee.Healed == 0 {
		t.Error("la soignée ne porte pas son éclair : rien ne dira au joueur " +
			"pourquoi son travail est perdu")
	}
	if w.Enemies().At(1).Healing == 0 {
		t.Error("le soigneur ne porte pas son éclair : rien ne dira au joueur " +
			"lequel aller chercher")
	}
}

// TestLeSecouristeNeSeSoignePas garde la moitié qui le rend abattable.
//
// Trois touches tombent vite une fois qu'on l'a rejoint, et c'est cette
// récompense qui paie le trajet dans la horde. Un soigneur qui se régénère la
// retirerait au moment de l'obtenir.
func TestLeSecouristeNeSeSoignePas(t *testing.T) {
	w, profil := areneDeSoin(t)
	soigneur := poser(t, w, "soigneur", FromInt(10), 0)
	soigneur.Hits = 1

	for range 3 * (profil.HealCooldown + 1) {
		w.Step(Vec{})
	}
	if soigneur.Hits != 1 {
		t.Errorf("résistance %d : il s'est soigné lui-même", soigneur.Hits)
	}
}

// TestLeSoinNeDepassePasLaResistance vérifie qu'un soin ne rend pas une créature
// plus solide que son profil.
//
// Sans plafond, un Badaud laissé longtemps auprès d'un Secouriste accumulerait
// des touches que rien dans la table ne lui donne, et sa résistance cesserait
// d'être une propriété du profil.
//
// **Le soin est porté à deux touches ici, et le cas n'existe qu'à cette
// condition.** Le manifeste en donne une, or `plusBlessee` n'élit que des
// créatures à qui il en manque au moins une : le plafond n'a alors jamais rien à
// couper, et le retirer laisse la suite au vert — ce qu'une mutation a montré.
// Deux touches rendues pour une seule perdue est un réglage que le format admet,
// et c'est le seul où la question se pose.
func TestLeSoinNeDepassePasLaResistance(t *testing.T) {
	w, profil := areneDeSoin(t)
	voisine := poser(t, w, "marcheur", FromInt(10), 0)
	poser(t, w, "soigneur", FromInt(10)+One, 0)
	plein := w.profils.Enemies[voisine.Profile].Hits

	profil.HealHits = 2
	voisine.Hits = plein - 1
	for range 4 * (profil.HealCooldown + 1) {
		w.Step(Vec{})
	}
	if voisine.Hits != plein {
		t.Errorf("résistance %d pour un profil qui en porte %d", voisine.Hits, plein)
	}
}

// TestLeSoinNeRessuscitePas ferme l'état mort-vivant que le bassin ne connaît
// pas.
//
// Une résistance nulle **est** la mort, et le nettoyage retire la créature au
// même tick. Choisir un mort comme cible du soin le rendrait vivant après coup,
// ce qui ferait exister un instant où une créature est morte et ne l'est plus —
// exactement ce que l'explosion de la Baudruche a évité en vivant à part.
func TestLeSoinNeRessuscitePas(t *testing.T) {
	w, profil := areneDeSoin(t)
	morte := poser(t, w, "marcheur", FromInt(10), 0)
	poser(t, w, "soigneur", FromInt(10)+One, 0)

	morte.Hits = 0
	for range profil.HealCooldown + 1 {
		w.Step(Vec{})
	}

	// Elle a quitté le bassin ; ce qui reste est le seul soigneur.
	if n := w.Enemies().Len(); n != 1 {
		t.Errorf("%d créature(s) vivantes, attendu le seul soigneur : un mort a "+
			"été remis debout", n)
	}
}

// TestUnProfilSansPorteeDeSoinNeSoignePas garde ce qui permet au soin de vivre
// dans la passe commune sans y peser.
//
// Toute la horde la traverse : si la portée nulle ne fermait pas le mécanisme,
// chaque Badaud remettrait ses voisines debout et plus rien ne mourrait.
//
// **Les deux cobayes sont exactement superposés, et c'est ce qui discrimine.**
// Une portée nulle rejette toute distance non nulle par la comparaison de
// `plusBlessee` : le garde y paraît redondant, et le retirer laisse la suite au
// vert — ce qu'une mutation a montré. La superposition est le seul écart qu'un
// rayon nul admet, et elle n'a rien de théorique : l'anneau d'apparition la
// produit, une meute de Molosses par construction.
//
// Le soin s'appelle directement plutôt que par un tick : la séparation écarte
// les deux corps dès le premier déplacement, et le cas serait éteint avant
// d'avoir posé sa question.
//
// **La propriété gardée n'est visible que par un effet secondaire**, ce qui est
// assez contre-intuitif pour être dit : un Badaud rend zéro touche, si bien que
// la résistance reste identique même quand le mécanisme s'ouvre à lui. Mesurer
// ce que le soin est censé changer ne peut donc rien voir ; le seul témoin est
// le retour visuel, qui s'allume sur un soin qui n'a pas eu lieu.
func TestUnProfilSansPorteeDeSoinNeSoignePas(t *testing.T) {
	w, _ := areneDeSoin(t)
	blessee := poser(t, w, "marcheur", FromInt(10), 0)
	poser(t, w, "marcheur", FromInt(10), 0)
	if w.profils.Enemies[blessee.Profile].HealRange != 0 {
		t.Fatal("le Badaud soigne, ce cas suppose le contraire")
	}

	blessee.Hits = 1
	w.soigner()
	if blessee.Hits != 1 {
		t.Errorf("résistance %d : un profil sans portée de soin a guéri la créature "+
			"superposée à lui", blessee.Hits)
	}

	// **Les éclairs sont ce qui discrimine vraiment.** Un Badaud rend zéro
	// touche, donc la résistance ne bouge pas même quand le mécanisme s'ouvre à
	// lui : ce qui se voit alors est un soin annoncé et sans effet, c'est-à-dire
	// un joueur qui part chercher un Secouriste qui n'existe pas.
	if blessee.Healed != 0 || w.Enemies().At(1).Healing != 0 {
		t.Errorf("éclairs %d et %d sur deux Badauds : un soin s'affiche là où "+
			"aucun n'a lieu", blessee.Healed, w.Enemies().At(1).Healing)
	}
}
