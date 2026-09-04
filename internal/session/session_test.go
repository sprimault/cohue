// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le montage de la partie publiée, sans rien injecter : la chaîne entière depuis
// l'embed jusqu'au joueur posé, ce qu'une relance efface, et la graine dont
// descend la suite des runs.

package session

import (
	"testing"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
)

// graineDeTest est celle sur laquelle les parties de ce fichier se montent.
//
// Une valeur quelconque, mais nommée : lue dans un appel, elle dit qu'on monte
// la partie livrée et non qu'on éprouve un cas de graine particulier.
const graineDeTest uint64 = 20260902

// TestPartieLivreeSeMonte monte le jeu publié par le chemin qu'empruntent les
// deux binaires, et n'injecte rien.
//
// Il va plus loin que `TestLieuLivre`, dans `internal/level`, qui s'arrête à la
// grille de coûts : ni profils, ni armes, ni joueur n'y sont montés, si bien
// qu'un manifeste de personnages devenu illisible le laisserait au vert. Celui-ci
// tombe. À l'inverse, il ne dit rien de ce que la cuisson a mis dans chaque case,
// que l'autre relève une par une — supprimer l'un des deux laisse donc une moitié
// de la chaîne sans épreuve.
//
// Ce qu'il garde et que rien d'autre ne garde : **le joueur est posé sur une
// case franchissable**. `placer` le met au centre du lieu sans rien vérifier, et
// `World.Place` ne rattrape rien par principe — un point de départ dans un mur
// est un défaut du niveau. Le jour où le lieu livré changera de dessin, c'est ce
// test qui le dira plutôt qu'une partie où l'on ne peut pas bouger.
func TestPartieLivreeSeMonte(t *testing.T) {
	partie, err := Open(cohue.Assets, cohue.StartingCampaign, graineDeTest)
	if err != nil {
		t.Fatalf("montage de la partie livrée : %v", err)
	}

	// **La horde n'existe pas au montage, elle s'achète** — et c'est la chaîne
	// entière que ces deux relevés éprouvent sur les données publiées : le
	// scénario lu dans le lieu, le budget accumulé, et l'anneau qui trouve une
	// case franchissable à dix-neuf tuiles du joueur. Zéro créature après
	// quelques secondes signifierait qu'il n'en trouve aucune, ce qu'un lieu trop
	// étroit ou un rayon mal réglé produirait en silence.
	//
	// Dix secondes, sans chercher le compte exact : la courbe livrée s'ouvre à
	// une pression d'un par seconde pour des créatures qui en coûtent trois, et
	// un relevé serré ici encoderait ce réglage-là plutôt que la propriété.
	if n := partie.World.Enemies().Len(); n != 0 {
		t.Errorf("%d créature(s) au premier tick, alors que la horde s'achète", n)
	}
	for range 10 * game.TPS {
		partie.World.Step(game.Vec{})
	}
	if n := partie.World.Enemies().Len(); n == 0 {
		t.Error("aucune créature après dix secondes de jeu sur le lieu livré")
	}

	if partie.Tile != [2]int{64, 32} {
		t.Errorf("tuile %v, attendu [64 32] — la taille du manifeste ne voyage pas",
			partie.Tile)
	}

	x, y := partie.World.Player()
	u, v := x.Floor(), y.Floor()
	if !partie.Grid.Passable(u, v) {
		t.Errorf("le joueur est posé en (%d, %d), qui ne se franchit pas", u, v)
	}
}

