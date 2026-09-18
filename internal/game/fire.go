// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les flammes qu'une charge pose au sol : ce qu'elles brûlent par impulsions
// tant qu'elles durent, et ce que le rendu lit pour les peindre.

package game

// Fire est une zone de flammes posée, et ce qu'il lui reste à brûler.
//
// **Un bassin propre, jamais une variante de `Blast`.** Ce que les deux
// partagent est la géométrie, que `emporter` porte ; ce qui les sépare est
// l'instant — une déflagration s'applique **à la fin** d'une mèche, une flaque
// **pendant** toute sa durée. Un champ qui dirait lequel des deux serait le
// drapeau qui change le sens de la structure, et c'est ce qui range déjà le
// cadavre à part plutôt qu'en ennemi marqué mort.
//
// **Elle ne dévie personne, et ce n'est pas l'invariant du coût qui le décide.**
// Ce serait aussi lui — la grille arrête son nombre de seaux au montage —, mais
// l'argument qui tranche est que le décor protège **par le fait** partout
// ailleurs : le pilier reçoit la charge du Molosse, le projectile de la Buse y
// meurt, sans qu'aucune condition soit vérifiée. Une horde qui traverse et fond
// est un passage interdit, et cela se lit mieux qu'une horde qui contournerait,
// où le joueur douterait de ce que ses flammes font.
type Fire struct {
	// X et Y sont le point de pose, où la flaque reste centrée.
	X, Y Fixed
	// Weapon est son rang dans la table, qui porte son rayon et sa cadence.
	Weapon int
	// Life est ce qui lui reste à brûler, en ticks. À zéro, elle s'éteint.
	//
	// **Un décompte et non une date de naissance**, comme la mèche d'un souffle
	// et la cadence d'une tourelle : ce que pose une lourde porte ce qui lui
	// reste. L'épave et la gemme font l'inverse, et c'est leur droit — rien de ce
	// qu'elles mesurent ne décide, quand ce compte-ci arrête une impulsion.
	Life Tick
	// Pulse est ce qui reste avant l'impulsion suivante, en ticks.
	Pulse Tick
}

// Fires rend le bassin des flammes posées.
func (w *World) Fires() *Pool[Fire] { return w.flammes }

// FireWeapon rend l'arme dont une flaque tient ses valeurs.
func (w *World) FireWeapon(f *Fire) *Weapon { return &w.armes.All[f.Weapon] }

// poserDesFlammes pose une flaque sous le joueur, et dit si elle tient.
//
// **Sous le joueur, comme la tourelle et non comme la grenade.** Ce qu'une
// flaque interdit est un passage, et le joueur le choisit en s'y tenant ; la
// poser sur une cible la mettrait là où la horde est déjà, c'est-à-dire trop
// tard pour l'arrêter.
//
// **Elle brûle dès le tick de sa pose**, sa première impulsion n'attendant pas
// la cadence : on la pose sur ce qui arrive, et une demi-seconde d'inertie
// suffirait à laisser passer ce qu'elle devait brûler.
//
// Le bassin plein refuse, et la charge n'est alors pas dépensée : c'est la règle
// que toutes les branches suivent, une charge dépensée pour un effet qui n'est
// pas parti coûtant au joueur une des trois choses qu'il possède.
func (w *World) poserDesFlammes(arme *Weapon, rang int) bool {
	_, ok := w.flammes.Spawn(Fire{
		X: w.playerX, Y: w.playerY,
		Weapon: rang,
		Life:   arme.Duration,
		Pulse:  0,
	})
	return ok
}

// bruler applique les impulsions des flaques et retire celles qui s'éteignent.
//
// **L'impulsion se consomme même à vide**, à la différence du tir d'une
// tourelle : ce qui s'épuise ici est une durée et non un compte, si bien qu'une
// impulsion retenue ne serait rendue à personne — elle brûlerait plus tard sur
// une flaque qui a moins de temps devant elle. La cadence qui ne se consomme pas
// à vide protège un stock ; il n'y en a pas ici.
//
// **Les flammes n'atteignent pas le joueur**, par la règle du chapitre 9 : ce
// qu'une lourde emporte n'emporte pas celui qui la déclenche. Il se tient dans
// sa propre flaque par construction, puisqu'elle se pose sous lui, et l'y
// blesser ferait d'un effet qu'il ne dirige pas une sanction sans recours.
//
// **Le retrait réexamine la place libérée**, comme celui d'une tourelle : cette
// passe ne fait pas avancer d'entité, et sauter celle que l'échange remonte
// laisserait une flaque éteinte brûler un tick de plus.
func (w *World) bruler() {
	for i := 0; i < w.flammes.Len(); {
		f := w.flammes.At(i)
		arme := w.lourdeDe(f.Weapon)

		if f.Pulse > 0 {
			f.Pulse--
		} else {
			w.emporter(f.X, f.Y, arme.Radius, arme.Hits)
			f.Pulse = arme.Cooldown
		}

		f.Life--
		if f.Life <= 0 {
			w.flammes.RemoveAt(i)
			continue
		}
		i++
	}
}

// FireBounds rend les cases que la flaque peut atteindre.
//
// Le même découpage que pour une explosion, et pour la même raison : le rendu
// n'a ni rayon ni table à connaître, et la zone qu'il peint est celle que les
// impulsions appliquent. Deux calculs de la même zone finiraient par peindre une
// case que le feu épargne.
func (w *World) FireBounds(f *Fire) (u0, v0, u1, v1 int) {
	rayon := w.lourdeDe(f.Weapon).Radius
	return (f.X - rayon).Floor(), (f.Y - rayon).Floor(),
		(f.X + rayon).Floor(), (f.Y + rayon).Floor()
}

// FireCovers dit si le centre d'une case brûle.
func (w *World) FireCovers(f *Fire, u, v int) bool {
	rayon := w.lourdeDe(f.Weapon).Radius
	ecart := Vec{X: FromInt(u) + One/2 - f.X, Y: FromInt(v) + One/2 - f.Y}
	return ecart.carres() <= int64(rayon)*int64(rayon)
}

// FireLeft rend ce qui reste de vie rapporté à sa durée totale, sur mille.
//
// **Il descend là que celui d'une mèche monte**, et c'est ce qui sépare les deux
// nappes à l'œil : un télégraphe annonce ce qui va venir et s'intensifie
// jusqu'au coup, une flaque montre ce qui brûle et faiblit en s'éteignant. Sans
// cette opposition, deux marquages de même forme diraient deux choses contraires
// dans la même teinte.
// **Le rapport est juste tant qu'aucun axe ne touche la durée.** Ce qu'il
// divise est la durée telle qu'elle est maintenant, quand `Life` porte celle
// qui a servi à poser la flaque : le jour où un palier allongerait une zone,
// une flaque posée avant lui rendrait un rapport sous mille sans avoir vieilli.
func (w *World) FireLeft(f *Fire) int {
	total := w.lourdeDe(f.Weapon).Duration
	if total <= 0 {
		return 0
	}
	return int(f.Life) * 1000 / int(total)
}
