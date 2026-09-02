// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les deux tests qui livrent l'étape 1 — trois cents poursuivants qui convergent
// en contournant des obstacles, mille ticks sans une allocation — et les cas du
// déplacement : le mur qu'on ne traverse pas, l'angle qu'un pas diagonal ne coupe
// pas, le coût qui divise la vitesse.

package game

import (
	"testing"

	"github.com/sprimault/cohue"
)

// mondeDEssai monte une partie sur les profils livrés et une carte à obstacles.
//
// Les profils viennent du manifeste publié et non d'une table forgée : c'est ce
// qui fait de ce fichier une épreuve du montage entier, et pas seulement de la
// boucle. Un profil dont la vitesse cesserait d'être convertie casse ici.
func mondeDEssai(t *testing.T, largeur, hauteur int) (*World, *Profiles) {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, "assets/personnages/manifeste.json")
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	g := NewCostGrid(largeur, hauteur)
	for u := range largeur {
		g.Set(u, 0, Blocked)
		g.Set(u, hauteur-1, Blocked)
	}
	for v := range hauteur {
		g.Set(0, v, Blocked)
		g.Set(largeur-1, v, Blocked)
	}
	// Trois piliers et une cloison percée : de quoi obliger au contournement
	// plutôt qu'à la ligne droite.
	for v := 8; v < hauteur-8; v++ {
		if v != hauteur/2 {
			g.Set(largeur/2, v, Blocked)
		}
	}
	for _, p := range [][2]int{{10, 10}, {12, 30}, {40, 18}} {
		g.Set(p[0], p[1], Blocked)
		g.Set(p[0]+1, p[1], Blocked)
	}

	armes, err := LoadWeapons(cohue.Assets, manifesteArmes)
	if err != nil {
		t.Fatalf("armes livrées : %v", err)
	}
	return NewWorld(profils, armes.Base, g, graineDeTest, 300, 256), profils
}

// graineDeTest est celle sur laquelle les champs d'essai se montent.
//
// Sa valeur est indifférente et doit le rester : aucun test de ce paquet ne tire
// au sort, et celui qui le fera devra dire de quelle graine il dépend plutôt que
// d'hériter de celle-ci.
const graineDeTest uint64 = 1

// indexDuProfil rend la place d'un profil dans la table, ou arrête le test.
func indexDuProfil(t *testing.T, profils *Profiles, cle string) int {
	t.Helper()
	for i := range profils.Enemies {
		if profils.Enemies[i].Key == cle {
			return i
		}
	}
	t.Fatalf("« %s » absent de la table", cle)
	return 0
}

// peupler pose des créatures sur toutes les cases franchissables d'une carte,
// jusqu'à en avoir posé le compte demandé.
func peupler(t *testing.T, w *World, profil, combien int) {
	t.Helper()
	poses := 0
	for v := 1; v < 63 && poses < combien; v++ {
		for u := 1; u < 63 && poses < combien; u++ {
			if !w.grille.Passable(u, v) {
				continue
			}
			if _, ok := w.SpawnEnemy(profil, FromInt(u)+One/2, FromInt(v)+One/2); !ok {
				t.Fatalf("bassin plein après %d créatures", poses)
			}
			poses++
		}
	}
	if poses != combien {
		t.Fatalf("%d créatures posées, %d demandées", poses, combien)
	}
}

// distanceMoyenne rend l'éloignement moyen des créatures au joueur, en tuiles.
func distanceMoyenne(w *World) float64 {
	px, py := w.Player()
	var somme float64
	for i := range w.Enemies().Active() {
		e := w.Enemies().At(i)
		somme += (Vec{e.X - px, e.Y - py}).Len().Float()
	}
	return somme / float64(w.Enemies().Len())
}

