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

// autreLourde ajoute une arme lourde à la table de la partie et rend son rang.
//
// **Elle ne sert plus qu'au troisième type**, celui qui déborde des deux
// emplacements : le fusil à pompe a rendu les deux premiers exprimables par les
// données livrées, et c'est par elles que les cas y arrivent maintenant. Ce qui
// reste forgé est ce qu'aucun manifeste ne produit — une arme de plus que le
// joueur ne peut en tenir —, et c'est le cas que la doctrine autorise : isoler
// un critère, jamais contourner le montage.
//
// Une copie de la grenade sous une autre clé : ce qu'on éprouve est le type, pas
// les valeurs.
func autreLourde(t *testing.T, w *World, cle string) int {
	t.Helper()
	autre := w.armes.All[rangDeLArme(t, w, "grenade")]
	autre.Key = cle

	w.armes.All = append(w.armes.All, autre)
	rang := len(w.armes.All) - 1
	w.armes.Heavy = append(w.armes.Heavy, rang)
	return rang
}

// tomberRang pose au sol l'arme d'un rang donné, sous le joueur.
func tomberRang(t *testing.T, w *World, rang int) {
	t.Helper()
	px, py := w.Player()
	if _, ok := w.armesAuSol.Spawn(Drop{X: px, Y: py, Weapon: rang}); !ok {
		t.Fatal("bassin des armes au sol plein")
	}
}

// TestUneArmeDuMemeTypeSAjouteAuStock garde ce qu'une partie jouée a demandé.
//
// **Deux cases de grenades n'étaient pas un choix, c'était une perte.** Le
// catalogue ne portant qu'une lourde, les deux emplacements ne pouvaient tenir
// que des doublons, et la troisième ramassée en faisait disparaître une — alors
// que la conception veut que contourner laisse l'arme, c'est-à-dire que le joueur
// décide. Une arme du même type s'ajoute donc au stock, et le second emplacement
// reste libre pour ce qui n'est pas elle.
func TestUneArmeDuMemeTypeSAjouteAuStock(t *testing.T) {
	w, _ := champDeTir(t)

	tomber(t, w, "grenade")
	w.ramasserUneArme()
	arme, pleine := w.HeldHeavy(0)

	tomber(t, w, "grenade")
	w.ramasserUneArme()

	if _, charges := w.HeldHeavy(0); charges != 2*pleine {
		t.Errorf("%d tir(s) après deux grenades, attendu %d", charges, 2*pleine)
	}
	if _, charges := w.HeldHeavy(1); charges != 0 {
		t.Errorf("le second emplacement porte %d tir(s) : une grenade y est allée "+
			"alors que la première case la tenait déjà", charges)
	}
	if w.armesAuSol.Len() != 0 {
		t.Error("une arme est restée au sol alors que son type était tenu")
	}
	if arme.Key != "grenade" {
		t.Errorf("tenue : %q, attendu grenade", arme.Key)
	}
}

// TestLeStockPleinLaisseLArmeAuSol garde le plafond et ce qu'il refuse de faire.
//
// **Le surplus n'est pas rogné, il attend.** Remplir jusqu'au plafond et faire
// disparaître le reste ferait perdre une arme sans qu'aucun écran ne le dise ;
// laissée au sol, elle se reprend quand le stock a baissé. C'est ce que le
// plafond veut dire, et c'est pourquoi la condition porte sur le stock entier.
//
// **La case pleine ne déborde pas non plus sur la seconde**, qui redeviendrait la
// deuxième case du même type que la règle vient de fermer.
func TestLeStockPleinLaisseLArmeAuSol(t *testing.T) {
	w, _ := champDeTir(t)

	tomber(t, w, "grenade")
	w.ramasserUneArme()
	arme, pleine := w.HeldHeavy(0)

	// Autant de ramassages que le plafond en accepte, sans l'écrire : le compte
	// suit le manifeste au lieu de le redire, et la boucle est bornée pour qu'un
	// ramassage muet s'arrête au lieu de tourner.
	plein := pleine
	for range arme.Stock {
		if plein+pleine > arme.Stock {
			break
		}
		tomber(t, w, "grenade")
		w.ramasserUneArme()

		_, apres := w.HeldHeavy(0)
		if apres == plein {
			t.Fatalf("le stock reste à %d sous un plafond de %d", plein, arme.Stock)
		}
		plein = apres
	}

	tomber(t, w, "grenade")
	w.ramasserUneArme()

	if _, reste := w.HeldHeavy(0); reste != plein {
		t.Errorf("%d tir(s) après un ramassage de trop, attendu %d", reste, plein)
	}
	if w.armesAuSol.Len() != 1 {
		t.Error("l'arme de trop n'est pas restée au sol")
	}
	if _, charges := w.HeldHeavy(1); charges != 0 {
		t.Error("le second emplacement a pris le surplus du premier")
	}
}

