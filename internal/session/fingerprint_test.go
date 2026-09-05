// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le test qui vend le projet : une graine, une run jouée sans rendu, trois
// instants comparés à un attendu versionné — et le drapeau qui le met à jour.

package session

import (
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
)

// majAttendus réécrit les fichiers de référence au lieu de les comparer.
//
// **Jamais automatique, et c'est la règle et non une précaution.** Un attendu
// régénéré parce qu'il ne correspond plus ne teste plus rien : il enregistre ce
// que le code fait, ce qui est vrai par construction. La mise à jour est donc un
// geste que quelqu'un demande, et qui s'accompagne d'une relecture du diff.
var majAttendus = flag.Bool("maj-attendus", false,
	"réécrit les fichiers de référence au lieu de les comparer")

// enteteAttendu est la mention de licence que porte le fichier de référence.
//
// Elle fait partie de ce qui est comparé plutôt que d'être retirée avant :
// l'écrire d'un côté et l'ignorer de l'autre serait deux descriptions du même
// fichier. `make entetes` l'exige de tout ce que git publie, données de test
// comprises — un contrôle qui saute ce que personne d'autre ne regarde ne
// signale plus l'écart, il le certifie.
const enteteAttendu = "# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>\n" +
	"# SPDX-License-Identifier: Apache-2.0\n\n"

// instantanes sont les instants où l'empreinte est prise, et ce qui les a fait
// choisir.
//
// **Des comptes fixes et non des états atteints**, à l'inverse de ce qu'on exige
// d'une planche de relecture. Une planche montre un état de jeu, et s'arrêter à
// un compte de pas y approxime l'événement ; ce test-ci ne montre rien, il
// compare deux exécutions entre elles. S'arrêter « à la mort » ferait dépendre la
// longueur de la run de la courbe de pression, et l'attendu bougerait au premier
// réglage d'équilibrage — pour une raison sans rapport avec le déterminisme, qui
// est la seule chose gardée ici.
//
// **Trois instants et non un seul, pour ce que le fichier montre.** Un état final
// unique ne portait ni gemme ni projectile — au tick 1800 il n'y en a aucun —, et
// un lecteur conclut du contenu au format : neuf lignes sans gemme se lisent
// « les gemmes ne sont pas dans l'empreinte ». Celui qui modifierait le butin ne
// s'attendrait alors pas à voir l'attendu bouger, donc il le régénérerait sans le
// relire. Ils gardent aussi ce qu'un instant final ne peut pas voir : une
// divergence qui apparaît puis se résorbe.
//
// **Les ticks sont choisis pour ce qu'ils contiennent, jamais pour être ronds.**
// Les ramener à 800, 2600 et 2700 perdrait exactement ce pour quoi ils sont là.
//
// **Ils se rechoisissent quand le pilote change**, et c'est arrivé : les trois
// premiers avaient été relevés sur une trajectoire constante, et le tour de
// l'octogone les a vidés de ce qu'ils décrivaient — le troisième annonçait un
// niveau 2 que la nouvelle run n'avait pas encore atteint. Régénérer l'attendu
// sans rouvrir cette table aurait laissé trois phrases fausses sous des chiffres
// justes, c'est-à-dire la pire des deux moitiés.
//
// **Et un instant se décrit depuis l'attendu, jamais depuis une exploration.**
// `Fingerprint` consomme trois tirages à chacun de ces instants, si bien qu'une
// course qui ne l'appelle pas joue une autre run : le quatrième instant y montrait
// six Molosses en charge, et n'en porte aucun ici. La déviation est connue dans
// l'autre sens — une exploration qui relève à chaque pas —, elle vaut aussi pour
// celle qui ne relève jamais.
var instantanes = []struct {
	tick     int
	pourquoi string
}{
	{780, "deux projectiles en vol et deux gemmes au sol, que nul autre instant ne montre ensemble"},
	{2700, "un aimant au sol, le niveau 2 franchi, la vie entamée et dix créatures vivantes"},
	{3163, "quatre gemmes et une horde de onze, le plus que ces trois instants montrent"},
	{15987, "les quatre profils que la run atteint — trois Vigiles, quatorze " +
		"Arpenteurs et onze Molosses dans une horde de 135 —, avec le niveau 6"},
}