// TestTroisCentsPoursuivantsConvergent est le critère de livraison de l'étape 1.
//
// Trois cents créatures, une cible qui se déplace, des obstacles à contourner.
// Ce n'est pas un test de valeurs mais d'une propriété : la horde se rapproche,
// et aucune de ses créatures n'a traversé un mur pour le faire.
//
// Il ne dit rien du dernier demi-tuile : une moyenne qui tombe de trente à
// quatorze reste vraie d'une horde qui s'arrête au bord de la case du joueur.
// C'est `TestLaHordeRejointVraimentLeJoueur` qui garde cette moitié-là.
func TestTroisCentsPoursuivantsConvergent(t *testing.T) {
	w, profils := mondeDEssai(t, 64, 64)
	marcheur := indexDuProfil(t, profils, "marcheur")

	w.Place(FromInt(32)+One/2, FromInt(32)+One/2)
	peupler(t, w, marcheur, 300)

	depart := distanceMoyenne(w)

	// Dix secondes, la cible dérivant vers le haut à gauche pour qu'aucune
	// créature ne l'atteigne par une simple ligne droite.
	const ticks = 10 * TPS
	for range ticks {
		w.Step(Vec{-One, -One})
	}

	arrivee := distanceMoyenne(w)
	if arrivee >= depart {
		t.Errorf("distance moyenne %0.2f au départ, %0.2f après %d ticks : la horde ne converge pas",
			depart, arrivee, ticks)
	}
	// La convergence doit être franche : une horde qui grignote un dixième de
	// tuile passerait la comparaison ci-dessus en ne poursuivant rien.
	if arrivee > depart/2 {
		t.Errorf("distance moyenne passée de %0.2f à %0.2f seulement", depart, arrivee)
	}

	for i := range w.Enemies().Active() {
		e := w.Enemies().At(i)
		if !w.passable(e.X, e.Y) {
			t.Fatalf("la créature %d est dans un mur, en (%0.2f, %0.2f)",
				i, e.X.Float(), e.Y.Float())
		}
	}
	t.Logf("distance moyenne : %0.2f -> %0.2f tuiles en %d ticks", depart, arrivee, ticks)
}

// distanceLaPlusCourte rend l'éloignement de la créature la plus proche.
func distanceLaPlusCourte(w *World) Fixed {
	px, py := w.Player()
	mini := FromInt(1 << 10)
	for i := range w.Enemies().Active() {
		e := w.Enemies().At(i)
		if d := (Vec{e.X - px, e.Y - py}).Len(); d < mini {
			mini = d
		}
	}
	return mini
}

// TestLaHordeRejointVraimentLeJoueur garde le dernier demi-tuile.
//
// **Ce que `TestTroisCentsPoursuivantsConvergent` ne garde pas.** Celui-là
// mesure une distance *moyenne* sur une cible en fuite, et il est resté vert
// pendant que la horde encerclait le joueur à un demi-tuile sans jamais le
// toucher : le champ de flux mène de case en case, sa direction est nulle dans
// celle de la cible, et la densité y éjectait les créatures qui y entraient.
// Une moyenne qui tombe de trente à quatorze ne dit rien du dernier demi-tuile,
// qui est pourtant le seul où le contact existe.
//
// Réciproquement, celui-ci ne dit rien du contournement d'obstacles ni de la
// convergence d'ensemble : une seule créature arrivée au but le ferait passer.
// Les deux gardent deux moitiés du même déplacement.
//
// La cible ne fuit pas, parce que c'est l'encerclement qu'on mesure : contre une
// cible plus rapide qu'elles, les créatures resteraient derrière sans que cela
// dise rien du rabattement.
func TestLaHordeRejointVraimentLeJoueur(t *testing.T) {
	w, profils := mondeDEssai(t, 64, 64)
	marcheur := indexDuProfil(t, profils, "marcheur")

	px, py := FromInt(32)+One/2, FromInt(32)+One/2
	w.Place(px, py)

	// Un anneau à trois tuiles plutôt que le semis d'ensemble : celui-ci part
	// d'un coin, et l'arme abat les arrivantes une à une, si bien qu'on
	// mesurerait la portée du tir au lieu de l'approche.
	for k := range 24 {
		a := Heading(k % Headings).Scale(FromInt(3))
		if _, ok := w.SpawnEnemy(marcheur, px+a.X, py+a.Y); !ok {
			t.Fatal("bassin plein")
		}
	}

	for range 5 * TPS {
		w.Step(Vec{})
	}

	// Le contact vaut la somme des deux rayons ; la plus proche doit être
	// franchement en deçà, sinon elle ne fait que le frôler par accident.
	contact := profils.Player.Radius + profils.Enemies[marcheur].Radius
	if mini := distanceLaPlusCourte(w); mini >= contact {
		t.Errorf("la plus proche est à %0.3f tuile, contact à %0.3f : la horde encercle sans toucher",
			mini.Float(), contact.Float())
	}
}

