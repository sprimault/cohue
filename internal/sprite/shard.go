// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Où en est un éclat de sa course : sa position dans le plan et sa hauteur
// au-dessus du sol, toutes deux dérivées de son rang et de son âge.

package sprite

import "github.com/sprimault/cohue/internal/game"

// Shards est le nombre d'éclats qu'une volée dessine.
//
// **Un compte de rendu et non de simulation.** Le bassin retient un événement
// par chose détruite ; ce qu'on en dessine est une décision d'apparence, et la
// changer ne déplace rien de ce que la partie retient. Huit fait une gerbe
// franche sans que le sol se couvre de bois à la troisième caisse.
const Shards = 8

// portee est la distance qu'un éclat parcourt sur toute sa vie, en tuiles.
//
// Un peu plus d'une demi-tuile : de quoi que la volée s'ouvre et se lise comme
// une gerbe, assez peu pour qu'elle reste ce qui vient de la caisse plutôt
// qu'une pluie sur la salle.
const portee game.Fixed = game.One * 5 / 8

// hauteurEclat est le sommet de la parabole d'un éclat, en pixels d'écran.
//
// En pixels et non en tuiles, parce qu'une hauteur n'est pas une distance du
// monde : l'élévation ne participe à aucun calcul de simulation, et c'est le
// rendu qui la pose. Douze pixels font trois quarts de la hauteur d'une tuile.
const hauteurEclat = 12

// Shard rend où en est un éclat : son écart au point d'émission dans le plan du
// monde, et sa hauteur au-dessus du sol en pixels.
//
// **Rien n'est stocké, tout se dérive du rang et de l'âge**, comme l'image d'un
// cycle se dérive du tick. Une volée de huit éclats n'est donc qu'un point et un
// décompte dans le bassin, et son apparence peut changer sans que la simulation
// bouge d'un pixel.
//
// **Les directions viennent de la table du monde**, au pas de trois qui sépare
// deux rangs voisins de cent trente-cinq degrés — celle-là même qui écarte deux
// entités superposées, employée ici pour que la gerbe ne parte pas en éventail
// serré. Aucune trigonométrie, donc aucun arrondi qui diffère d'une machine à
// l'autre : ce n'est pas exigé d'un cosmétique, mais une planche de relecture
// doit rendre les mêmes octets d'une exécution à l'autre.
//
// **La course est linéaire et la hauteur parabolique.** Un éclat s'éloigne à
// vitesse constante — rien ne le freine à cette échelle — et monte puis retombe,
// ce qui est la seule des deux courbes que l'œil lit comme une chute.
func Shard(rang int, age, total game.Tick) (ecart game.Vec, hauteur int) {
	if total <= 0 {
		return game.Vec{}, 0
	}
	avance := min(max(int64(age), 0), int64(total))

	// Le rapport en virgule fixe et non par une division d'entiers larges : la
	// seconde demanderait de rétrécir un `int64` vers le type, et une conversion
	// qu'aucun outil ne peut borner devient une alerte à taire. `Div` sature là
	// où elle déborderait, ce qui est le comportement voulu et le seul écrit.
	part := game.FromInt(int(avance)).Div(game.FromInt(int(total)))

	direction := game.Heading(rang * 3)
	distance := portee.Mul(part)

	// La parabole passe par zéro aux deux bouts et culmine au milieu :
	// `4·h·t·(1−t)` vaut `h` en `t = 1/2`. Le calcul se fait en entiers, l'ordre
	// des facteurs plaçant la division en dernier.
	reste := int64(total) - avance
	return game.Vec{
			X: direction.X.Mul(distance),
			Y: direction.Y.Mul(distance),
		},
		int(4 * hauteurEclat * avance * reste / (int64(total) * int64(total)))
}
