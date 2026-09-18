// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'obstacle fragile : ce que l'auteur d'un lieu pose dans une ouverture, ce qui
// le casse, et ce qu'il laisse. Il bloque tant qu'il tient, et c'est le joueur
// qui le force à la touche — jamais son arme.

package game

import (
	"fmt"
	"slices"

	"github.com/sprimault/cohue/internal/manifest"
)

// Les deux sens dans lesquels un obstacle se pose.
const (
	// AxisU est le sens par défaut, celui où le manifeste dessine l'objet.
	AxisU = "u"
	// AxisV le fait courir le long de l'autre axe, par le dessin que son entrée
	// nomme sous `pivote`.
	AxisV = "v"
)

// porteeFrappe est la distance au centre de l'obstacle en deçà de laquelle le
// joueur le frappe.
//
// **Une tuile, et c'est la géométrie qui la fixe, pas un réglage.** L'obstacle
// bloque sa case : le joueur, qui ne s'y enfonce pas, s'en approche au mieux à
// une demi-tuile de face et à sept dixièmes par un coin. Une portée posée sur
// l'une de ces bornes serait fausse en permanence — elle paraîtrait marcher et
// ne toucherait jamais. Une tuile les dépasse toutes deux, et s'arrête avant le
// centre de la case voisine : il faut venir contre, ce que la conception
// demande.
const porteeFrappe = One

// BreakableSpec est le semis d'obstacles fragiles tel qu'un lieu l'écrit.
type BreakableSpec []BreakablePlacementSpec

// BreakablePlacementSpec pose un obstacle fragile à une case donnée.
type BreakablePlacementSpec struct {
	manifest.Commentable
	// Object est son nom au catalogue des objets.
	Object string `json:"objet"`
	// At est la case qu'il ferme, en coordonnées de lieu.
	At *[2]int `json:"position"`
	// Axis est le sens dans lequel il court, `u` quand il est absent.
	Axis string `json:"orientation,omitempty"`
}

// BreakablePlacement est un obstacle compilé : sa sorte, sa position, son sens
// et le sol qu'il recouvre.
type BreakablePlacement struct {
	// Kind est son rang dans la table des sortes.
	Kind int
	// X et Y sont sa position, au centre de la case écrite.
	X, Y Fixed
	// Floor est le coût de la case avant qu'il s'y dresse, relevé ici pour la
	// raison qui le fait relever pour une caisse : au montage, la case porte
	// déjà son blocage.
	Floor Cost
	// Across dit qu'il court le long de v.
	Across bool
}

// Breakable est un obstacle fragile debout.
type Breakable struct {
	// X et Y sont sa position dans le monde, au centre de sa case.
	X, Y Fixed
	// Kind est son rang dans la table des sortes.
	Kind int
	// Hits est ce qu'il encaisse encore, en touches. À zéro il cède, et la valeur
	// est l'état — comme la résistance d'une créature.
	//
	// **Rien ne le remet à neuf quand le joueur s'écarte**, à l'inverse de
	// l'appui d'une caisse. Là-bas le délai existe pour qu'on ne casse pas en
	// passant ; ici chaque touche coûte déjà un arrêt sous la horde, et revenir
	// finir un rideau entamé est une décision que le jeu doit payer, pas effacer.
	Hits int
	// Floor est le coût qu'il rend à sa case en cédant.
	Floor Cost
	// Across dit qu'il court le long de v. La simulation ne le lit pas : une
	// case bloquée l'est dans les deux sens.
	Across bool
	// Flash est ce qui reste de l'éclair de la dernière frappe, en ticks.
	//
	// C'est le seul retour d'une touche qui ne casse pas encore, et sans lui
	// cinq secondes contre un rideau de fer ressembleraient à une touche qui ne
	// fait rien.
	Flash Tick
}

