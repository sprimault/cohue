// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la gemme : la mort qui en laisse une là où elle a eu lieu, la
// volée qui ne s'empile pas, et le ramassage qui s'arrête à la portée.

package game

import "testing"

// TestLaMortLaisseSaGemmeSurPlace vérifie où tombe le butin d'une créature.
//
// **La position exacte est ce qui se garde ici**, et non le seul fait qu'une
// gemme apparaisse : le tas dit au joueur où il a tué, donc où revenir pour
// récolter, et c'est ce lien qui donne son prix au trajet de collecte — il va
// contre le trajet de fuite. Une gemme posée ailleurs, au centre de la case ou
// dispersée, casserait ce lien sans casser aucun autre test.
//
// L'appel passe par `toucher` plutôt que par un tick complet, parce qu'une
// créature se déplace plus tôt dans le tick : sa position au moment de mourir
// n'est pas celle où on l'a posée, et le test ne saurait pas à quoi comparer.
//
// **La créature est posée hors du centre de sa case, et c'est ce qui rend la
// mesure discriminante.** Le joueur se tient au centre du lieu, donc au centre
// d'une case ; une créature posée à un nombre entier de tuiles de lui y est
// aussi, si bien qu'une gemme arrondie à la case tomberait au même point que la
// gemme posée sur place, et le test passerait sur les deux comportements.
func TestLaMortLaisseSaGemmeSurPlace(t *testing.T) {
	w, profils := champSansTir(t)
	px, py := w.Player()

	profil := indexDuProfil(t, profils, "marcheur")
	if _, ok := w.SpawnEnemy(profil, px+FromInt(3)+One/8, py+One/16); !ok {
		t.Fatal("créature refusée")
	}
	e := w.Enemies().At(0)
	e.Hits = 1
	x, y := e.X, e.Y

	depart := Vec{X: x - One/8, Y: y}
	if !w.toucher(depart, &Projectile{
		X: depart.X, Y: depart.Y,
		Step:      Vec{X: One / 8},
		Remaining: FromInt(4),
		Hits:      1,
	}) {
		t.Fatal("le projectile n'a touché personne")
	}

	if n := w.Gems().Len(); n != 1 {
		t.Fatalf("%d gemme(s) au sol, attendu 1", n)
	}
	if g := w.Gems().At(0); g.X != x || g.Y != y {
		t.Errorf("gemme en (%v, %v), attendu (%v, %v) : elle ne tombe pas où la "+
			"créature est morte", g.X, g.Y, x, y)
	}
}

// TestDeuxProjectilesNeDonnentQuUneVolee éprouve le butin sur le cas limite que
// la doctrine de test nomme : deux tirs atteignant la même créature dans le même
// tick.
//
// Le butin est attaché à la **transition** et non à l'état : une créature dont
// la résistance est tombée reste en place jusqu'à la fin du tick, et rattacher
// la volée à « sa résistance est nulle » la ferait repartir à chaque projectile
// qui la traverse. Le joueur ramasserait deux fois ce qu'il a tué une fois.
//
// Le second projectile survit et va chercher derrière, ce que
// `TestUnMortCesseDEtreUneCibleSansQuitterLeBassin` garde de son côté : celui-ci
// ne dit rien du projectile, seulement de ce qui tombe au sol.
//
// **Trois projectiles et non deux, et le compte est ce qui rend le tick
// discriminant.** Un projectile qui touche libère sa place, où la suppression
// par échange remonte le dernier du bassin — lequel attend le tick suivant. À
// deux, le second n'est donc jamais traité dans le même tick que le premier, et
// le test passerait même si le butin repartait à chaque touche. À trois, la
// place remontée est occupée par le troisième et le deuxième garde la sienne.
func TestDeuxProjectilesNeDonnentQuUneVolee(t *testing.T) {
	w, profils := champSansTir(t)
	px, py := w.Player()

	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(3), py); !ok {
		t.Fatal("créature refusée")
	}
	e := w.Enemies().At(0)
	e.Hits = 1

	for range 3 {
		if _, ok := w.Shots().Spawn(Projectile{
			X: e.X - One/8, Y: e.Y,
			Step:      Vec{X: One / 8},
			Remaining: FromInt(4),
			Hits:      1,
		}); !ok {
			t.Fatal("projectile refusé")
		}
	}

	w.Step(Vec{})

	if n := w.Gems().Len(); n != 1 {
		t.Errorf("%d gemme(s) pour une seule mort : le butin suit l'état et non "+
			"la transition", n)
	}
}

