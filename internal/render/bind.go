// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce qui relie une entité de la simulation à celle de la scène : une poignée par
// identifiant de bassin, et le solde des morts à la fin de chaque image.

package render

import (
	"github.com/sprimault/lozengine/scene"

	"github.com/sprimault/cohue/internal/game"
)

// attache tient la poignée de scène de chaque entité d'un bassin.
//
// **Indexée par l'identifiant du bassin, et jamais par la place.** C'est ce qui
// rend l'échange à la suppression invisible ici : une entité qui change de place
// garde son identifiant, donc sa poignée reste à la même entrée. Indexer par la
// place aurait obligé le rendu à suivre des échanges que la simulation ne lui
// montre pas, c'est-à-dire à savoir ce que `docs/go.md` lui interdit de savoir.
//
// **L'identifiant est borné par la capacité du bassin**, qui le recycle dans
// `[0, Cap())` : une tranche de cette longueur suffit, et rien ne s'alloue
// ensuite. Ce n'est pas une propriété qu'on espère — `Pool` indexe déjà sa
// redirection et ses générations dessus.
//
// **Les deux générations ne s'accordent pas, et n'ont pas à le faire.** Celle du
// bassin dit qu'une référence de cohue est périmée, celle du moteur qu'une
// poignée de scène l'est. Cette tranche est le seul lien entre les deux, et il
// n'y a donc pas de troisième indirection à tenir.
type attache struct {
	// poignees[identifiant] désigne l'entité de scène, et sa valeur zéro dit
	// qu'aucune n'est posée — les générations du moteur partant à un.
	poignees []scene.Handle
	// vues[identifiant] est le tick où l'entité a été vue vivante. C'est ce qui
	// permet de retirer les mortes sans que la simulation ait à les signaler :
	// une entité disparue cesse simplement d'être marquée.
	vues []game.Tick
}

// nouvelleAttache prépare la correspondance d'un bassin de la capacité donnée.
//
// Tout est alloué ici et plus rien ne l'est ensuite, comme pour le bassin
// qu'elle double : une tranche qui croîtrait au milieu d'une horde ferait passer
// le ramasse-miettes là où il se voit le plus.
func nouvelleAttache(capacite int) *attache {
	return &attache{
		poignees: make([]scene.Handle, capacite),
		vues:     make([]game.Tick, capacite),
	}
}

// poser range une entité dans la scène, ou met à jour celle qui y est déjà.
//
// **Le même appel pose et déplace**, parce que l'appelant n'a pas à savoir
// laquelle des deux choses il fait : ce qu'il tient est une entité vivante à une
// position, et c'est la poignée déjà présente ou non qui tranche. Lui faire
// porter la distinction l'obligerait à relire l'attache avant chaque pose.
//
// Un refus de la scène — capacité atteinte, couche ou étage hors de ce que
// l'ordre de dessin distingue — laisse l'entité sans poignée, donc absente de
// l'image. C'est ce que le moteur en dit : un refus et non une erreur, puisqu'il
// arrive en cours de partie.
// La poignée rendue est celle sous laquelle la scène connaît l'entité, et zéro
// dit qu'elle n'y est pas. C'est ce que le relevé de recouvrement désigne, et le
// seul endroit où elle se lit sans relire l'attache.
func (a *attache) poser(sc *scene.Scene, id int, e scene.Entity, tick game.Tick) scene.Handle {
	a.vues[id] = tick

	if p := a.poignees[id]; p != 0 {
		sc.Move(p, e.X, e.Y, e.Z)
		sc.SetAspect(p, e.Aspect)
		return p
	}
	if p, pose := sc.Place(e); pose {
		a.poignees[id] = p
		return p
	}
	return 0
}