// CompileBreakables résout un semis d'obstacles contre le catalogue et la carte
// cuite.
//
// Elle rend tout ce qui l'empêche de valoir, comme les caisses. Les refus qui
// leur sont propres :
//
//   - **un nom qui n'est pas un obstacle fragile.** La liste des noms admis
//     accompagne le refus, puisque c'est elle que l'auteur cherchait ;
//   - **une case qui ne se franchit pas.** Un obstacle ferme une ouverture, et
//     posé dans un mur il ne fermerait rien ;
//   - **une case déjà prise**, par un autre obstacle, une caisse ou un figurant.
//     Un obstacle bloque sa case, si bien que ce qui s'y trouverait serait
//     emmuré : une caisse qu'on n'atteint plus, un figurant qui ne sort plus ;
//   - **un sens inconnu**, les deux admis étant nommés.
func CompileBreakables(brut BreakableSpec, sortes []BreakableKind, carte *CostGrid,
	caisses []CratePlacement, ambiance []AmbientPlacement) ([]BreakablePlacement, []string) {
	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	// Ce que les autres semis occupent, pour dire lequel emmurerait quoi.
	prises := make(map[[2]int]string, len(brut)+len(caisses)+len(ambiance))
	for i, c := range caisses {
		prises[[2]int{c.X.Floor(), c.Y.Floor()}] = fmt.Sprintf("caisses[%d]", i)
	}
	for i, a := range ambiance {
		prises[[2]int{a.X.Floor(), a.Y.Floor()}] = fmt.Sprintf("ambiance[%d]", i)
	}

	pose := make([]BreakablePlacement, 0, len(brut))
	for i, o := range brut {
		ou := fmt.Sprintf("destructibles[%d]", i)
		sorte := slices.IndexFunc(sortes, func(s BreakableKind) bool { return s.Key == o.Object })
		if sorte < 0 {
			dire("%s.objet : « %s » n'est pas un obstacle fragile ; connus : %s",
				ou, o.Object, nomsDe(sortes))
		}
		if o.Axis != "" && o.Axis != AxisU && o.Axis != AxisV {
			dire("%s.orientation : « %s », attendu %s ou %s", ou, o.Axis, AxisU, AxisV)
		}
		if o.At == nil {
			dire("%s.position : absente, un obstacle se place", ou)
			continue
		}

		u, v := o.At[0], o.At[1]
		switch occupant, occupee := prises[[2]int{u, v}]; {
		case !carte.InBounds(u, v):
			dire("%s.position : (%d, %d) hors du lieu, qui fait %d sur %d",
				ou, u, v, carte.Width(), carte.Height())
		case !carte.Passable(u, v):
			dire("%s.position : (%d, %d) est dans un mur ; un obstacle ferme une ouverture",
				ou, u, v)
		case occupee:
			dire("%s.position : (%d, %d) porte déjà %s, qu'il emmurerait", ou, u, v, occupant)
		case sorte >= 0:
			prises[[2]int{u, v}] = ou
			pose = append(pose, BreakablePlacement{
				Kind:   sorte,
				X:      FromInt(u) + One/2,
				Y:      FromInt(v) + One/2,
				Floor:  carte.At(u, v),
				Across: o.Axis == AxisV,
			})
		}
	}
	return pose, manques
}

// nomsDe rend les noms d'une table de sortes, pour un message.
func nomsDe(sortes []BreakableKind) string {
	noms := make([]string, len(sortes))
	for i, s := range sortes {
		noms[i] = s.Key
	}
	return fmt.Sprint(noms)
}

// StampBreakables bloque dans la grille les cases que les obstacles ferment.
//
// **Avant que le monde soit bâti**, comme les caisses, et pour la même raison
// d'ordre. Rien ne s'y oppose ici : un blocage n'entre pas dans le compte des
// seaux, qui ne voit que ce qui se franchit.
func StampBreakables(grille *CostGrid, poses []BreakablePlacement) {
	for _, o := range poses {
		grille.Set(o.X.Floor(), o.Y.Floor(), Blocked)
	}
}

