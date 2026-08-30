// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package game

import "math/rand/v2"

// Stream est une source de hasard destinée à un usage, et à un seul.
//
// Elle n'expose que des entiers : `Float64` reste hors d'atteinte, faute de quoi
// un flottant rentrerait dans la simulation par la porte de derrière, après
// qu'on a payé la virgule fixe pour l'en tenir dehors.
type Stream struct {
	source *rand.Rand
}

// Streams porte les quatre sources d'une partie, dérivées d'une même graine.
//
// Quatre et non une : le test central du projet joue une run **sans rendu**, et
// une exécution qui ne tirerait pas les teintes de vêtement décalerait tous les
// tirages suivants. Les vagues d'une run simulée cesseraient alors de
// correspondre à celles de la même graine jouée à l'écran, c'est-à-dire que
// l'outil d'équilibrage mesurerait autre chose que le jeu.
//
// Ce qui n'a aucun effet sur la simulation est confiné dans Cosmetic. Ce qui est
// interdit n'est donc pas le flux, mais qu'une décision lise ce qu'il alimente —
// et cela se vérifie plutôt que de se surveiller : deux runs sur la même graine,
// teintes forcées différentes, rendent la même empreinte d'état.
type Streams struct {
	// Waves achète les créatures dans le budget de pression.
	Waves *Stream
	// Positions place les apparitions sur l'anneau et tout ce qui relève du lieu.
	Positions *Stream
	// Loot décide du contenu des caisses et de ce que lâche une créature.
	Loot *Stream
	// Cosmetic ne décide de rien : teintes de vêtement, variantes d'éclat.
	Cosmetic *Stream
}

// Les identifiants de flux, passés à PCG comme numéro de suite : deux suites
// d'une même graine sont indépendantes, c'est ce à quoi sert ce paramètre.
//
// Des constantes littérales et non le rang dans une énumération. Insérer un flux
// au milieu décalerait tous les suivants et invaliderait d'un coup toutes les
// graines publiées — un numéro attribué ne se réattribue jamais, et un flux
// nouveau prend le suivant. C'est la même règle que les numéros d'étape de la
// feuille de route, pour la même raison.
const (
	suiteWaves     uint64 = 1
	suitePositions uint64 = 2
	suiteLoot      uint64 = 3
	suiteCosmetic  uint64 = 4
)

// NewStreams dérive les quatre flux de la graine d'une partie.
//
// PCG, et le nom compte autant que le choix : l'algorithme est spécifié et
// stable d'une version de Go à l'autre, là où le générateur global ne l'est pas
// et se trouve proscrit par le déterminisme de la run. En changer invaliderait
// toutes les graines publiées.
func NewStreams(graine uint64) *Streams {
	nouveau := func(suite uint64) *Stream {
		// #nosec G404 -- un générateur cryptographique serait le défaut : on
		// veut rejouer la même partie deux fois, pas résister à un adversaire.
		// Rien de ce qui est tiré ici ne protège quoi que ce soit.
		return &Stream{source: rand.New(rand.NewPCG(graine, suite))}
	}
	return &Streams{
		Waves:     nouveau(suiteWaves),
		Positions: nouveau(suitePositions),
		Loot:      nouveau(suiteLoot),
		Cosmetic:  nouveau(suiteCosmetic),
	}
}

// IntN rend un entier de [0, n).
func (s *Stream) IntN(n int) int {
	return s.source.IntN(n)
}

// Fixed rend une longueur de [0, max).
//
// Par les entiers : tirer un flottant puis le convertir ferait dépendre le
// résultat de l'arrondi de la conversion, et la longueur cesserait d'être
// identique d'une architecture à l'autre.
func (s *Stream) Fixed(max Fixed) Fixed {
	if max <= 0 {
		return 0
	}
	return Fixed(s.source.Int32N(int32(max)))
}

// Pick rend un index de [0, n), et 0 pour un ensemble vide plutôt qu'une panique.
//
// Le cas arrive : un scénario dont aucun profil n'est autorisé, un butin dont la
// table est vide. Il vaut mieux un choix inerte qu'un arrêt du jeu, et l'appelant
// qui tient à le savoir teste son ensemble avant.
func (s *Stream) Pick(n int) int {
	if n <= 0 {
		return 0
	}
	return s.source.IntN(n)
}
