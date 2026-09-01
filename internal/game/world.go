// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La partie en cours — bassins, champ de flux, densité, profils, arme — et
// l'ordre d'un tick. Rien n'y est alloué après le montage : la réutilisation des
// tableaux d'un tick à l'autre est ce qui tient le budget d'allocation.

package game

// flowPeriod est le nombre de ticks entre deux calculs du champ de flux.
//
// Le champ ne dépend que du joueur et des obstacles, et six ticks font un
// dixième de seconde : à cinq tuiles par seconde, le joueur y parcourt une
// demi-tuile. Le recalculer à chaque image paierait plein tarif pour une
// information qui n'a pas bougé d'une case. C'est un
// compromis de coût, pas un réglage de jeu — le baisser ne rend pas la horde
// plus maligne, il la rend plus chère.
const flowPeriod Tick = 6

// separationScale convertit la pente de densité en poussée.
//
// La grille rend une pente en entités par cellule ; l'attirance du champ vaut
// une tuile. Sans ce facteur, deux voisines suffiraient à retourner une créature
// contre le joueur. Le poids de séparation du profil multiplie ce vecteur et
// reste un rapport — c'est ici qu'est l'échelle, et c'est un des chiffres qu'on
// cherchera au jalon 3, avec la vitesse relative et le plafond de dégâts.
const separationScale = One / 8

// World est une partie en cours : tout ce qu'un tick lit et modifie.
//
// La struct porte les tableaux plutôt que de les rendre à l'appelant, parce que
// leur réutilisation d'un tick à l'autre est ce qui tient le budget
// d'allocation. Rien n'y est alloué après la construction.
type World struct {
	profils *Profiles
	arme    Weapon
	grille  *CostGrid
	flux    *FlowField
	densite *DensityGrid
	ennemis *Pool[Enemy]
	tirs    *Pool[Projectile]

	playerX, playerY Fixed
	// vie est ce qu'il reste au joueur, en points. À zéro, il est mort — la
	// valeur est l'état, comme la résistance d'une créature.
	vie int
	// degatsSubis est le reste de l'accumulateur de contact, en points-ticks.
	// Il porte ce qui n'a pas encore fait un point entier ; sa raison d'être est
	// dans `subir`.
	degatsSubis int
	// cooldown est ce qui reste à attendre avant le prochain tir. Il descend à
	// zéro et y demeure : sans cible, l'arme ne tire pas et ne consomme rien.
	cooldown Tick
	tick     Tick
}

// NewWorld monte une partie sur une carte et une table de profils.
//
// Les deux capacités sont celles des bassins — les ennemis, puis les
// projectiles — et ne changent plus après le montage. Les plafonds eux-mêmes et
// ce qui les justifie vivent dans `internal/session`, qui monte les parties : ce
// sont des valeurs de jeu, pas un paramètre que chaque appelant choisirait.
func NewWorld(profils *Profiles, arme Weapon, grille *CostGrid, capacite, tirs int) *World {
	return &World{
		profils: profils,
		arme:    arme,
		vie:     profils.Player.Health,
		grille:  grille,
		flux:    NewFlowField(grille),
		densite: NewDensityGrid(grille.Width(), grille.Height()),
		ennemis: NewPool[Enemy](capacite),
		tirs:    NewPool[Projectile](tirs),
	}
}

// Enemies rend le bassin des ennemis, pour le parcourir ou le peupler.
func (w *World) Enemies() *Pool[Enemy] { return w.ennemis }

// Shots rend le bassin des projectiles en vol.
func (w *World) Shots() *Pool[Projectile] { return w.tirs }

// Player rend la position du joueur.
func (w *World) Player() (Fixed, Fixed) { return w.playerX, w.playerY }

// Place pose le joueur, au montage du lieu.
//
// Sans projection ni contrôle de passabilité : c'est au chargeur de savoir où
// commence un lieu, et un point de départ dans un mur est un défaut du niveau,
// pas un cas que la simulation rattrape.
func (w *World) Place(x, y Fixed) {
	w.playerX, w.playerY = x, y
	w.flux.Rebuild(x, y)
}

// SpawnEnemy pose une créature d'un profil donné.
//
// L'index est celui de la table, jamais une copie de ses valeurs. Le second
// résultat est faux quand le bassin est plein.
func (w *World) SpawnEnemy(profil int, x, y Fixed) (Handle, bool) {
	return w.ennemis.Spawn(Enemy{
		Profile: profil,
		X:       x,
		Y:       y,
		Hits:    w.profils.Enemies[profil].Hits,
	})
}

// Step avance la simulation d'un tick, le joueur suivant la direction voulue.
//
// L'ordre est celui de la conception, et il est écrit une fois : les entrées, le
// champ de flux si c'est son tick, la densité, les intentions et leur
// projection, les dégâts de contact, le tir puis le vol des projectiles avec ce
// qu'ils touchent, et les suppressions en dernier. Ce qui manque encore y
// prendra sa place — les apparitions, entre les entrées et le champ.
//
// **Le contact se constate après le déplacement et non avant**, sinon une
// créature qui vient de se coller ne blesserait qu'au tick suivant et le joueur
// verrait la horde le traverser sans effet pendant une image.
//
// **La simulation continue de tourner après la mort**, et `subir` cesse
// seulement d'appliquer des dégâts. Figer le monde ici serait une décision
// d'écran — ce que la mort suspend, ce qu'elle laisse courir — et elle
// appartient à qui l'affichera, pas à la boucle.
//
// Les intentions et la projection tiennent en une seule passe alors que la
// conception les énumère séparément, et c'est équivalent — **à une condition qui
// vaut mieux que la conclusion** : aucune intention ne lit l'état d'une autre
// entité. Aujourd'hui elle ne lit que le champ et la densité, tous deux figés
// avant la passe, si bien que rien de ce qu'une créature déplacée modifie n'est
// lu par les suivantes.
//
// Le jour où une intention devra tenir compte d'une autre créature — un ennemi
// qui évite celui qui le précède —, l'équivalence tombe et il faudra les deux
// passes, donc une tranche d'intentions à préallouer.
func (w *World) Step(voulu Vec) {
	w.deplacerJoueur(voulu)

	if w.tick%flowPeriod == 0 {
		w.flux.Rebuild(w.playerX, w.playerY)
	}
	w.compterDensite()
	w.deplacerEnnemis()
	w.subir()
	w.tirer()
	w.deplacerTirs()
	w.retirerLesMorts()

	w.tick++
}