// Erect dresse les obstacles du lieu, au montage et à chaque relance.
//
// **La table des sortes entre ici et non au montage du monde**, parce qu'elle ne
// sert qu'à ce qu'on dresse : une partie sans obstacle n'a rien à en faire, et
// la réclamer à chaque appelant de `NewWorld` la ferait porter à des dizaines de
// cas qui n'en posent aucun.
func (w *World) Erect(sortes []BreakableKind, poses []BreakablePlacement) {
	w.sortesObstacles = sortes
	for _, o := range poses {
		w.obstacles.Spawn(Breakable{
			X: o.X, Y: o.Y,
			Kind:   o.Kind,
			Hits:   sortes[o.Kind].Hits,
			Floor:  o.Floor,
			Across: o.Across,
		})
	}
}

// Interact dit que le joueur tient la touche d'interaction pendant ce tick.
//
// **Une entrée, comme la direction voulue**, que le tick consomme : elle ne vaut
// que pour le pas qui suit, si bien qu'une touche relâchée ne frappe plus sans
// qu'il faille le dire.
func (w *World) Interact() { w.frapper = true }

// Breakables rend le bassin des obstacles debout, que le rendu parcourt.
func (w *World) Breakables() *Pool[Breakable] { return w.obstacles }

// BreakableKey rend le nom au catalogue d'une sorte d'obstacle.
func (w *World) BreakableKey(sorte int) string { return w.sortesObstacles[sorte].Key }

// BreakableHits rend ce qu'une sorte d'obstacle encaisse intacte, en touches :
// le décompte seul ne dit pas d'où il part.
func (w *World) BreakableHits(sorte int) int { return w.sortesObstacles[sorte].Hits }

// forcer frappe l'obstacle que le joueur tient, et vide celui qui cède.
//
// **Une touche à la cadence de l'arme de base au premier niveau**, celle de la
// table et non celle que les paliers ont montée : c'est l'unité dans laquelle
// les touches sont écrites, et un rideau de fer qui cèderait plus vite à la
// dixième minute mesurerait la build au lieu du prix de l'ouverture.
//
// **Le décompte descend que la touche soit tenue ou non.** Remis à zéro au
// relâchement, il rendrait le martèlement plus rapide que l'appui tenu — la
// corvée que le geste tenu existe pour épargner.
//
// **L'arme ne le frappe jamais**, et c'est la règle de la caisse : rangé parmi
// les cibles, il détournerait la visée automatique, qui prend la plus proche
// sans que le joueur choisisse.
func (w *World) forcer() {
	if w.frappe > 0 {
		w.frappe--
	}
	for i := range w.obstacles.Active() {
		if o := w.obstacles.At(i); o.Flash > 0 {
			o.Flash--
		}
	}
	if !w.frapper || w.frappe > 0 || !w.Alive() {
		return
	}

	// Le plus proche à portée, et à égalité le premier du bassin : deux
	// obstacles qui se touchent — une vitrine dans un angle — ne se frappent pas
	// ensemble, sans quoi tenir contre deux diviserait le prix de chacun.
	cible, meilleur := -1, int64(porteeFrappe)*int64(porteeFrappe)
	for i := range w.obstacles.Active() {
		o := w.obstacles.At(i)
		ecart := Vec{X: w.playerX - o.X, Y: w.playerY - o.Y}
		if d := ecart.carres(); d < meilleur {
			cible, meilleur = i, d
		}
	}
	if cible < 0 {
		return
	}

	w.frappe = w.armes.Base.Cooldown
	o := w.obstacles.At(cible)
	o.Hits--
	o.Flash = eclairImpact
	if o.Hits > 0 {
		return
	}

	// La case rendue au sol avant tout le reste, pour la raison qui vaut pour la
	// caisse : un blocage laissé derrière arrêterait la horde sur un point
	// qu'aucun pixel ne montre plus.
	w.grille.Set(o.X.Floor(), o.Y.Floor(), o.Floor)
	w.ceder(o.X, o.Y, w.sortesObstacles[o.Kind].Key, o.Across)
	w.obstacles.RemoveAt(cible)
}
