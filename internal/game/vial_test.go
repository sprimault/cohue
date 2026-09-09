// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la fiole : ce qu'une caisse en laisse, le ramassage en marchant
// dessus, le stock qui plafonne, et ce que boire rend — y compris à pleine vie,
// où le surplus se perd.

package game

import "testing"

// poserUneFiole pose une fiole sous les pieds du joueur, par le bassin.
//
// Le pendant de `tomber` pour l'arme lourde, et pour la même raison : ce que ces
// cas éprouvent est ce qui suit la chute, et passer par une caisse ferait
// dépendre chacun d'un tirage. La chute est gardée par
// `TestUneCaisseLaisseParfoisUneFiole`.
func poserUneFiole(t *testing.T, w *World) {
	t.Helper()
	px, py := w.Player()
	if !w.SpawnVial(px, py) {
		t.Fatal("bassin des fioles au sol plein")
	}
}

// TestUneFioleSeRamasseEnMarchantDessus garde ce que la conception veut du geste.
//
// Aucun menu, aucune touche : la touche ne sert qu'à boire, et prendre se fait
// en passant dessus — comme une arme lourde sur un emplacement libre.
func TestUneFioleSeRamasseEnMarchantDessus(t *testing.T) {
	w, _ := champDeTir(t)
	poserUneFiole(t, w)

	w.ramasserUneFiole()

	if n := w.VialStock(); n != 1 {
		t.Errorf("%d fiole(s) tenue(s) après avoir marché dessus, attendu 1", n)
	}
	if w.Vials().Len() != 0 {
		t.Error("la fiole est restée au sol après avoir été prise")
	}
}

// TestUneFioleDeTropResteAuSol garde le plafond, et ce qu'il veut dire.
//
// **Le surplus n'est pas rogné, il attend.** En rogner ferait perdre au joueur
// ce qu'il vient de trouver ; la ramasser au-delà du plafond ferait de la fiole
// un second socle, quand la conception veut que la vie reste la seule ressource
// rare. Ce qui reste est la troisième voie : on revient la chercher une fois
// qu'on a bu.
func TestUneFioleDeTropResteAuSol(t *testing.T) {
	w, _ := champDeTir(t)
	stock := w.fiole.Stock
	if stock < 1 {
		t.Fatalf("le catalogue livré annonce un stock de %d", stock)
	}

	for range stock + 1 {
		poserUneFiole(t, w)
		w.ramasserUneFiole()
	}

	if n := w.VialStock(); n != stock {
		t.Errorf("%d fiole(s) tenue(s) pour un plafond de %d", n, stock)
	}
	if n := w.Vials().Len(); n != 1 {
		t.Errorf("%d fiole(s) au sol, attendu celle que le plafond a refusée", n)
	}
}

// TestBoireRendCeQueLaFioleAnnonce garde le soin et sa dépense d'un seul coup.
//
// **La vie manquante est prise plus grande que le soin**, ce qui sépare ce cas
// du suivant : contre une vie pleine, un soin qui ne rendrait rien passerait
// aussi bien qu'un soin juste.
func TestBoireRendCeQueLaFioleAnnonce(t *testing.T) {
	w, _ := champDeTir(t)
	soin := w.fiole.Heal
	if soin < 1 {
		t.Fatalf("le catalogue livré annonce un soin de %d", soin)
	}
	poserUneFiole(t, w)
	w.ramasserUneFiole()
	w.vie = w.profils.Player.Health - 2*soin

	w.Drink()

	if attendu := w.profils.Player.Health - soin; w.vie != attendu {
		t.Errorf("vie de %d après un soin de %d, attendu %d", w.vie, soin, attendu)
	}
	if n := w.VialStock(); n != 0 {
		t.Errorf("%d fiole(s) tenue(s) après avoir bu, attendu 0", n)
	}
}

