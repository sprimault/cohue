// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le repère : la touche qui horodate un instant de la partie, ce qu'elle écrit
// au journal, et l'accusé qui le confirme sans faire quitter la horde des yeux.

package render

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/game"
)

// repere est la touche qui pose un repère.
//
// **Espace, qui ne fait rien d'autre pendant qu'on joue.** Il valide une carte
// et relance après la mort, deux états où `Update` a rendu la main plus haut :
// les trois usages ne se rencontrent jamais dans le même tick. Le geste doit se
// faire sans regarder le clavier, ce qu'aucune touche de bord ne permet.
var repere = []ebiten.Key{ebiten.KeySpace}

// dureeRepere est le temps que l'accusé reste à l'écran.
//
// Deux secondes : assez pour être lu au prochain moment de répit, trop court
// pour qu'on le prenne pour une lecture permanente du bandeau.
const dureeRepere = 2 * game.TPS

// poserRepere horodate l'instant et dit ce que le joueur avait à ce moment-là.
//
// **Le joueur marque, le jeu horodate**, et la division est ce qui rend la
// mesure utilisable : ce qu'un jalon demande d'observer est un ressenti, que
// personne ne peut détecter à sa place, mais lire un minuteur en pleine horde
// revient à cesser de jouer pour mesurer.
//
// Ce qu'elle écrit va au-delà de l'heure, et c'est là qu'est son intérêt : les
// paliers acquis nomment ce qui a produit ce qu'on vient de ressentir, et ils ne
// se restituent pas de mémoire une fois la partie finie.
//
// Elle alloue, ce qui est admis parce qu'elle part d'un enfoncement de touche et
// non d'une image — la règle du journal, comme celle du budget, vise ce qui
// tombe soixante fois par seconde.
func (s *Screen) poserRepere() {
	s.finRepere = s.monde.Tick() + dureeRepere

	axes := s.monde.Axes()
	pris := s.monde.TiersTaken()
	paliers := make([]string, 0, len(axes))
	for i := range axes {
		paliers = append(paliers, fmt.Sprintf("%s %d/%d", axes[i].Name, pris[i], axes[i].Tiers))
	}

	slog.Info("repère posé",
		"time", minuteur(s.monde.Tick()),
		"level", s.monde.Level(),
		"tiers", strings.Join(paliers, ", "))
}

// marque rend l'accusé à afficher, vide quand il n'y a rien à confirmer.
//
// L'instant se relit dans `finRepere` plutôt que d'être gardé à côté : deux
// champs pour une même pose finiraient par se contredire, et celui-ci suffit à
// retrouver l'autre.
func (s *Screen) marque() string {
	if s.finRepere <= s.monde.Tick() {
		return ""
	}
	return "repère " + minuteur(s.finRepere-dureeRepere)
}
