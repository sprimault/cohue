// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La fiole : ce qui la fait tomber d'une caisse, ce qu'on en tient, et ce que
// boire rend. La seule chose du jeu qui redonne de la vie hors d'une montée de
// niveau.

package game

// Vial est une fiole posée au sol, en attente d'être prise.
//
// Elle ne porte que sa place, comme l'aimant : ce qu'elle rend appartient à la
// partie, et un objet qui porterait son effet en ferait une seconde description
// le jour où une deuxième sorte de consommable arriverait.
//
// **Elle ne s'efface pas avec le temps**, à la différence d'une gemme et pour la
// raison qui garde une arme lourde au sol : l'effacement d'une gemme donne sa
// contre-force à l'aimant, quand une fiole est une trouvaille rare que le joueur
// laisse là où il ne peut pas aller tout de suite.
type Vial struct {
	// X et Y sont le point où elle attend.
	X, Y Fixed
}

// Vials rend le bassin des fioles au sol.
func (w *World) Vials() *Pool[Vial] { return w.fiolesAuSol }

// VialStock rend le nombre de fioles que le joueur tient.
func (w *World) VialStock() int { return w.fioles }

// VialKey rend la clé de catalogue de la fiole.
//
// Le bandeau en a besoin pour poser son icône, et il n'a pas à connaître le
// manifeste de progression où le nom est écrit — c'est la même frontière que
// pour l'arme d'un emplacement, dont il reçoit la clé sans savoir d'où elle
// vient.
func (w *World) VialKey() string { return w.progression.VialObject }

// tirerUneFiole décide si une caisse en portera une.
//
// **Elle n'est consultée que lorsque aucune arme n'est sortie**, ce qui fait de
// sa chance une chance sur les caisses restantes et non sur toutes. C'est
// `tirerLeButin` qui tient cet ordre, et il y est écrit.
func (w *World) tirerUneFiole() bool {
	chance := w.progression.VialOdds
	if chance <= 0 {
		return false
	}
	return w.hasard.Loot.IntN(chance) == 0
}

// lacherUneFiole pose une fiole au sol.
//
// Le bassin plein la perd, comme il perd une arme, et pour la même raison : une
// fiole qui apparaîtrait plus tard et ailleurs ne se relierait à aucune caisse.
func (w *World) lacherUneFiole(x, y Fixed) {
	w.fiolesAuSol.Spawn(Vial{X: x, Y: y})
}

// SpawnVial pose une fiole au sol, sans passer par une caisse.
//
// Elle sert à monter un état — une planche de relecture, un cas qui éprouve ce
// qui suit la chute sans éprouver le tirage —, jamais à jouer : la boucle ne
// l'appelle pas. C'est le pendant de `SpawnDrop`, et le second résultat dit le
// bassin plein pour la même raison qu'elle : qui monte une scène chercherait un
// défaut d'affichage là où il n'y a plus de place.
func (w *World) SpawnVial(x, y Fixed) bool {
	_, ok := w.fiolesAuSol.Spawn(Vial{X: x, Y: y})
	return ok
}

// ramasserUneFiole prend la fiole sous les pieds du joueur, s'il a de la place.
//
// **Marcher dessus suffit**, ce que la conception veut d'un consommable : aucun
// menu, aucune touche pour prendre. La touche ne sert qu'à boire.
//
// **Le stock plein la laisse au sol**, comme le plafond d'une arme lourde laisse
// celle qui ne tient pas entière : on revient la chercher une fois qu'on a bu.
// En rogner le surplus ferait perdre au joueur ce qu'il vient de trouver, et la
// ramasser au-delà du plafond ferait de la fiole un second socle — la vie
// cesserait d'être la ressource rare que la conception veut qu'elle soit.
func (w *World) ramasserUneFiole() {
	if !w.Alive() || w.fioles >= w.fiole.Stock {
		return
	}
	portee := w.progression.PickupRange
	for i := range w.fiolesAuSol.Active() {
		f := w.fiolesAuSol.At(i)
		if (Vec{X: f.X - w.playerX, Y: f.Y - w.playerY}).carres() > int64(portee)*int64(portee) {
			continue
		}
		w.fioles++
		w.fiolesAuSol.RemoveAt(i)
		// Une seule par tick, la place libérée n'étant pas réexaminée : c'est la
		// règle des passes qui avancent, et deux fioles posées l'une sur l'autre
		// se prennent donc à un tick d'écart, ce qui ne se voit pas.
		return
	}
}

// Drink boit une fiole et rend ce qu'elle porte de vie.
//
// **Sans effet quand le stock est vide ou le joueur mort**, comme `Trigger` et
// `Attract` : l'appelant est un clavier, et une touche pressée à vide ne doit ni
// consommer ni avertir.
//
// **Boire à pleine vie est permis, et gaspille.** C'est une décision de la
// conception et non un oubli : le surplus perdu est ce qui donne au joueur une
// raison de ne pas boire tout de suite, donc ce qui fait de la fiole une
// décision plutôt qu'un cadeau. Un refus à pleine vie retirerait cette
// décision, et il a exactement la forme d'une garde qu'on ajoute en croyant
// corriger un défaut — c'est le pendant de l'aimant déclenché à vide, que le
// même chapitre assume.
//
// Le seuil d'alerte du profil vaut ce que rend un soin, si bien qu'en dessous
// boire ne gaspille rien : l'alerte annonce la décision qu'elle doit
// déclencher, et c'est ce qui interdit de régler la rareté du soin en baissant
// ce que rend une fiole — `Progression.VialOdds` porte l'autre bout de cette
// règle.
func (w *World) Drink() {
	if !w.Alive() || w.fioles <= 0 {
		return
	}
	w.fioles--
	w.vie = min(w.vie+w.fiole.Heal, w.profils.Player.Health)
}
