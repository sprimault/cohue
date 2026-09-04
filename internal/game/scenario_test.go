// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas de la compilation d'un scénario : la frise stricte, les profils qui
// doivent exister et être hostiles, l'ordre des phases, le prix d'une meute au
// plancher du report, et la salle sans horde qui reste un lieu valide.

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

// reportDeTest est la borne du budget reporté que ces cas donnent à la
// compilation.
//
// Trois secondes, comme la progression livrée : ce qui se règle ici est le refus
// d'un profil trop cher pour son plafond, et le mesurer contre une borne
// inventée ne dirait rien du lieu réel.
const reportDeTest = 3 * TPS

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
	}}, profilsLivres(t), reportDeTest)

	if !contient(manques, "badaud") {
		t.Errorf("le profil inconnu passe : %v", manques)
	}
}

// TestUnProfilRefuseNommeLaCleEtLeNom vérifie que le refus donne de quoi le
// corriger.
//
// **Le cas est écrit avec le mot qu'un auteur essaie vraiment.** Tout ce qu'un
// humain lit ailleurs porte le nom de fiction — la règle du jeu, la table des
// rôles —, quand le fichier attend la clé du moteur. Un auteur arrive donc avec
// « arpenteur », et un refus qui se contente de le déclarer inconnu le laisse
// chercher `flanqueur` dans un manifeste qu'il n'a aucune raison d'ouvrir.
//
// Les deux moitiés sont exigées ensemble : la clé seule ne se relie à rien de ce
// qu'il a lu, le nom seul ne s'écrit pas dans un fichier.
func TestUnProfilRefuseNommeLaCleEtLeNom(t *testing.T) {
	_, manques := CompileScenario(WaveScenario{Phases: []WavePhase{
		phaseEcrite("0:00", 8, "arpenteur"),
	}}, profilsLivres(t), reportDeTest)

	for _, attendu := range []string{"« flanqueur »", "(Arpenteur)"} {
		if !contient(manques, attendu) {
			t.Errorf("le refus ne porte pas %s : %v", attendu, manques)
		}
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
	}}, profilsLivres(t), reportDeTest)

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
	}}, profilsLivres(t), reportDeTest)

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
	}}, profilsLivres(t), reportDeTest)

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
	scenario, manques := CompileScenario(WaveScenario{}, profilsLivres(t), reportDeTest)
	if len(manques) > 0 {
		t.Fatalf("un lieu sans vagues est refusé : %v", manques)
	}
	if len(scenario.Phases) != 0 {
		t.Errorf("%d phase(s) sorties d'un scénario absent", len(scenario.Phases))
	}
}

// TestUnProfilTropCherPourSaPhaseEstRefuse garde le silence le plus coûteux de
// la courbe.
//
// Un profil qu'une phase autorise sans pouvoir le payer est écrit dans le
// fichier et n'arrive jamais : le budget monte, bute sur son plafond de report,
// et le spawner n'a rien à signaler puisqu'il achète ce qu'il peut. C'est le
// défaut qui a fait qu'un Molosse placé à la première minute du lieu livré
// n'apparaissait dans aucune run de dix minutes.
//
// **Le second cas est le plus fin, et il est déjà décrit dans la godoc de
// `Cheapest`** : une pression de quatre reportée sur trois secondes donne un
// plafond de 11,9998 en virgule fixe pour une meute qui en coûte douze — la
// conversion par tick le place un millième en dessous. Personne ne retrouverait
// ce cas par hasard, et c'est celui qui décide si le refus se fait au bon
// endroit.
//
// **Le Badaud accompagne le Molosse dans les deux cas, et ce n'est pas un
// décor** : `Cheapest` relève le plafond au prix du moins cher, si bien qu'un
// Molosse seul serait sauvé par son propre prix. C'est la présence d'un profil
// bon marché qui laisse le cher inatteignable, et c'est ce qui rend le défaut
// invisible.
func TestUnProfilTropCherPourSaPhaseEstRefuse(t *testing.T) {
	for _, c := range []struct {
		nom      string
		pression int
	}{
		{"franchement au-dessus", 3},
		{"un millième au-dessus, par l'arrondi du tick", 4},
	} {
		t.Run(c.nom, func(t *testing.T) {
			_, manques := CompileScenario(WaveScenario{Phases: []WavePhase{
				phaseEcrite("0:00", c.pression, "marcheur", "sprinteur"),
			}}, profilsLivres(t), reportDeTest)

			if !contient(manques, "n'apparaîtrait jamais") {
				t.Errorf("un profil impayable passe : %v", manques)
			}
			if !contient(manques, "sprinteur") {
				t.Errorf("le refus ne nomme pas le profil fautif : %v", manques)
			}
		})
	}
}

// TestLePlancherDeReportCompteLaMeute vérifie que le prix retenu pour la phase
// est celui d'une apparition et non d'une créature.
//
// C'est le second endroit où le prix intervient, et l'oublier ne se voit pas :
// une phase qui ne convoque que le Molosse plafonnerait son report au prix d'un
// chien, si bien que le budget monterait, buterait sous les douze de la meute et
// n'achèterait jamais rien — la salle resterait vide sans qu'aucun refus ne le
// dise, ce qui est exactement le défaut que `Cheapest` existe pour fermer.
func TestLePlancherDeReportCompteLaMeute(t *testing.T) {
	profils := profilsLivres(t)
	scenario, manques := CompileScenario(WaveScenario{Phases: []WavePhase{
		phaseEcrite("0:00", 4, "sprinteur"),
	}}, profils, reportDeTest)
	if len(manques) > 0 {
		t.Fatalf("phase refusée : %v", manques)
	}

	// L'attendu se recompose ici plutôt que d'appeler `PackCost` : adossé à la
	// méthode, ce cas resterait vert le jour où elle-même rendrait le prix d'une
	// créature, et ne garderait plus que l'accord de deux erreurs.
	molosse := &profils.Enemies[indexDuProfil(t, profils, "sprinteur")]
	veut := FromInt(molosse.PressureCost * molosse.Group)
	if scenario.Phases[0].Cheapest != veut {
		t.Errorf("prix de la phase %v, attendu %v : le plancher retient le prix "+
			"unitaire", scenario.Phases[0].Cheapest, veut)
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
		profilsLivres(t), reportDeTest)
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