// TestLaRelanceNeConserveRienDeLaPartie fixe la règle du remontage.
//
// **Ce qui compte n'est pas que la partie se remonte, mais ce qu'elle emporte.**
// Une relance qui garderait la vie entamée, la horde en place ou le tick courant
// serait une reprise et non une nouvelle run, et le joueur ne le verrait
// qu'après coup — au moment où il meurt deux fois plus vite qu'il ne devrait.
//
// **L'absence de survivant est énumérée plutôt que supposée**, et c'est ce qui
// fait de ce test une garde : le premier élément de méta-progression devra le
// modifier pour entrer, donc le décider. Ce qui traverse aujourd'hui est ce que
// la partie n'a pas touché — les tables du manifeste et le lieu cuit —, et rien
// de cela n'est un état de jeu.
//
// **La graine échappe à cette énumération parce qu'elle n'est pas conservée mais
// remplacée**, et c'est `TestLaSuiteDesRunsDescendDeLaGraineDeDepart` qui dit ce
// qu'elle devient. Sans ce renvoi, ce test-ci se lirait comme si rien du tout ne
// reliait deux runs, ce qui est faux depuis que la relance en dérive une.
func TestLaRelanceNeConserveRienDeLaPartie(t *testing.T) {
	partie, err := Open(cohue.Assets, cohue.StartingCampaign, graineDeTest)
	if err != nil {
		t.Fatalf("montage de la partie livrée : %v", err)
	}

	monde, grille, tuile := partie.World, partie.Grid, partie.Tile
	vie := monde.Health()
	semis := monde.Enemies().Len()

	// Jouer assez pour que tout ait bougé : le tick avance et le spawner a posé
	// de quoi changer le compte des vivants. Dix secondes plutôt que trois, la
	// courbe livrée commençant à une pression d'un par seconde pour des créatures
	// qui en coûtent trois — un compte serré ici encoderait ce réglage-là.
	for range 10 * game.TPS {
		monde.Step(game.Vec{})
	}
	if monde.Tick() == 0 || monde.Enemies().Len() == semis {
		t.Fatal("la partie n'a pas assez avancé : la relance n'aurait rien à effacer")
	}

	partie.Restart()

	if partie.World == monde {
		t.Fatal("la relance rend le même monde : l'état précédent y survit en entier")
	}
	if got := partie.World.Health(); got != vie {
		t.Errorf("vie de %d après la relance, attendu %d", got, vie)
	}
	if got := partie.World.Tick(); got != 0 {
		t.Errorf("tick à %d après la relance, attendu 0", got)
	}
	if got := partie.World.Enemies().Len(); got != semis {
		t.Errorf("%d créatures après la relance, attendu le semis de %d", got, semis)
	}
	if got := partie.World.Shots().Len(); got != 0 {
		t.Errorf("%d projectiles encore en vol après la relance", got)
	}

	// Le lieu et sa taille de tuile traversent, parce que la partie ne les a pas
	// touchés : les recuire rendrait les mêmes octets pour le prix d'un
	// décodage complet.
	if partie.Grid != grille || partie.Tile != tuile {
		t.Error("la relance recharge le lieu, qu'aucune partie ne modifie")
	}
}

// TestLaRelanceRefermeLaPorte garde la raison pour laquelle l'ouverture ne vit
// pas dans la carte.
//
// **La carte est partagée par toutes les runs d'une session**, ce que le
// remontage énonce comme une propriété par construction : ce qui survit est ce
// que la partie n'a pas touché. Ouvrir la porte en changeant le coût de sa case
// aurait été la voie la plus courte, et elle transformait la grille en état de
// jeu — la run suivante serait partie porte ouverte, gagnée sans avoir rien
// abattu.
//
// Le cas force l'objectif plutôt que de l'atteindre en jouant : ce qu'il éprouve
// est ce que la relance efface, pas la façon dont la porte s'ouvre.
func TestLaRelanceRefermeLaPorte(t *testing.T) {
	partie, err := Open(cohue.Assets, cohue.StartingCampaign, graineDeTest)
	if err != nil {
		t.Fatalf("montage de la partie livrée : %v", err)
	}
	sortie := partie.World.Exit()
	if sortie == nil {
		t.Fatal("le lieu livré n'a pas de porte : ce cas ne garde rien")
	}

	for !partie.World.DoorOpen() {
		if !partie.World.Alive() {
			t.Fatalf("mort avant l'ouverture, %d abattus sur %d",
				partie.World.Kills(), sortie.Kills)
		}
		partie.World.Step(Pilot(partie.World.Tick()))
	}

	partie.Restart()

	if partie.World.DoorOpen() {
		t.Error("la porte est ouverte au premier tick de la relance")
	}
	if got := partie.World.Kills(); got != 0 {
		t.Errorf("%d abattus après la relance, attendu 0", got)
	}
	if partie.World.Exit() == nil {
		t.Error("la relance a perdu la porte, que le lieu porte et non la partie")
	}
}

