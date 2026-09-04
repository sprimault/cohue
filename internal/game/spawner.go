// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le spawner : un budget de pression qui s'accumule tick après tick, des
// créatures achetées dedans par meutes indivisibles, et un anneau hors du champ
// de vision où les poser. Trois façons de ne rien poser, et elles ne font pas la
// même chose du budget.

package game

// tentativesApparition est le nombre de directions tirées avant d'abandonner
// une apparition.
//
// Le même budget fixe que l'aimant, et pour la même raison : une boucle jusqu'à
// trouver consommerait un nombre de tirages qui dépend du lieu, si bien que deux
// salles différentes désynchroniseraient le flux. Abandonner est ce que la
// conception demande — plutôt aucune créature qu'une créature surgie d'un mur.
const tentativesApparition = 8

// apparaitre achète et pose les créatures du tick.
//
// **Trois façons de ne rien poser, et le budget ne finit pas au même endroit.**
// L'anneau bouché reporte, parce qu'un couloir étroit ne doit pas être un abri
// où la pression tombe. Le plafond d'effectif perd, sinon on retrouve le mur
// d'ennemis différé. Les plafonds de simultanéité perdent aussi : le budget
// n'achèterait rien d'autre, et le garder reviendrait à faire payer plus tard un
// Secouriste qu'on n'a pas eu.
//
// Elle vient juste après les entrées et avant le champ de flux, où la conception
// range les apparitions. Le champ ne dépend que du joueur et des obstacles, donc
// une créature apparue après son calcul n'y perd rien ; la densité, elle, dépend
// des ennemis, et deux créatures nées au même endroit se superposeraient le temps
// d'une image sans que personne ne retrouve jamais l'origine du scintillement.
func (w *World) apparaitre() {
	if len(w.scenario.Phases) == 0 {
		return
	}

	// **Les deux calculs passent par `int64` et saturent.** La pression d'une
	// phase vient d'un fichier que le dépôt ne produit pas : une courbe qui
	// demanderait un million par seconde ferait déborder le produit par la borne
	// de report, et un budget devenu négatif n'achèterait plus rien du reste de la
	// partie. Saturer donne un budget absurde mais monotone, que le plafond
	// d'effectif borne de toute façon.
	phase := w.scenario.phase(w.tick)
	accorde := phase.budget(w.tick)
	w.budget = borner(int64(w.budget) + int64(accorde))

	// **La borne ne descend jamais sous le prix d'une créature.** Elle limite
	// l'accumulation ; l'empêcher d'atteindre un seul achat ne serait plus une
	// limite mais un arrêt, et une phase à faible pression cesserait de produire
	// quoi que ce soit sans qu'aucun refus ne le dise.
	plafond := borner(int64(accorde) * int64(w.progression.CarryOver))
	if plafond < phase.Cheapest {
		plafond = phase.Cheapest
	}
	if w.budget > plafond {
		w.budget = plafond
	}

	w.compterLesVivants()
	for {
		sousPlafond, abordables := w.etalage(phase)
		if sousPlafond == 0 {
			w.budget = 0
			return
		}
		if abordables == 0 {
			return
		}

		profil := w.achetables[w.hasard.Waves.Pick(abordables)]
		x, y, trouve := w.placeApparition()
		if !trouve {
			return
		}

		// **Une seule position pour toute la meute.** Elles arrivent ensemble ou
		// ce ne sont que trois créatures du même profil, et un tirage par membre
		// les ferait naître aux quatre coins de l'anneau. Elles se superposent
		// donc à l'apparition, ce que la conception admet au chapitre 4 en
		// donnant une direction au vecteur nul ; la séparation les écarte au
		// même tick, tant que les apparitions y précèdent le comptage de
		// densité.
		meute := &w.profils.Enemies[profil]
		for range meute.Group {
			w.SpawnEnemy(profil, x, y)
		}
		w.budget -= meute.PackCost()
		w.vivants[profil] += meute.Group
	}
}

// etalage range dans `w.achetables` les profils que la phase autorise, dont la
// meute entière tient dans le bassin et sous leur plafond de simultanéité, et
// que le budget paie.
//
// Elle rend aussi le nombre de profils qui tiennent, budget mis à part, et c'est
// ce compte qui distingue les deux façons de ne rien acheter : plus rien qui
// tienne est une impasse dont le budget ne sortira pas, tandis que rien
// d'abordable se règle en attendant un tick de plus.
//
// **Une meute qui ne tient pas est écartée entière**, jamais rognée : le Molosse
// n'apparaît jamais seul, et un bassin presque plein est précisément le moment
// où l'exception s'écrirait. C'est aussi pour cela que la place restante se juge
// ici plutôt que dans la boucle appelante — un seul endroit décide de ce qu'on
// peut acheter, et c'est lui qui borne la boucle quand le bassin se remplit.
func (w *World) etalage(phase *Phase) (sousPlafond, abordables int) {
	w.achetables = w.achetables[:0]
	place := w.ennemis.Cap() - w.ennemis.Len()
	for _, p := range phase.Profiles {
		profil := &w.profils.Enemies[p]
		if profil.Group > place {
			continue
		}
		if profil.MaxAlive > 0 && w.vivants[p]+profil.Group > profil.MaxAlive {
			continue
		}
		sousPlafond++
		if profil.PackCost() <= w.budget {
			w.achetables = append(w.achetables, p)
		}
	}
	return sousPlafond, len(w.achetables)
}

// compterLesVivants recompte la horde par profil.
//
// **Un parcours complet plutôt qu'un compteur tenu à jour**, et ce n'est pas une
// négligence : trois cents incréments par tick ne coûtent rien, là où deux
// compteurs à maintenir de part et d'autre d'une suppression par échange sont
// exactement le genre d'état qui se désynchronise sans qu'aucun test ne le voie.
// Le budget d'allocation, lui, tient : la tranche est montée une fois.
func (w *World) compterLesVivants() {
	for i := range w.vivants {
		w.vivants[i] = 0
	}
	for i := range w.ennemis.Active() {
		w.vivants[w.ennemis.At(i).Profile]++
	}
}

// placeApparition tire une position sur l'anneau, hors du champ de vision.
//
// **La direction se tire dans le disque puis se normalise**, et non sur le carré
// qui l'enferme : une direction tirée sur un carré puis ramenée à l'unité arrive
// plus souvent près des axes que des diagonales, et la horde naîtrait de biais
// sans que rien ne l'explique. Le rejet coûte un tirage sur cinq et rend un angle
// uniforme sans une ligne de trigonométrie — que la virgule fixe proscrit de
// toute façon, `sin` et `cos` n'étant pas correctement arrondis.
//
// Le rayon est constant : l'anneau est un cercle et non une couronne. Une
// épaisseur donnerait un second chiffre à régler pour une variété que personne ne
// verrait, les créatures arrivant toutes du hors-champ.
func (w *World) placeApparition() (Fixed, Fixed, bool) {
	rayon := w.progression.SpawnRadius
	for range tentativesApparition {
		tire := Vec{
			X: w.hasard.Positions.Fixed(2*One) - One,
			Y: w.hasard.Positions.Fixed(2*One) - One,
		}
		if carres := tire.carres(); carres == 0 || carres > int64(One)*int64(One) {
			continue
		}

		// L'index de `Direction` ne sert qu'au vecteur nul, que le rejet
		// ci-dessus a déjà écarté.
		vers := tire.Direction(0)
		x, y := w.playerX+vers.X.Mul(rayon), w.playerY+vers.Y.Mul(rayon)
		if !w.passable(x, y) {
			continue
		}
		return x, y, true
	}
	return 0, 0, false
}
