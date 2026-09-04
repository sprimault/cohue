// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La courbe de pression d'un lieu, telle qu'un auteur l'écrit et telle que le
// spawner la lit : des phases datées sur une frise, un budget par seconde, et
// les profils qu'elles autorisent. La compilation résout les noms en index et
// les instants en ticks, une seule fois.

package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sprimault/cohue/internal/manifest"
)

// WaveScenario est le scénario tel qu'il s'écrit dans le lieu.
//
// **Il est exporté et vit ici, alors que c'est `internal/level` qui décode le
// fichier.** Le scénario est une table de jeu — des profils, un budget, des
// instants — et son sens appartient à qui le consomme. Le décrire dans le
// paquet du chargeur en aurait fait une seconde description, à tenir d'accord
// avec celle-ci ; `internal/level` importe déjà `internal/game`, si bien que le
// porter ici ne coûte aucune dépendance nouvelle.
type WaveScenario struct {
	manifest.Commentable
	// Phases sont les paliers de la courbe, par instant croissant.
	Phases []WavePhase `json:"phases"`
}

// WavePhase est un palier de la courbe, tel qu'il s'écrit.
type WavePhase struct {
	manifest.Commentable
	// Start est l'instant où la phase prend effet, sur la frise « m:ss ».
	Start string `json:"debut"`
	// Pressure est le budget accordé par seconde.
	Pressure *int `json:"pression"`
	// Profiles nomme les profils que le spawner peut acheter dans cette phase.
	Profiles []string `json:"profils"`
	// Peak est la pointe de la phase, absente le plus souvent.
	Peak *WavePeak `json:"pic,omitempty"`
	// Toughness multiplie la résistance des créatures de la phase, un par défaut.
	//
	// **Fractionnaire, et pas un entier.** La première valeur entière au-dessus
	// de un est deux : toute progression de dureté commencerait par un Badaud à
	// six touches au lieu de trois, ce qui n'est pas un palier mais un mur. Un
	// virgule trois en donne quatre, qui est le pas naturel.
	Toughness *float64 `json:"resistance,omitempty"`
}

// WavePeak est une pointe de pression à l'intérieur d'une phase.
type WavePeak struct {
	manifest.Commentable
	// At est l'instant où la pointe commence, sur la même frise.
	At string `json:"a"`
	// Multiplier est ce par quoi le budget est multiplié pendant la pointe.
	Multiplier *int `json:"multiplicateur"`
	// Seconds est la durée de la pointe, en secondes.
	Seconds *int `json:"duree_s"`
}

// Scenario est la courbe de pression compilée : ce que le spawner dépense, et
// parmi quels profils.
//
// **Un scénario exprime un débit, jamais des compteurs.** Un fichier qui dirait
// « à 7 min, 120 marcheurs » rendrait injouable ou vide le premier lieu venu
// d'un auteur tiers, dont on ne connaît ni l'aire ouverte ni la géométrie. Le
// spawner achète dans un budget, ce qui reste cohérent quelle que soit la salle.
type Scenario struct {
	// Phases sont les paliers, par instant croissant. Une tranche vide est un
	// lieu sans horde, ce que le format admet : une salle de passage ou une
	// boutique n'a rien à acheter.
	Phases []Phase
}

// Phase est un palier compilé de la courbe.
type Phase struct {
	// Start est le tick où la phase prend effet.
	Start Tick
	// Pressure est le budget accordé à chaque tick.
	Pressure Fixed
	// Profiles sont les index des profils achetables, dans `Profiles.Enemies`.
	Profiles []int
	// Peak est la pointe de la phase.
	Peak Peak
	// Toughness multiplie la résistance de ce que la phase fait apparaître.
	//
	// **Il s'applique à l'apparition et jamais après**, ce qui est ce qui garde
	// la résistance en touches. Une créature qui durcirait pendant qu'on la
	// frappe demanderait plus de coups qu'au coup précédent : « trois touches »
	// cesserait d'être une unité pour redevenir un nombre, et le joueur ne
	// pourrait plus compter.
	Toughness Fixed
	// Cheapest est le prix de l'apparition la moins chère de la phase, meute
	// comprise.
	//
	// **Il ne sert qu'à empêcher la borne de report de tuer la phase.** Une
	// pression d'un par seconde et un report de trois secondes plafonnent le
	// budget à trois, c'est-à-dire au prix exact d'un Badaud — et l'arrondi de la
	// conversion par tick le place un millième en dessous. La phase n'achète alors
	// jamais rien, sans qu'aucun refus ne le dise : le budget monte, bute sur la
	// borne, et redescend jamais. Le cas est arrivé en réglant la courbe.
	Cheapest Fixed
}