// TestUneVoleeNeSEmpilePas vérifie que plusieurs gemmes d'une même mort occupent
// des positions distinctes.
//
// **Aucune donnée livrée ne produit ce cas** : tous les profils du manifeste
// déclarent une seule gemme, la quantité par profil étant un réglage
// d'équilibrage que rien n'a encore mesuré. Le profil est donc monté ici à cinq,
// et ce test ne prouve pas que des volées existent en jeu — seulement que le
// chemin qui les pose ne superpose rien. Le lire comme la preuve du contraire
// est exactement le renversement contre lequel la doctrine de test met en garde.
//
// **Dix, et le compte n'est pas arbitraire.** Le vrai risque n'est pas que deux
// gemmes voisines se collent — la table les écarte de cent trente-cinq degrés —
// mais qu'elle **repasse** : elle n'a que huit entrées, et le rayon doit croître
// d'un tour au suivant. La première gemme reste au centre et les rangs un à sept
// prennent sept directions distinctes ; le rang huit prend la huitième, encore
// libre ; c'est le rang **neuf** qui retombe sur la direction du rang un. En
// deçà de dix gemmes, un rayon constant passerait ce test.
func TestUneVoleeNeSEmpilePas(t *testing.T) {
	const volee = 10

	w, profils := champSansTir(t)
	px, py := w.Player()

	profil := indexDuProfil(t, profils, "marcheur")
	profils.Enemies[profil].Gems = volee

	if _, ok := w.SpawnEnemy(profil, px+FromInt(3), py); !ok {
		t.Fatal("créature refusée")
	}
	e := w.Enemies().At(0)
	e.Hits = 1

	depart := Vec{X: e.X - One/8, Y: e.Y}
	if !w.toucher(depart, &Projectile{
		X: depart.X, Y: depart.Y,
		Step:      Vec{X: One / 8},
		Remaining: FromInt(4),
		Hits:      1,
	}) {
		t.Fatal("le projectile n'a touché personne")
	}

	if n := w.Gems().Len(); n != volee {
		t.Fatalf("%d gemme(s) au sol, attendu %d", n, volee)
	}
	for i := range w.Gems().Active() {
		for j := range w.Gems().Active()[:i] {
			a, b := w.Gems().At(i), w.Gems().At(j)
			if a.X == b.X && a.Y == b.Y {
				t.Errorf("gemmes %d et %d au même point (%v, %v)", j, i, a.X, a.Y)
			}
		}
	}
}

// TestLeRamassageSArreteALaPortee éprouve les deux côtés de la portée dans le
// même tick.
//
// **Les deux, parce qu'un seul ne prouverait rien.** Un ramassage qui viderait le
// bassin sans regarder la distance passerait un test qui ne pose qu'une gemme à
// portée ; c'est celle qui reste au sol qui établit que la portée décide.
//
// Elle est posée à trois tuiles, soit largement au-delà de ce que le joueur
// atteint, pour que le test ne dépende pas du réglage exact — lequel se fera en
// jouant, avec la durée de vie d'une gemme dont la conception dit qu'elle forme
// un couple avec lui.
func TestLeRamassageSArreteALaPortee(t *testing.T) {
	w, _ := champSansTir(t)
	px, py := w.Player()

	if _, ok := w.Gems().Spawn(Gem{X: px, Y: py}); !ok {
		t.Fatal("gemme refusée")
	}
	if _, ok := w.Gems().Spawn(Gem{X: px + FromInt(3), Y: py}); !ok {
		t.Fatal("gemme refusée")
	}

	w.Step(Vec{})

	if n := w.Gems().Len(); n != 1 {
		t.Fatalf("%d gemme(s) au sol après un tick, attendu 1", n)
	}
	if g := w.Gems().At(0); g.X == px && g.Y == py {
		t.Error("la gemme ramassée est celle des pieds du joueur, l'autre a disparu")
	}
}
