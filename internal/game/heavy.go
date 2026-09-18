// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'arme lourde que le joueur tient : ce qu'il lui reste, ce qu'un déclenchement
// dépense, et la déflagration qu'elle pose.

package game

import "slices"

// Heavy est l'arme lourde tenue, et ce qu'il lui reste de charges.
//
// **Zéro charge veut dire qu'aucune arme n'est tenue**, et ce n'est pas une
// valeur d'absence qui coïncide avec une valeur valide : la conception veut
// qu'une lourde soit **jetée à vide**, donc une arme à zéro n'existe jamais. Le
// zéro d'un champ oublié dit alors exactement ce qu'il doit dire.
//
// **Un rang et aucune copie de l'arme.** Elle en portait une, prise au
// ramassage ; les paliers se lisant maintenant à l'usage, cette copie serait la
// seule version de l'arme qui ne les connaîtrait pas, et rien n'aurait dit
// laquelle des deux fait foi. Ce qui la remplace se dérive à chaque lecture et
// ne peut donc pas se périmer.
type Heavy struct {
	// Charges est ce qui reste à dépenser. À zéro, l'arme est partie.
	Charges int
	// rang est sa place dans la table, d'où tout le reste se dérive.
	rang int
}

// Slots est le nombre d'emplacements d'armes lourdes.
//
// **Deux, et ce n'est pas un réglage.** La conception en fait une règle : le
// joueur a une décision — laquelle garder — et non une gestion. Un troisième
// emplacement retirerait le choix qu'une trouvaille pose, et un champ de
// manifeste inviterait à le bouger.
const Slots = 2

// HeldHeavy rend l'arme d'un emplacement et ce qu'il lui reste.
//
// Par valeur : l'interface a besoin d'un nom et d'un compte, pas d'une prise sur
// ce que la partie modifie. Une arme vide rend une `Weapon` nulle, ce que la
// valeur zéro de `Heavy` dit déjà — une lourde à zéro charge n'existe pas.
func (w *World) HeldHeavy(place int) (Weapon, int) {
	if place < 0 || place >= len(w.lourdes) || w.lourdes[place].Charges <= 0 {
		return Weapon{}, 0
	}
	return w.lourdeDe(w.lourdes[place].rang), w.lourdes[place].Charges
}

// Trigger dépense une charge et pose la déflagration de l'arme tenue.
//
// **Sans effet quand rien n'est tenu**, comme `Attract` sans charge d'aimant :
// l'appelant est un clavier, et une touche pressée à vide ne doit ni consommer ni
// avertir.
//
// **Ce qui vise le fait dans sa branche, et non ici.** Deux des trois effets
// exigent une cible — une grenade et une gerbe parties dans le vide se liraient
// comme un défaut, pour la raison qui fait que la cadence ne se consomme pas à
// vide —, mais une tourelle se pose pour tenir un passage, souvent avant que la
// horde n'arrive. La garde vivait ici tant qu'aucun effet n'en voulait pas.
//
// **La cadence d'une lourde n'est consultée que par ce qui tire sans le
// joueur.** Elle ne l'était par personne tant que tout partait d'une touche —
// c'est le joueur qui espaçait ses déclenchements ; la tourelle est ce qui a
// rendu cette phrase fausse, et elle dit maintenant à quelle condition elle vaut.
func (w *World) Trigger(place int) {
	if place < 0 || place >= len(w.lourdes) || w.lourdes[place].Charges <= 0 {
		return
	}
	tenue := &w.lourdes[place]

	// **L'effet décide, et la charge se dépense après lui.** Chaque branche peut
	// renoncer — un bassin plein, une cible absente —, et une charge dépensée
	// pour un effet qui n'est pas parti coûterait au joueur une des trois choses
	// qu'il possède, pour une raison qu'aucun écran ne peut lui montrer.
	if !w.declencher(tenue) {
		return
	}

	tenue.Charges--
	if tenue.Charges == 0 {
		// **Jetée à vide**, ce que la conception exige : l'arme quitte
		// l'emplacement au lieu d'y rester inerte, et le joueur voit sa place se
		// libérer plutôt qu'un compteur à zéro.
		*tenue = Heavy{}
	}
}

