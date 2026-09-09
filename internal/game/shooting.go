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

	_ = w.salve(&w.arme, w.interception(w.ennemis.At(cible)).Direction(0))
	w.cooldown = w.arme.Cooldown
}

// salve pose les projectiles d'une arme depuis le joueur.
//
// Le cas ordinaire de `salveDe` : ce qui tire est le joueur, et son point de
// départ est le sien.
func (w *World) salve(arme *Weapon, vers Vec) int {
	return w.salveDe(arme, w.playerX, w.playerY, vers)
}

// salveDe pose les projectiles d'une arme depuis un point, dans une direction.
//
// **Paramétrée par l'arme et non par le socle**, depuis qu'une lourde tire : le
// fusil à pompe emploie ce mécanisme tel quel, avec ses propres nombre, front et
// ouverture. Recopier ces trente lignes en aurait fait deux endroits où tenir la
// répartition dans la salve, dont le premier a déjà coûté un palier qui faisait
// perdre.
//
// **Et par un point depuis qu'une tourelle tire**, qui n'est pas là où le joueur
// se tient. C'est le second paramètre que ce mécanisme a pris plutôt qu'une
// copie, et le même geste : une origine, pas une abstraction.
//
// Ce qu'elle ne fait pas est décider **quand** : la cadence appartient à
// l'appelant — un compteur pour le socle, une charge dépensée pour une lourde,
// le sien pour une tourelle.
//
// Elle rend le nombre de projectiles réellement posés, que le socle ignore et
// que les deux autres regardent : ceux-là ne dépensent leur charge que si
// quelque chose est parti, quand le socle perd les tirs qu'un bassin plein
// refuse.
func (w *World) salveDe(arme *Weapon, x, y Fixed, vers Vec) int {
	cote := vers.Perp()
	n, poses := arme.Projectiles, 0
	for k := range n {
		depart := cote.Scale(ecartDansLaSalve(arme.Front, k, n))

		// **La course reste `vers` tant que l'ouverture est nulle**, ce qui rend
		// l'identité vraie par construction plutôt que par ce qu'une
		// renormalisation rendrait. `Direction` passe par un flottant et un
		// arrondi ; sur la run de référence elle rend bien `vers` au bit près —
		// éprouvé en supprimant cette garde, qui laisse l'attendu inchangé —, mais
		// ce n'est pas une propriété qu'on puisse écrire pour toute direction et
		// toute portée. Elle évite le calcul, et surtout la question.
		// **La gerbe inverse le signe de l'écartement au point visé**, et rien de
		// plus : la salve part de son front et converge là où elle vise, au lieu
		// de s'en écarter. C'est ce qui la rend écrivable sans mécanisme neuf, et
		// ce qui répare ce que le front coûte à cible unique.
		ouverture := arme.Spread
		if arme.Spray {
			ouverture = -ouverture
		}

		pas := vers
		if ouverture != 0 {
			// Le point visé est à la portée, décalé de l'ouverture : c'est ce qui
			// donne l'éventail sans jamais écrire d'angle.
			pas = vers.Scale(arme.Range).
				Add(cote.Scale(ecartDansLaSalve(ouverture, k, n))).
				Direction(k)
		}

		if _, ok := w.tirs.Spawn(Projectile{
			X:         x + depart.X,
			Y:         y + depart.Y,
			Step:      pas.Scale(arme.ProjectileSpeed),
			Remaining: arme.Range,
			Hits:      arme.Hits,
			Pierce:    arme.Pierce,
			Bounces:   arme.Bounces,
			Rail:      arme.Rail,
		}); !ok {
			// Bassin plein : le tir est perdu, pas différé. Une file d'attente
			// rendrait la cadence élastique, et l'arme rattraperait son retard
			// par une salve que rien n'a demandée.
			break
		}
		poses++
	}
	return poses
}

