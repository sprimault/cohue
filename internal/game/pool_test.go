// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Les cas du bassin : la référence qui survit à un échange, celle qui meurt avec
// son entité, la place recyclée qui ne ressuscite personne, la référence zéro
// qui ne désigne rien, le plein qui refuse sans écraser, l'identité qui ne suit
// pas la place, et le budget.

package game

import (
	"slices"
	"testing"
)

// TestHandleSurvitALEchange est le test pour lequel la table de redirection
// existe.
//
// Il reconnaît l'entité à un champ et jamais à sa place : un test qui
// comparerait des places dirait « l'index a changé », ce qui est vrai et normal
// après un échange, et il finirait assoupli. Ce qu'il faut constater est une
// erreur d'identité — la référence d'une entité vivante qui se met à désigner
// quelqu'un d'autre parce qu'une **autre** entité est morte.
//
// L'écriture par la référence est ce qui achève la démonstration : une place
// périmée pointe encore sur une copie que l'échange a laissée derrière lui, et
// la lecture seule s'y laisserait prendre.
func TestHandleSurvitALEchange(t *testing.T) {
	p := NewPool[Enemy](4)
	premier, _ := p.Spawn(Enemy{Profile: 10})
	p.Spawn(Enemy{Profile: 11})
	dernier, _ := p.Spawn(Enemy{Profile: 12})

	p.Remove(premier)

	place, vivante := p.Slot(dernier)
	if !vivante {
		t.Fatal("la référence du dernier est morte, alors que c'est le premier qu'on a retiré")
	}
	if place >= p.Len() {
		t.Fatalf("la référence désigne la place %d, hors des %d vivantes : "+
			"l'échange n'a pas mis à jour la redirection de l'entité qui remonte", place, p.Len())
	}
	if profil := p.At(place).Profile; profil != 12 {
		t.Fatalf("la référence désigne le profil %d, attendu 12", profil)
	}

	// Écrire par la référence, puis relire par le parcours dense : les deux
	// doivent voir la même entité.
	p.At(place).Profile = 99
	vus := profils(p)
	if !slices.Contains(vus, 99) {
		t.Errorf("le parcours voit %v : l'écriture par la référence est allée "+
			"dans une copie que l'échange a laissée derrière lui", vus)
	}
}

// TestReferenceDUneEntiteRetireeMeurt vérifie que la génération invalide bien la
// référence de celle qui part, et d'elle seule.
func TestReferenceDUneEntiteRetireeMeurt(t *testing.T) {
	p := NewPool[Enemy](4)
	premier, _ := p.Spawn(Enemy{Profile: 10})
	second, _ := p.Spawn(Enemy{Profile: 11})

	p.Remove(premier)

	if p.Alive(premier) {
		t.Error("la référence de l'entité retirée est encore valide")
	}
	if !p.Alive(second) {
		t.Error("la référence de l'entité restante est morte avec l'autre")
	}
	if p.Remove(premier) {
		t.Error("retirer deux fois la même entité a réussi la seconde fois")
	}
}

// TestPlaceRecycleeNeRessuscitePas éprouve la génération là où elle sert.
//
// L'entité suivante occupe la place libérée, et sans génération incrémentée la
// référence de la morte la désignerait — une entité qui reviendrait à la vie
// sous une autre identité, ce qu'aucun symptôme ne relierait à sa cause.
func TestPlaceRecycleeNeRessuscitePas(t *testing.T) {
	p := NewPool[Enemy](4)
	morte, _ := p.Spawn(Enemy{Profile: 10})
	p.Remove(morte)

	if _, ok := p.Spawn(Enemy{Profile: 11}); !ok {
		t.Fatal("la place libérée n'a pas été rendue")
	}
	if p.Alive(morte) {
		t.Error("la référence de l'entité retirée désigne celle qui a pris sa place")
	}
}

// TestHandleZeroNestValideDansAucunBassin vérifie que le champ oublié se voit.
//
// La valeur zéro d'une structure est ce qu'on obtient d'un champ qu'on a omis de
// remplir. Si elle désignait la première entité, l'oubli passerait pour un
// ciblage valide.
func TestHandleZeroNestValideDansAucunBassin(t *testing.T) {
	p := NewPool[Enemy](4)
	if p.Alive(Handle{}) {
		t.Error("la référence zéro est valide dans un bassin vide")
	}
	p.Spawn(Enemy{Profile: 10})
	if p.Alive(Handle{}) {
		t.Error("la référence zéro désigne la première entité posée")
	}
}

