// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la compilation d'un scénario : la frise stricte, les profils qui
// doivent exister et être hostiles, l'ordre des phases, et la salle sans horde
// qui reste un lieu valide.

package game

import (
	"strings"
	"testing"

	"github.com/sprimault/cohue"
)

// profilsLivres rend la table de créatures publiée.
func profilsLivres(t *testing.T) *Profiles {
	t.Helper()
	profils, err := LoadProfiles(cohue.Assets, manifestePersonnages)
	if err != nil {
		t.Fatalf("profils livrés : %v", err)
	}
	return profils
}

// phaseEcrite rend une phase minimale, datée et budgetée.
func phaseEcrite(debut string, pression int, profils ...string) WavePhase {
	return WavePhase{Start: debut, Pressure: &pression, Profiles: profils}
}

// TestLaFriseSeLitStrictement vérifie qu'un instant mal écrit est refusé plutôt
// qu'interprété.
//
// **Deviner serait le pire choix ici** : la frise décide de tout le rythme d'un
// lieu, et `0:60` comme `1:5` ont chacun deux lectures plausibles. Un auteur qui
// se trompe doit l'apprendre au chargement, pas au bout de sept minutes de jeu.
func TestLaFriseSeLitStrictement(t *testing.T) {
	for _, c := range []struct {
		frise string
		veut  Tick
	}{
		{"0:00", 0},
		{"0:45", 45 * TPS},
		{"1:30", 90 * TPS},
		{"15:00", 900 * TPS},
	} {
		a, err := instant(c.frise)
		if err != nil {
			t.Errorf("« %s » refusé : %v", c.frise, err)
		} else if a != c.veut {
			t.Errorf("« %s » vaut %d ticks, attendu %d", c.frise, a, c.veut)
		}
	}

	// `1:005` et `0:0` encadrent la règle des deux chiffres, qui sans elle
	// laisserait passer les deux ; `130` est la même durée sans séparateur, que
	// personne n'a écrite exprès.
	for _, frise := range []string{"0:60", "1:5", "1:005", "0:0", "130", "1:3a",
		"+1:00", "-1:00", "", ":30", "1:"} {
		if a, err := instant(frise); err == nil {
			t.Errorf("« %s » accepté, et vaut %d ticks", frise, a)
		}
	}
}

// TestUnProfilQuiNExistePasEstRefuse vérifie qu'une faute de frappe ne donne pas
// une phase muette.
//
// Ignorer le nom inconnu aurait laissé la phase acheter ce qui reste, ou rien du
// tout, et son auteur aurait cherché longtemps pourquoi sa vague n'arrive pas.
func TestUnProfilQuiNExistePasEstRefuse(t *testing.T) {
	_, manques := CompileScenario(WaveScenario{Phases: []WavePhase{
		phaseEcrite("0:00", 8, "marcheur", "badaud"),
	}}, profilsLivres(t))

	if !contient(manques, "badaud") {
		t.Errorf("le profil inconnu passe : %v", manques)
	}
}

// TestLePassantNEstPasAchetable vérifie que ce qui n'est pas hostile n'entre
// dans aucun compte.
//
// **Le Passant existe dans le manifeste des personnages**, avec le rôle
// `ambiance` : il n'est pas dans la table des ennemis, donc pas dans un budget de
// pression. Sans ce refus, un lieu peuplé de Passants verrait sa pression
// dépensée pour des figurants, et son auteur n'aurait aucun moyen de comprendre
// pourquoi sa salle est calme.
func TestLePassantNEstPasAchetable(t *testing.T) {
	_, manques := CompileScenario(WaveScenario{Phases: []WavePhase{
		phaseEcrite("0:00", 8, "civil"),
	}}, profilsLivres(t))

	if !contient(manques, "civil") {
		t.Errorf("le Passant s'achète : %v", manques)
	}
}

