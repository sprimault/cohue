// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le soin du Secouriste : une voisine par impulsion, la plus blessée à portée,
// et les deux éclairs qui disent qui soigne et qui a été soigné.

package game

// soigner rend des touches aux voisines des créatures qui savent le faire.
//
// **Le seul profil qui crée une priorité de cible.** Les six autres cassent le
// kiting par la position ; celui-ci annule le travail tant qu'il vit, et comme le
// tir prend le plus proche sans que le joueur puisse choisir, l'abattre demande
// d'aller vers lui — donc de quitter sa position pour entrer dans la horde.
//
// Elle vient après les dégâts et avant les suppressions : une créature soignée
// dans le même tick où elle est touchée doit voir les deux s'appliquer, et une
// résistance remontée au-dessus de zéro avant le nettoyage est précisément ce
// qui « annule le travail ».
func (w *World) soigner() {
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		if e.Healing > 0 {
			e.Healing--
		}
		if e.Healed > 0 {
			e.Healed--
		}

		profil := &w.profils.Enemies[e.Profile]
		if profil.HealRange == 0 || e.Hits <= 0 {
			continue
		}
		if e.HealTimer > 0 {
			e.HealTimer--
			continue
		}

		blessee := w.plusBlessee(i, profil)
		if blessee < 0 {
			// Rien à soigner : la cadence reste prête, comme celle d'une arme
			// sans cible. La consommer à vide ferait dépendre le premier soin du
			// temps passé au milieu d'une horde intacte.
			continue
		}

		cible := w.ennemis.At(blessee)
		cible.Hits = min(cible.Hits+profil.HealHits, cible.MaxHits)
		cible.Healed = eclairSoin
		e.Healing = eclairSoin
		e.HealTimer = profil.HealCooldown
	}
}

// plusBlessee rend la place de la voisine à qui le soin manque le plus.
//
// **Elle-même exclue**, ce qui est la moitié de ce qui rend le Secouriste
// abattable : ses trois touches tombent vite une fois qu'on l'a rejoint, et c'est
// cette récompense qui paie le trajet.
//
// **Un mort n'est jamais choisi.** Une résistance nulle *est* la mort, et le
// nettoyage la retire au même tick : soigner un cadavre créerait l'état
// mort-vivant que le bassin ne connaît pas, et que l'explosion de la Baudruche a
// déjà évité en vivant dans son propre bassin.
//
// **L'entame se mesure contre la résistance d'apparition, jamais contre celle de
// la table.** Le durcissement d'une phase multiplie les touches à la naissance :
// mesurée contre le manifeste, une créature née sous 1,7 paraîtrait entamée de
// plusieurs touches sans avoir jamais été touchée, et le soin irait à la plus
// dure de la horde plutôt qu'à la plus blessée. Le lieu livré n'ouvre le
// Secouriste qu'à la dixième minute, c'est-à-dire dans les seules phases qui
// durcissent.
//
// À manque égal, la première rencontrée l'emporte : l'ordre du bassin est stable
// à l'intérieur d'un tick, ce qui suffit au déterminisme.
func (w *World) plusBlessee(soigneur int, profil *EnemyProfile) int {
	choix, manque := -1, 0
	for i := range w.ennemis.Active() {
		if i == soigneur {
			continue
		}
		e := w.ennemis.At(i)
		if e.Hits <= 0 {
			continue
		}
		perdu := e.MaxHits - e.Hits
		if perdu <= manque {
			continue
		}

		s := w.ennemis.At(soigneur)
		ecart := Vec{X: e.X - s.X, Y: e.Y - s.Y}
		if ecart.carres() > int64(profil.HealRange)*int64(profil.HealRange) {
			continue
		}
		choix, manque = i, perdu
	}
	return choix
}
