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

// tirerUneArme décide si une caisse portera une arme lourde, et laquelle.
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
// **Elle tire à l'apparition de la caisse**, et non au moment où celle-ci cède :
// une caisse annonce ce qu'elle porte, et on ne montre pas ce qui n'est pas
// décidé.
// **Le rang est rendu tel quel, le second résultat portant l'absence.** Il était
// décalé de un tant que la caisse le rangeait seul ; c'est la sorte qui porte
// cette question maintenant, et garder le décalage en aurait fait deux réponses
// à la même.
func (w *World) tirerUneArme() (int, bool) {
	chance := w.progression.HeavyOdds
	if chance <= 0 || len(w.armes.Heavy) == 0 {
		return 0, false
	}
	if w.hasard.Loot.IntN(chance) != 0 {
		return 0, false
	}
	return w.armes.Heavy[w.hasard.Loot.Pick(len(w.armes.Heavy))], true
}

// lacherUneArme pose au sol l'arme d'un rang donné.
//
// Le bassin plein perd l'arme plutôt que de la différer. C'est le cas d'un joueur
// qui a laissé traîner ses trouvailles, et une arme qui apparaîtrait plus tard,
// ailleurs, ne se relierait à aucune caisse.
//
// **Elle perd donc une arme que la caisse avait annoncée**, ce qui est assumé :
// l'annonce vaut pour ce que le joueur décide d'aller chercher, et deux
// trouvailles laissées au sol sont déjà son choix.
func (w *World) lacherUneArme(rang int, x, y Fixed) {
	w.armesAuSol.Spawn(Drop{X: x, Y: y, Weapon: rang})
}

// ramasserUneArme prend l'arme sous les pieds du joueur, s'il a une place.
//
// **Marcher dessus suffit quand un emplacement l'accueille**, ce que la
// conception veut : aucun menu, aucune touche. Les deux tenus par d'autres
// armes, celle-ci reste au sol et c'est la touche d'un emplacement qui l'échange
// — voir `TakeDrop`.
func (w *World) ramasserUneArme() {
	if !w.Alive() {
		return
	}
	i, sur := w.armeSousLesPieds()
	if !sur {
		return
	}
	if place := w.emplacementPour(w.armesAuSol.At(i).Weapon); place >= 0 {
		w.prendre(i, place)
	}
}

// emplacementPour rend la place qui accueille une arme d'un rang donné : celle
// qui en tient déjà et où le stock tient encore, sinon la première libre, sinon
// aucune.
//
// **Le même type avant le vide**, et c'est ce qui rend une case unique par arme.
// Une grenade ramassée alors qu'une place est libre irait sinon la remplir, et le
// joueur se retrouverait avec deux cases de grenades dont il ne peut vider que
// l'une — ce qu'une partie a signalé, et qui n'avait de sens que du jour où le
// catalogue portera plusieurs lourdes.
//
// **Une case pleine ne renvoie pas au second emplacement.** Il en ferait la
// seconde case du même type que la règle vient de fermer ; l'arme reste au sol,
// et c'est ce que le plafond veut dire — on revient la chercher.
func (w *World) emplacementPour(rang int) int {
	arme := &w.armes.All[rang]
	for i := range w.lourdes {
		if w.lourdes[i].Charges == 0 || w.lourdes[i].rang != rang {
			continue
		}
		if w.lourdes[i].Charges+arme.Charges > arme.Stock {
			return -1
		}
		return i
	}
	for i := range w.lourdes {
		if w.lourdes[i].Charges == 0 {
			return i
		}
	}
	return -1
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
//
// **Une arme du même type s'ajoute au lieu de remplacer**, ce qui fait du
// contenu d'un emplacement un stock de tirs plutôt qu'un exemplaire. C'est ce
// que la conception demande depuis qu'une partie a montré le cas qu'elle n'avait
// pas prévu : le catalogue ne porte qu'une lourde, si bien que les deux
// emplacements ne pouvaient tenir que des doublons, et la troisième grenade
// ramassée en faisait perdre une.
func (w *World) prendre(sol, place int) {
	d := w.armesAuSol.At(sol)
	arme := &w.armes.All[d.Weapon]
	tenue := &w.lourdes[place]

	if tenue.Charges > 0 && tenue.rang == d.Weapon {
		tenue.Charges += arme.Charges
	} else {
		*tenue = Heavy{Weapon: *arme, Charges: arme.Charges, rang: d.Weapon}
	}
	w.armesAuSol.RemoveAt(sol)
}