// TestLesPhasesSuiventLaFrise vérifie que l'ordre d'écriture est celui du temps.
//
// Le palier en vigueur est le dernier dont l'instant est passé : une phase
// écrite avant sa devancière ne serait donc jamais en vigueur, et sa courbe
// sauterait en arrière sans que rien ne le dise.
func TestLesPhasesSuiventLaFrise(t *testing.T) {
	_, manques := CompileScenario(WaveScenario{Phases: []WavePhase{
		phaseEcrite("0:00", 4, "marcheur"),
		phaseEcrite("2:00", 8, "marcheur"),
		phaseEcrite("1:00", 12, "marcheur"),
	}}, profilsLivres(t))

	if !contient(manques, "ne vient pas après") {
		t.Errorf("une phase en arrière passe : %v", manques)
	}
}

// TestLaPremierePhaseDateLeDebut vérifie qu'aucune seconde n'est sans palier.
//
// Une courbe qui commencerait à trente secondes laisserait le spawner sans
// budget avant, ce qui se lirait comme une salle vide plutôt que comme un
// fichier incomplet.
func TestLaPremierePhaseDateLeDebut(t *testing.T) {
	_, manques := CompileScenario(WaveScenario{Phases: []WavePhase{
		phaseEcrite("0:30", 4, "marcheur"),
	}}, profilsLivres(t))

	if !contient(manques, "la première phase commence à 0:00") {
		t.Errorf("une courbe qui commence en retard passe : %v", manques)
	}
}

// TestUneSalleSansHordeResteUnLieu vérifie qu'un scénario absent n'est pas une
// erreur de fichier.
//
// Le format doit porter la boutique et le passage aussi bien que l'arène. Ce qui
// garde le lieu livré d'y tomber par mégarde est un cas de conformité, dans
// `internal/level`, et non un refus ici.
func TestUneSalleSansHordeResteUnLieu(t *testing.T) {
	scenario, manques := CompileScenario(WaveScenario{}, profilsLivres(t))
	if len(manques) > 0 {
		t.Fatalf("un lieu sans vagues est refusé : %v", manques)
	}
	if len(scenario.Phases) != 0 {
		t.Errorf("%d phase(s) sorties d'un scénario absent", len(scenario.Phases))
	}
}

// TestLaPointeSeCompileEnFenetre vérifie qu'une durée en secondes devient deux
// bornes de ticks.
//
// La durée est la seule chose du scénario qui ne se lise pas sur la frise, et
// c'est voulu : déplacer une pointe ne doit pas obliger à recalculer sa fin.
func TestLaPointeSeCompileEnFenetre(t *testing.T) {
	multiplicateur, duree := 3, 25
	phase := phaseEcrite("0:00", 8, "marcheur")
	phase.Peak = &WavePeak{At: "2:10", Multiplier: &multiplicateur, Seconds: &duree}

	scenario, manques := CompileScenario(WaveScenario{Phases: []WavePhase{phase}},
		profilsLivres(t))
	if len(manques) > 0 {
		t.Fatalf("pointe refusée : %v", manques)
	}

	pic := scenario.Phases[0].Peak
	if pic.At != 130*TPS || pic.Until != 155*TPS {
		t.Errorf("pointe [%d, %d[, attendu [%d, %d[", pic.At, pic.Until, 130*TPS, 155*TPS)
	}

	// **Le multiplicateur est consommé au chargement**, pas gardé pour le tick :
	// ce que la pointe porte est le budget de vingt-quatre par seconde qu'un
	// multiplicateur de trois donne à une phase de huit.
	if veut := parTick(24); pic.Pressure != veut {
		t.Errorf("budget de pointe %v, attendu %v", pic.Pressure, veut)
	}
}

// contient dit si l'un des manquements porte le fragment donné.
func contient(manques []string, fragment string) bool {
	for _, m := range manques {
		if strings.Contains(m, fragment) {
			return true
		}
	}
	return false
}