// Peak est une pointe compilée.
//
// **Elle porte un budget et non un multiplicateur.** Le fichier écrit un
// multiplicateur, qui se lit bien ; la multiplication se fait une fois au
// chargement, ce qui évite au tick une opération par image et retire du même
// coup une valeur d'origine extérieure du chemin de l'arithmétique en virgule
// fixe. Une phase sans pointe a une fenêtre vide, si bien que ce budget n'est
// jamais lu et n'est pas un cas particulier à écrire.
type Peak struct {
	// At et Until bornent la fenêtre, `Until` exclu.
	At, Until Tick
	// Pressure est le budget accordé à chaque tick de la fenêtre.
	Pressure Fixed
}

// budget rend ce que la phase accorde au tick donné, pointe comprise.
func (p *Phase) budget(t Tick) Fixed {
	if t >= p.Peak.At && t < p.Peak.Until {
		return p.Peak.Pressure
	}
	return p.Pressure
}

// phase rend le palier en vigueur au tick donné.
//
// Le dernier dont l'instant est passé, et non le premier à venir : une phase
// vaut jusqu'à ce que la suivante la remplace, ce qui donne au dernier palier une
// durée illimitée sans qu'on ait à lui écrire une fin.
func (s *Scenario) phase(t Tick) *Phase {
	courante := &s.Phases[0]
	for i := range s.Phases {
		if s.Phases[i].Start > t {
			break
		}
		courante = &s.Phases[i]
	}
	return courante
}

// durcissement rend le multiplicateur de résistance en vigueur.
//
// Un lieu sans horde n'en a aucun, et `phase` ne saurait pas quoi rendre sur une
// courbe vide : la garde vit ici plutôt que chez elle, parce que c'est le seul
// appelant qui puisse être interrogé hors d'un tick d'achat — `SpawnEnemy` est
// exportée, et les cas qui posent une créature à la main n'ont pas de scénario.
func (w *World) durcissement() Fixed {
	if len(w.scenario.Phases) == 0 {
		return One
	}
	return w.scenario.phase(w.tick).Toughness
}

// CompileScenario résout un scénario écrit contre la table des profils.
//
// Elle rend tout ce qui l'empêche de valoir, plutôt que le premier écart :
// l'appelant les joint aux manquements du lieu, et qui met au point un niveau
// veut la liste.
//
// Un scénario absent — aucune phase — est admis et ne produit aucune apparition.
// Le refuser ferait d'une salle de passage une erreur de fichier.
func CompileScenario(brut WaveScenario, profils *Profiles, report Tick) (*Scenario, []string) {
	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	scenario := &Scenario{Phases: make([]Phase, 0, len(brut.Phases))}
	precedent := Tick(-1)
	for i, p := range brut.Phases {
		ou := fmt.Sprintf("vagues.phases[%d]", i)
		var phase Phase

		debut, err := instant(p.Start)
		if err != nil {
			dire("%s.debut : %v", ou, err)
		} else {
			phase.Start = debut
			if debut <= precedent {
				// Sans l'ordre, `phase` rendrait le dernier palier dont
				// l'instant est passé et non celui qu'on croit lire ; l'auteur
				// verrait une courbe qui saute en arrière.
				dire("%s.debut : « %s » ne vient pas après la phase précédente", ou, p.Start)
			}
			precedent = debut
		}
		if i == 0 && debut != 0 {
			// La première phase date le début de la partie : sans elle, les
			// premières secondes n'auraient aucun palier en vigueur.
			dire("%s.debut : « %s », la première phase commence à 0:00", ou, p.Start)
		}

		pression := exige(ou, "pression", p.Pressure, dire)
		phase.Pressure = parTick(float64(pression))
		if p.Pressure != nil && pression < 1 {
			dire("%s.pression : %d, une phase sans budget n'achète rien", ou, pression)
		}

		// Le durcissement absent vaut un : une phase qui ne dit rien laisse les
		// profils à la résistance de leur table, et c'est ce qui rend le champ
		// facultatif sans lui donner de valeur d'absence à distinguer.
		phase.Toughness = One
		if p.Toughness != nil {
			phase.Toughness = FromFloat(*p.Toughness)
			if phase.Toughness < One {
				dire("%s.resistance : %v, une courbe durcit et n'adoucit pas — un "+
					"profil d'une touche y tomberait à zéro", ou, *p.Toughness)
			}
		}

		phase.Profiles = profilsAutorises(ou, p.Profiles, profils, dire)
		for _, rang := range phase.Profiles {
			if prix := profils.Enemies[rang].PackCost(); phase.Cheapest == 0 ||
				prix < phase.Cheapest {
				phase.Cheapest = prix
			}
		}
		payables(ou, &phase, profils, pression, report, dire)
		phase.Peak = pointe(ou, p.Peak, pression, dire)
		scenario.Phases = append(scenario.Phases, phase)
	}
	return scenario, manques
}

