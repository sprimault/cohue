// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le tir automatique : le ciblage du plus proche à portée, la cadence qui ne se
// consomme qu'en tirant, le vol des projectiles et ce qu'ils touchent.

package game

// tirer arme le tir automatique du joueur.
//
// La cadence descend à zéro et y demeure tant qu'aucune cible n'est à portée :
// l'arme prête reste prête. La consommer à vide rendrait la première salve d'un
// couloir dégagé dépendante du temps passé sans rien à viser — un décalage que
// rien à l'écran n'expliquerait, et qui ferait du comportement de l'arme une
// fonction du passé récent.
func (w *World) tirer() {
	if w.cooldown > 0 {
		w.cooldown--
		return
	}

	cible, trouvee := w.plusProche()
	if !trouvee {
		return
	}

	e := w.ennemis.At(cible)
	vers := (Vec{e.X - w.playerX, e.Y - w.playerY}).Direction(0)
	for range w.arme.Projectiles {
		if _, ok := w.tirs.Spawn(Projectile{
			X:         w.playerX,
			Y:         w.playerY,
			Step:      vers.Scale(w.arme.ProjectileSpeed),
			Remaining: w.arme.Range,
			Hits:      w.arme.Hits,
		}); !ok {
			// Bassin plein : le tir est perdu, pas différé. Une file d'attente
			// rendrait la cadence élastique, et l'arme rattraperait son retard
			// par une salve que rien n'a demandée.
			break
		}
	}
	w.cooldown = w.arme.Cooldown
}

// plusProche rend la place de la créature la plus proche à portée.
//
// La visée est omnidirectionnelle et le joueur ne choisit pas : c'est ce qui
// donnera son rôle au Secouriste, dont le seul moyen de se débarrasser est
// d'aller vers lui. Réintroduire un moyen de viser le désactiverait sans que
// personne ne touche au Secouriste.
//
// La comparaison porte sur les carrés des distances : une racine par créature et
// par tick, pour un classement que le carré donne aussi bien.
func (w *World) plusProche() (int, bool) {
	portee := w.arme.Range
	if portee <= 0 {
		// Une arme sans portée n'atteint rien. Sans cette ligne elle viserait ce
		// qui est exactement superposé au joueur, à la seule distance qu'un
		// carré nul admet.
		return 0, false
	}
	meilleure := int64(portee) * int64(portee)
	choix := -1

	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		if d := (Vec{e.X - w.playerX, e.Y - w.playerY}).carres(); d <= meilleure {
			meilleure = d
			choix = i
		}
	}
	return choix, choix >= 0
}

// deplacerTirs avance les projectiles et résout ce qu'ils touchent.
//
// Un projectile disparaît par le même chemin qu'il ait touché ou qu'il ait
// épuisé sa portée : deux causes, une seule suppression. Les écrire en deux
// branches finirait par en laisser une oublier de libérer sa place.
func (w *World) deplacerTirs() {
	for i := 0; i < w.tirs.Len(); i++ {
		p := w.tirs.At(i)
		p.X += p.Step.X
		p.Y += p.Step.Y
		p.Remaining -= p.Step.Len()

		if w.toucher(p) || p.Remaining <= 0 || !w.passable(p.X, p.Y) {
			w.tirs.RemoveAt(i)
			// L'entité remontée dans la place libérée attend le tick suivant :
			// la réexaminer maintenant la ferait avancer deux fois, et le
			// déterminisme dépendrait du sens du parcours.
		}
	}
}

// toucher applique un projectile à la première créature qu'il atteint.
//
// Une créature dont la résistance est tombée **cesse d'être une cible sans
// quitter le bassin** : elle y reste jusqu'à la fin du tick, pour que les index
// tiennent, et c'est la passe de nettoyage qui l'en retire. Deux projectiles
// arrivant sur elle dans le même tick ne la tuent donc qu'une fois, et le second
// va chercher derrière.
//
// La retirer sur-le-champ donnerait le même résultat visible aujourd'hui, mais
// pour une raison accidentelle : les projectiles et les ennemis vivent dans deux
// bassins distincts, si bien que supprimer dans l'un ne dérange pas le parcours
// de l'autre. Les dégâts de contact, eux, parcourront le bassin des ennemis en
// les tuant — et une suppression en cours de passe y changerait les index sous
// les pieds de la boucle. La garde, elle, tient quel que soit le bassin parcouru.
func (w *World) toucher(p *Projectile) bool {
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		if e.Hits <= 0 {
			continue
		}
		rayon := w.profils.Enemies[e.Profile].Radius
		if (Vec{e.X - p.X, e.Y - p.Y}).carres() > int64(rayon)*int64(rayon) {
			continue
		}

		// La transition, et non l'état : c'est ici que se brancheront le butin,
		// les points et le cadavre, une seule fois chacun.
		e.Hits -= p.Hits
		return true
	}
	return false
}