// Tick rend le numéro du tick courant.
func (w *World) Tick() Tick { return w.tick }

// deplacerJoueur applique la direction voulue, à la vitesse de son profil.
func (w *World) deplacerJoueur(voulu Vec) {
	if voulu == (Vec{}) {
		return
	}
	pas := voulu.Direction(0).Scale(w.vitesse(w.profils.Player.Speed, w.playerX, w.playerY))
	w.playerX, w.playerY = w.glisser(w.playerX, w.playerY, pas)
}

// compterDensite refait le comptage des ennemis, cellule par cellule.
func (w *World) compterDensite() {
	w.densite.Clear()
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		u, v := w.flux.Cell(e.X, e.Y)
		w.densite.Add(u, v)
	}
}

// deplacerEnnemis calcule l'intention de chaque créature et la projette.
//
// L'index de l'entité sert deux fois : il désigne son profil dans la table, et
// il départage les directions quand la somme des forces s'annule.
func (w *World) deplacerEnnemis() {
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		profil := &w.profils.Enemies[e.Profile]

		u, v := w.flux.Cell(e.X, e.Y)
		attirance := w.flux.Direction(u, v)
		repulsion := w.densite.Gradient(u, v).Scale(profil.SeparationWeight).Scale(separationScale)

		voulu := attirance.Sub(repulsion)
		if profil.Tangential != 0 {
			// La dérive latérale d'un flanqueur se prend sur l'attirance et non
			// sur la somme : sur la somme, une créature serrée par ses voisines
			// tournerait autour d'elles au lieu de tourner autour du joueur.
			voulu = voulu.Add(attirance.Perp().Scale(profil.Tangential))
		}

		pas := voulu.Direction(i).Scale(w.vitesse(profil.Speed, e.X, e.Y))
		e.X, e.Y = w.glisser(e.X, e.Y, pas)
	}
}

// vitesse rend la vitesse effective sur la case du point d'appui.
//
// Le coût de la case divise la vitesse, sans quoi le parcours pondéré serait une
// superstition : le champ contournerait au prix de deux cases ce qui ne coûte
// rien à traverser, et l'écart entre le chemin choisi et le chemin payé ne se
// verrait nulle part.
//
// La case est celle d'avant le pas. Diviser par celle d'arrivée serait
// circulaire — le pas plein entre dans la flaque, la division le raccourcit, il
// n'y entre plus, et l'entité tremble à la frontière.
func (w *World) vitesse(base Fixed, x, y Fixed) Fixed {
	cout := w.grille.At(x.Floor(), y.Floor())
	if cout <= Free {
		return base
	}
	return base.Div(FromInt(int(cout)))
}

// glisser applique un pas en annulant ce qui entrerait dans un obstacle.
//
// Les deux axes se traitent séparément : un déplacement qui mène dans un mur
// perd sa composante bloquée et l'entité longe la paroi au lieu de s'y enfoncer.
// Les deux bloqués, elle ne bouge pas.
//
// La poussée de séparation ne décide donc pas de la position finale, et c'est ce
// qui permet à vingt Badauds de pousser un Vigile contre une cloison sans le
// faire passer au travers : ils s'entassent derrière lui, ce qui est tout
// l'intérêt d'un couloir bouché.
func (w *World) glisser(x, y Fixed, pas Vec) (Fixed, Fixed) {
	nx, ny := x+pas.X, y+pas.Y
	if !w.passable(nx, y) {
		nx = x
	}
	if !w.passable(nx, ny) {
		ny = y
	}
	return nx, ny
}

// retirerLesMorts ferme le tick en vidant le bassin de ce qui n'a plus de
// résistance.
//
// Une créature morte a cessé d'être une cible dès l'instant où sa résistance est
// tombée, mais elle est restée en place : c'est ce qui permet aux boucles du tick
// de garder leurs index, et c'est pourquoi les suppressions viennent en dernier.
//
// La place libérée **est** réexaminée ici, à l'inverse de ce que fait une passe
// de mise à jour : celle-ci ferait avancer deux fois l'entité remontée, alors que
// ce nettoyage ne fait que filtrer — la sauter y laisserait un mort jusqu'au tick
// suivant.
func (w *World) retirerLesMorts() {
	for i := 0; i < w.ennemis.Len(); {
		if w.ennemis.At(i).Hits <= 0 {
			w.ennemis.RemoveAt(i)
			continue
		}
		i++
	}
}

// passable dit si une position du monde tombe sur une case franchissable.
func (w *World) passable(x, y Fixed) bool {
	return w.grille.Passable(x.Floor(), y.Floor())
}