// payables refuse les profils qu'une phase autorise sans jamais pouvoir les
// payer.
//
// **C'est un réglage mort et parfaitement silencieux.** Le budget d'une phase
// s'accumule jusqu'à son plafond de report et pas au-delà : un profil dont
// l'apparition coûte davantage n'est jamais acheté, alors qu'il est écrit dans le
// fichier. L'auteur voit un nom dans sa courbe et ne voit jamais la créature ; le
// spawner, lui, n'a rien à signaler — il achète ce qu'il peut, ce qui est son
// travail.
//
// `Cheapest` ne le ferme pas, et c'est ce qui rend le cas surprenant : il retient
// le prix **le moins cher** de la phase, si bien qu'un Badaud à trois satisfait
// la garde et laisse un Molosse à douze inatteignable dans la même phase.
//
// **Le message nomme les trois nombres**, parce que c'est leur relation qui est
// fausse et non l'un d'eux : le prix, le plafond, et la pression qui le produit.
// Avec les trois, l'auteur choisit de monter la pression, d'allonger le report ou
// de déplacer le profil ; avec le seul refus, il ne sait pas quoi toucher.
//
// La pression de base sert de référence, jamais celle d'une pointe : un profil
// qui n'apparaîtrait que pendant vingt-cinq secondes de pointe serait presque
// aussi muet, et rien dans le format ne dit qu'on peut réserver un profil à une
// pointe.
func payables(ou string, phase *Phase, profils *Profiles, pression int, report Tick,
	dire func(string, ...any)) {
	plafond := PlafondDeReport(phase.Pressure, report, phase.Cheapest)
	for _, rang := range phase.Profiles {
		profil := &profils.Enemies[rang]
		prix := profil.PackCost()
		if prix <= plafond {
			continue
		}
		dire("%s.profils : « %s » coûte %d par apparition, quand une pression de %d "+
			"par seconde reportée sur %d s plafonne le budget à %s — il "+
			"n'apparaîtrait jamais. Monter la pression, allonger le report, ou "+
			"déplacer ce profil vers une phase plus dense",
			ou, profil.Key, profil.PressureCost*profil.Group, pression,
			int(report)/TPS, tronque(plafond))
	}
}

// tronque écrit un budget en unités de jeu, arrondi vers le bas au centième.
//
// **Vers le bas et non au plus proche**, parce que le cas qui a motivé ce message
// se joue à un cheveu : une pression de quatre reportée sur trois secondes donne
// un plafond de 11,9998 pour une meute qui coûte douze, la conversion par tick le
// plaçant un millième en dessous. Arrondi au plus proche, le message dirait
// « plafond 12,00 » face à « coûte 12 », et l'auteur chercherait longtemps.
func tronque(f Fixed) string {
	centiemes := int64(f) * 100 / int64(One)
	return fmt.Sprintf("%d,%02d", centiemes/100, centiemes%100)
}

