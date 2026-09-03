// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le bassin d'entités de capacité fixe, et le Handle qui survit à ses échanges.
// Le mécanisme est le même pour tous les bassins qu'annoncent les invariants,
// d'où le paramètre de type : une copie par sorte ferait autant d'endroits où
// tenir une règle qu'aucun test ne verrait manquer.

package game

// Handle est une référence à une entité qui survit à plusieurs images.
//
// Il porte un identifiant stable, et non la place de l'entité dans le bassin :
// la suppression par échange déplace la dernière entité vers le trou, si bien
// qu'une place ne désigne pas deux fois la même chose. Une référence indexée par
// la place se briserait donc parce qu'une **autre** entité est morte — la cible
// d'un cracheur changerait à chaque mort ailleurs, et le comportement
// deviendrait inexplicable à l'observation.
//
// Les deux champs sont privés : un Handle se fabrique en posant une entité et se
// résout par le bassin, jamais en écrivant deux nombres. Sa valeur zéro n'est
// valide dans aucun bassin, ce qui rend un champ oublié détectable.
type Handle struct {
	id  int
	gen uint32
}

// Pool est un bassin d'entités de capacité fixe.
//
// Le motif est le même pour toutes les sortes d'entités, des ennemis aux
// cadavres, et il est générique pour cette raison : ce n'est pas une abstraction
// bâtie sur un cas, c'est la mise en facteur d'un mécanisme que l'invariant 1
// fixe et qui ne peut pas différer d'un bassin à l'autre. Une copie par sorte
// ferait autant d'endroits où tenir la règle, et une copie qui la manquerait ne
// ferait échouer aucun test : elle ferait qu'une référence périmée désigne une
// entité vivante.
//
// `T` n'a rien à savoir du bassin — ni méthode, ni contrainte, ni champ de
// génération. Une entité qui porterait la sienne devrait l'exposer, et
// quelqu'un finirait par la lire au lieu de passer par le Handle.
type Pool[T any] struct {
	// entities a une capacité fixe et n'est jamais réallouée. Les vivantes
	// occupent [0, active), sans trou.
	entities []T
	active   int

	// La correspondance entre une place et une identité, dans les deux sens.
	// Aucun des trois tableaux ne se compacte : une entité supprimée doit
	// laisser derrière elle une génération que la suivante trouvera incrémentée.
	ids   []int    // place -> identifiant
	slots []int    // identifiant -> place
	gens  []uint32 // identifiant -> génération

	// libres est la pile des identifiants disponibles. Elle est allouée à la
	// capacité du bassin et ne peut pas la dépasser, donc son `append` ne
	// réalloue jamais.
	libres []int
}

// NewPool prépare un bassin qui contiendra au plus capacite entités.
//
// Tout est alloué ici et plus rien ne l'est ensuite : c'est le seul moment où ce
// type touche au tas.
func NewPool[T any](capacite int) *Pool[T] {
	p := &Pool[T]{
		entities: make([]T, capacite),
		ids:      make([]int, capacite),
		slots:    make([]int, capacite),
		gens:     make([]uint32, capacite),
		libres:   make([]int, capacite),
	}
	for i := range p.gens {
		// Les générations partent de 1, ce qui suffit à rendre `Handle{}`
		// invalide partout : sa génération zéro n'égale celle d'aucun
		// identifiant, y compris l'identifiant zéro.
		p.gens[i] = 1
		// Pile dépilée par la fin : les identifiants sortent dans l'ordre
		// croissant, ce qui rend un bassin neuf lisible en test.
		p.libres[i] = capacite - 1 - i
	}
	return p
}

// Cap rend le nombre d'entités que le bassin peut contenir.
func (p *Pool[T]) Cap() int { return len(p.entities) }

// Len rend le nombre d'entités vivantes.
func (p *Pool[T]) Len() int { return p.active }

// Active rend les entités vivantes, dans l'ordre des places.
//
// C'est le parcours de la boucle de mise à jour, et il ne traverse aucune
// redirection : les identifiants ne servent qu'aux références longues. En
// itérant, prendre l'adresse plutôt que la copie — `e := &pool.Active()[i]`.
func (p *Pool[T]) Active() []T { return p.entities[:p.active] }