// ecartDansLaSalve répartit le k-ième d'une salve de n sur une largeur, autour de
// l'axe de visée.
//
// **Le rang zéro reste sur l'axe, quel que soit le nombre**, et les suivants
// s'écartent par paires de part et d'autre : le pas vaut la demi-largeur divisée
// par le nombre de paires, si bien que la salve ne dépasse jamais `largeur` et se
// densifie de deux paliers en deux paliers.
//
// **La première version répartissait à intervalles égaux du premier au dernier,
// et elle faisait perdre.** Le centre n'y était occupé que pour un nombre impair
// ; à deux et à quatre projectiles, la salve encadrait la cible sans jamais la
// toucher — un Quidam a un rayon de 0,125 tuile pour un front d'une tuile, si
// bien que le premier palier de l'axe faisait passer d'une touche à zéro. La
// conception pose qu'à trois projectiles on ne fait pas trois fois un contre une
// créature isolée ; elle ne dit pas *moins d'une fois un*, et ce projet a
// explicitement refusé un palier qui fait perdre.
//
// **Le prix est l'asymétrie d'un rang pair** : il est le rang impair précédent
// plus un projectile à une extrémité. Une salve centrée et symétrique ne peut pas
// tenir les deux à la fois, et c'est le centre qui décide de ce qui touche.
//
// **Deux largeurs la traversent, et c'est ce qui l'a fait extraire.** `Front`
// écarte les départs au canon, `Spread` écarte les points visés à la portée : la
// même répartition, sur deux distances. Un éventail se dit alors sans angle, ce
// que le déterminisme exige — voir `Weapon.Spread`. Le centre tenu leur profite
// à toutes les deux : un éventail qui écarterait tout de l'axe ferait rater ce
// qu'il vise, exactement comme le front.
//
// **Aucune des deux ne dépend du nombre**, si bien qu'un palier de projectiles
// resserre la salve au lieu de l'étaler. La conséquence qui compte pour
// l'éventail est le cas `n < 2` : il ne fait rien sur un tir seul, et la carte
// reste offerte quand même.
func ecartDansLaSalve(largeur Fixed, k, n int) Fixed {
	if n < 2 {
		return 0
	}
	// La multiplication précède la division : l'inverse arrondirait le pas avant
	// de le multiplier, et les extrêmes n'atteindraient plus la demi-largeur.
	ecart := largeur.Mul(FromInt((k + 1) / 2)).Div(FromInt(2 * (n / 2)))
	if k%2 == 0 {
		return -ecart
	}
	return ecart
}

// reductionDuDos ramène deux longueurs au 1/256 de tuile avant de les
// multiplier.
//
// **Les carrés de deux longueurs en virgule fixe frôlent la borne de l'`int64`
// dès qu'on les multiplie entre eux** : six tuiles au carré valent déjà 1,5e11,
// et le produit de deux tels carrés déborde. Un huitième de degré de précision ne
// manque à personne pour décider d'un secteur, et le décalage arithmétique reste
// déterministe — c'est ce que la virgule fixe demande, pas plus.
const reductionDuDos = 8

// Le seuil du secteur arrière, `cos²(22,5°)` en dix-millièmes.
//
// **C'est la demi-largeur d'une bande de sprite, et non un angle choisi.** Le
// manifeste des personnages déclare huit directions, donc quarante-cinq degrés
// chacune ; celle qui fait face à l'opposé du pas couvre vingt-deux degrés et
// demi de part et d'autre. Écarter ce secteur écarte exactement la bande qui
// montrait le dos, et pas un degré de plus — le chiffre se dérive du nombre de
// directions au lieu de se régler.
const (
	dosNumerateur   = 8536
	dosDenominateur = 10000
)