// profilsAutorises résout les noms d'une phase en index de la table des profils.
//
// Un nom inconnu est refusé plutôt qu'ignoré : une faute de frappe donnerait
// sinon une phase muette, et l'auteur chercherait longtemps pourquoi sa vague
// n'arrive pas.
func profilsAutorises(ou string, noms []string, profils *Profiles,
	dire func(string, ...any)) []int {
	if len(noms) == 0 {
		dire("%s.profils : une phase qui n'autorise aucun profil n'achète rien", ou)
		return nil
	}

	index := make([]int, 0, len(noms))
	for _, nom := range noms {
		rang := -1
		for i := range profils.Enemies {
			if profils.Enemies[i].Key == nom {
				rang = i
				break
			}
		}
		if rang < 0 {
			// Le Passant n'est pas dans cette table : son rôle est `ambiance`,
			// et ce qui n'est pas hostile n'entre dans aucun compte.
			dire("%s.profils : « %s » n'est pas un profil d'ennemi, attendu %s",
				ou, nom, listeDesEnnemis(profils))
			continue
		}
		index = append(index, rang)
	}
	return index
}

// pointe compile la pointe d'une phase, absente ou non.
//
// Absente, elle rend une fenêtre vide : `Phase.budget` la traverse sans avoir à
// distinguer les deux cas.
func pointe(ou string, brute *WavePeak, pression int, dire func(string, ...any)) Peak {
	if brute == nil {
		return Peak{}
	}

	multiplicateur := exige(ou, "pic.multiplicateur", brute.Multiplier, dire)
	if brute.Multiplier != nil && (multiplicateur < 2 || multiplicateur > 10) {
		// À un, la pointe ne fait rien et se lit pourtant comme un événement ; à
		// zéro ou moins, elle serait une accalmie déguisée en pointe. Au-delà de
		// dix, ce n'est plus une pointe mais une autre phase, et le plafond
		// d'effectif absorberait tout ce qu'elle achète.
		dire("%s.pic.multiplicateur : %d, une pointe multiplie entre deux et dix fois",
			ou, multiplicateur)
	}

	debut, err := instant(brute.At)
	if err != nil {
		dire("%s.pic.a : %v", ou, err)
	}

	secondes := exige(ou, "pic.duree_s", brute.Seconds, dire)
	if brute.Seconds != nil && secondes < 1 {
		dire("%s.pic.duree_s : %d, une pointe dure au moins une seconde", ou, secondes)
	}
	duree, err := TicksFromSeconds(secondes)
	if err != nil {
		dire("%s.pic.duree_s : %v", ou, err)
	}

	return Peak{At: debut, Until: debut + duree, Pressure: parTick(float64(pression * multiplicateur))}
}

// instant convertit un point de la frise d'un scénario en ticks.
//
// **La frise est la seule durée d'un manifeste qui ne s'écrit pas en
// millisecondes**, et la raison tient à qui l'écrit : ce n'est pas une cadence
// de mécanisme sortie d'un générateur, c'est un déroulé que son auteur relit
// comme une minuterie. `"2:10"` se place sur une courbe de quinze minutes,
// `130000` non.
//
// **Elle est stricte au point de refuser ce qui se laisserait interpréter** :
// `0:60` n'est pas une minute, `1:5` n'est pas cinq secondes, et un fichier qui
// les porte est faux. Les accepter reviendrait à deviner, sur un chiffre qui
// décide de tout le rythme d'un lieu.
func instant(frise string) (Tick, error) {
	minutes, secondes, coupe := strings.Cut(frise, ":")
	if !coupe {
		return 0, fmt.Errorf("« %s » : un instant s'ecrit m:ss", frise)
	}
	if len(minutes) < 1 || len(minutes) > 2 || !chiffres(minutes) {
		return 0, fmt.Errorf("« %s » : les minutes s'ecrivent avec un ou deux chiffres", frise)
	}
	if len(secondes) != 2 || !chiffres(secondes) {
		return 0, fmt.Errorf("« %s » : les secondes s'ecrivent avec deux chiffres", frise)
	}

	m, _ := strconv.Atoi(minutes)
	s, _ := strconv.Atoi(secondes)
	if s > 59 {
		return 0, fmt.Errorf("« %s » : %d secondes, une minute en compte soixante", frise, s)
	}
	return TicksFromSeconds(m*60 + s)
}

// chiffres dit si une chaîne n'est faite que de chiffres décimaux.
//
// `strconv.Atoi` accepte un signe, et `+1:00` comme `-1:00` se seraient chargés
// en donnant une frise qui recule.
func chiffres(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
