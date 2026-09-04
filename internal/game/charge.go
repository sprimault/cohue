// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La charge : trois phases sur l'entité, une direction figée au départ, et trois
// façons d'en sortir — le mur, le joueur touché, la durée épuisée. Aucune ne lit
// le comportement du profil, seulement ses paramètres.

package game

// ChargePhase est où une créature en est de son cycle de charge.
//
// **Un état sur l'entité, jamais une branche sur `Behaviour`.** Le tangentiel de
// l'Arpenteur donne le patron : il produit un contournement sans qu'aucune ligne
// ne lise le comportement, parce que son paramètre vaut zéro partout ailleurs.
// La charge demande en plus de la mémoire d'un tick à l'autre, ce que la
// conception autorise à hauteur de trois ou quatre états — la mémoire vit donc
// dans la créature, et le profil ne porte toujours que des nombres.
type ChargePhase uint8

// Les phases du cycle. `ChargeNone` vaut zéro : c'est l'état d'une créature
// qu'on vient de poser, et celui de tout profil qui ne charge pas.
const (
	ChargeNone ChargePhase = iota
	ChargeTelegraph
	ChargeRun
	ChargeRecover
)

// charger avance le cycle de charge d'une créature et rend le pas qu'il impose,
// plus un booléen disant s'il remplace celui de la poursuite.
//
// Elle est appelée avant que l'intention ordinaire soit calculée, et c'est ce
// qui la rend simple à lire : pendant le télégraphe et la récupération, la
// créature est immobile ; pendant la course, elle suit sa direction figée et
// ignore le champ de flux comme la poussée de ses voisines. Le reste du temps
// elle ne décide de rien.
//
// **Le déclenchement ne vérifie aucune ligne de vue.** Un pilier entre les deux
// n'empêche pas la charge : elle part, elle s'y arrête, et la créature paie sa
// récupération pour rien. C'est le seul usage défensif que le décor ait, et le
// vérifier le supprimerait.
func (w *World) charger(e *Enemy, profil *EnemyProfile) (Vec, bool) {
	if profil.ChargeRange == 0 {
		return Vec{}, false
	}

	if e.ChargeTimer > 0 {
		e.ChargeTimer--
	}

	switch e.ChargePhase {
	case ChargeNone:
		ecart := Vec{X: w.playerX - e.X, Y: w.playerY - e.Y}
		if ecart.carres() > int64(profil.ChargeRange)*int64(profil.ChargeRange) {
			return Vec{}, false
		}
		e.ChargePhase = ChargeTelegraph
		e.ChargeTimer = profil.Telegraph
		return Vec{}, true

	case ChargeTelegraph:
		if e.ChargeTimer > 0 {
			return Vec{}, true
		}
		// La direction se prend à la fin de l'anticipation et non à son début :
		// c'est ce qui rend l'esquive latérale possible, puisque se décaler
		// pendant le télégraphe déplacerait sinon une charge déjà visée.
		e.ChargePhase = ChargeRun
		e.ChargeTimer = profil.ChargeDuration
		e.ChargeDir = Vec{X: w.playerX - e.X, Y: w.playerY - e.Y}.Direction(0)
		return e.ChargeDir.Scale(w.vitesse(profil.Speed, e.X, e.Y)), true

	case ChargeRun:
		if e.ChargeTimer == 0 {
			w.finirLaCharge(e, profil)
			return Vec{}, true
		}
		return e.ChargeDir.Scale(w.vitesse(profil.Speed, e.X, e.Y)), true

	default: // ChargeRecover
		if e.ChargeTimer > 0 {
			return Vec{}, true
		}
		e.ChargePhase = ChargeNone
		return Vec{}, false
	}
}

// finirLaCharge clôt la course et ouvre la récupération.
//
// Les trois fins passent par ici — la durée épuisée, le mur, le joueur touché —
// et c'est ce qui garde la récupération vraie dans les trois cas. Elle suit
// **toute** fin de course : sans cela, une charge aboutie enchaînerait sur la
// suivante et la créature n'aurait aucun moment vulnérable, quand la conception
// lui oppose l'esquive latérale.
func (w *World) finirLaCharge(e *Enemy, profil *EnemyProfile) {
	e.ChargePhase = ChargeRecover
	e.ChargeTimer = profil.Recovery
	e.ChargeDir = Vec{}
}

// arreterAuMur termine la course quand un obstacle a mordu le pas.
//
// **La détection compare le pas obtenu au pas voulu du tick, et strictement.**
// `glisser` n'arrondit pas et annule une composante entière plutôt que de la
// raccourcir, si bien que l'écart est en tout ou rien ; la densité ne change que
// la direction du pas, et le coût du terrain agit des deux côtés de la
// comparaison. Comparer au pas **nominal** du profil, en revanche, arrêterait la
// charge à la première flaque.
//
// Elle tient donc à une propriété de `glisser`, qui le dit de son côté : le jour
// où il raccourcirait un pas au lieu d'annuler une composante, cette égalité
// deviendrait fausse sans que rien ne le signale.
func (w *World) arreterAuMur(e *Enemy, profil *EnemyProfile, voulu Vec) {
	if e.ChargePhase == ChargeRun && e.Step != voulu {
		w.finirLaCharge(e, profil)
	}
}

// Charging dit si une créature est dans sa course, télégraphe et récupération
// exclus.
//
// Le rendu la lit pour distinguer ce qu'il annonce de ce qui arrive ; ce qu'il
// en fait — la teinte, le contour — lui appartient.
func (e *Enemy) Charging() bool { return e.ChargePhase == ChargeRun }

// Telegraphing dit si une créature annonce sa charge sans l'avoir commencée.
//
// C'est la seule des trois phases que le joueur doit lire : une charge annoncée
// qu'on ne voit pas venir n'est qu'un coup qui tombe, et le décor ne sert alors
// à rien.
func (e *Enemy) Telegraphing() bool { return e.ChargePhase == ChargeTelegraph }
