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

	vers := w.interception(w.ennemis.At(cible)).Direction(0)
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

// tirerLaHorde fait tirer les créatures dont le profil porte une portée.
//
// **La cadence ne se consomme pas hors de portée**, exactement comme celle de
// l'arme du joueur : une Buse qui voit le joueur réapparaître tirerait sinon
// avec un retard fonction du temps qu'elle a passé sans cible, et rien à l'écran
// ne l'expliquerait.
//
// **Rien ne vérifie que la voie est libre.** Le projectile part et meurt sur le
// pilier, par le même chemin qu'un tir du joueur : le décor protège par le fait,
// pas par une condition — c'est ce que la charge fait déjà, et pour la même
// raison.
func (w *World) tirerLaHorde() {
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		profil := &w.profils.Enemies[e.Profile]
		if profil.Range == 0 {
			continue
		}

		vers, portee := w.viseeDe(e, profil)
		if !portee {
			continue
		}
		if e.ShotTimer > 0 {
			e.ShotTimer--
			continue
		}

		// Le tir vise où le joueur est et non où il sera : c'est ce qui le rend
		// esquivable, donc ce qui punit le camping sans punir le déplacement.
		if _, ok := w.tirsEnnemis.Spawn(Projectile{
			X:         e.X,
			Y:         e.Y,
			Step:      vers.Direction(i).Scale(profil.ShotSpeed),
			Remaining: profil.Range,
			Hits:      profil.ShotDamage,
		}); !ok {
			continue
		}
		e.ShotTimer = profil.ShotCooldown
	}
}

// deplacerTirsEnnemis avance les projectiles de la horde et les applique au
// joueur.
//
// Elle double la passe des tirs du joueur plutôt que de la partager, et la
// raison n'est pas le bassin mais la cible : celle-là cherche la première
// créature atteinte dans un bassin de trois cents, celle-ci compare à un seul
// point. Les fondre demanderait un paramètre disant quoi toucher, pour deux
// corps qui n'ont en commun que le déplacement.
func (w *World) deplacerTirsEnnemis() {
	for i := 0; i < w.tirsEnnemis.Len(); i++ {
		p := w.tirsEnnemis.At(i)
		depart := Vec{p.X, p.Y}
		p.X += p.Step.X
		p.Y += p.Step.Y
		p.Remaining -= p.Step.Len()

		touche := w.Alive() && w.auContactDu(p.X, p.Y)
		if touche {
			// Hors du plafond de dégâts, comme le choc d'une charge : ce que le
			// plafond rend lisible est l'encerclement, pas un projectile qu'on
			// a vu venir.
			w.blesser(p.Hits)
		}
		if touche || p.Remaining <= 0 || w.traverse(depart, p.X, p.Y) {
			w.tirsEnnemis.RemoveAt(i)
		}
	}
}

// viseeDe rend l'écart au joueur et dit s'il est à portée de tir.
//
// **Un seul prédicat pour les deux endroits qui en dépendent** : celui qui tire,
// et celui qui immobilise la créature pour qu'elle tire. Écrits séparément, une
// borne stricte d'un côté et large de l'autre donneraient une Buse arrêtée à la
// distance exacte où elle refuse de tirer — un blocage qu'on chercherait dans le
// champ de flux.
func (w *World) viseeDe(e *Enemy, profil *EnemyProfile) (Vec, bool) {
	vers := Vec{X: w.playerX - e.X, Y: w.playerY - e.Y}
	if profil.Range == 0 {
		return vers, false
	}
	return vers, vers.carres() <= int64(profil.Range)*int64(profil.Range)
}