// TestLaBoucleNalloueRien est l'autre moitié du critère de l'étape.
//
// La mesure porte sur des ticks entiers et non sur la seule mise à jour des
// entités : le champ de flux, la grille de densité et les seaux ont chacun leurs
// tableaux, et c'est leur réutilisation d'un tick à l'autre qui est en jeu. Une
// allocation par tick ne se verrait pas sur une passe isolée.
//
// Le champ ne se reconstruit qu'un tick sur `flowPeriod`, donc la mesure couvre
// forcément des ticks avec et sans reconstruction — c'est voulu : c'est
// justement la reconstruction qui est le candidat le plus probable.
//
// **Ce n'est pas un doublon de `TestLeChampNalloueRien`**, malgré l'apparence.
// Celui-là garde une propriété du champ pris seul ; celui-ci garde qu'aucun
// assemblage n'alloue — un tri, une tranche d'indices, un tampon partagé entre
// deux sous-systèmes qui vont bien chacun de leur côté. Le second peut tomber
// alors que le premier passe, et c'est le second qu'on serait tenté de
// supprimer, parce qu'il est le plus lent à exécuter.
//
// `TestBassinNalloueRien` ne le remplace pas davantage : un tick ordinaire ne
// supprime pas assez d'entités pour éprouver l'échange du bassin, que celui-là
// exerce en propre.
func TestLaBoucleNalloueRien(t *testing.T) {
	w, profils := mondeDEssai(t, 64, 64)
	w.Place(FromInt(32)+One/2, FromInt(32)+One/2)
	peupler(t, w, indexDuProfil(t, profils, "marcheur"), 300)

	// Quelques ticks de préchauffage : les seaux atteignent leur capacité au
	// premier parcours, et c'est une allocation qu'on ne veut pas compter.
	for range 3 * flowPeriod {
		w.Step(Vec{One, 0})
	}

	moyenne := testing.AllocsPerRun(1000, func() {
		w.Step(Vec{One, 0})
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par tick à 300 entités, attendu aucune", moyenne)
	}
}

// TestRienNeTraverseUnMur éprouve la projection sur la passabilité.
//
// Une créature poussée droit dans une cloison perd sa composante bloquée et
// longe la paroi : c'est ce qui permet à la horde d'entasser un bloqueur contre
// un mur sans le faire passer au travers.
func TestRienNeTraverseUnMur(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, "assets/personnages/manifeste.json")
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	// Un couloir d'une case de haut, la cible derrière le mur du fond.
	g := NewCostGrid(6, 3)
	for u := range 6 {
		g.Set(u, 0, Blocked)
		g.Set(u, 2, Blocked)
	}
	g.Set(5, 1, Blocked)

	// Arme inerte : ce test isole le déplacement, et un joueur qui abat la
	// créature dont on suit la trajectoire ne mesurerait plus rien.
	w := NewWorld(profils, Weapon{}, g, graineDeTest, 4, 1)
	w.Place(FromInt(4)+One/2, FromInt(1)+One/2)
	if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), One/2+One, One/2+One); !ok {
		t.Fatal("créature refusée")
	}

	for range 5 * TPS {
		w.Step(Vec{})
		e := w.Enemies().At(0)
		if !w.passable(e.X, e.Y) {
			t.Fatalf("la créature est entrée dans un mur, en (%0.2f, %0.2f)", e.X.Float(), e.Y.Float())
		}
	}
}