// dansLeDos dit si un écart tombe dans le secteur arrière d'un pas.
//
// **Le secteur strict et non le demi-plan, et la nuance est tout le lot.** Le
// chapitre 9 a écarté un cône avant avec un argument qui tient toujours : « kiter,
// c'est avoir la horde derrière soi, un cône avant ferait de la fuite un moment
// sans dégâts ». Mesuré sur la run de référence, il coûtait les trois quarts des
// éliminations — cent abattus tombaient à vingt-six, et le pilote mourait avant
// la porte. Ce qu'une partie jouée reprochait n'était pas de tirer derrière soi
// mais de **marcher à reculons**, relevé à 23,7 % des images sur la bande à
// l'opposé exact du pas.
//
// Les quatre secteurs mesurés, sur les cent abattus qu'ouvrir la porte demande :
//
//	dos écarté   180°   90°   60°   45°
//	abattus       26     80    93   100
//
// Retirer la seule bande opposée règle ce qui se voit sans rien coûter au
// kiting : on tire encore sur les flancs et le trois-quarts arrière, donc en
// fuyant.
//
// **Sans cosinus, ce que le déterminisme exige** : le signe du produit scalaire
// donne le côté, et la comparaison de son carré au produit des carrés donne
// l'angle. Aucune racine, aucun flottant.
func dansLeDos(ecart, pas Vec) bool {
	a := Vec{ecart.X >> reductionDuDos, ecart.Y >> reductionDuDos}
	b := Vec{pas.X >> reductionDuDos, pas.Y >> reductionDuDos}

	s := a.scalaire(b)
	if s >= 0 {
		return false
	}
	return dosDenominateur*s*s > dosNumerateur*a.carres()*b.carres()
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
// champ de flux. Deux **mesures** différentes seraient pires encore : l'arrêt se
// décide par cellule, si bien qu'une créature arrêtée à six cases peut être à
// sept tuiles réelles, et resterait figée hors de sa propre portée.
//
// **Deux bornes, et chacune garde ce que l'autre ne peut pas.** La distance à
// vol d'oiseau garde que le projectile porte — il vole la portée du profil et
// pas une tuile de plus. Le champ garde qu'un chemin existe : à vol d'oiseau
// seul, une Buse que six tuiles de mur séparent du joueur se croyait arrivée,
// se figeait, et tirait dans la paroi où le projectile meurt. C'est un profil
// dont tout le rôle est de blesser de loin, réduit à une statue par un obstacle
// qu'il n'avait qu'à contourner. C'est l'isodistance que la conception pose au
// chapitre 4 : le même champ pour tous les comportements, les uns descendant son
// gradient, celui-ci se stabilisant sur une de ses courbes.
//
// **Le champ seul ne suffisait pas, et c'est ce qui a fixé la forme.** Il compte
// par cellule et à quatre voisins : une créature arrêtée sur la case qui porte sa
// portée est au-delà en ligne droite, si bien que ses tirs mouraient tous avant
// d'arriver. Retenir la plus contraignante des deux ramène le cas ordinaire à ce
// qu'il était et ne change que celui du détour.
//
// Conséquence à connaître : le chemin étant orthogonal, une approche en diagonale
// s'arrête plus près qu'une approche par un axe. C'est ce que mesurer un chemin
// veut dire, et le même écart fait déjà contourner ce qui ralentit.
//
// L'écart, lui, reste géométrique — c'est la direction du tir, et elle vise où le
// joueur est.
func (w *World) viseeDe(e *Enemy, profil *EnemyProfile) (Vec, bool) {
	vers := Vec{X: w.playerX - e.X, Y: w.playerY - e.Y}
	if profil.Range == 0 {
		return vers, false
	}
	if vers.carres() > int64(profil.Range)*int64(profil.Range) {
		return vers, false
	}
	u, v := w.flux.Cell(e.X, e.Y)
	return vers, w.flux.Distance(u, v) <= porteeDuChamp(profil.Range)
}

// porteeDuChamp convertit une portée en tuiles vers l'unité du champ de flux.
//
// Une case ordinaire y coûte `Free`, donc une tuile de chemin. La troncature est
// ce qui garde la conversion du bon côté : une portée de six et demie autorise
// six cases et jamais sept, et une case de plus vaudrait une tuile entière
// au-delà de ce que le projectile sait couvrir.
//
// Ce que le champ compte est un coût et non un nombre de cases : une flaque vaut
// ce que le manifeste lui donne, si bien qu'une Buse s'arrête plus loin derrière
// un sol coûteux. C'est ce que « contourner ce qui ralentit » veut dire, appliqué
// à un profil qui ne contourne pas.
func porteeDuChamp(portee Fixed) uint32 {
	// Le chargement refuse déjà une portée négative, mais la conversion ne s'y
	// adosse pas : non bornée, elle rendrait une portée immense là où le manifeste
	// dit l'inverse, et la créature ne s'arrêterait plus nulle part.
	return uint32(max(portee.Floor(), 0)) * uint32(Free) // #nosec G115 -- borné par le max
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

// plusProche rend la place de la créature la plus proche à portée, devant le
// joueur quand il marche.
//
// **Elle reste omnidirectionnelle à un secteur près.** Le sprite s'oriente sur
// la visée, si bien qu'une visée libre faisait marcher le joueur à reculons près
// d'une image sur quatre. Le cône arrière strict est donc écarté quand il marche
// — voir `dansLeDos`, qui dit pourquoi il est étroit plutôt qu'un demi-plan.
//
// **À l'arrêt, plus rien n'est écarté**, et cette exception est ce qui rend la
// règle sûre. La conception justifie le corps solide du Vigile ainsi : « un
// joueur coincé entre un Vigile et un mur tire nécessairement dessus, puisque la
// visée prend le plus proche, et douze touches finissent par tomber ». Un joueur
// bloqué pousse dos à lui et ne le viserait plus — il mourrait coincé sans
// qu'un pixel dise pourquoi. Or `playerStep` porte le pas **obtenu** : pousser
// contre un corps ou contre un mur y rend zéro, et le cas se referme sans avoir
// à le reconnaître.
//
// **Le Secouriste est intact.** Il n'est écarté que s'il se tient exactement dans
// le dos, et il vient au joueur : le seul moyen de l'abattre reste d'aller vers
// lui, ce que le chapitre 4 lui donne pour rôle.
//
// La comparaison porte sur les carrés des distances : une racine par créature et
// par tick, pour un classement que le carré donne aussi bien.
func (w *World) plusProche() (int, bool) {
	return w.plusProcheDe(w.playerX, w.playerY, w.arme.Range, Handle{}, w.playerStep)
}

// plusProcheDe rend la place de la créature vivante la plus proche d'un point.
//
// **Le point n'est pas toujours le joueur**, et c'est le ricochet qui l'a
// demandé : un projectile qui repart cherche autour de son impact, et il exclut
// celle qu'il vient de frapper — sans quoi il rebondirait sur elle, qui est par
// construction la plus proche.
//
// **Les mortes sont écartées**, ce que la recherche autour du joueur ne faisait
// pas. Une résistance tombée est la mort et la passe de nettoyage n'a pas encore
// eu lieu : un rebond les comptant repartirait vers un cadavre, et une salve
// tirée dans le même tick viserait un mort plutôt que ce qui menace.
//
// `Handle{}` ne désigne aucune entité — les générations partent à un —, si bien
// que la recherche sans exclusion n'a pas de cas à part.
//
// **`devant` restreint la recherche au demi-plan qu'il ouvre, et le vecteur nul
// ne restreint rien.** Ce n'est pas une sentinelle : un pas nul est un mobile à
// l'arrêt, qui n'a pas de devant, et c'est exactement le cas où l'on cherche tout
// autour. Le ricochet et la déflagration passent zéro pour la même raison — un
// projectile qui repart et une grenade qui tombe n'ont pas de dos à protéger.
func (w *World) plusProcheDe(x, y, portee Fixed, sauf Handle, devant Vec) (int, bool) {
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
		if e.Hits <= 0 || w.ennemis.HandleAt(i) == sauf {
			continue
		}
		ecart := Vec{e.X - x, e.Y - y}
		if devant != (Vec{}) && dansLeDos(ecart, devant) {
			continue
		}
		if d := ecart.carres(); d <= meilleure {
			meilleure = d
			choix = i
		}
	}
	return choix, choix >= 0
}

// rebondir réoriente un projectile vers une autre cible et dit s'il en a une.
//
// **Le rayon de recherche est la portée restante**, ce qui n'ajoute aucun
// réglage : un projectile ne repart que vers ce qu'il pouvait déjà atteindre, et
// la portée continue de descendre. Un rebond qui la rechargerait rendrait le
// projectile perpétuel dans une foule, chaque créature en amenant une autre.
//
// La vitesse se relit sur le pas plutôt que sur l'arme : un projectile en vol ne
// renvoie pas vers ce qui l'a tiré, et l'arme peut avoir monté de niveau depuis.
func (w *World) rebondir(p *Projectile) bool {
	// Sans restriction de côté : un projectile qui repart cherche autour de son
	// impact, et lui imposer le demi-plan de sa course en ferait un tir de plus
	// plutôt qu'un rebond.
	cible, trouvee := w.plusProcheDe(p.X, p.Y, p.Remaining, p.LastHit, Vec{})
	if !trouvee {
		return false
	}
	e := w.ennemis.At(cible)
	p.Step = (Vec{e.X - p.X, e.Y - p.Y}).Direction(cible).Scale(p.Step.Len())
	return true
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
		if e.Hits <= 0 || w.ennemis.HandleAt(i) == p.LastHit {
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
	w.compter(e.X, e.Y, p.Hits)
	if e.Hits <= 0 {
		w.lacher(e)
		w.amorcer(e)
	}
	p.LastHit = w.ennemis.HandleAt(touchee)

	// **La perforation prolonge la course, le rebond la redirige**, et c'est ce
	// qui fixe leur ordre : rebondir d'abord ferait qu'un projectile ayant les
	// deux ne traverserait jamais rien, et l'axe du perforant ne servirait à rien
	// chez qui l'a pris. `Projectile.Pierce` porte la décision.
	switch {
	case p.Rail:
		// Le rail ne décompte rien : ce que la fusion retire est la borne, pas le
		// nombre. Le tir s'arrête sur un mur ou au bout de sa portée, comme les
		// autres.
		return false
	case p.Pierce > 0:
		p.Pierce--
		return false
	case p.Bounces > 0:
		p.Bounces--
		// Sans cible à portée, le projectile meurt là où il a frappé plutôt que
		// de poursuivre tout droit : le rebond a été dépensé, et le laisser
		// filer donnerait à un tir sans suite la course d'un tir ordinaire.
		return !w.rebondir(p)
	default:
		return true
	}
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
// laquelle un Quidam parcourt une tuile et demie pour un rayon de un huitième. Une
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