// auContactDu dit si un point touche le joueur.
//
// Le rayon est celui du profil, comme pour le contact d'une créature : c'est la
// même mesure, et un projectile qui aurait la sienne rendrait la hitbox du
// joueur dépendante de ce qui l'atteint.
func (w *World) auContactDu(x, y Fixed) bool {
	portee := w.profils.Player.Radius
	ecart := Vec{X: x - w.playerX, Y: y - w.playerY}
	return ecart.carres() < int64(portee)*int64(portee)
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
		depart := Vec{p.X, p.Y}
		p.X += p.Step.X
		p.Y += p.Step.Y
		p.Remaining -= p.Step.Len()

		if w.toucher(depart, p) || p.Remaining <= 0 || w.traverse(depart, p.X, p.Y) {
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
func (w *World) toucher(depart Vec, p *Projectile) bool {
	// **La mesure porte sur le segment parcouru, pas sur le point d'arrivée.**
	// Un projectile avance de 0,2 tuile par tick pour un rayon de créature de
	// 0,125 : une cible qui tombe entre deux positions échantillonnées n'est
	// jamais dans le rayon au moment du test, et le tir la traverse sans effet.
	// Le cas ne se voyait pas tant que la horde arrivait de loin et de face, où
	// l'un des points finit par tomber dedans ; il devient systématique dès
	// qu'une créature touche le joueur, puisque le projectile naît sur elle et
	// l'a dépassée avant la première mesure.
	pas := p.Step
	long := pas.carres()

	touchee, avancee := -1, int64(0)
	for i := range w.ennemis.Active() {
		e := w.ennemis.At(i)
		if e.Hits <= 0 {
			continue
		}

		vers := Vec{e.X - depart.X, e.Y - depart.Y}
		projection := int64(vers.X)*int64(pas.X) + int64(vers.Y)*int64(pas.Y)

		// La distance se prend au point le plus proche du segment, borné à ses
		// deux extrémités : un projectile ne touche ni ce qui est derrière son
		// point de départ, ni ce qui est au-delà de son point d'arrivée.
		var carre int64
		switch {
		case long == 0 || projection <= 0:
			carre, projection = vers.carres(), 0
		case projection >= long:
			carre, projection = (Vec{vers.X - pas.X, vers.Y - pas.Y}).carres(), long
		default:
			carre = vers.carres() - projection*projection/long
		}

		rayon := int64(w.profils.Enemies[e.Profile].Radius)
		if carre > rayon*rayon {
			continue
		}

		// **La première rencontrée le long du segment, pas la première du
		// bassin.** Un projectile n'en touche qu'une : la retenir dans l'ordre du
		// bassin la ferait tuer à travers une créature qui la précède, ce qui se
		// verra dès que la horde sera dense.
		if touchee < 0 || projection < avancee {
			touchee, avancee = i, projection
		}
	}

	if touchee < 0 {
		return false
	}

	// La transition, et non l'état : c'est ici que se branchent le butin et, plus
	// tard, les points et le cadavre — une seule fois chacun. La boucle
	// ci-dessus ayant écarté ce qui n'a plus de résistance, une créature ne
	// franchit ce seuil qu'une fois.
	e := w.ennemis.At(touchee)
	e.Hits -= p.Hits
	e.Flash = eclairImpact
	if e.Hits <= 0 {
		w.lacher(e)
	}
	return true
}

// traverse dit si le pas d'un projectile entre dans une case qui l'arrête.
//
// **Le point d'arrivée ne suffit pas.** Un pas de deux dixièmes de tuile qui
// coupe l'angle où quatre cases se rencontrent entre dans l'une d'elles et en
// ressort sans qu'aucun de ses deux bouts n'y tombe : le tir traversait le muret,
// rarement et sans raison visible. Il faut raser l'angle, ce qui explique qu'on
// ne l'ait vu qu'une fois en jouant.
//
// C'est la leçon que `toucher` avait déjà tirée pour les cibles — la mesure porte
// sur le segment parcouru et non sur son extrémité —, et qui n'avait pas été
// portée jusqu'aux murs. Les deux cases obliques suffisent à la couvrir : un pas
// plus court qu'une demi-case ne peut en croiser d'autres.
func (w *World) traverse(depart Vec, x, y Fixed) bool {
	if !w.passable(x, y) {
		return true
	}

	du, dv := depart.X.Floor(), depart.Y.Floor()
	au, av := x.Floor(), y.Floor()
	if du == au || dv == av {
		return false
	}
	return !w.grille.Passable(au, dv) || !w.grille.Passable(du, av)
}

// interception rend le vecteur qui va du joueur à l'endroit où la cible sera
// quand le projectile y arrivera.
//
// **Viser où la cible est se voit en jouant.** Le projectile vole à douze tuiles
// par seconde sur une portée de six : une demi-seconde au plus loin, pendant
// laquelle un Badaud parcourt une tuile et demie pour un rayon de un huitième. Une
// créature qui traverse était donc manquée de douze fois son rayon, et seules
// celles qui venaient droit sur le joueur étaient touchées de façon fiable — ce
// que le champ de flux rend fréquent, d'où un tir qui rate « parfois » plutôt
// que toujours.
//
// **Un tir manqué n'est pas un signal d'adresse ici**, la visée étant
// automatique : c'est du bruit, et le joueur n'a aucun moyen de le corriger.
//
// **Deux passes plutôt qu'une équation.** Le point d'interception exact est la
// racine d'un trinôme, dont le discriminant demanderait une seconde racine carrée
// et une branche pour le cas où la cible est plus rapide que le tir. L'itération
// converge en deux tours parce que le projectile va quatre fois plus vite que ce
// qu'il poursuit : la première passe estime le vol sur la distance actuelle, la
// seconde sur la distance corrigée, et ce qui reste est en deçà du rayon d'une
// créature.
func (w *World) interception(e *Enemy) Vec {
	vers := Vec{e.X - w.playerX, e.Y - w.playerY}
	if w.arme.ProjectileSpeed <= 0 {
		return vers
	}

	for range 2 {
		ticks := vers.Len().Div(w.arme.ProjectileSpeed)
		vers = Vec{
			X: e.X + e.Step.X.Mul(ticks) - w.playerX,
			Y: e.Y + e.Step.Y.Mul(ticks) - w.playerY,
		}
	}
	return vers
}
