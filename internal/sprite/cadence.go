// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Où en est une animation : l'image d'un cycle qui boucle, celle d'un cycle qui
// s'achève. Rien n'est stocké, tout se dérive de ce que la simulation tient
// déjà.

package sprite

import "github.com/sprimault/cohue/internal/game"

// Loop rend l'image d'un cycle qui boucle, au tick donné et pour un décalage.
//
// **Il n'y a pas d'état d'animation, et c'est la décision qui compte ici.** Un
// compteur par entité aurait posé deux questions dont aucune n'a de bonne
// réponse : entre-t-il dans l'empreinte d'une run, lui qui est cosmétique ? et
// qui l'avance, la simulation qui n'en a que faire ou le rendu qui n'a nulle
// part où le ranger, les places d'un bassin changeant à chaque mort ? Dérivée du
// tick, la question disparaît au lieu d'être tranchée.
//
// **Le décalage vient de l'identifiant de l'entité dans son bassin.** Sans lui,
// une horde entière marche au pas cadencé, ce qui se voit immédiatement ; avec
// lui, deux créatures voisines sont à des images différentes sans qu'aucun
// tirage soit consommé. C'est le troisième emploi de l'identifiant comme source
// de variation déterministe, après la direction d'un vecteur dégénéré et
// l'étalement d'une volée de gemmes.
//
// Une cadence nulle rend la première image : le chargement refuse une durée sous
// le pas, mais un cycle bâti à la main dans un test n'a pas traversé ce refus.
func Loop(c game.Cycle, tick game.Tick, decalage int) int {
	if c.Frames <= 1 || c.Duration <= 0 {
		return 0
	}
	return (int(tick/c.Duration) + decalage) % c.Frames
}

// Once rend l'image d'un cycle qui ne boucle pas, d'après ce qui reste à l'état
// qui le porte.
//
// **L'animation s'achève exactement quand l'état s'achève**, et c'est ce que le
// décompte permet de dire sans rien stocker : la simulation tient déjà `Flash`,
// `ChargeTimer` et `Healing`, qui sont des décomptes et non des dates. Une
// anticipation dont la pose se complète à l'instant du départ est ce qu'un
// télégraphe doit montrer.
//
// **Ancrer sur la fin est la seule forme sans état, et ce n'est pas un choix.**
// Au premier tick d'un état, tout ce qu'on lit est le décompte entier : rien n'y
// distingue un état long qui commence d'un état court, si bien qu'« où en est-on
// depuis le début » n'est pas calculable. Ancrer sur le début exigerait la durée
// totale de l'état, que le rendu n'a pas et n'aura qu'en la lui donnant.
//
// D'où le régime à connaître : **si le décompte est plus court que le cycle,
// l'animation entre en cours de route** et se termine quand même. C'est ce que
// le décompte signifie — un état écourté est un état interrompu, donc une
// animation qui n'a pas eu le temps de se dérouler entière, et en montrer la fin
// plutôt que le début est ce qui correspond au sens de ce qu'on lit.
func Once(c game.Cycle, reste game.Tick) int {
	if c.Frames <= 1 || c.Duration <= 0 {
		return 0
	}
	return max(0, c.Frames-1-int(reste/c.Duration))
}
