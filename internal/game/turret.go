// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La tourelle : ce qu'une charge pose au sol, et ce qu'elle tire toute seule
// jusqu'à épuisement. La première chose du jeu qui agit sans le joueur.

package game

// Turret est une tourelle posée, et ce qu'il lui reste à tirer.
//
// **Elle n'est visée par personne**, et c'est la règle du figurant et de la
// caisse : la horde ne vise que le joueur, et le tir du joueur ne prend que ce
// qui menace. La rendre destructible demanderait de la ranger parmi les cibles,
// ce qui détournerait la visée automatique — celle-là même dont dépend la
// mécanique du Secouriste.
//
// **Elle ne bloque pas non plus.** Écrire un coût dans la grille en cours de
// partie est interdit : le champ de flux dérive son nombre de seaux au montage,
// et un coût apparu après n'aurait pas de seau où entrer.
type Turret struct {
	// X et Y sont le point où elle a été posée, et d'où elle tire.
	X, Y Fixed
	// Weapon est son rang dans la table, qui porte ce qu'elle tire.
	Weapon int
	// Shots est ce qu'il lui reste à tirer. À zéro, elle disparaît.
	//
	// **Un compte et non une durée** : une tourelle posée dans un couloir vide
	// attend au lieu de s'éteindre pour rien, ce qui est ce qui rend une charge
	// toujours rendue. `Weapon.Shots` porte le motif complet.
	Shots int
	// Cooldown est ce qui reste avant son tir suivant, en ticks.
	Cooldown Tick
}

// Turrets rend le bassin des tourelles posées.
func (w *World) Turrets() *Pool[Turret] { return w.tourelles }

// TurretWeapon rend l'arme qu'une tourelle emploie.
//
// Le rendu en a besoin pour poser la bonne image, comme pour une arme au sol :
// il lit une arme et n'a pas à connaître la table.
func (w *World) TurretWeapon(t *Turret) *Weapon { return &w.armes.All[t.Weapon] }

// poserUneTourelle place une tourelle sous le joueur, et dit si elle tient.
//
// **Sous le joueur et non sur une cible**, à la différence de la déflagration :
// ce qu'une tourelle tient est une position, et le joueur la choisit en s'y
// tenant. C'est la seule décision de placement que le chapitre 9 lui accorde,
// puisqu'il ne dirige rien d'autre.
//
// Le bassin plein refuse, et la charge n'est alors pas dépensée : six tourelles
// debout dans une salle sont déjà six charges employées, et la septième doit se
// voir refusée plutôt que perdue.
func (w *World) poserUneTourelle(arme *Weapon, rang int) bool {
	_, ok := w.tourelles.Spawn(Turret{
		X: w.playerX, Y: w.playerY,
		Weapon: rang,
		Shots:  arme.Shots,
		// Prête au tick suivant : une tourelle qui attendrait sa cadence entière
		// avant son premier tir se poserait inerte au moment précis où on la pose
		// pour parer.
		Cooldown: 0,
	})
	return ok
}

// tirerLesTourelles fait tirer ce que le joueur a posé, et retire ce qui est
// épuisé.
//
// **Elle vient avec le tir du joueur, parce que ce qui tire en son nom tire avec
// lui.** L'ordre entre les deux est sans conséquence, la mort étant un état :
// une créature abattue cesse d'être une cible dans la même passe, si bien que
// deux tirs ne peuvent pas la tuer deux fois et que le second va chercher
// derrière.
//
// **Le retrait se fait dans la même passe et réexamine la place libérée**, à la
// différence de ce qui avance : cette moitié-ci ne fait que trier, et sauter
// l'entité remontée par l'échange y laisserait une tourelle vide un tick de plus.
func (w *World) tirerLesTourelles() {
	for i := 0; i < w.tourelles.Len(); {
		t := w.tourelles.At(i)
		if t.Cooldown > 0 {
			t.Cooldown--
			i++
			continue
		}

		arme := &w.armes.All[t.Weapon]
		cible, trouvee := w.plusProcheDe(t.X, t.Y, arme.Range, Handle{}, Vec{})
		if !trouvee {
			// Rien à portée : le tir n'est pas consommé, comme la cadence de
			// l'arme de base. C'est ce qui fait qu'une tourelle posée d'avance
			// tient toujours ses tirs quand la horde arrive.
			i++
			continue
		}

		e := w.ennemis.At(cible)
		vers := Vec{X: e.X - t.X, Y: e.Y - t.Y}.Direction(i)
		if w.salveDe(arme, t.X, t.Y, vers) == 0 {
			// Bassin de projectiles plein : le tir est perdu comme celui du
			// socle, mais il n'est pas décompté — une tourelle ne doit pas
			// s'épuiser sur ce que le moteur n'a pas su poser.
			i++
			continue
		}

		t.Shots--
		t.Cooldown = arme.Cooldown
		if t.Shots <= 0 {
			w.tourelles.RemoveAt(i)
			continue
		}
		i++
	}
}
