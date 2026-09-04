// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La caisse : ce qui la casse, et ce qu'elle laisse. Elle est posée par le lieu
// comme les figurants et la porte, et elle n'est pas une cible — c'est le joueur
// qui va la chercher, jamais son arme qui la trouve.

package game

import (
	"fmt"

	"github.com/sprimault/cohue/internal/manifest"
)

// CrateSpec est le semis de caisses tel qu'un lieu l'écrit.
type CrateSpec []CratePlacementSpec

// CratePlacementSpec pose une caisse à une case donnée.
//
// Une position et non un nombre, pour la raison qui a déjà tranché le
// peuplement : un semis tiré au sort abandonne en silence ce qui tombe dans un
// mur, et sauterait d'un coin à l'autre entre deux relances.
type CratePlacementSpec struct {
	manifest.Commentable
	// At est la case où la caisse repose, en coordonnées de lieu.
	At *[2]int `json:"position"`
}

// CratePlacement est une caisse compilée : une position, et rien d'autre.
type CratePlacement struct {
	// X et Y sont sa position, au centre de la case écrite.
	X, Y Fixed
}

// Crate est une caisse posée dans le lieu.
//
// **Elle n'a ni résistance ni état**, et ce n'est pas une simplification : elle
// se casse au premier contact, si bien qu'un compteur de coups n'aurait qu'une
// valeur. Le délai de contact et le ralentissement que la conception lui promet
// appartiennent à l'étape 7, avec le blocage du champ de flux et les
// emplacements de consommables.
type Crate struct {
	// X et Y sont sa position dans le monde, en tuiles.
	X, Y Fixed
}

// CompileCrates résout un semis de caisses contre la carte cuite.
//
// Elle rend tout ce qui l'empêche de valoir plutôt que le premier écart, comme
// les vagues, le peuplement et la sortie.
//
// **Une caisse se pose sur une case franchissable, à l'inverse d'une porte.**
// Elle ne bloque pas le champ de flux — l'étape 7 le lui donnera —, si bien
// qu'une caisse dans un mur serait un objet qu'on voit sans jamais l'atteindre.
func CompileCrates(brut CrateSpec, carte *CostGrid) ([]CratePlacement, []string) {
	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	pose := make([]CratePlacement, 0, len(brut))
	for i, c := range brut {
		ou := fmt.Sprintf("caisses[%d]", i)
		if c.At == nil {
			dire("%s.position : absente, une caisse se place", ou)
			continue
		}

		u, v := c.At[0], c.At[1]
		switch {
		case !carte.InBounds(u, v):
			dire("%s.position : (%d, %d) hors du lieu, qui fait %d sur %d",
				ou, u, v, carte.Width(), carte.Height())
		case !carte.Passable(u, v):
			dire("%s.position : (%d, %d) est dans un mur", ou, u, v)
		default:
			pose = append(pose, CratePlacement{
				X: FromInt(u) + One/2,
				Y: FromInt(v) + One/2,
			})
		}
	}
	return pose, manques
}

// Stock pose les caisses du lieu, au montage et à chaque relance.
//
// Le pendant de `Populate` pour les figurants : les positions viennent du
// fichier et sont déjà vérifiées, cette passe ne décide de rien. Une salle dont
// les caisses resteraient cassées après une mort ne serait pas la même salle.
func (w *World) Stock(caisses []CratePlacement) {
	for _, c := range caisses {
		w.SpawnCrate(c.X, c.Y)
	}
}

// SpawnCrate pose une caisse.
//
// **Par ses coordonnées et non par le placement compilé**, bien que les deux
// portent les mêmes champs — et l'analyseur propose d'ailleurs la conversion.
// Ce qu'il voit est une similitude de champs ; ce qu'il ne voit pas est que
// `CratePlacement` est une donnée d'entrée validée au chargement et `Crate` un
// état de simulation. Elles coïncident par accident, pas par identité.
//
// Les confondre ferait ce que le projet a refusé pour le figurant et refusera
// pour le cadavre : un type qui signifie deux choses selon l'endroit où on le
// lit. Que l'étape 7 donne un délai de contact à l'une et pas à l'autre n'est
// que la date à laquelle l'accident cessera.
func (w *World) SpawnCrate(x, y Fixed) (Handle, bool) {
	return w.caisses.Spawn(Crate{X: x, Y: y})
}

// Crates rend le bassin des caisses.
func (w *World) Crates() *Pool[Crate] { return w.caisses }

// casser vide les caisses que le joueur atteint.
//
// **Le joueur les casse, jamais son arme.** Une caisse rangée parmi les cibles
// détournerait la visée automatique, qui prend la plus proche sans que le joueur
// choisisse : chaque salve partirait vers le décor, et la mécanique du
// Secouriste — dont tout l'intérêt tient à cette visée — s'effondrerait avec.
// C'est la même règle que pour le figurant, et pour la même raison.
//
// Elle vient avec les ramassages, dont elle est le voisin naturel : ce qu'elle
// produit est une volée de gemmes, que la même passe ramasse au tick suivant.
func (w *World) casser() {
	if !w.Alive() {
		return
	}

	portee := w.progression.CrateRange
	for i := 0; i < w.caisses.Len(); {
		c := w.caisses.At(i)
		ecart := Vec{X: w.playerX - c.X, Y: w.playerY - c.Y}
		if ecart.carres() >= int64(portee)*int64(portee) {
			i++
			continue
		}

		// Les gemmes se posent avant la suppression : après, la caisse a quitté
		// le bassin et sa position vient d'une copie qu'on aurait gardée pour
		// rien.
		w.lacherEn(c.X, c.Y, w.progression.CrateGems)
		w.caisses.RemoveAt(i)
	}
}
