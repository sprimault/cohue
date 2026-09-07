// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'arme lourde que le joueur tient : ce qu'il lui reste, ce qu'un déclenchement
// dépense, et la déflagration qu'elle pose.

package game

// Heavy est l'arme lourde tenue, et ce qu'il lui reste de charges.
//
// **Zéro charge veut dire qu'aucune arme n'est tenue**, et ce n'est pas une
// valeur d'absence qui coïncide avec une valeur valide : la conception veut
// qu'une lourde soit **jetée à vide**, donc une arme à zéro n'existe jamais. Le
// zéro d'un champ oublié dit alors exactement ce qu'il doit dire.
type Heavy struct {
	// Weapon est la copie de l'arme, comme `World.arme` l'est du socle.
	Weapon Weapon
	// Charges est ce qui reste à dépenser. À zéro, l'arme est partie.
	Charges int
	// rang est sa place dans la table, que la déflagration désigne.
	rang int
}

// HeldHeavy rend l'arme lourde tenue et ce qu'il lui reste.
//
// Par valeur : l'interface a besoin d'un nom et d'un compte, pas d'une prise sur
// ce que la partie modifie.
func (w *World) HeldHeavy() (Weapon, int) { return w.lourde.Weapon, w.lourde.Charges }

// GiveHeavy pose une arme lourde entre les mains du joueur, par sa clé.
//
// **C'est la sonde du lot qui l'introduit, pas la façon dont on en obtient une.**
// La conception veut qu'une lourde se trouve dans une caisse et se ramasse au
// sol ; tant que rien ne la pose dans le lieu, un déclenchement n'a rien à
// éprouver. Elle disparaîtra avec le ramassage, comme la porte donnée d'office a
// disparu quand un objectif l'a ouverte.
//
// Le second résultat est faux quand la clé ne désigne pas une lourde : l'appelant
// se trompe de table, et le silence lui ferait chercher un défaut de
// déclenchement.
func (w *World) GiveHeavy(cle string) bool {
	for i := range w.armes.All {
		arme := &w.armes.All[i]
		if arme.Key != cle || arme.Role != roleHeavy {
			continue
		}
		w.lourde = Heavy{Weapon: *arme, Charges: arme.Charges, rang: i}
		return true
	}
	return false
}

// Trigger dépense une charge et pose la déflagration de l'arme tenue.
//
// **Sans effet quand rien n'est tenu**, comme `Attract` sans charge d'aimant :
// l'appelant est un clavier, et une touche pressée à vide ne doit ni consommer ni
// avertir.
//
// **La grenade tombe sur la cible la plus proche à portée, et rien ne part sans
// cible.** Le joueur ne dirige pas son lancer — il ne contrôle que son
// déplacement —, donc le point d'arrivée se choisit comme le tir automatique
// choisit sa cible. Sans cible, la charge n'est pas dépensée, pour la raison qui
// fait que la cadence ne se consomme pas à vide : une arme à trois charges dont
// une part dans le vide se lirait comme un défaut.
//
// La cadence de l'arme lourde n'est pas consultée ici. Elle vaut pour ce qui tire
// tout seul, et une lourde ne tire que sur une touche — c'est le joueur qui
// espace ses déclenchements.
func (w *World) Trigger() {
	if w.lourde.Charges <= 0 {
		return
	}
	cible, trouvee := w.plusProcheDe(w.playerX, w.playerY, w.lourde.Weapon.Range, Handle{})
	if !trouvee {
		return
	}

	e := w.ennemis.At(cible)
	if _, ok := w.souffles.Spawn(Blast{
		X: e.X, Y: e.Y,
		Source: BlastWeapon,
		Index:  w.lourde.rang,
		Fuse:   w.lourde.Weapon.Fuse,
	}); !ok {
		// Bassin plein : la charge n'est pas dépensée. Une déflagration perdue
		// coûterait au joueur une des trois choses qu'il possède, pour une raison
		// qu'aucun écran ne peut lui montrer.
		return
	}

	w.lourde.Charges--
	if w.lourde.Charges == 0 {
		// **Jetée à vide**, ce que la conception exige : l'arme quitte
		// l'emplacement au lieu d'y rester inerte, et le joueur voit sa place se
		// libérer plutôt qu'un compteur à zéro.
		w.lourde = Heavy{}
	}
}

// emporter applique la déflagration d'une arme à la horde autour d'un point.
//
// **Elle n'emporte que la horde, et jamais le joueur.** L'inverse créerait une
// décision de placement — reculer avant de lancer — que le joueur ne peut pas
// prendre : il ne dirige pas son lancer, et le punir d'un geste qu'il n'oriente
// pas serait une sanction sans recours. C'est ce qui la sépare de la Baudruche,
// qu'on voit venir et qu'on esquive. La question se rouvrira le jour où une arme
// lourde deviendra dirigeable, et ce sera le bon moment.
//
// La transition de mort passe par le même chemin qu'un tir : le butin, et
// l'amorce d'une Baudruche qui explose à son tour.
func (w *World) emporter(arme *Weapon, x, y Fixed) {
	rayon := int64(arme.BurstRadius) * int64(arme.BurstRadius)
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		if e.Hits <= 0 {
			continue
		}
		if (Vec{X: e.X - x, Y: e.Y - y}).carres() > rayon {
			continue
		}

		e.Hits -= arme.BurstHits
		e.Flash = eclairImpact
		if e.Hits <= 0 {
			w.lacher(e)
			w.amorcer(e)
		}
	}
}
