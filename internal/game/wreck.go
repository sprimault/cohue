// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'épave d'une caisse cédée : la trace qu'elle laisse au sol, et rien d'autre.
// La simulation la pose et ne l'interroge plus.

package game

// Wreck est ce qu'une caisse laisse en cédant.
//
// **Une nature à part, et non une caisse à drapeau.** C'est le raisonnement qui
// a déjà donné son bassin au cadavre : un type qui signifierait deux choses
// selon un booléen obligerait chaque boucle écrite ensuite à demander « est-elle
// cassée ? », et l'oubli se paierait dans une boucle qui n'existe pas encore.
// Une épave ne se casse pas, ne coûte rien à traverser, ne se vise pas et ne
// compte nulle part.
//
// **Elle ne s'efface pas non plus**, à la différence d'un effet bref : ce
// qu'elle dit est qu'une caisse a été ouverte ici, et cette information reste
// vraie jusqu'à la fin de la run. C'est aussi ce qui borne son bassin sans
// qu'aucun plafond nouveau soit écrit — il y a exactement autant d'épaves
// possibles que le semis du lieu compte de caisses, et chaque épave remplace
// celle qui l'a produite.
//
// **Rien ici n'entre dans l'empreinte d'une run.** Le bassin est entièrement
// cosmétique, comme celui des effets : une run simulée sans rendu peut ne pas
// l'alimenter sans que rien ne diverge.
type Wreck struct {
	// X et Y sont le point où la caisse se tenait.
	X, Y Fixed
	// Born est le tick où elle a cédé.
	//
	// **Une date et non un décompte**, à l'inverse de ce que porte un effet. Le
	// cycle de rupture s'achève sur l'épave et l'épave demeure : il n'y a donc
	// rien qui finisse, donc rien à décompter, et le rendu tire l'avancement de
	// l'âge comme il le fait déjà pour l'extinction d'une gemme.
	Born Tick
}

// Wrecks rend le bassin des épaves, que le rendu parcourt.
func (w *World) Wrecks() *Pool[Wreck] { return w.epaves }

// WreckAge rend le nombre de ticks écoulés depuis qu'une caisse a cédé.
func (w *World) WreckAge(e *Wreck) Tick { return w.tick - e.Born }
