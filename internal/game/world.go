// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package game

// flowPeriod est le nombre de ticks entre deux calculs du champ de flux.
//
// Le champ ne dépend que du joueur et des obstacles, et le joueur ne traverse
// pas une tuile en six centièmes de seconde : le recalculer à chaque image
// paierait plein tarif pour une information qui n'a pas changé. C'est un
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
	grille  *CostGrid
	flux    *FlowField
	densite *DensityGrid
	ennemis *Pool[Enemy]

	playerX, playerY Fixed
	tick             Tick
}

// NewWorld monte une partie sur une carte et une table de profils.
//
// La capacité du bassin est fixée ici et pour de bon : c'est le plafond que le
// spawner rencontrera, et il vaut mieux qu'il le rencontre plutôt que de laisser
// la horde croître jusqu'à ce que l'image s'effondre.
func NewWorld(profils *Profiles, grille *CostGrid, capacite int) *World {
	return &World{
		profils: profils,
		grille:  grille,
		flux:    NewFlowField(grille),
		densite: NewDensityGrid(grille.Width(), grille.Height()),
		ennemis: NewPool[Enemy](capacite),
	}
}

// Enemies rend le bassin des ennemis, pour le parcourir ou le peupler.
func (w *World) Enemies() *Pool[Enemy] { return w.ennemis }

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
	return w.ennemis.Spawn(Enemy{Profile: profil, X: x, Y: y})
}

// Step avance la simulation d'un tick, le joueur suivant la direction voulue.
//
// L'ordre est celui de la conception, et il est écrit une fois : les entrées, le
// champ de flux si c'est son tick, la densité, puis les intentions et leur
// projection. Ce qui manque encore y prendra sa place — les apparitions entre
// les entrées et le champ, les contacts et les suppressions à la fin.
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

// passable dit si une position du monde tombe sur une case franchissable.
func (w *World) passable(x, y Fixed) bool {
	return w.grille.Passable(x.Floor(), y.Floor())
}