// TestBoireAPleineVieGaspille garde une décision, pas un défaut.
//
// **Le surplus perdu est ce qui donne au joueur une raison de ne pas boire tout
// de suite**, donc ce qui fait de la fiole une décision plutôt qu'un cadeau. Un
// refus à pleine vie paraîtrait une correction évidente — la touche ne ferait
// rien plutôt que de gâcher — et retirerait cette décision. C'est le pendant de
// l'aimant déclenché à vide, que la conception assume de la même façon.
//
// Ce cas n'a donc pas de mutation qui le justifie : ce qu'il garde est le
// comportement qu'on croirait faux, et sa raison d'être est d'arrêter la main de
// qui viendrait le « corriger ».
func TestBoireAPleineVieGaspille(t *testing.T) {
	w, _ := champDeTir(t)
	poserUneFiole(t, w)
	w.ramasserUneFiole()

	w.Drink()

	if w.vie != w.profils.Player.Health {
		t.Errorf("vie de %d après un soin à pleine vie, attendu %d",
			w.vie, w.profils.Player.Health)
	}
	if n := w.VialStock(); n != 0 {
		t.Errorf("%d fiole(s) tenue(s) : le surplus n'a pas été perdu", n)
	}
}

// TestBoireSansFioleNeFaitRien garde ce qu'une touche pressée à vide doit faire.
//
// L'appelant est un clavier : ni consommation, ni avertissement. Le joueur est
// blessé pour que le cas discrimine — à pleine vie, un soin appliqué sans stock
// ne se verrait pas.
func TestBoireSansFioleNeFaitRien(t *testing.T) {
	w, _ := champDeTir(t)
	w.vie = 1

	w.Drink()

	if w.vie != 1 {
		t.Errorf("vie de %d après avoir bu sans fiole, attendu 1", w.vie)
	}
}

// TestUneCaisseLaisseParfoisUneFiole garde le second tirage du flux « butin ».
//
// Le compte n'est pas vérifié — une chance sur deux n'a pas de fréquence exacte
// sur un échantillon —, seulement qu'il en tombe et qu'il n'en tombe pas à
// chaque fois. Sans la seconde moitié, un code qui en lâcherait toujours
// passerait.
func TestUneCaisseLaisseParfoisUneFiole(t *testing.T) {
	w, _ := champDeTir(t)

	portees, essais := 0, 60
	for range essais {
		if w.tirerUneFiole() {
			portees++
		}
	}

	if portees == 0 {
		t.Errorf("aucune fiole sur %d caisses, la chance déclarée est de une sur %d",
			essais, w.progression.VialOdds)
	}
	if portees == essais {
		t.Error("chaque caisse a laissé une fiole : le tirage ne départage rien")
	}
}

// TestUneCaisseALaFioleEnLacheUne garde l'accord entre l'annonce et le butin.
//
// **C'est le seul endroit où l'annonce peut mentir.** Une caisse tire son
// contenu à l'apparition et le pose en cédant : si les deux se décidaient
// séparément, l'icône dirait une chose et le sol en donnerait une autre — un
// mensonge qu'aucun test de tirage ne verrait, les deux étant individuellement
// corrects.
func TestUneCaisseALaFioleEnLacheUne(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	c := reposerJusqua(t, w, x, y, LootVial)
	if cle, porte := w.CrateContent(c); !porte || cle != w.VialKey() {
		t.Fatalf("la caisse annonce « %s », attendu « %s »", cle, w.VialKey())
	}

	casserLaCaisse(t, w, x, y)

	if n := w.Vials().Len(); n != 1 {
		t.Fatalf("%d fiole(s) au sol après une caisse qui en annonçait une", n)
	}
	if n := w.Drops().Len(); n != 0 {
		t.Errorf("%d arme(s) au sol pour une caisse qui annonçait une fiole", n)
	}
}

// TestUneCaisseALArmeNeLachePasDeFiole garde que la casse pose une chose et une
// seule.
//
// **Ce n'est pas le tirage qu'il éprouve, et l'avoir cru est instructif.** La
// sorte porte l'exclusivité par construction — un champ ne vaut qu'une valeur —,
// si bien qu'inverser la priorité de `tirerLeButin` ne fait tomber personne : une
// caisse d'arme reste une caisse d'arme. Ce qui peut mentir est la pose, où deux
// branches se suivent : ajouter une fiole sur la branche de l'arme fait tomber ce
// cas, et rien d'autre.
func TestUneCaisseALArmeNeLachePasDeFiole(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	reposerJusqua(t, w, x, y, LootHeavy)

	casserLaCaisse(t, w, x, y)

	if n := w.Vials().Len(); n != 0 {
		t.Errorf("%d fiole(s) au sol pour une caisse qui annonçait une arme", n)
	}
}
