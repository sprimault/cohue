// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'aimant : l'objet qui apparaît régulièrement, la charge que le joueur garde,
// et la ruée des gemmes qu'il déclenche.

package game

// tentativesAimant est le nombre de cases tirées avant d'abandonner une
// apparition.
//
// **Un budget fixe et non une boucle jusqu'à trouver.** Une boucle consommerait
// un nombre de tirages qui dépend du lieu, si bien que deux salles différentes
// désynchroniseraient le flux — et une salle où presque rien ne convient
// tournerait longtemps. Abandonner est la même réponse que celle de l'anneau
// d'apparition : plutôt aucun aimant qu'un aimant sur un mur.
const tentativesAimant = 8

// Magnet est l'aimant posé au sol.
//
// Il ne porte que sa place : ce qu'il fait appartient à la partie, et un objet
// qui porterait son effet ferait deux descriptions du même mécanisme le jour où
// une seconde sorte apparaîtrait.
type Magnet struct {
	X, Y Fixed
}

// Magnets rend le bassin des aimants au sol.
//
// Un bassin de capacité un plutôt qu'un champ optionnel : le rendu parcourt les
// bassins pour composer sa scène, et une entité rangée autrement y serait un cas
// particulier à écrire deux fois.
func (w *World) Magnets() *Pool[Magnet] { return w.aimants }

// Charged dit si le joueur tient une charge.
func (w *World) Charged() bool { return w.charge }

// Charge donne une charge au joueur, sans passer par un objet au sol.
//
// Elle existe pour la planche de relecture, qui doit montrer une ruée sans
// jouer les trente secondes d'une apparition — au bout desquelles la horde a
// tué le joueur. Ce n'est pas une voie rapide du jeu : la boucle ne l'appelle
// jamais, et un aimant se ramasse.
func (w *World) Charge() { w.charge = true }

// poserAimant fait apparaître un aimant quand son heure vient.
//
// **Un seul à la fois au sol, une seule charge gardée.** La conception parle
// d'une charge au singulier, et rien n'apparaît tant que celle qu'on tient n'est
// pas dépensée : le joueur ne voit donc jamais au sol un aimant qu'il ne peut pas
// prendre.
//
// **Ce qui tient la première moitié de cette règle est la capacité du bassin**,
// qui vaut un : `Spawn` refuse quand il est plein, et l'unicité ne dépend donc
// d'aucune vigilance. Le retour anticipé ci-dessous n'est pas cette garantie mais
// une économie de tirages — sans lui, chaque période consommerait huit positions
// du flux pour une apparition que le bassin ou la charge refusera de toute façon.
//
// La période se compte depuis la dernière apparition et non depuis le début de la
// partie : un aimant qui reste au sol ne fait pas s'accumuler les suivants, et le
// joueur qui tarde ne trouve pas trois aimants d'un coup à son retour.
func (w *World) poserAimant() {
	if w.tick-w.dernierAimant < w.progression.MagnetPeriod {
		return
	}
	if w.aimants.Len() > 0 || w.charge {
		return
	}

	x, y, trouve := w.placeAimant()
	if !trouve {
		// L'heure est passée quand même : réessayer au tick suivant ferait tirer
		// à chaque image dans un lieu encombré, ce qui viderait le flux à une
		// vitesse que rien ne borne.
		w.dernierAimant = w.tick
		return
	}
	w.aimants.Spawn(Magnet{X: x, Y: y})
	w.dernierAimant = w.tick
}

// placeAimant tire une case passable assez loin du joueur.
//
// Elle consomme le flux des positions, celui où le spawner puise aussi, puisque
// les deux placent quelque chose dans le lieu. Le nombre de tirages est borné et
// connu d'avance, ce qui laisse le flux dans un état prévisible que la salle soit
// dégagée ou non.
func (w *World) placeAimant() (Fixed, Fixed, bool) {
	mini := int64(w.progression.MagnetMinRange)
	for range tentativesAimant {
		u := w.hasard.Positions.IntN(w.grille.Width())
		v := w.hasard.Positions.IntN(w.grille.Height())
		if !w.grille.Passable(u, v) {
			continue
		}

		x, y := FromInt(u)+One/2, FromInt(v)+One/2
		if (Vec{X: x - w.playerX, Y: y - w.playerY}).carres() < mini*mini {
			continue
		}
		return x, y, true
	}
	return 0, 0, false
}

// prendreAimant ramasse l'aimant que le joueur atteint.
//
// À la même portée qu'une gemme : deux portées de ramassage différentes seraient
// deux règles là où le joueur n'en lit qu'une, et il jugerait la seconde
// capricieuse.
func (w *World) prendreAimant() {
	if w.charge || w.aimants.Len() == 0 {
		return
	}
	portee := int64(w.progression.PickupRange)
	a := w.aimants.At(0)
	if (Vec{X: a.X - w.playerX, Y: a.Y - w.playerY}).carres() <= portee*portee {
		w.aimants.RemoveAt(0)
		w.charge = true
	}
}

// Attract dépense la charge et lance la ruée des gemmes.
//
// **Sans effet quand rien n'est chargé**, et le silence est voulu : l'appelant
// est un clavier, et une touche pressée sans charge ne doit ni consommer ni
// avertir. Le joueur voit son emplacement vide, ce qui suffit.
//
// Elle ne collecte rien elle-même. Les gemmes convergent, et c'est la passe de
// ramassage qui les prend à l'arrivée — un aimant qui les ferait disparaître
// d'un coup escamoterait ce que la conception nomme le moment de plaisir maximal
// du genre.
func (w *World) Attract() {
	if !w.charge {
		return
	}
	w.charge = false
	for i := range w.gemmes.Active() {
		w.gemmes.At(i).Pulled = true
	}
}

// attirer avance d'un pas les gemmes que l'aimant a saisies.
//
// La direction se recalcule à chaque tick plutôt que de se figer au
// déclenchement : le joueur continue de courir pendant la ruée, et des gemmes
// qui viseraient l'endroit où il était le manqueraient toutes.
func (w *World) attirer() {
	pas := w.progression.PullSpeed
	for i := range w.gemmes.Active() {
		g := w.gemmes.At(i)
		if !g.Pulled {
			continue
		}
		vers := Vec{X: w.playerX - g.X, Y: w.playerY - g.Y}.Direction(i).Scale(pas)
		g.X, g.Y = g.X+vers.X, g.Y+vers.Y
	}
}
