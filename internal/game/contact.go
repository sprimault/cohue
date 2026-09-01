// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le contact : ce que la horde retire au joueur tant qu'elle le touche, et le
// plafond par seconde sans lequel un encerclement tuerait instantanément.

package game

// subir accumule les dégâts du contact et en retire des points de vie.
//
// **Les dégâts sont continus, pas un choc à l'impact.** Une créature collée
// retire sa part chaque tick tant qu'elle touche ; se dégager arrête la perte
// sans qu'aucun coup n'ait été « porté ». C'est ce qui rend la fuite payante à
// tout instant plutôt qu'entre deux frappes.
//
// **Le plafond vaut pour la somme, jamais par créature.** Vingt par seconde
// quel que soit le nombre d'ennemis au contact : sans lui, un encerclement à
// douze tue en une demi-seconde et le joueur ne peut pas se raconter sa mort.
// C'est la règle du chapitre 5 de la conception, et la seule de ce fichier qui
// ne soit pas mécanique.
func (w *World) subir() {
	if !w.Alive() {
		return
	}

	somme := 0
	for i := range w.ennemis.Len() {
		e := w.ennemis.At(i)
		if !w.auContact(e) {
			continue
		}
		somme += w.profils.Enemies[e.Profile].ContactDamage
	}
	if plafond := w.profils.Player.DamageCap; somme > plafond {
		somme = plafond
	}
	if somme == 0 {
		return
	}

	// **L'accumulateur compte en points-ticks, et c'est ce qui évite une
	// troisième échelle.** Vingt points par seconde à soixante ticks font un
	// tiers de point par tick, qu'aucun entier ne représente : passer la vie en
	// virgule fixe la ferait entrer dans le type des tuiles, dont les bornes et
	// l'échelle sont celles d'une position. En accumulant `points × ticks` et en
	// retirant un point chaque fois qu'on atteint `TPS`, la perte est exacte sur
	// une seconde et le reste se garde d'un tick à l'autre.
	w.degatsSubis += somme

	// La condition d'arrêt porte la vie autant que l'accumulateur, et c'est elle
	// qui interdit une vie négative — sans garde-fou à côté, qui serait du code
	// mort et qu'un test finirait par déclarer nécessaire.
	//
	// La vie à zéro **est** la mort, comme la résistance d'une créature : pas de
	// drapeau à synchroniser. Ce qui se déclenche une fois est la transition, et
	// c'est cette boucle qui la produit — l'écran et la relance s'y brancheront,
	// jamais sur un tick où la vie est déjà nulle.
	for w.degatsSubis >= TPS && w.vie > 0 {
		w.degatsSubis -= TPS
		w.vie--
	}
}

// auContact dit si une créature touche le joueur.
//
// Les rayons viennent des profils et jamais d'une distance choisie : ce sont eux
// que le manifeste porte, et c'est la même mesure qui décide du contact et de la
// séparation. La comparaison se fait sur les carrés, pour n'extraire aucune
// racine dans la boucle.
func (w *World) auContact(e *Enemy) bool {
	portee := w.profils.Player.Radius + w.profils.Enemies[e.Profile].Radius
	ecart := Vec{X: e.X - w.playerX, Y: e.Y - w.playerY}
	return ecart.carres() < int64(portee)*int64(portee)
}

// Health rend les points de vie restants.
func (w *World) Health() int { return w.vie }

// Alive dit si la partie continue.
//
// Elle se lit plutôt qu'elle ne se retient : la vie à zéro est la mort, donc
// aucun état parallèle ne peut en diverger.
func (w *World) Alive() bool { return w.vie > 0 }
