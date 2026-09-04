// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// L'explosion de la Baudruche : une entité posée à sa mort, une mèche qui brûle
// pendant que le télégraphe l'annonce, puis une détonation qui ne blesse que le
// joueur.

package game

// Blast est une explosion amorcée, qui attend sa détonation.
//
// **Une nature à part, jamais une créature morte qu'on garderait.** La mort est
// la résistance nulle et rien d'autre — c'est ce que le bassin des ennemis
// suppose partout —, si bien qu'y laisser une Baudruche le temps de sa mèche
// obligerait `subir`, `deplacerEnnemis` et le nettoyage à reconnaître chacun un
// état de mort-vivante. Trois passes amendées pour un état que la conception
// refuse, quand une entité distincte n'en demande aucune. C'est ce qui range le
// cadavre à part plutôt qu'en ennemi porteur d'un drapeau.
type Blast struct {
	// X et Y sont le point de la mort, où la déflagration reste centrée. Elle ne
	// suit pas ce qui l'a produite, qui n'existe plus.
	X, Y Fixed
	// Profile est l'index du profil qui l'a laissée, d'où viennent son rayon et
	// ses dégâts. L'index et non les valeurs, comme pour toute entité : régler
	// une Baudruche ne doit pas dépendre de celles déjà amorcées.
	Profile int
	// Fuse est ce qui reste à brûler, en ticks. À zéro, elle détone et part.
	Fuse Tick
}

// amorcer pose une explosion là où une créature vient de mourir.
//
// Sans effet sur les profils qui n'explosent pas, dont le rayon est nul : le
// mécanisme se ferme par une donnée, jamais par un test sur le comportement.
//
// Le bassin plein perd l'explosion plutôt que de la différer, comme un tir
// perdu : une déflagration qui arriverait en retard détonerait sur un joueur qui
// s'est déjà écarté, ce qui est pire que rien.
func (w *World) amorcer(e *Enemy) {
	profil := &w.profils.Enemies[e.Profile]
	if profil.BurstRadius == 0 {
		return
	}
	w.souffles.Spawn(Blast{X: e.X, Y: e.Y, Profile: e.Profile, Fuse: profil.Fuse})
}

// detoner fait brûler les mèches et applique celles qui arrivent au bout.
//
// **L'explosion ne blesse que le joueur, et ce n'est pas un raccourci
// d'implémentation.** Qu'elle emporte la horde autour serait plus imitatif du
// réel et retournerait la mécanique : la Baudruche existe pour punir le
// nettoyage à l'aveugle en mêlée, or une déflagration qui nettoie les voisines
// récompense exactement le geste qu'elle devait décourager — tuer sans regarder
// ce qu'on tue deviendrait la bonne façon d'éclaircir une foule. La raison est
// écrite ici parce que le contraire paraîtra naturel à qui reprendra ce fichier
// sans la conception sous les yeux.
func (w *World) detoner() {
	for i := 0; i < w.souffles.Len(); i++ {
		b := w.souffles.At(i)
		if b.Fuse > 0 {
			b.Fuse--
			continue
		}

		profil := &w.profils.Enemies[b.Profile]
		ecart := Vec{X: w.playerX - b.X, Y: w.playerY - b.Y}
		if w.Alive() && ecart.carres() <= int64(profil.BurstRadius)*int64(profil.BurstRadius) {
			w.blesser(profil.BurstDamage)
		}
		w.souffles.RemoveAt(i)
	}
}

// BlastBounds rend les cases que l'emprise d'une explosion peut atteindre.
//
// Un rectangle englobant, que `BlastCovers` affine : le rendu n'a alors ni rayon
// ni table de profils à connaître, et l'emprise qu'il peint est celle que la
// détonation appliquera. Deux calculs de la même zone finiraient par marquer une
// case que l'explosion épargne.
func (w *World) BlastBounds(b *Blast) (u0, v0, u1, v1 int) {
	rayon := w.profils.Enemies[b.Profile].BurstRadius
	return (b.X - rayon).Floor(), (b.Y - rayon).Floor(),
		(b.X + rayon).Floor(), (b.Y + rayon).Floor()
}

// BlastCovers dit si le centre d'une case tombe sous une explosion.
//
// Le centre plutôt qu'un recouvrement partiel : ce que la simulation mesure est
// une distance depuis un point, et marquer une case dont le centre est hors du
// rayon annoncerait un danger qui n'arrivera pas.
func (w *World) BlastCovers(b *Blast, u, v int) bool {
	rayon := w.profils.Enemies[b.Profile].BurstRadius
	ecart := Vec{X: FromInt(u) + One/2 - b.X, Y: FromInt(v) + One/2 - b.Y}
	return ecart.carres() <= int64(rayon)*int64(rayon)
}

// FuseLeft rend ce qui reste de mèche rapporté à sa durée totale, sur mille.
//
// **Un rapport et non deux nombres**, parce que ce que le télégraphe montre est
// une progression : le rendu n'a pas à savoir combien de ticks une amorce dure,
// seulement où elle en est. Mille plutôt que cent pour qu'une mèche courte garde
// des paliers distincts, et un entier pour qu'aucun flottant n'entre dans ce que
// la simulation expose.
func (w *World) FuseLeft(b *Blast) int {
	total := w.profils.Enemies[b.Profile].Fuse
	if total <= 0 {
		return 0
	}
	return int(b.Fuse) * 1000 / int(total)
}