// TestLesDeuxEmplacementsSeRemplissentPuisSArretent garde la borne.
//
// **Deux et pas trois** : la conception en fait une règle, le joueur ayant une
// décision — laquelle garder — et non une gestion. La troisième arme reste au
// sol, ce qui est la condition de l'échange.
//
// **Trois types, puisque la borne porte sur eux et non sur les exemplaires.**
// Deux du même type s'empilent désormais dans une seule case, si bien qu'une
// borne éprouvée sur des grenades ne dirait plus rien.
func TestLesDeuxEmplacementsSeRemplissentPuisSArretent(t *testing.T) {
	w, _ := champDeTir(t)

	tomber(t, w, "grenade")
	w.ramasserUneArme()
	tomber(t, w, "fusil")
	w.ramasserUneArme()

	for place := range Slots {
		if _, charges := w.HeldHeavy(place); charges == 0 {
			t.Errorf("emplacement %d vide après deux types ramassés", place)
		}
	}

	troisieme := autreLourde(t, w, "troisieme_lourde")
	tomberRang(t, w, troisieme)
	w.ramasserUneArme()
	if w.armesAuSol.Len() != 1 {
		t.Error("un troisième type a été pris alors que les deux places sont tenues")
	}
}

// TestLaToucheEchangeQuandLesDeuxPlacesSontTenues garde la règle unique.
//
// **La touche d'un emplacement, pressée sur une arme au sol, y met cette arme** —
// vide ou plein. Les trois autres lectures possibles — remplacer la plus
// ancienne, la moins chargée ou la première — font perdre une arme sans que le
// joueur sache laquelle, et aucun aperçu au sol ne peut le lui dire à l'avance.
//
// **Deux types, depuis que le même s'ajoute.** Une grenade posée sur une case de
// grenades n'échange plus rien, elle s'y ajoute : ce que la touche tranche est le
// cas où l'arme au sol n'est d'aucun des deux types tenus.
func TestLaToucheEchangeQuandLesDeuxPlacesSontTenues(t *testing.T) {
	w, _ := champDeTir(t)

	tomber(t, w, "grenade")
	w.ramasserUneArme()
	tomber(t, w, "fusil")
	w.ramasserUneArme()

	// Une arme entamée dans la seconde place : c'est ce qui rend l'échange
	// visible, la neuve arrivant pleine.
	w.lourdes[1].Charges = 1
	troisieme := autreLourde(t, w, "troisieme_lourde")
	tomberRang(t, w, troisieme)

	if !w.TakeDrop(1) {
		t.Fatal("l'échange a été refusé alors qu'une arme est sous les pieds")
	}
	arme, charges := w.HeldHeavy(1)
	if charges != arme.Charges {
		t.Errorf("la place échangée porte %d charge(s), attendu une arme pleine", charges)
	}
	if arme.Key != "troisieme_lourde" {
		t.Errorf("la place échangée tient %q, attendu l'arme ramassée", arme.Key)
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
//
// **Le tirage se relève sur ce qu'une caisse porte et non sur ce qu'elle lâche**,
// depuis qu'elle l'annonce avant de céder : ce qui décide est l'apparition, et
// la casse ne fait plus que poser au sol ce qui était déjà décidé.
func TestUneCaisseLaisseParfoisUneArme(t *testing.T) {
	w, _ := champDeTir(t)

	portees, essais := 0, 60
	for range essais {
		if _, tiree := w.tirerUneArme(); tiree {
			portees++
		}
	}

	if portees == 0 {
		t.Errorf("aucune arme sur %d caisses, la chance déclarée est de une sur %d",
			essais, w.progression.HeavyOdds)
	}
	if portees == essais {
		t.Error("chaque caisse a laissé une arme : le tirage ne départage rien")
	}
}

// TestUneCaisseLacheCeQuElleAnnonce garde l'accord entre les deux moments.
//
// **C'est ce que l'annonce promet, et le seul endroit où elle peut mentir.** Une
// caisse tire son contenu à l'apparition et le pose en cédant : si les deux se
// décidaient séparément, l'icône dirait une chose et le sol en donnerait une
// autre — un mensonge qu'aucun test de tirage ne verrait, les deux étant
// individuellement corrects.
func TestUneCaisseLacheCeQuElleAnnonce(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	c := reposerJusqua(t, w, x, y, LootHeavy)
	annoncee, porte := w.CrateContent(c)
	if !porte {
		t.Fatal("la caisse porte une arme que sa lecture ne rend pas")
	}

	casserLaCaisse(t, w, x, y)

	if n := w.Drops().Len(); n != 1 {
		t.Fatalf("%d arme(s) au sol après une caisse qui en annonçait une", n)
	}
	if lachee := w.DropWeapon(w.Drops().At(0)); lachee.Key != annoncee {
		t.Errorf("« %s » au sol pour « %s » annoncée", lachee.Key, annoncee)
	}
}

// reposerJusqua repose la caisse du lieu jusqu'à en obtenir une qui porte une
// arme, ou qui n'en porte pas.
//
// **Elle repose plutôt qu'elle n'écrit le champ**, ce qui garde le tirage dans
// le chemin : une caisse forgée à la main éprouverait ce que la casse fait d'un
// contenu, jamais l'accord entre ce qui est tiré et ce qui est lâché.
//
// **Et elle est bornée**, ce qu'une mutation a montré : à contenu forcé à vide,
// une boucle qui attend une arme ne s'arrête jamais, et le cas pendait au lieu
// d'échouer. Une chance sur deux rend trente essais suffisants au-delà de tout
// doute, et la borne dit ce qu'elle attendait plutôt que de laisser lire une
// interruption.
//
// **Elle attend une sorte et non un booléen**, depuis que la fiole en fait une
// troisième : « garnie ou non » ne savait plus dire laquelle des deux garnitures
// on cherchait.
func reposerJusqua(t *testing.T, w *World, x, y Fixed, sorte LootKind) *Crate {
	t.Helper()
	for range 30 {
		c := w.Crates().At(0)
		if c.Content.Kind == sorte {
			return c
		}
		w.Crates().RemoveAt(0)
		if _, ok := w.SpawnCrate(x, y, Free); !ok {
			t.Fatal("bassin de caisses plein")
		}
	}
	t.Fatalf("trente caisses posées sans en obtenir une de sorte %d, pour une "+
		"arme sur %d et une fiole sur %d de ce qui reste",
		sorte, w.progression.HeavyOdds, w.progression.VialOdds)
	return nil
}

// TestUneCaisseSansArmeNenLachePas garde l'autre moitié de l'annonce.
//
// Sans elle, un code qui lâcherait toujours une arme passerait le cas
// précédent : il aurait posé au sol ce que la caisse annonçait, et aussi ce
// qu'aucune n'annonçait.
func TestUneCaisseSansArmeNenLachePas(t *testing.T) {
	w, x, y := salleAvecCaisse(t)
	reposerJusqua(t, w, x, y, LootNothing)

	casserLaCaisse(t, w, x, y)

	if n := w.Drops().Len(); n != 0 {
		t.Errorf("%d arme(s) au sol pour une caisse qui n'en annonçait aucune", n)
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

// TestUnDeclenchementDeSalveTireEtDepense ferme le chemin de la touche aux
// projectiles, comme son voisin le ferme jusqu'à l'explosion.
//
// **Il vérifie aussi qu'aucune déflagration ne part**, ce qui est la moitié que
// l'aiguillage peut manquer : une branche qui tomberait dans le cas par défaut
// laisserait le bassin des souffles vide et celui des tirs aussi, mais une
// branche qui ferait les deux passerait un cas qui ne compterait que les tirs.
func TestUnDeclenchementDeSalveTireEtDepense(t *testing.T) {
	w, profils := champDeTir(t)
	tomber(t, w, "fusil")
	w.ramasserUneArme()
	px, py := w.Player()
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), px+FromInt(2), py); !ok {
		t.Fatal("créature refusée")
	}
	arme, avant := w.HeldHeavy(0)

	w.Trigger(0)

	if _, apres := w.HeldHeavy(0); apres != avant-1 {
		t.Errorf("%d charge(s) après un déclenchement, attendu %d", apres, avant-1)
	}
	if n := w.tirs.Len(); n != arme.Projectiles {
		t.Errorf("%d projectile(s) en vol, attendu les %d de la salve", n, arme.Projectiles)
	}
	if n := w.souffles.Len(); n != 0 {
		t.Errorf("%d déflagration(s) pour une arme qui tire", n)
	}
}

// TestUneSalveNeDepenseRienSansCible garde ce que la conception refuse.
//
// **La règle vaut pour tous les effets et non pour la seule grenade** : une arme
// à huit charges dont une part dans le vide se lit comme un défaut, quel que
// soit ce qu'elle produit. Le cas part d'une salle vide de créatures, la portée
// courte du fusil rendant l'absence de cible ordinaire plutôt qu'exceptionnelle.
func TestUneSalveNeDepenseRienSansCible(t *testing.T) {
	w, _ := champDeTir(t)
	tomber(t, w, "fusil")
	w.ramasserUneArme()
	_, avant := w.HeldHeavy(0)

	w.Trigger(0)

	if _, apres := w.HeldHeavy(0); apres != avant {
		t.Errorf("%d charge(s) après un déclenchement à vide, attendu %d", apres, avant)
	}
	if n := w.tirs.Len(); n != 0 {
		t.Errorf("%d projectile(s) partis sans cible", n)
	}
}