// declencher produit ce que l'arme tenue déclare, et dit si c'est parti.
//
// **Un aiguillage sur une donnée, jamais une branche par arme.** Ce que le
// moteur connaît est une liste close d'effets ; deux armes qui déclareraient le
// même partageraient ce chemin sans qu'une ligne s'ajoute, ce qui est la règle
// des données qui ne sont pas du code.
//
// **Chaque effet cherche ce dont il a besoin**, et la cible n'en fait pas
// partout partie : deux la veulent, et elle n'y veut pas dire la même chose — la
// déflagration s'y pose, la salve s'y dirige. La tourelle, elle, se pose sous le
// joueur et vise plus tard, depuis là où elle se tient.
func (w *World) declencher(tenue *Heavy) bool {
	arme := w.lourdeDe(tenue.rang)
	switch arme.Effect {
	case effetDeflagration:
		cible, trouvee := w.cibleDe(&arme)
		if !trouvee {
			return false
		}
		_, ok := w.souffles.Spawn(Blast{
			X: cible.X, Y: cible.Y,
			Source: BlastWeapon,
			Index:  tenue.rang,
			Fuse:   arme.Fuse,
		})
		return ok
	case effetSalve:
		cible, trouvee := w.cibleDe(&arme)
		if !trouvee {
			return false
		}
		// **Vers la cible et non vers son interception**, à la différence du tir
		// de base : un fusil à pompe étale des plombs sur une largeur, et viser où
		// la cible sera n'aurait de sens que pour un projectile unique. Ce qui
		// touche est le front, pas l'anticipation.
		vers := Vec{X: cible.X - w.playerX, Y: cible.Y - w.playerY}.Direction(0)
		return w.salve(&arme, vers) > 0
	case effetTourelle:
		return w.poserUneTourelle(&arme, tenue.rang)
	case effetFlammes:
		return w.poserDesFlammes(&arme, tenue.rang)
	}
	return false
}

// lourdeDe rend l'arme d'un rang, telle que les paliers pris la font.
//
// **Dérivée à chaque lecture plutôt que tenue à jour.** Une copie prise au
// ramassage aurait à être rafraîchie à chaque montée de niveau, pour chacun des
// deux emplacements et pour tout ce qu'une charge a posé au sol ; celle-ci ne
// peut pas se périmer, puisqu'il n'y a rien à tenir d'accord. C'est ce que le
// chapitre 9 veut par ailleurs : une arme est sensible à ce qu'on a pris
// **depuis** qu'on la tient, sans quoi une trouvaille de la douzième minute
// serait plus faible que le tir de base.
//
// **Seuls les axes que l'arme déclare l'atteignent.** Cadence et portée valent
// pour toutes, le nombre de projectiles n'a aucun sens pour une tourelle qui
// tire seule : c'est la donnée qui décide, arme par arme, et ce champ était
// déclaré et contrôlé sans que rien ne le lise.
//
// **Les fusions n'y entrent pas, et c'est une absence et non un oubli.** Une
// recette pose un effet de tir — le rail, la gerbe — et ne se déclare nulle
// part par arme : `axes` ne porte que la liste close des axes. Une lourde qui
// devrait en profiter demanderait d'abord un champ pour le dire.
func (w *World) lourdeDe(rang int) Weapon {
	arme := w.armes.All[rang]
	for i := range w.passifs.Axes {
		axe := &w.passifs.Axes[i]
		if !slices.Contains(arme.Axes, axe.Axis) {
			continue
		}
		porter(&arme, axe, w.paliers[i])
	}
	return arme
}

// cibleDe rend la créature qu'une lourde vise depuis le joueur.
//
// **Sans restriction de côté, à la différence du tir de base** : une grenade
// tombe où le joueur la lance et non où il regarde, et le chapitre 9 pose qu'il
// ne dirige pas son lancer. Le fusil suit la même règle — ce qui s'oriente est
// le tir automatique, jamais ce qu'une touche déclenche.
func (w *World) cibleDe(arme *Weapon) (*Enemy, bool) {
	place, trouvee := w.plusProcheDe(w.playerX, w.playerY, arme.Range, Handle{}, Vec{})
	if !trouvee {
		return nil, false
	}
	return w.ennemis.At(place), true
}

// emporter retire des touches à la horde dans un rayon autour d'un point.
//
// **Elle n'emporte que la horde, et jamais le joueur.** L'inverse créerait une
// décision de placement — reculer avant de lancer — que le joueur ne peut pas
// prendre : il ne dirige pas son lancer, et le punir d'un geste qu'il n'oriente
// pas serait une sanction sans recours. C'est ce qui la sépare de la Baudruche,
// qu'on voit venir et qu'on esquive. La question se rouvrira le jour où une arme
// lourde deviendra dirigeable, et ce sera le bon moment.
//
// **Le rayon et les touches sont des paramètres et non une arme**, pour la
// raison qui avait fait passer une origine à `salve` : deux effets appellent
// cette géométrie avec des valeurs qu'ils ne lisent pas au même endroit — la
// déflagration ce qu'une explosion retire, les flammes ce qu'une impulsion
// retire. Un paramètre, pas une abstraction ; troisième fois que ce mécanisme
// s'étend ainsi.
//
// La transition de mort passe par le même chemin qu'un tir : le butin, et
// l'amorce d'une Baudruche qui explose à son tour.
func (w *World) emporter(x, y, rayon Fixed, touches int) {
	portee := int64(rayon) * int64(rayon)
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		if e.Hits <= 0 {
			continue
		}
		if (Vec{X: e.X - x, Y: e.Y - y}).carres() > portee {
			continue
		}

		e.Hits -= touches
		e.Flash = eclairImpact
		w.compter(e.X, e.Y, touches)
		if e.Hits <= 0 {
			w.lacher(e)
			w.amorcer(e)
		}
	}
}