// jouerLaRun monte la partie livrée sur une graine et rend l'empreinte des trois
// instants.
//
// Le déplacement vient de `Pilot`, qui ne lit rien du monde : ce que le test
// garde est le déterminisme, pas une trajectoire intéressante, et une entrée qui
// ne dépend que du tick se relit sans se demander ce qu'elle valait au 412e.
//
// **Elle était constante et menait au mur.** Le joueur mourait vers 2:00, ce qui
// bornait la run bien avant les paliers tardifs de la courbe : allonger les
// instants aurait fait jouer dix minutes à un cadavre. Le tour de l'octogone ne
// fait pas un bon joueur, il fait un joueur qui traverse la courbe.
//
// **Elle dit les profils qu'elle a visités**, ce que la doctrine exige d'une run
// aléatoire : une course ne couvre que ce que sa courbe lui présente, et un
// profil qui n'apparaît qu'au sixième palier n'est visité par aucun instant. Le
// compte se lit dans la sortie du test plutôt que de se déduire d'un attendu de
// deux cents lignes.
func jouerLaRun(t *testing.T, graine uint64) string {
	t.Helper()
	s, err := Open(cohue.Assets, cohue.StartingCampaign, graine)
	if err != nil {
		t.Fatalf("montage de la partie livrée : %v", err)
	}

	var b strings.Builder
	vus := map[int]bool{}
	suivant := 0
	for tick := 1; tick <= instantanes[len(instantanes)-1].tick; tick++ {
		s.World.Step(Pilot(game.Tick(tick)))

		// Relevé sur la horde vivante et non sur les apparitions : c'est ce que
		// les instants peuvent voir, et donc ce que l'attendu peut garder.
		horde := s.World.Enemies()
		for i := range horde.Active() {
			vus[horde.At(i).Profile] = true
		}

		if tick != instantanes[suivant].tick {
			continue
		}
		fmt.Fprintf(&b, "# choisi pour %s\n%s\n", instantanes[suivant].pourquoi,
			s.World.Fingerprint())
		suivant++
	}
	t.Logf("profils visités par la run : %v", slices.Sorted(maps.Keys(vus)))
	return b.String()
}

// TestUneGraineRejoueLaMemeRun est le test qui vend le projet.
//
// **Il tourne sur les trois cibles natives de l'intégration continue**, dont deux
// arm64 : c'est là que se vérifie ce que la virgule fixe achète, et c'est ce qui
// sépare un invariant d'une discipline. Un déterminisme qui ne tiendrait que sur
// la machine qui l'a écrit ne porterait ni le classement par graine, ni le
// partage d'un défi.
//
// L'attendu est versionné et se met à jour par `-maj-attendus`, jamais tout seul.
func TestUneGraineRejoueLaMemeRun(t *testing.T) {
	empreinte := enteteAttendu + jouerLaRun(t, graineDeTest)
	chemin := filepath.Join("testdata", "empreinte.txt")

	if *majAttendus {
		if err := os.WriteFile(chemin, []byte(empreinte), 0o600); err != nil {
			t.Fatalf("écriture de l'attendu : %v", err)
		}
		t.Logf("attendu réécrit : %s", chemin)
		return
	}

	// **L'absence de l'attendu fait échouer, elle ne le crée pas.** Un contrôle
	// privé de son entrée échoue, il ne passe pas : un fichier écrit à la volée
	// enregistrerait ce que le code fait, ce qui est vrai par construction.
	attendu, err := os.ReadFile(chemin)
	if err != nil {
		t.Fatalf("attendu introuvable : %v — le poser avec « go test ./internal/session -maj-attendus », puis le relire", err)
	}
	if string(attendu) != empreinte {
		t.Errorf("l'état des trois instants diffère de l'attendu.\n--- attendu ---\n%s\n--- obtenu ---\n%s",
			attendu, empreinte)
	}
}

// TestDeuxExecutionsDeLaMemeGraineCoincident garde ce que l'attendu ne peut pas.
//
// **L'attendu versionné ne dit rien d'une divergence entre deux exécutions du
// même binaire.** Il compare à un état figé, donc il tombe aussi bien pour un
// changement d'équilibrage voulu que pour un déterminisme cassé, et sa mise à
// jour efface les deux. Celui-ci ne compare qu'à lui-même : il ne peut tomber
// que si la même graine, jouée deux fois dans le même processus, a divergé.
//
// C'est la moitié qui reste vraie le jour où l'attendu est régénéré à tort.
func TestDeuxExecutionsDeLaMemeGraineCoincident(t *testing.T) {
	if a, b := jouerLaRun(t, graineDeTest), jouerLaRun(t, graineDeTest); a != b {
		t.Errorf("deux runs de la graine %d divergent.\n--- première ---\n%s\n--- seconde ---\n%s",
			graineDeTest, a, b)
	}
}