// At rend l'entité qui occupe une place, pour la lire ou la modifier.
//
// Le pointeur vaut pour l'image en cours et pas au-delà : la première
// suppression venue déplacera ce qu'il désigne. Ce qui doit vivre plus longtemps
// est un Handle.
func (p *Pool[T]) At(place int) *T { return &p.entities[place] }

// IDAt rend l'identifiant de l'entité qui occupe une place.
//
// L'identifiant seul, jamais le Handle : privé de sa génération, il ne résout
// rien et ne permet donc pas de contourner le bassin. C'est délibéré — un Handle
// rendu ici finirait conservé quelque part, et l'invariant tomberait par la porte
// qu'on aurait ouverte pour un tri.
//
// Ce qu'il apporte est la seule propriété que la place n'a pas : il ne bouge
// jamais. La suppression par échange ramène la dernière entité dans le trou, si
// bien que deux entités changent de place sans que rien ne leur soit arrivé. Un
// ordre qui se départagerait sur la place changerait donc parce qu'une troisième
// est morte ailleurs, ce qui se voit à l'écran et se relie très mal à sa cause.
func (p *Pool[T]) IDAt(place int) int { return p.ids[place] }

// Spawn pose une entité et rend la référence qui la désignera.
//
// Le second résultat est faux quand le bassin est plein. Il l'est vraiment :
// rien n'est alloué et aucune entité vivante n'est écrasée, parce qu'une vague
// qui déborde ne doit pas faire disparaître ce qui est déjà à l'écran.
//
// L'entité est passée par valeur plutôt que construite à travers un pointeur
// rendu par le bassin : c'est ce qui garantit qu'aucun pointeur n'en sort au
// moment même où il serait le plus tentant de le garder.
func (p *Pool[T]) Spawn(entite T) (Handle, bool) {
	if p.active == len(p.entities) {
		return Handle{}, false
	}
	id := p.libres[len(p.libres)-1]
	p.libres = p.libres[:len(p.libres)-1]

	place := p.active
	p.entities[place] = entite
	p.ids[place] = id
	p.slots[id] = place
	p.active++
	return Handle{id: id, gen: p.gens[id]}, true
}

// RemoveAt retire l'entité qui occupe une place, en y ramenant la dernière.
//
// La redirection des **deux** entités concernées est mise à jour : celle qui
// part rend son identifiant avec une génération incrémentée, celle qui remonte
// apprend sa nouvelle place. Oublier la seconde ne casserait aucune compilation
// et ne ferait échouer aucun parcours — cela ferait seulement qu'un Handle
// désigne quelqu'un d'autre, ce qui est le défaut le plus difficile à relier à
// sa cause de tout ce que ce bassin peut produire.
func (p *Pool[T]) RemoveAt(place int) {
	dernier := p.active - 1

	mort := p.ids[place]
	p.gens[mort]++
	p.libres = append(p.libres, mort)

	if place != dernier {
		p.entities[place] = p.entities[dernier]
		deplace := p.ids[dernier]
		p.ids[place] = deplace
		p.slots[deplace] = place
	}
	p.active = dernier
}

// Remove retire l'entité que désigne une référence, si elle est encore vivante.
func (p *Pool[T]) Remove(h Handle) bool {
	place, vivante := p.Slot(h)
	if !vivante {
		return false
	}
	p.RemoveAt(place)
	return true
}

// Slot rend la place courante d'une référence, et faux si l'entité est morte.
//
// La place change au fil des suppressions ; c'est pour cela qu'elle se demande à
// chaque usage plutôt que de se conserver.
func (p *Pool[T]) Slot(h Handle) (int, bool) {
	if h.id < 0 || h.id >= len(p.gens) || p.gens[h.id] != h.gen {
		return 0, false
	}
	return p.slots[h.id], true
}

// Alive dit si une référence désigne encore l'entité qu'elle désignait.
func (p *Pool[T]) Alive(h Handle) bool {
	_, vivante := p.Slot(h)
	return vivante
}
