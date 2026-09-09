// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La caisse : ce qui la casse, et ce qu'elle laisse. Elle est posée par le lieu
// comme les figurants et la porte, et elle n'est pas une cible — c'est le joueur
// qui va la chercher, jamais son arme qui la trouve.

package game

import (
	"fmt"

	"github.com/sprimault/cohue/internal/manifest"
)

// CrateSpec est le semis de caisses tel qu'un lieu l'écrit.
type CrateSpec []CratePlacementSpec

// CratePlacementSpec pose une caisse à une case donnée.
//
// Une position et non un nombre, pour la raison qui a déjà tranché le
// peuplement : un semis tiré au sort abandonne en silence ce qui tombe dans un
// mur, et sauterait d'un coin à l'autre entre deux relances.
type CratePlacementSpec struct {
	manifest.Commentable
	// At est la case où la caisse repose, en coordonnées de lieu.
	At *[2]int `json:"position"`
}

// CratePlacement est une caisse compilée : une position et le sol qu'elle
// recouvre.
type CratePlacement struct {
	// X et Y sont sa position, au centre de la case écrite.
	X, Y Fixed
	// Floor est le coût que la case portait avant qu'une caisse s'y pose.
	//
	// **Relevé ici parce que c'est le dernier moment où la case est encore celle
	// du sol.** Le montage écrit le coût de la caisse dans la grille avant que le
	// champ de flux se bâtisse, donc avant qu'une `Crate` existe : une caisse qui
	// lirait sa case à l'apparition mémoriserait son propre coût comme étant
	// celui du sol, et le lui rendrait en cédant.
	Floor Cost
}

// Crate est une caisse posée dans le lieu.
//
// **Elle porte maintenant un état de partie**, ce qui la sépare enfin de
// `CratePlacement` : le décompte d'appui vit le temps d'une run, quand la
// position et le sol viennent du fichier et n'en bougent pas.
type Crate struct {
	// X et Y sont sa position dans le monde, en tuiles.
	X, Y Fixed
	// Press est ce qu'il reste d'appui avant qu'elle cède, en ticks.
	//
	// **Il repart entier dès que le joueur s'écarte.** Sans cela trois passages
	// successifs la casseraient, c'est-à-dire exactement ce que le délai existe
	// pour interdire : on ne casse pas en passant, on casse en décidant d'y
	// aller.
	Press Tick
	// Floor est le coût qu'elle rend à sa case en cédant, venu de son placement.
	Floor Cost
}

// CompileCrates résout un semis de caisses contre la carte cuite.
//
// Elle rend tout ce qui l'empêche de valoir plutôt que le premier écart, comme
// les vagues, le peuplement et la sortie.
//
// **Une caisse se pose sur une case franchissable, à l'inverse d'une porte.**
// Elle coûte à traverser plutôt qu'elle n'arrête, si bien qu'une caisse dans un
// mur serait un objet qu'on voit sans jamais l'atteindre.
//
// **Deux caisses sur une même case sont refusées, et c'est le coût qui l'exige.**
// Elles étaient jusqu'ici une redondance sans conséquence ; maintenant qu'une
// caisse écrit dans la grille, la première cassée rend la case au sol sous la
// seconde, qui coûterait alors ce que coûte le sol et ralentirait ce qu'elle
// devrait ralentir.
//
// La carte reçue est celle du lieu cuit, avant qu'aucune caisse n'y ait écrit :
// c'est de là que chaque placement tire le sol qu'il recouvre.
func CompileCrates(brut CrateSpec, carte *CostGrid) ([]CratePlacement, []string) {
	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	prises := make(map[[2]int]int, len(brut))
	pose := make([]CratePlacement, 0, len(brut))
	for i, c := range brut {
		ou := fmt.Sprintf("caisses[%d]", i)
		if c.At == nil {
			dire("%s.position : absente, une caisse se place", ou)
			continue
		}

		u, v := c.At[0], c.At[1]
		switch premiere, occupee := prises[[2]int{u, v}]; {
		case !carte.InBounds(u, v):
			dire("%s.position : (%d, %d) hors du lieu, qui fait %d sur %d",
				ou, u, v, carte.Width(), carte.Height())
		case !carte.Passable(u, v):
			dire("%s.position : (%d, %d) est dans un mur", ou, u, v)
		case occupee:
			dire("%s.position : (%d, %d) porte déjà caisses[%d] ; la première cassée "+
				"rendrait la case au sol sous la seconde", ou, u, v, premiere)
		default:
			prises[[2]int{u, v}] = i
			pose = append(pose, CratePlacement{
				X:     FromInt(u) + One/2,
				Y:     FromInt(v) + One/2,
				Floor: carte.At(u, v),
			})
		}
	}
	return pose, manques
}

