// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le pas de simulation et la conversion des durées de manifeste. La simulation
// ne connaît jamais le temps écoulé et n'accepte aucun delta : elle compte des
// ticks, et une durée trop courte pour en faire un est refusée.

package game

import (
	"errors"
	"fmt"
	"math"
)

// TPS est le nombre de pas de simulation par seconde. Fixe, et jamais réglable.
//
// Ebitengine appelle la mise à jour à cadence constante et rattrape un retard en
// l'appelant plusieurs fois d'affilée : la simulation n'a donc jamais à
// connaître le temps écoulé, et n'accepte aucun delta.
const TPS = 60

// Tick est l'unité de temps du paquet, et la seule.
//
// Un int32 couvre plus de quatre cents jours de jeu continu à cette cadence,
// alors qu'une run en dure quinze minutes.
type Tick int32

// ErrDurationTooShort refuse une durée qui n'atteint pas un pas de simulation.
//
// Refusée et non relevée à un tick : la relever produirait un fichier qui ment,
// où huit millisecondes auraient l'air de valoir du 125 Hz sans que rien ne le
// dise. Le refus est tenable parce que ces durées ne sont saisies par personne —
// elles sortent des générateurs, donc une telle valeur est un défaut dans un
// script.
var ErrDurationTooShort = errors.New("duree sous le pas de simulation")

// TicksFromMs convertit une durée de manifeste en ticks, arrondie au plus proche.
//
// Une seule fois au chargement, jamais à l'usage : à 60 Hz, 330 ms valent 19,8
// pas, et convertir à chaque appel ferait céder la caisse au 19e ou au 20e selon
// qui écrit le code.
func TicksFromMs(ms int) (Tick, error) {
	if ms*TPS < 1000 {
		return 0, fmt.Errorf("%w : %d ms, il en faut %d", ErrDurationTooShort, ms, msParTick())
	}
	ticks := (int64(ms)*TPS + 500) / 1000
	if ticks > math.MaxInt32 {
		return 0, fmt.Errorf("duree de %d ms : au-dela de ce qu un compteur de ticks porte", ms)
	}
	return Tick(ticks), nil // #nosec G115 -- borné à la ligne précédente
}

// TicksFromSeconds convertit un nombre entier de secondes en ticks.
//
// **Elle existe pour la frise d'un scénario de vagues, la seule chose du projet
// qui ne s'écrive pas en millisecondes** — un déroulé que son auteur relit comme
// une minuterie, et non une cadence de mécanisme sortie d'un générateur.
//
// Zéro est admis, là où `TicksFromMs` le refuse, et la différence est celle
// d'une durée à un instant : une frise commence à 0:00, ce qui n'est pas une
// durée trop courte mais son origine.
func TicksFromSeconds(secondes int) (Tick, error) {
	ticks := int64(secondes) * TPS
	if ticks > math.MaxInt32 {
		return 0, fmt.Errorf("%d secondes : au-dela de ce qu un compteur de ticks porte", secondes)
	}
	return Tick(ticks), nil // #nosec G115 -- borné à la ligne précédente
}

// msParTick rend la durée d'un pas, arrondie au plus proche, pour les messages.
func msParTick() int {
	return (1000 + TPS/2) / TPS
}

// Seconds rend la durée en secondes, pour l'affichage d'un score ou d'un temps
// de référence. Jamais pour décider dans la simulation, qui compte en ticks.
func (t Tick) Seconds() float64 {
	return float64(t) / TPS
}