// poserBassin pose dans la scène les entités vivantes d'un bassin.
//
// **Générique parce que le geste est le même pour douze bassins**, et qu'il ne
// dépend pas de ce qu'ils portent : parcourir les places vivantes, décrire
// chacune, ranger la description sous l'identifiant du bassin. Douze copies
// seraient douze endroits où tenir la correspondance, et une copie qui
// oublierait l'identifiant poserait une créature à la place d'une autre à la
// première mort — sans qu'aucun test ne tombe.
//
// `decrire` rend faux pour ce qui n'a rien à montrer : un cycle que le profil ne
// porte pas, une image absente. L'entité n'est alors pas marquée, donc elle
// quitte la scène au solde — ce qui est juste, puisqu'elle n'y serait pas
// dessinée.
func poserBassin[T any](s *Screen, quoi sorte, p *game.Pool[T],
	decrire func(*T, int) (scene.Entity, trace, bool)) {
	suivi := quoi == sorteTirHorde
	for place := range p.Len() {
		id := p.IDAt(place)
		e, t, montre := decrire(p.At(place), id)
		if !montre {
			continue
		}
		e.Measure = suivi
		if poignee := s.attaches[quoi].poser(s.scene, id, e, s.monde.Tick()); suivi {
			s.retenir(poignee, t, quoi)
		}
	}
}

// attacher prépare une correspondance par bassin, à la capacité de chacun.
//
// **Le décor n'en a pas, et c'est tout ce que sa sorte dit encore.** Ses
// cellules sont posées une fois au montage, donc rien ne les repose par
// identifiant ; garder l'entrée vide coûte un pointeur nul et évite un
// aiguillage que personne ne consulterait.
func (s *Screen) attacher(monde *game.World) {
	s.attaches = [sorteLisere + 1]*attache{
		sorteDecor:      nouvelleAttache(0),
		sorteObstacle:   nouvelleAttache(monde.Breakables().Cap()),
		sorteEpave:      nouvelleAttache(monde.Wrecks().Cap()),
		sorteEnnemi:     nouvelleAttache(monde.Enemies().Cap()),
		sorteAmbiance:   nouvelleAttache(monde.Ambients().Cap()),
		sorteTir:        nouvelleAttache(monde.Shots().Cap()),
		sorteTirHorde:   nouvelleAttache(monde.EnemyShots().Cap()),
		sorteGemme:      nouvelleAttache(monde.Gems().Cap()),
		sorteAimant:     nouvelleAttache(monde.Magnets().Cap()),
		sorteCaisse:     nouvelleAttache(monde.Crates().Cap()),
		sorteArmeAuSol:  nouvelleAttache(monde.Drops().Cap()),
		sorteFioleAuSol: nouvelleAttache(monde.Vials().Cap()),
		sorteTourelle:   nouvelleAttache(monde.Turrets().Cap()),
		sorteEclair:     nouvelleAttache(monde.Enemies().Cap()),
		sorteJoueur:     nouvelleAttache(1),
		sorteLisere:     nouvelleAttache(len(contourAutour)),
	}
}

// solder retire de la scène les entités qui n'ont pas été vues à ce tick.
//
// **Une entité disparue ne se signale pas**, et c'est pourquoi le solde se fait
// par marquage : la simulation supprime par échange, sans rien dire à personne,
// si bien que la seule façon de reconnaître une morte est de constater qu'on ne
// l'a pas posée. Le passage est linéaire sur la capacité et non sur les vivants,
// ce qui est le prix de n'avoir rien à signaler.
//
// L'entrée est remise à zéro en même temps : un identifiant recyclé par le
// bassin doit retrouver une attache vierge, faute de quoi la nouvelle entité
// hériterait de la poignée de la précédente et se dessinerait à sa place.
func (a *attache) solder(sc *scene.Scene, tick game.Tick) {
	for id, p := range a.poignees {
		if p == 0 || a.vues[id] == tick {
			continue
		}
		sc.Remove(p)
		a.poignees[id] = 0
	}
}

// vider retire de la scène tout ce que l'attache tient, et la rend vierge.
//
// **Le solde ne saurait pas le faire**, et c'est pourquoi ce geste existe : il
// reconnaît une morte à ce qu'elle n'a pas été marquée *à ce tick*, or une run
// neuve repart au tick zéro. Une entité posée au premier tick de la précédente
// et jamais revue y porterait la marque zéro, donc le solde la garderait — une
// silhouette abandonnée à la place qu'elle occupait dans une partie finie, et
// rien ne la reprendrait jamais.
//
// Les tranches ne sont pas réallouées : une relance doit tenir dans la seconde
// que la conception lui donne, et c'est tout l'intérêt de relier plutôt que de
// remonter.
func (a *attache) vider(sc *scene.Scene) {
	for id, p := range a.poignees {
		if p != 0 {
			sc.Remove(p)
			a.poignees[id] = 0
		}
		a.vues[id] = 0
	}
}