// StampCrates écrit le coût de traversée des caisses dans la grille.
//
// **Avant que le monde soit bâti, et c'est une contrainte d'ordre** : le champ
// de flux dérive son nombre de seaux du plus grand coût qu'il trouve, si bien
// qu'un coût posé après lui n'aurait pas de seau où entrer. Ce qui suit ne fait
// que descendre — une caisse cassée rend sa case au sol —, et c'est ce qui rend
// l'ordre suffisant plutôt que fragile.
//
// **Elle vit ici et non dans le montage, à côté de ce qui rend la case.** Les
// deux moitiés du geste sont alors sous les yeux l'une de l'autre : `casser`
// réécrit exactement ce que cette passe a recouvert, et une caisse qui laisserait
// son coût derrière elle se verrait à deux lignes d'écart.
//
// La grille reçue est celle de la run, jamais le lieu cuit : écrire dans la carte
// partagée laisserait le coût d'une caisse à la run suivante.
func StampCrates(grille *CostGrid, caisses []CratePlacement, regles CrateRules) {
	for _, c := range caisses {
		grille.Set(c.X.Floor(), c.Y.Floor(), regles.Cost)
	}
}

// Stock pose les caisses du lieu, au montage et à chaque relance.
//
// Le pendant de `Populate` pour les figurants : les positions viennent du
// fichier et sont déjà vérifiées, cette passe ne décide de rien. Une salle dont
// les caisses resteraient cassées après une mort ne serait pas la même salle.
func (w *World) Stock(caisses []CratePlacement) {
	for _, c := range caisses {
		w.SpawnCrate(c.X, c.Y, c.Floor)
	}
}

// SpawnCrate pose une caisse sur un sol dont elle retient le coût.
//
// **Par ses valeurs et non par le placement compilé**, bien que les deux se
// ressemblent — et l'analyseur proposait d'ailleurs la conversion tant que les
// champs coïncidaient. Ce qu'il voyait était une similitude ; ce qu'il ne voyait
// pas est que `CratePlacement` est une donnée d'entrée validée au chargement et
// `Crate` un état de simulation. Le décompte d'appui a fini par les séparer,
// comme annoncé.
//
// L'appui part entier : une caisse qu'on trouve déjà entamée serait une caisse
// que quelqu'un a poussée avant que la partie commence.
func (w *World) SpawnCrate(x, y Fixed, sol Cost) (Handle, bool) {
	return w.caisses.Spawn(Crate{X: x, Y: y, Press: w.appuiCaisse, Floor: sol})
}

// Crates rend le bassin des caisses.
func (w *World) Crates() *Pool[Crate] { return w.caisses }

// CratePress rend le temps d'appui qu'une caisse intacte porte, en ticks.
//
// Le rendu en a besoin pour reconnaître une caisse qu'on pousse : le décompte
// seul ne dit pas d'où il part, et comparer à la valeur pleine est ce qui
// distingue « entamée » de « au repos ».
func (w *World) CratePress() Tick { return w.appuiCaisse }

// casser décompte l'appui du joueur sur les caisses qu'il touche, et vide celles
// qui cèdent.
//
// **Le joueur les casse, jamais son arme.** Une caisse rangée parmi les cibles
// détournerait la visée automatique, qui prend la plus proche sans que le joueur
// choisisse : chaque salve partirait vers le décor, et la mécanique du
// Secouriste — dont tout l'intérêt tient à cette visée — s'effondrerait avec.
// C'est la même règle que pour le figurant, et pour la même raison.
//
// Elle vient avec les ramassages, dont elle est le voisin naturel : ce qu'elle
// produit est une volée de gemmes, que la même passe ramasse au tick suivant.
//
// **La case rendue au sol avant tout le reste.** Une caisse qui laisserait son
// coût derrière elle ralentirait la horde sur un point qu'aucun pixel ne montre
// plus, et le champ de flux le contournerait au prochain rafraîchissement — un
// détour autour de rien.
func (w *World) casser() {
	if !w.Alive() {
		return
	}

	portee := w.progression.CrateRange
	for i := 0; i < w.caisses.Len(); {
		c := w.caisses.At(i)
		ecart := Vec{X: w.playerX - c.X, Y: w.playerY - c.Y}
		if ecart.carres() >= int64(portee)*int64(portee) {
			c.Press = w.appuiCaisse
			i++
			continue
		}
		// Le décompte tombe avant d'être lu : c'est ce qui fait céder la caisse au
		// vingtième tick d'appui et non au vingt et unième — un délai lu avant
		// d'être consommé compte le tick où l'on arrive comme un tick d'attente.
		c.Press--
		if c.Press > 0 {
			i++
			continue
		}

		// Les gemmes et les éclats se posent avant la suppression : après, la
		// caisse a quitté le bassin et sa position vient d'une copie qu'on aurait
		// gardée pour rien.
		w.grille.Set(c.X.Floor(), c.Y.Floor(), c.Floor)
		w.lacherEn(c.X, c.Y, w.progression.CrateGems)
		w.lacherUneArme(c.X, c.Y)
		w.emettre(c.X, c.Y, FxCrate)
		w.epaves.Spawn(Wreck{X: c.X, Y: c.Y, Born: w.tick})
		w.caisses.RemoveAt(i)
		// **La place libérée n'est pas réexaminée**, à la différence de ce que
		// faisait cette boucle quand une caisse n'avait pas d'état : la
		// suppression par échange y remonte la dernière, et la revoir ici lui
		// décompterait deux ticks d'appui en un. Elle attend le tick suivant,
		// comme le veut la règle des passes de mise à jour — deux caisses qui
		// cèdent au même tick cèdent donc à un tick d'écart, ce qui ne se voit
		// pas.
		i++
	}
}