// TestBassinPleinNecrasePersonne éprouve le refus plutôt que le débordement.
//
// Une vague qui déborde ne doit pas faire disparaître ce qui est déjà à l'écran :
// le spawner apprend qu'il ne peut plus acheter, et dépense ailleurs.
func TestBassinPleinNecrasePersonne(t *testing.T) {
	p := NewPool[Enemy](2)
	p.Spawn(Enemy{Profile: 10})
	garde, _ := p.Spawn(Enemy{Profile: 11})

	if _, ok := p.Spawn(Enemy{Profile: 12}); ok {
		t.Fatal("un bassin plein a accepté une entité de plus")
	}
	if p.Len() != 2 {
		t.Errorf("%d vivantes après le refus, attendu 2", p.Len())
	}
	if !p.Alive(garde) {
		t.Error("le refus a emporté une entité vivante")
	}
	if vus := profils(p); !slices.Contains(vus, 10) || !slices.Contains(vus, 11) {
		t.Errorf("le bassin porte %v, attendu les deux premières", vus)
	}
}

// TestParcoursSansTrou vérifie que les vivantes restent contiguës, quel que soit
// l'ordre des suppressions.
//
// C'est ce que la boucle de mise à jour suppose : elle parcourt une tranche, pas
// une liste d'occupation.
func TestParcoursSansTrou(t *testing.T) {
	p := NewPool[Enemy](5)
	refs := make([]Handle, 5)
	for i := range refs {
		refs[i], _ = p.Spawn(Enemy{Profile: i})
	}

	// Le premier, puis le dernier, puis un du milieu : les trois cas de
	// l'échange, dont celui où la place retirée est déjà la dernière.
	p.Remove(refs[0])
	p.Remove(refs[4])
	p.Remove(refs[2])

	if p.Len() != 2 {
		t.Fatalf("%d vivantes, attendu 2", p.Len())
	}
	vus := profils(p)
	if !slices.Contains(vus, 1) || !slices.Contains(vus, 3) {
		t.Errorf("le bassin porte %v, attendu les profils 1 et 3", vus)
	}
	for _, i := range []int{1, 3} {
		if !p.Alive(refs[i]) {
			t.Errorf("la référence du profil %d est morte", i)
		}
	}
}

// TestLIdentiteNeSuitPasLaPlace garde ce que `IDAt` apporte, et qui n'est
// justement pas la place.
//
// Trois entités, celle du milieu retirée : la dernière remonte dans le trou par
// échange, donc sa place change alors que rien ne lui est arrivé. Son identifiant
// ne bouge pas, et c'est toute la raison d'être de la méthode — un ordre
// d'affichage qui s'y adosse ne change pas parce qu'une autre est morte.
//
// La mutation qu'il attrape est la seule qu'on écrirait : rendre la place, qui
// est juste tant qu'aucune suppression n'a eu lieu.
func TestLIdentiteNeSuitPasLaPlace(t *testing.T) {
	p := NewPool[Enemy](3)
	for i := range 3 {
		p.Spawn(Enemy{Profile: i})
	}

	// L'identité de la dernière, relevée avant qu'elle ne change de place.
	avant := p.IDAt(2)

	p.RemoveAt(1)

	if p.Len() != 2 {
		t.Fatalf("%d vivantes, attendu 2", p.Len())
	}
	if p.At(1).Profile != 2 {
		t.Fatalf("la place 1 porte le profil %d, attendu 2 — l'échange n'a pas eu lieu",
			p.At(1).Profile)
	}
	if apres := p.IDAt(1); apres != avant {
		t.Errorf("identité %d après l'échange, %d avant", apres, avant)
	}
}

// TestBassinNalloueRien est le jumeau de `TestLaBoucleNalloueRien`, à l'échelle
// du bassin.
//
// C'est l'invariant le plus facile à casser sans s'en apercevoir : un `append`
// qui réalloue, une conversion qui échappe. `AllocsPerRun` le dit en un chiffre.
//
// Ce que la boucle ne garde pas : elle ne supprime pas assez pour éprouver
// l'échange et la redirection des identités, qui sont l'endroit où un `append`
// se glisse. Celui-ci remplit le bassin, puis retire, et c'est ce parcours-là
// qu'aucun tick ordinaire ne reproduit.
func TestBassinNalloueRien(t *testing.T) {
	p := NewPool[Enemy](300)
	refs := make([]Handle, 300)

	moyenne := testing.AllocsPerRun(1000, func() {
		for i := range refs {
			refs[i], _ = p.Spawn(Enemy{Profile: i})
		}
		for _, h := range refs {
			p.Remove(h)
		}
	})
	if moyenne != 0 {
		t.Errorf("%v allocation(s) par tour de 300 entités, attendu aucune", moyenne)
	}
}

// profils relève les profils des entités vivantes, dans l'ordre des places.
func profils(p *Pool[Enemy]) []int {
	vus := make([]int, 0, p.Len())
	for i := range p.Active() {
		vus = append(vus, p.At(i).Profile)
	}
	return vus
}