// TestLaRelanceRepose LesCaisses garde ce que la salle redevient après une mort.
//
// Une salle dont les caisses resteraient cassées ne serait pas la même salle :
// le joueur qui relance trouverait un lieu vidé de ce qu'il a déjà pris, sans
// que rien ne l'explique — et la deuxième run serait plus difficile que la
// première pour une raison qui n'appartient ni à la graine ni à la courbe.
func TestLaRelanceReposeLesCaisses(t *testing.T) {
	partie, err := Open(cohue.Assets, cohue.StartingCampaign, graineDeTest)
	if err != nil {
		t.Fatalf("montage de la partie livrée : %v", err)
	}
	semis := partie.World.Crates().Len()
	if semis == 0 {
		t.Fatal("le lieu livré ne pose aucune caisse : ce cas ne garde rien")
	}

	for partie.World.Crates().Len() == semis {
		if !partie.World.Alive() {
			t.Fatal("mort sans avoir cassé une seule caisse")
		}
		partie.World.Step(Pilot(partie.World.Tick()))
	}

	partie.Restart()

	if got := partie.World.Crates().Len(); got != semis {
		t.Errorf("%d caisse(s) après la relance, attendu le semis de %d", got, semis)
	}
}

// TestLaSuiteDesRunsDescendDeLaGraineDeDepart garde ce que la relance fait de la
// graine, et non ce qu'elle en calcule.
//
// Deux propriétés, mesurées sur les tirages d'une partie montée et non sur le
// champ `Seed` : c'est ce qui distingue une graine branchée d'une graine tenue à
// côté. Une dérivation parfaite que `monter` n'utiliserait pas laisserait toutes
// les runs identiques, et une lecture du seul champ ne le dirait pas.
//
// **La première run tourne sur la graine reçue**, non sur sa dérivée : ouvrir une
// session sur une graine et en jouer une autre se verrait au premier lieu de
// défi partagé, c'est-à-dire trop tard.
//
// **Les runs d'une session tirent différemment**, sinon mourir ne changerait
// rien. **Et deux sessions ouvertes sur la même graine tirent pareil**, sinon la
// suite ne se rejoue pas — c'est cette moitié-là qui rend une mort injuste
// reproductible.
//
// `TestLaGraineDeriveeEstStableEtNouvelle`, dans `internal/game`, garde la
// dérivation elle-même : ce test-ci ne verrait pas un cycle court, puisqu'il ne
// déroule que quatre runs.
func TestLaSuiteDesRunsDescendDeLaGraineDeDepart(t *testing.T) {
	// Quatre runs, dont le premier tirage de vagues est relevé avant de relancer.
	// Le flux n'est lu par personne d'autre : ce qu'on relève est bien le premier
	// tirage de la run et non un état laissé par la précédente.
	suite := func() []int {
		partie, err := Open(cohue.Assets, cohue.StartingCampaign, graineDeTest)
		if err != nil {
			t.Fatalf("montage de la partie livrée : %v", err)
		}
		if partie.Seed != graineDeTest {
			t.Fatalf("la première run tourne sur %d, et non sur la graine reçue %d",
				partie.Seed, graineDeTest)
		}

		tirages := make([]int, 4)
		for i := range tirages {
			tirages[i] = partie.World.Streams().Waves.IntN(1 << 30)
			partie.Restart()
		}
		return tirages
	}

	a, b := suite(), suite()

	for i, x := range a {
		for j, y := range a[:i] {
			if x == y {
				t.Errorf("les runs %d et %d tirent la même chose (%d) : la relance ne change pas de graine",
					j, i, x)
			}
		}
		if b[i] != x {
			t.Errorf("run %d : %d puis %d sur la même graine de départ", i, x, b[i])
		}
	}
}