// TestLeGlissementNeCoupeAucunAngle éprouve ce qu'un pas diagonal franchirait
// sans qu'aucune des deux cases qu'il longe ne l'arrête.
//
// Il garde la même règle que `TestLaDirectionNeCoupePasUnAngle`, à l'autre bout
// du tick. Celui-là garde ce que le champ **propose** — l'orientation d'une
// cellule ne désigne jamais une diagonale fermée par deux murs — et ne dit rien
// de ce qu'une entité fait de cette proposition ; celui-ci garde ce que le
// déplacement **applique**, y compris pour qui n'obéit pas au champ, à commencer
// par le joueur. Supprimer l'un laisse l'autre moitié sans protection.
//
// Les deux cas ne se gardent pas contre la même faute, et c'est pourquoi il en
// faut deux :
//
//   - **l'angle** éprouve la version qui ne testerait que la case d'arrivée.
//     Elle est libre, donc le pas passerait, en traversant l'arête où deux murs
//     se touchent — le défaut se voit tout de suite à l'écran, et se relie au
//     déplacement bien plus tard ;
//   - **le pilier** éprouve l'ordre des deux tests. Le second porte sur `nx`
//     déjà corrigé, jamais sur `x` : sur `x`, une entité qui longe un obstacle
//     par le côté libre finirait dans l'obstacle même.
func TestLeGlissementNeCoupeAucunAngle(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, "assets/personnages/manifeste.json")
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	cas := []struct {
		nom      string
		grille   []string
		depart   [2]int
		interdit [2]int
	}{
		{
			nom: "l'angle de deux murs",
			grille: []string{
				"..#.",
				".#..",
				"....",
				"....",
			},
			depart:   [2]int{1, 0},
			interdit: [2]int{2, 1},
		},
		{
			nom: "le coin d'un pilier",
			grille: []string{
				"....",
				"....",
				"..#.",
				"....",
			},
			depart:   [2]int{1, 1},
			interdit: [2]int{2, 2},
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			// Arme inerte et bassin d'une place : ce test isole le déplacement.
			w := NewWorld(profils, Weapon{}, grilleDepuis(c.grille...), graineDeTest, 1, 1)
			w.Place(FromInt(c.depart[0])+One/2, FromInt(c.depart[1])+One/2)

			// Vers le sud-est du monde, c'est-à-dire vers la case en diagonale.
			for range 2 * TPS {
				w.Step(Vec{X: One, Y: One})
				x, y := w.Player()
				if u, v := x.Floor(), y.Floor(); u == c.interdit[0] && v == c.interdit[1] {
					t.Fatalf("le joueur a franchi l'angle : (%d, %d) au tick %d",
						u, v, w.Tick())
				}
			}
		})
	}
}

// TestLeCoutDeLaCaseDiviseLaVitesse vérifie que le prix du terrain se paie au
// déplacement et pas seulement au calcul du chemin.
//
// Sans cela, le parcours pondéré serait une superstition : la horde
// contournerait la flaque et la traverserait à la même vitesse que le sol.
func TestLeCoutDeLaCaseDiviseLaVitesse(t *testing.T) {
	profils, err := LoadProfiles(cohue.Assets, "assets/personnages/manifeste.json")
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}

	// Deux couloirs identiques, l'un ordinaire et l'autre entièrement en flaque.
	parcours := func(cout Cost) Fixed {
		g := NewCostGrid(12, 3)
		for u := range 12 {
			g.Set(u, 0, Blocked)
			g.Set(u, 2, Blocked)
			g.Set(u, 1, cout)
		}
		w := NewWorld(profils, Weapon{}, g, graineDeTest, 1, 1)
		w.Place(FromInt(1)+One/2, FromInt(1)+One/2)
		depart := FromInt(10) + One/2
		if _, ok := w.SpawnEnemy(indexDuProfil(t, profils, "marcheur"), depart, FromInt(1)+One/2); !ok {
			t.Fatal("créature refusée")
		}
		for range TPS {
			w.Step(Vec{})
		}
		return depart - w.Enemies().At(0).X
	}

	ordinaire := parcours(Free)
	flaque := parcours(3)
	if flaque >= ordinaire {
		t.Errorf("parcouru %0.2f tuile(s) dans la flaque contre %0.2f sur le sol",
			flaque.Float(), ordinaire.Float())
	}
	t.Logf("en une seconde : %0.2f tuiles sur le sol, %0.2f dans la flaque",
		ordinaire.Float(), flaque.Float())
}
