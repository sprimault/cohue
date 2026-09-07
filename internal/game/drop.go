// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'arme lourde tombée d'une caisse : ce qu'elle attend au sol, comment on la
// prend en marchant dessus, et l'échange quand les deux emplacements sont pleins.

package game

// Drop est une arme lourde posée au sol, en attente d'être prise.
//
// **Elle porte le rang de son arme et non ses valeurs**, comme toute entité : ce
// qu'on ramasse est une arme de la table, et régler une grenade ne doit pas
// dépendre de celles déjà tombées.
//
// Elle ne s'efface pas avec le temps, à la différence d'une gemme. L'effacement
// d'une gemme existe pour donner sa contre-force à l'aimant et pour que ramasser
// coûte un trajet ; une arme lourde est une trouvaille rare, et la faire
// disparaître punirait un joueur occupé à survivre au moment où elle tombe.
type Drop struct {
	// X et Y sont le point où elle attend.
	X, Y Fixed
	// Weapon est son rang dans la table des armes.
	Weapon int
}

// Drops rend le bassin des armes au sol, en lecture.
func (w *World) Drops() *Pool[Drop] { return w.armesAuSol }

// DropWeapon rend l'arme que désigne une chose tombée.
//
// Le rendu en a besoin pour poser la bonne image, et il n'a pas à connaître la
// table : il lit une arme, comme il lit un profil pour une créature.
func (w *World) DropWeapon(d *Drop) *Weapon { return &w.armes.All[d.Weapon] }

// SpawnDrop pose une arme lourde nommée au sol, sans passer par le hasard.
//
// **Elle sert à monter un état, jamais à jouer** : une planche de relecture doit
// pouvoir montrer une arme au sol sans dépendre d'un tirage, et un test éprouver
// ce qui suit la chute sans éprouver la chute. Ce qui la produit en partie est
// `lacherUneArme`, depuis une caisse.
//
// Le second résultat est faux quand la clé ne désigne pas une lourde ou que le
// bassin est plein : l'appelant monte une scène, et le silence lui ferait
// chercher un défaut d'affichage.
func (w *World) SpawnDrop(cle string, x, y Fixed) bool {
	for _, rang := range w.armes.Heavy {
		if w.armes.All[rang].Key != cle {
			continue
		}
		_, ok := w.armesAuSol.Spawn(Drop{X: x, Y: y, Weapon: rang})
		return ok
	}
	return false
}

// lacherUneArme pose une arme lourde au sol, tirée parmi celles de la table.
//
// **C'est le premier lecteur du flux `butin`**, qui attendait le sien depuis sa
// déclaration : sa godoc annonçait « au futur, et rien ne l'alimente encore », et
// seul le témoin de l'empreinte le gardait numéroté. Le même moment que les
// figurants pour le flux cosmétique.
//
// **Une chance sur n plutôt qu'une probabilité en flottant** : la simulation ne
// tire que des entiers, et un flottant y rentrerait par la porte que la virgule
// fixe a fermée.
//
// Le bassin plein perd l'arme plutôt que de la différer. C'est le cas d'un joueur
// qui a laissé traîner ses trouvailles, et une arme qui apparaîtrait plus tard,
// ailleurs, ne se relierait à aucune caisse.
func (w *World) lacherUneArme(x, y Fixed) {
	chance := w.progression.HeavyOdds
	if chance <= 0 || len(w.armes.Heavy) == 0 {
		return
	}
	if w.hasard.Loot.IntN(chance) != 0 {
		return
	}

	rang := w.armes.Heavy[w.hasard.Loot.Pick(len(w.armes.Heavy))]
	w.armesAuSol.Spawn(Drop{X: x, Y: y, Weapon: rang})
}

// ramasserUneArme prend l'arme sous les pieds du joueur, s'il a une place.
//
// **Marcher dessus suffit quand un emplacement est libre**, ce que la conception
// veut : aucun menu, aucune touche. Les deux pleins, l'arme reste au sol et c'est
// la touche d'un emplacement qui l'échange — voir `TakeDrop`.
func (w *World) ramasserUneArme() {
	if !w.Alive() {
		return
	}
	place := -1
	for i := range w.lourdes {
		if w.lourdes[i].Charges == 0 {
			place = i
			break
		}
	}
	if place < 0 {
		return
	}

	if i, sur := w.armeSousLesPieds(); sur {
		w.prendre(i, place)
	}
}

// TakeDrop met dans un emplacement l'arme sous les pieds du joueur.
//
// **Une seule règle pour les deux cas** : la touche d'un emplacement, pressée sur
// une arme au sol, y met cette arme — que l'emplacement soit vide ou plein. Le
// ramassage au contact fait la même chose sur la première place libre, si bien
// que les deux mènent au même état et que la touche n'est indispensable que
// lorsque les deux emplacements sont pleins.
//
// **Ce que l'emplacement tenait est perdu, jamais reposé au sol.** Le rendre
// ferait un échange sans fin — on repasserait dessus, on la reprendrait — et le
// joueur ne saurait plus laquelle des deux il tient.
//
// Sans effet quand rien n'est sous les pieds : l'appelant est un clavier, et la
// touche sert aussi à déclencher.
func (w *World) TakeDrop(place int) bool {
	if place < 0 || place >= len(w.lourdes) || !w.Alive() {
		return false
	}
	i, sur := w.armeSousLesPieds()
	if !sur {
		return false
	}
	w.prendre(i, place)
	return true
}

// armeSousLesPieds rend la place d'une arme au sol sous le joueur.
//
// La portée est celle du ramassage d'une gemme : ce qu'on prend en marchant
// dessus se prend à la même distance, et deux portées pour un même geste
// finiraient par se contredire.
func (w *World) armeSousLesPieds() (int, bool) {
	portee := w.progression.PickupRange
	for i := range w.armesAuSol.Active() {
		d := w.armesAuSol.At(i)
		if (Vec{X: d.X - w.playerX, Y: d.Y - w.playerY}).carres() <= int64(portee)*int64(portee) {
			return i, true
		}
	}
	return 0, false
}

// prendre pose l'arme d'une place du sol dans un emplacement.
func (w *World) prendre(sol, place int) {
	d := w.armesAuSol.At(sol)
	arme := &w.armes.All[d.Weapon]
	w.lourdes[place] = Heavy{Weapon: *arme, Charges: arme.Charges, rang: d.Weapon}
	w.armesAuSol.RemoveAt(sol)
}
