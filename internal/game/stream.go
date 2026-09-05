// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les quatre flux aléatoires d'une partie, dérivés de sa graine et nommés par
// leur usage. Un flux unique suffirait à rejouer une partie jouée et casserait
// la run simulée sans rendu, où chaque tirage manquant décale tous les suivants.

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
// et cela se vérifie plutôt que de se surveiller : `TestLeCosmetiqueNeDecideDeRien`
// joue deux fois la même graine en décalant ce flux, et attend la même empreinte
// d'état. Son jumeau `TestLesFigurantsBougentAvecLeCosmetique` garde l'autre
// moitié — un flux que plus personne ne consommerait rendrait le premier vert
// sans rien séparer, ce qu'il a fait tant qu'aucune entité n'y puisait.
type Streams struct {
	// Waves achète les créatures dans le budget de pression.
	Waves *Stream
	// Positions place les apparitions sur l'anneau et tout ce qui relève du lieu.
	Positions *Stream
	// Loot décidera du contenu des caisses et de ce que lâche une créature.
	//
	// **Au futur, et rien ne l'alimente encore** : ce qu'une caisse laisse est un
	// nombre fixe de gemmes, et une créature laisse celui de son profil. Son seul
	// appel est le témoin que l'empreinte imprime — c'est ce qui garde le flux
	// numéroté, donc ce qui empêche qu'une graine publiée change de sens le jour
	// où un butin se tirera vraiment.
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
//
// La dernière ne nomme pas un flux de partie mais la dérivation d'une graine à
// la suivante : elle occupe le même espace de numéros parce que c'est la même
// contrainte — une suite réattribuée changerait ce que rejoue une graine publiée.
const (
	suiteWaves     uint64 = 1
	suitePositions uint64 = 2
	suiteLoot      uint64 = 3
	suiteCosmetic  uint64 = 4
	suiteRelance   uint64 = 5
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

// NextSeed rend la graine de la run suivante, dérivée de celle qui s'achève.
//
// Deux exigences se tiennent, et la dérivation est ce qui les concilie : une
// relance ne rejoue pas la run précédente, sinon mourir deux fois montrerait deux
// fois la même chose ; et la suite des runs d'une session reste rejouable, parce
// que tout y descend de la graine de départ. Tirer la suivante de l'horloge
// donnerait la première sans la seconde, et l'invariant du déterminisme la
// proscrit de toute façon.
//
// Elle prend sa propre suite PCG. Puiser dans un flux de partie ferait dépendre
// la graine suivante de ce que celle-ci a tiré, donc du trajet du joueur : deux
// morts au même endroit rapprocheraient les runs qui les suivent.
func NextSeed(graine uint64) uint64 {
	return rand.NewPCG(graine, suiteRelance).Uint64()
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
