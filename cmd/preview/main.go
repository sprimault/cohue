// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La planche de relecture du rendu : les vues qu'elle écrit, le monde qu'elle
// pose pour chacune, et le détour par le cycle Ebitengine que le dessin impose.

// Preview écrit des vues du rendu dans `.tmp/apercus/`, à regarder.
//
// `internal/render` n'a pas de test et n'en aura pas : importer Ebitengine
// initialise GLFW, qui panique sans écran. Le rendu se juge donc à l'œil, et une
// planche est ce qui rend ce jugement possible ailleurs que dans une partie —
// on y compare une image d'avant et d'après un changement, ce qu'une observation
// au clavier ne permet pas.
//
// **Elle pilote le `render.Screen` du jeu, jamais une scène montée à côté.** Une
// planche qui dessinerait autrement que le jeu relirait la planche, et c'est le
// même piège qu'un test qui bâtit son entrée au lieu de passer par le chemin qui
// la produit.
//
// **Elle est déterministe, et doit le rester.** Deux exécutions écrivent des
// octets identiques, sans quoi comparer une planche d'avant et d'après un
// changement ne dirait rien. Ce qui la tient n'est plus l'absence de tirage — le
// spawner en fait à chaque tick — mais la graine fixée d'avance, et des pas
// joués avec une direction nulle. C'est pour cette échéance qu'elle avait été
// écrite ainsi.
//
// Elle exige un écran et ne tourne donc pas en intégration continue. Ce n'est
// pas un contrôle mais une planche : ce qui se vérifie mécaniquement vit du côté
// simulation, qui n'importe pas Ebitengine.
package main

import (
	"errors"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
	"github.com/sprimault/cohue/internal/level"
	"github.com/sprimault/cohue/internal/render"
	"github.com/sprimault/cohue/internal/session"
	"github.com/sprimault/cohue/internal/sprite"
)

// sortie est le dossier des planches.
//
// Dans `.tmp/` et non dans `assets/`, dont le contrôle des ressources compare le
// contenu à ce que les générateurs produisent : une planche y deviendrait un
// écart à excepter, et elle pèserait dans le binaire de qui l'a régénérée.
const sortie = ".tmp/apercus"

// echelle agrandit le tampon avant l'écriture, en entier.
//
// **La planche montre ce que le joueur verra, et non le tampon nu.** Une image
// écrite à la taille du tampon se relit à la loupe, et une image agrandie d'un
// facteur qui n'est pas entier épaissit un pixel sur deux : un glyphe y paraît
// bancal alors que le jeu le rend net, et le jugement porte sur l'agrandissement
// plutôt que sur ce qu'on croit juger.
//
// Deux et non trois parce que c'est le facteur d'une fenêtre de 1080p, le cas
// courant. Ce que la fenêtre du jeu fera de son côté est un réglage d'affichage
// que l'étape 15 tranchera ; ici, on montre le tampon multiplié.
const echelle = 2

// graine est celle sur laquelle chaque vue monte sa partie.
//
// Fixée ici plutôt que reçue : la planche se compare d'une exécution à l'autre,
// et la simulation tire à chaque tick — les vagues du spawner, la place d'un
// aimant —, si bien qu'une graine qui varierait rendrait la comparaison muette.
// C'est une exigence de la planche et non un réglage partagé avec le jeu, qui a
// la sienne pour une autre raison.
const graine uint64 = 1

// repere est un endroit du lieu, résolu sur sa grille plutôt qu'écrit en cases.
//
// **Les cases étaient écrites, et elles ont menti dès que le lieu a grandi.**
// Les quatre coins du losange étaient ceux d'une carte de trente-deux cases : la
// planche a continué de les nommer coins en les posant au premier neuvième d'un
// lieu qui en fait trois fois plus, là où la caméra ne bute sur rien. Un repère
// se résout sur la grille reçue, donc il suit.
type repere int

// Les endroits que la planche sait nommer. Le centre vaut zéro : une vue qui ne
// dit rien s'y pose, ce qui est le cas de toutes celles qui jugent autre chose
// que le cadrage.
const (
	centre repere = iota
	nord
	ouest
	est
	sud
	// porte pose le joueur devant la sortie. Elle se résout sur le lieu comme
	// les autres, et pour la même raison : la porte est écrite dans le fichier
	// du lieu, et l'écrire ici en ferait une seconde description qui mentirait
	// au premier déplacement.
	porte
	// Les quatre milieux d'arête, où le joueur vient s'appuyer contre l'enceinte.
	//
	// **Les quatre sommets ne montrent pas ce qui se passe le long d'un bord.**
	// Ils posent le joueur à deux cases du mur, ce qui juge le cadrage et rien
	// d'autre ; une partie jouée a signalé que longer l'enceinte donnait au
	// personnage l'air d'être passé dessus, et aucune vue ne le donnait à voir.
	// Un milieu d'arête est le seul endroit où l'on soit contre le mur sans être
	// dans un coin, c'est-à-dire sans que deux bords se répondent.
	murNordOuest
	murNordEst
	murSudEst
	murSudOuest
)

// margeDuCoin est la distance au bord à laquelle un repère de coin se pose.
//
// Deux cases : le trottoir du lieu en fait autant, si bien que le joueur y est
// posé sur du sol quel que soit le bloc qui touche le coin.
const margeDuCoin = 2

// cases résout un repère sur une grille et la sortie du lieu.
//
// La porte est murée — le chargement l'exige —, donc le joueur se pose sur la
// case franchissable qui la touche. La chercher plutôt que l'écrire est ce qui
// permet à un lieu de poser sa sortie sur n'importe lequel de ses quatre bords.
func (r repere) cases(g *game.CostGrid, sortie *game.Exit) (int, int) {
	loinU, loinV := g.Width()-1-margeDuCoin, g.Height()-1-margeDuCoin
	switch r {
	case nord:
		return margeDuCoin, margeDuCoin
	case ouest:
		return margeDuCoin, loinV
	case est:
		return loinU, margeDuCoin
	case sud:
		return loinU, loinV
	case porte:
		for _, pas := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			if u, v := sortie.U+pas[0], sortie.V+pas[1]; g.Passable(u, v) {
				return u, v
			}
		}
		return sortie.U, sortie.V
	case murNordOuest:
		return margeDuCoin, g.Height() / 2
	case murNordEst:
		return g.Width() / 2, margeDuCoin
	case murSudEst:
		return g.Width() - 1 - margeDuCoin, g.Height() / 2
	case murSudOuest:
		return g.Width() / 2, g.Height() - 1 - margeDuCoin
	default:
		return g.Width() / 2, g.Height() / 2
	}
}

// versLeMur rend la direction dans laquelle un repère d'arête pousse, et faux
// pour tous les autres.
//
// **La position finale se joue, elle ne s'écrit pas.** Le joueur s'arrête à
// vingt-cinq millièmes de tuile du mur, et poser ce nombre ici en ferait une
// seconde description de ce que `projeter` décide — celle qui mentirait au
// premier réglage de vitesse.
func (r repere) versLeMur() (game.Vec, bool) {
	switch r {
	case murNordOuest:
		return game.Vec{X: -game.One}, true
	case murNordEst:
		return game.Vec{Y: -game.One}, true
	case murSudEst:
		return game.Vec{X: game.One}, true
	case murSudOuest:
		return game.Vec{Y: game.One}, true
	default:
		return game.Vec{}, false
	}
}

// ticksDeContact est ce qu'on joue pour venir buter contre le mur.
//
// Deux cases de marge à cinq tuiles par seconde en demandent vingt-quatre ; la
// moitié en plus laisse la place à un réglage de vitesse sans que la vue cesse
// de montrer un contact.
const ticksDeContact = 36

// dureeDeRupture est le temps au bout duquel une épave est posée, en ticks.
//
// **Elle doit couvrir le cycle de rupture**, que le manifeste des objets déclare
// en trois images de quatre-vingt-dix millisecondes, soit seize ticks. Trente
// laissent la marge d'un réglage de cadence sans que la vue cesse de montrer une
// épave — en deçà, elle donnerait à relire une caisse en train d'éclater.
const dureeDeRupture game.Tick = 30

// vue est une scène à écrire : où poser le joueur, combien de pas jouer avant de
// dessiner, ce qu'on pose par-dessus, et le nom du fichier qui en sort.
type vue struct {
	nom   string
	ou    repere
	ticks int
	// videLaHorde retire les créatures à chaque pas.
	//
	// **À chaque pas et non une fois avant**, depuis que le spawner existe : la
	// pression rachète au tick suivant ce qu'on vient de retirer, et une salle
	// vidée d'un coup se repeuple pendant les pas joués.
	//
	// **Ce que ce dégagement contourne a changé de nature.** Il servait à un semis
	// qui rendait toute montée de niveau inatteignable ; la courbe de pression a
	// levé cela, et une partie jouée atteint le niveau deux avant la minute. Ce
	// qui reste est que la planche joue avec une direction nulle : un joueur
	// immobile meurt à la vingtième seconde, quand le plancher de temps en demande
	// quarante-cinq. Le lever demanderait de donner une trajectoire à ces pas, et
	// une trajectoire de fuite dépend de l'équilibrage que le lot du réglage va
	// changer — la vue mesurerait alors autre chose à chaque retouche.
	//
	// La salle vide reste la salle du jeu, dessinée par son écran : ce que la
	// planche met de côté est ce qui empêche d'y arriver, pas ce qu'elle donne à
	// juger.
	videLaHorde bool
	// semeDesGemmes pose une rangée de gemmes d'âges échelonnés, hors de portée
	// du joueur.
	//
	// **L'extinction ne se juge pas autrement.** Une gemme vit six secondes et se
	// ramasse dès qu'on la frôle : dans une partie jouée, aucune ne reste assez
	// longtemps au sol pour qu'on voie sa teinte descendre, et celles qui y
	// restent ont toutes le même âge. Il faut donc les poser, comme la vue de
	// mêlée doit jouer des pas pour que la horde soit arrivée.
	semeDesGemmes bool
	// chargeLAimant donne une charge sans la dépenser, pour que l'emplacement du
	// bandeau se juge plein.
	chargeLAimant bool
	// poseDesArmes met une arme lourde dans un emplacement et une autre au sol.
	//
	// Sans passer par une caisse : la chute est un tirage, et une vue qui en
	// dépendrait montrerait un sol vide une fois sur trois.
	poseDesArmes bool
	// declencheUneArme dépense une charge de l'emplacement tenu, pour que la
	// déflagration porte et que ses chiffres jaillissent.
	declencheUneArme bool
	// surUneCaisse pose le joueur contre la première caisse du lieu, à portée de
	// contact, si bien que l'appui commence au premier pas.
	//
	// **Contre et non dessus, et c'est la vue de l'appui qui l'a montré.** Posé
	// au centre de la case, le joueur recouvre exactement ce qu'on venait relire :
	// son sprite fait deux fois la caisse, et la déformation disparaissait sous
	// lui. Il est donc décalé du côté d'où l'on vient, ce qui met la caisse devant
	// lui dans le tri en profondeur — et c'est aussi ce qu'une partie produit,
	// puisqu'on casse en arrivant dessus et non en s'y tenant.
	//
	// **Elle retient aussi le pilote, et les deux ne se séparent pas.** L'appui
	// ne va au bout que si l'on reste : le tour d'octogone emportait le joueur au
	// deuxième pas, et la caisse finissait par céder onze secondes plus tard, à
	// un passage que rien n'avait voulu. Un second drapeau aurait laissé écrire
	// la moitié qui ne marche pas.
	surUneCaisse bool
	// loinDUneCaisse pose le joueur à trois tuiles d'elle, hors de portée de
	// contact.
	//
	// Ce qu'elle donne à relire est la décision d'y aller — la caisse intacte et
	// l'icône qui dit ce qu'elle porte —, jamais ce qui se passe une fois dessus.
	// C'est aussi la seule distance à laquelle l'annonce sert : au contact, on a
	// déjà décidé.
	loinDUneCaisse bool
	// poseUneBaudruche en fait apparaître une au pied du joueur, que son arme
	// abat aussitôt : c'est le seul chemin qui produise une déflagration sans
	// forger un état que la partie ne connaît pas.
	poseUneBaudruche bool
	// declencheLAimant lance la ruée après les pas joués, et dessine l'image en
	// plein vol.
	//
	// La charge est donnée plutôt que ramassée : ce que cette vue montre est la
	// convergence, et jouer les trente secondes qu'une apparition demande y
	// ajouterait une horde qui a tué le joueur entre-temps.
	declencheLAimant bool
	// Les trois arrêts par événement. Le nombre de ticks devient alors un
	// plafond : il tient la planche déterministe même si l'équilibrage change ce
	// qu'il faut de temps pour y arriver.
	//
	// **Une vue qui s'arrête à un compte relit l'équilibrage du jour où on l'a
	// écrit.** Les trois vues jouées ont toutes fini par mentir de cette façon :
	// quatre secondes suffisaient à une horde semée d'un bloc et ne suffisent plus
	// à une pression qui monte, vingt secondes tuaient un joueur immobile que la
	// courbe réglée laisse vivre trois minutes et demie. Ce qu'elles attendent est
	// un état, alors elles l'attendent.
	//
	// jusquAuChoix s'arrête au premier panneau de montée de niveau.
	jusquAuChoix bool
	// jusquAuxDegats s'arrête quand la horde a coûté un quart de la vie. La
	// première touche serait trop tôt : ce que la mêlée juge est le personnage
	// entouré, pas le premier contact.
	jusquAuxDegats bool
	// jusquAuTir s'arrête sur un tir du joueur détaché de son canon.
	//
	// **À distance et non à l'émission.** Un projectile qui vient de partir a sa
	// traînée sous le personnage, qui la recouvre : ce que la vue donne à relire
	// est le trait entre le tireur et ce qu'il vise, jamais l'instant du départ.
	// C'est la règle de l'effet à mi-vie, sur un objet qui dure plus longtemps.
	jusquAuTir bool
	// jusquAlEffet s'arrête au premier effet bref d'une sorte donnée.
	//
	// **Un effet dure une fraction de seconde et rien ne le rejoue** : une vue
	// qui s'arrêterait à un compte de pas ne l'attraperait qu'avec de la chance,
	// et cesserait de l'attraper au premier réglage de durée. C'est le cas type
	// de l'artefact qui doit attendre son événement.
	jusquAlEffet *game.FxKind
	// jusquAMiAppui s'arrête quand une caisse a encaissé la moitié de son délai.
	//
	// **À mi-appui pour la même raison qu'un effet se relit à mi-vie** : au
	// premier tick la déformation n'a pas commencé, au dernier la caisse a déjà
	// cédé. Ce que la vue doit donner à relire est l'écrasement en cours, qui est
	// le seul signal disant au joueur qu'il casse quelque chose plutôt qu'il bute
	// sur un mur.
	jusquAMiAppui bool
	// jusquALepave s'arrête sur une épave dont la rupture est finie.
	//
	// **Elle attend un âge et non un compte de pas de la vue**, ce qui la garde
	// juste si le délai d'appui change : ce qui compte est le temps écoulé depuis
	// que la caisse a cédé. Le seuil doit couvrir le cycle de rupture, que le
	// manifeste des objets déclare en trois images de quatre-vingt-dix
	// millisecondes — le jour où cette bande s'allonge, ce nombre suit, sinon la
	// vue montre une caisse en train d'éclater là où elle promet une épave.
	jusquALepave bool
	// jusquAuDanger s'arrête au franchissement du seuil d'alerte.
	//
	// Ni la mêlée ni la mort ne montrent la vignette : la première s'arrête à
	// trois quarts de vie, bien au-dessus du seuil, et la seconde au dernier
	// point, quand l'alerte s'est déjà tue au profit de l'écran de mort. Un
	// signal qu'aucune vue ne produit ne se relit pas.
	jusquAuDanger bool
	// jusquALaMort s'arrête au dernier point de vie.
	jusquALaMort bool
	// texte pose les échantillons nus, pour juger la police seule ; hud pose
	// l'interface, pour juger ce qui l'entoure. Les deux sont séparés parce
	// qu'un cadre sous un texte change ce qu'on lit de la police.
	texte bool
	hud   bool
}

// vues énumère ce que la planche donne à relire.
//
// Le centre montre la projection sans rien qui la borde ; les quatre suivantes
// posent le joueur près d'un coin du losange, là où la caméra bute et où le
// joueur se décentre. Ce sont les seuls endroits où le cadrage décide de quelque
// chose, et ils couvrent les deux axes dans les deux sens.
//
// **La mêlée est la seule qui juge l'exception du joueur**, et c'est pourquoi
// elle joue des pas : il faut qu'une créature soit arrivée pour qu'on voie si le
// personnage reste devant ce qui l'entoure. Une horde qui approche encore ne le
// dirait pas — elle est derrière lui ou devant lui, jamais autour.
//
// **Elle s'arrête quand la horde a coûté un quart de la vie**, et non à un
// compte de secondes. Ce n'est pas un raffinement : la courbe s'ouvre à une
// pression d'un par seconde, si bien que le moment où le joueur est entouré
// dépend d'un réglage qui bougera encore. Ce que la vue attend est un état.
var vues = []vue{
	{nom: "centre"},
	{nom: "nord", ou: nord},
	{nom: "ouest", ou: ouest},
	{nom: "est", ou: est},
	{nom: "sud", ou: sud},

	// **La porte fermée est le cas qu'une partie jouée a raté.** Elle est une
	// case de l'enceinte, teintée plutôt que remplacée : ce que cette vue juge
	// est l'écart entre elle et le mur qui l'entoure, sur toute la longueur du
	// bord. L'état ouvert demanderait cent créatures abattues, et c'est le
	// fermé qui décide — une porte qu'on ne trouve pas ne s'ouvre jamais.
	{nom: "porte", ou: porte},

	// **Les quatre bords, joueur appuyé contre l'enceinte.** Ce qui s'y relit
	// n'est pas le cadrage mais le personnage lui-même : deux des arêtes le
	// placent devant le mur, les deux autres derrière, et un mur de quatre-vingt-
	// seize pixels le couvre alors en entier. C'est là que la silhouette décide
	// de ce qu'on lit — « derrière le mur » ou « dessus ».
	{nom: "mur-nord-ouest", ou: murNordOuest},
	{nom: "mur-nord-est", ou: murNordEst},
	{nom: "mur-sud-est", ou: murSudEst},
	{nom: "mur-sud-ouest", ou: murSudOuest},

	{nom: "melee", ticks: 300 * game.TPS, jusquAuxDegats: true},

	// **Le tir du joueur n'avait aucune vue**, et c'est ce qui l'a laissé
	// invisible jusqu'à ce qu'une partie jouée le signale : les projectiles sont
	// à l'image sur près d'un quart d'une partie, ils portent et ils tuent, et
	// rien ne les donnait à relire. La horde reste, puisque l'arme ne part que
	// s'il y a une cible à portée.
	{nom: "tirs", ticks: 300 * game.TPS, jusquAuTir: true},

	// **Les armes lourdes, tenues et au sol.** Ce que cette vue relit est ce
	// qu'aucune mesure ne dit : que l'icône tienne dans sa case, que les pastilles
	// se comptent d'un coup d'œil, et qu'une arme posée au sol se voie parmi le
	// décor. La horde est retirée pour que rien ne passe devant.
	{nom: "armes", videLaHorde: true, poseDesArmes: true},

	// **Les chiffres de dégâts, et la graduation qui les distingue.** Ce qui se
	// relit ici est ce qu'aucune mesure ne dit : qu'un « 1 » de tir de base reste
	// discret dans une horde dense, et qu'un « 6 » de grenade s'en détache. La
	// horde reste, puisque c'est la masse qui rend la question intéressante — un
	// chiffre seul sur un sol vide serait toujours lisible.
	{nom: "degats", ticks: 300 * game.TPS, poseDesArmes: true,
		declencheUneArme: true, jusquAlEffet: &effetChiffre},

	// La vignette de danger, qui ne se juge que sur ce qu'elle laisse voir : la
	// horde doit rester lisible au centre, sans quoi le signal coûte la fuite
	// qu'il réclame. C'est ce qu'un aplat plein écran avait fait perdre.
	{nom: "danger", ticks: 600 * game.TPS, jusquAuDanger: true},
	{nom: "texte", texte: true},
	{nom: "interface", hud: true},

	// **Le choix ne se pose pas non plus, il s'obtient** — comme la mort, et pour
	// la même raison : une maquette de cartes divergerait du panneau que le jeu
	// peint dès le premier déplacement de colonne. La horde est retirée parce
	// qu'elle rend la montée inatteignable, et c'est alors le plancher de temps
	// qui la donne. Le plafond de ticks est large : c'est le choix qui arrête les
	// pas, pas le compte, si bien que la vue tient encore quand le plancher change.
	{nom: "cartes", ticks: 90 * game.TPS,
		videLaHorde: true, jusquAuChoix: true},

	// **L'extinction se juge sur une rangée, pas sur une gemme.** Ce qu'on
	// regarde n'est pas une teinte mais un écart : deux âges voisins doivent se
	// distinguer, sinon l'information continue que l'effacement promet n'arrive
	// pas au joueur. Une seule gemme, si pâle soit-elle, ne dirait rien de ça.
	{nom: "gemmes", videLaHorde: true, semeDesGemmes: true,
		chargeLAimant: true},

	// **La ruée est ce qui juge une anticipation du projet.** Les gemmes sont
	// entrées dans la séquence de tri en profondeur parce qu'on prévoyait que
	// l'aimant les ferait traverser la horde à hauteur de torse ; rien ne l'avait
	// vérifié. Cette vue est le premier endroit où ça se voit : la horde reste,
	// les gemmes sont semées au loin, et l'aimant est déclenché.
	//
	// Elle attend la même arrivée que la mêlée, et pour la même raison : des
	// gemmes qui traverseraient une salle vide ne diraient rien.
	{nom: "ruee", ticks: 300 * game.TPS, jusquAuxDegats: true,
		semeDesGemmes: true, declencheLAimant: true},

	// **Les deux effets brefs s'attendent, ils ne se posent pas.** Une caisse
	// cassée crache ses éclats sur un demi-tick et une déflagration son onde sur
	// trois dixièmes : les poser à la main donnerait un état qu'aucune partie ne
	// produit, et les attendre à un compte de pas relirait le réglage du jour.
	//
	// La volée s'obtient en posant le joueur sur une caisse, qui cède au premier
	// contact. L'onde demande une Baudruche, que la vue fait apparaître au pied du
	// joueur : l'arme la prend pour cible la plus proche, la mèche brûle, et
	// l'explosion part. La horde est retirée dans les deux cas — ce qui est jugé
	// est un effet, pas ce qui l'entoure.
	{nom: "eclats", ticks: 30 * game.TPS, videLaHorde: true,
		surUneCaisse: true, jusquAlEffet: &effetCaisse},

	// **Les deux moments d'une caisse qui cède, que rien ne montrait.** Le délai
	// d'appui n'a de sens que s'il se voit : sans la déformation, un joueur
	// ralenti sur une caisse croit avoir buté sur un mur, et c'est ce que la
	// première donne à relire. La seconde montre ce qui reste — l'épave que le
	// cycle de rupture laisse au sol, et qui dit que la salle a été fouillée ici.
	//
	// La horde est retirée dans les deux cas : ce qui est jugé est un dessin, pas
	// ce qui l'entoure.
	// **La seule distance à laquelle l'annonce sert.** Au contact, le joueur a
	// déjà décidé d'y aller : ce que l'icône doit faire est se lire de loin, au
	// moment où il choisit entre le détour et la horde.
	{nom: "caisse-annonce", ticks: 2, videLaHorde: true, loinDUneCaisse: true},
	{nom: "caisse-appui", ticks: 30 * game.TPS, videLaHorde: true,
		surUneCaisse: true, jusquAMiAppui: true},
	{nom: "caisse-epave", ticks: 30 * game.TPS, videLaHorde: true,
		surUneCaisse: true, jusquALepave: true},
	// **Celle-ci ne vide pas la horde**, contrairement à sa voisine : le
	// dégagement tourne à chaque pas, et il emportait la Baudruche avant que
	// l'arme ait eu le temps de l'abattre. Les premières secondes de la courbe
	// n'achètent presque rien, si bien que la posée reste la cible la plus
	// proche sans qu'on ait à retirer quoi que ce soit.
	{nom: "souffle", ticks: 30 * game.TPS,
		poseUneBaudruche: true, jusquAlEffet: &effetSouffle},

	// La mort ne se pose pas, elle s'obtient : le joueur reste immobile au milieu
	// du lieu et la horde finit par l'avoir. Dix minutes de plafond, parce que la
	// courbe réglée laisse un joueur immobile vivre trois minutes et demie et que
	// ce chiffre bougera encore — la vue attend l'événement, pas son heure.
	{nom: "mort", ticks: 600 * game.TPS, jusquALaMort: true},
}

// echantillons sont les chaînes que la vue de texte affiche.
//
// **Ce sont des échantillons de mesure, et ils le restent.** Cette note
// promettait qu'ils disparaîtraient le jour où le jeu aurait ses vrais libellés ;
// il les a, et « Espace pour relancer » diverge déjà de ce que l'écran de mort
// affiche. Ce n'est pas une dette : ce que la vue juge est une police, donc il
// lui faut un texte qui couvre ce qu'elle doit rendre — accents, tiret cadratin,
// pourcentage, espace insécable —, et non ce que le jeu se trouve écrire.
//
// Les espaces insécables y sont écrites par leur code : posées en littéral,
// elles se confondent avec des espaces ordinaires dans le source. Elles sont ce
// que le français impose devant un pourcentage et entre les milliers, et la vue
// est le premier usage réel du glyphe que la table déclare.
var echantillons = []string{
	"Niveau 5 " + string(rune(0x2014)) + " choisissez une amélioration",
	"Rafale",
	"+1 projectile",
	"Trois projectiles au lieu de deux.",
	"Cadence +15" + string(rune(0x00A0)) + "%",
	"Portée +2 tuiles",
	"L'arme tire plus souvent.",
	"Espace pour relancer",
	"ÀÉÈÊÇÎÔÙŸŒÆ «»",
}

// planche écrit toutes les vues au premier pas, puis demande l'arrêt.
//
// Elle est un jeu Ebitengine sans l'être : le dessin et la lecture de pixels ne
// sont valides que dans le cycle de la bibliothèque, si bien qu'un programme qui
// écrirait ses images depuis `main` n'obtiendrait rien. D'où ce détour, dont
// `Draw` reste vide — la fenêtre qui s'ouvre une fraction de seconde n'a rien à
// montrer.
//
// **Elle n'appelle jamais `render.Screen.Update`, et c'est délibéré.** Cette
// méthode lit les touches : une direction pressée pendant l'écriture déplacerait
// le joueur, et les images cesseraient d'être comparables d'une exécution à
// l'autre — le déterminisme reposerait sur la précaution de ne pas toucher au
// clavier plutôt que sur une propriété. Ne pas l'appeler n'est pas une garde
// qu'on pourrait contourner, c'est un chemin qui n'existe pas.
//
// Ce qu'elle en perd est nul : monter un écran cadre déjà sur le joueur, donc un
// écran monté après la scène est cadré juste. Ce qui doit avancer d'un pas passe
// par `World.Step`, où la direction est écrite et non lue.
//
// **Chaque vue monte sa propre partie**, et n'hérite donc pas de ce que les
// précédentes ont joué. Une vue qui avance de quatre secondes laisserait sinon
// une horde déplacée aux suivantes, et l'ordre de la table déciderait de ce
// qu'on voit — ce qui est exactement le genre de dépendance cachée qu'une
// planche ne doit pas avoir. Le montage coûte quelques millisecondes.
type planche struct {
	tampon *ebiten.Image
	// agrandi porte le tampon multiplié par `echelle`, et c'est lui qu'on écrit.
	agrandi *ebiten.Image
	hud     *render.HUD
	// tuiles est le catalogue des formes du décor, lu une fois : il ne dépend
	// que du manifeste, là où le terrain d'une vue dépend du lieu qu'elle monte.
	// Le décoder par vue coûterait soixante et une images onze fois.
	tuiles *sprite.Tileset
	// troupe est celle de toutes les vues : les bandes ne dépendent que du
	// manifeste des personnages, là où la partie d'une vue lui est propre.
	troupe *render.Cast
	// objets est le catalogue des ramassables et des projectiles, lu une fois
	// pour la même raison que la troupe : il ne dépend que de son manifeste.
	objets *render.Stage
	ecrit  bool
}

// Update écrit les vues, puis rend la fin de partie.
func (p *planche) Update() error {
	if p.ecrit {
		return ebiten.Termination
	}
	for _, v := range vues {
		if err := p.vue(v); err != nil {
			return err
		}
	}
	p.ecrit = true
	return nil
}

// Draw ne dessine rien : la fenêtre n'est ouverte que pour le contexte
// graphique.
func (p *planche) Draw(*ebiten.Image) {}

// Layout donne au tampon la taille de celui du jeu.
func (p *planche) Layout(_, _ int) (int, int) { return render.Width, render.Height }

// vue monte une partie, y pose la scène, la dessine et écrit le fichier.
//
// Le joueur est posé au centre de sa case et non sur son coin, faute de quoi la
// planche montrerait un cas qu'aucune partie ne produit. Les pas se jouent
// ensuite, avec une direction nulle : la horde avance, le joueur ne bouge pas, et
// rien de ce qui se passe ne dépend de ce qu'on presse. L'écran vient en dernier,
// puisque c'est son montage qui cadre.
func (p *planche) vue(v vue) error {
	partie, err := session.Open(cohue.Assets, cohue.StartingCampaign, graine)
	if err != nil {
		return err
	}
	pu, pv := v.ou.cases(partie.Grid, partie.World.Exit())
	partie.World.Place(game.FromInt(pu)+game.One/2, game.FromInt(pv)+game.One/2)
	if v.surUneCaisse || v.loinDUneCaisse {
		c, err := caisseGarnie(partie.World, v.nom)
		if err != nil {
			return err
		}
		// Au contact : trois cinquièmes de tuile sur chaque axe, soit un écart de
		// 0,85 — sous la portée que la progression déclare, et assez pour que la
		// caisse se peigne devant le personnage. À distance : trois tuiles, très
		// au-delà de cette portée, avec la caisse du même côté pour la même raison.
		ecart := game.One * 3 / 5
		if v.loinDUneCaisse {
			ecart = game.FromInt(3)
		}
		partie.World.Place(c.X-ecart, c.Y-ecart)
	}
	if v.poseUneBaudruche {
		if err := poserUneBaudruche(partie); err != nil {
			return err
		}
	}
	// Le contact avant les pas joués : c'est la position de départ de la vue, et
	// non un état qu'on atteindrait en chemin.
	if vers, contre := v.ou.versLeMur(); contre {
		for range ticksDeContact {
			partie.World.Step(vers)
		}
	}
	// **La mort arrête les pas dès qu'une vue en dépend.** `World.Step` continue
	// de tourner après elle — c'est l'écran qui fige, et la planche l'appelle
	// directement —, si bien qu'un cadavre continue de tirer et de ramasser. La
	// vue du choix montrait ainsi un niveau gagné trente secondes après la mort.
	for tick := range v.ticks {
		if v.arrive(partie.World) {
			break
		}

		// **Le jeu n'avance pas pendant qu'un choix est ouvert, et la planche non
		// plus.** Elle appelle `Step` directement, là où l'écran l'arrête : sans
		// cette prise, une vue qui joue plus d'une minute finit avec un panneau de
		// cartes ouvert par-dessus la scène qu'elle vient montrer, et une horde
		// qui a continué de converger pendant une pause qui n'en était pas une.
		// C'est arrivé dès que la courbe a rendu la montée de niveau atteignable.
		if partie.World.Choosing() {
			partie.World.Choose(0)
			continue
		}
		if v.videLaHorde {
			degagerLaHorde(partie.World)
		}

		// **Le personnage tourne au lieu de rester planté.** Immobile, il mourait
		// à 1:28 — avant l'entrée du deuxième profil —, si bien qu'aucune planche
		// ne pouvait montrer une charge, une meute superposée, un souffle amorcé
		// ni un éclair de soin. Le pilote est celui du test de déterminisme, pour
		// que les deux relectures voient la même chose.
		if v.surUneCaisse {
			partie.World.Step(game.Vec{})
			continue
		}
		partie.World.Step(session.Pilot(game.Tick(tick)))
	}
	if v.videLaHorde {
		degagerLaHorde(partie.World)
	}

	// **Les gemmes se sèment après les pas, jamais avant.** Elles vivent six
	// secondes : semées d'abord, celles des vues qui jouent des pas se seraient
	// éteintes avant qu'on dessine, et la vue de la ruée montrait un sol vide.
	if v.semeDesGemmes {
		semerDesAges(partie.World, pu, pv)
	}
	if v.chargeLAimant {
		partie.World.Charge()
	}

	// **Une tenue et une au sol**, parce que ce sont deux choses distinctes à
	// relire : l'emplacement avec son icône et ses pastilles, et le dessin qui
	// attend d'être ramassé. Une seule des deux ne montrerait que la moitié de ce
	// que le lot a ajouté — et c'est l'absence d'une vue de ce genre qui a laissé
	// livrer une arme au sol que rien ne dessinait.
	if v.poseDesArmes {
		px, py := partie.World.Player()
		// La première est ramassée par le tick qui suit, la seconde reste au sol :
		// une pose et un tick suffisent, et un échec ne s'annonce pas — la vue
		// montrerait alors une case vide, ce qui se voit.
		partie.World.SpawnDrop("grenade", px, py)
		partie.World.Step(game.Vec{})
		partie.World.SpawnDrop("grenade", px+game.FromInt(2), py)
	}

	// **Après les pas, parce que la déflagration ne dure qu'une mèche.**
	// Déclenchée avant, elle aurait détoné pendant la course et ses chiffres
	// seraient éteints à l'image. Les ticks joués ici sont ceux de la mèche, plus
	// de quoi laisser les chiffres monter.
	if v.declencheUneArme {
		attendreUnGrosChiffre(partie.World)
	}

	// La ruée se dessine en plein vol : quelques ticks suffisent à ce que les
	// gemmes soient parties sans être arrivées, ce qui est le seul état où la
	// convergence se juge.
	if v.declencheLAimant {
		partie.World.Charge()
		partie.World.Attract()
		for range game.TPS / 4 {
			partie.World.Step(game.Vec{})
		}
	}

	sol, err := render.NewTerrain(partie.Tiles, p.tuiles)
	if err != nil {
		return err
	}
	render.NewScreen(partie.World, partie.Grid, sol, p.troupe, p.objets).
		WithHUD(p.hud).Draw(p.tampon)
	if v.texte {
		p.poser()
	}
	if v.hud {
		p.poserInterface()
	}

	// L'agrandissement se fait au plus proche voisin, qui est le filtre par
	// défaut : un lissage rendrait la planche inutilisable pour juger du pixel
	// art, qui est précisément ce qu'elle donne à relire.
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(echelle, echelle)
	p.agrandi.DrawImage(p.tampon, &op)

	chemin := filepath.Join(sortie, v.nom+".png")

	// Le chemin n'a aucune part variable : `sortie` est une constante et le nom
	// vient de la table ci-dessus. `os.Root` ne fermerait donc rien qui soit
	// ouvert, et l'ajouter pour taire l'avertissement mettrait un mécanisme là
	// où il n'y a pas de question.
	f, err := os.OpenFile(chemin, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304
	if err != nil {
		return err
	}

	// La fermeture est signalée, mais elle ne masque pas l'écriture : un encodage
	// qui échoue dit ce qui ne va pas, là où la fermeture d'un fichier déjà
	// fautif ne dirait que le symptôme.
	err = png.Encode(f, p.agrandi)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("%s: %w", chemin, err)
	}
	fmt.Println(chemin)
	return nil
}

// poser écrit les échantillons sur la scène déjà dessinée.
//
// Trois situations, et ce sont elles qui font la vue : une colonne de texte nu
// sur le décor, deux chiffres contourés posés là où le fond est le plus clair,
// et deux lignes alignées sur le bord droit par mesure plutôt qu'à distance
// fixe. La troisième est ce qui manquait à la planche qui a fait écrire la règle
// — un minuteur placé au jugé déborde dès qu'il s'allonge.
func (p *planche) poser() {
	h := p.hud
	y := 12
	for _, s := range echantillons {
		h.Font.Draw(p.tampon, s, 12, y, h.Color("texte"))
		y += h.Font.Height() + 4
	}

	// Sur le monde et non sur un cadre : c'est le cas que la conception vise
	// quand elle exige un contour.
	//
	// **Le fond le plus hostile n'est pas à l'image, et ne peut pas y être
	// aujourd'hui.** Le contour existe pour un chiffre posé sur du décor clair,
	// or le rendu provisoire ne peint que trois gris et un bleu : le cas se
	// jugera quand les sprites entreront, à l'étape 5. Ce que la vue montre est
	// que le contour ne nuit pas sur fond moyen, pas qu'il suffit sur fond clair.
	contour := h.Color("texte_contour")
	h.Font.DrawOutlined(p.tampon, "247", render.Width/2-40, render.Height/2-60,
		h.Color("texte_valeur"), contour)
	h.Font.DrawOutlined(p.tampon, "12", render.Width/2+30, render.Height/2-20,
		h.Color("texte"), contour)

	for i, s := range []string{"07:41", "1" + string(rune(0x00A0)) + "340 pts"} {
		x := render.Width - 12 - h.Font.Advance(s)
		h.Font.Draw(p.tampon, s, x, 12+i*(h.Font.Height()+2), h.Color("texte"))
	}
}

// poserInterface compose l'écran de jeu à partir des primitives.
//
// **Chaque élément se dimensionne sur son contenu**, jamais sur une constante :
// la carte prend la largeur de sa plus longue ligne, la case le côté de ce
// qu'elle contient, le minuteur sa place mesurée depuis le bord droit. C'est ce
// que la planche doit donner à juger — si une dimension y était écrite, on
// jugerait le chiffre plutôt que la règle qui le produit.
func (p *planche) poserInterface() {
	h := p.hud
	marge, hauteur := h.Margin(), h.Font.Height()

	// Le bandeau n'est pas remaquetté ici : `Screen.Draw` vient de le poser, avec
	// les valeurs de la partie montée. Une maquette qui le doublerait
	// superposerait deux jeux de chiffres — c'est ce qui est arrivé, et l'image
	// l'a montré tout de suite : les glyphes se chevauchaient au point de rendre
	// « 62 / 100 » illisible. Les vues qui le jugent sur des valeurs parlantes
	// sont `melee` et `mort`, qui jouent assez de ticks pour cela.

	// Les emplacements, sous les jauges. Le contenu vaut deux lignes faute
	// d'icône : elles viendront à l'étape 5, et c'est leur taille réelle qui
	// fixera le côté de la case — d'où un côté calculé plutôt que réglé.
	x := 12
	for _, touche := range []string{"1", "2"} {
		x += h.Slot(p.tampon, x, 62, hauteur*2, touche) + marge
	}

	// Le score, sous le minuteur du bandeau. Il n'est pas encore une lecture du
	// jeu — rien ne le compte —, et la planche le porte pour que sa place soit
	// jugée avec le reste.
	score := "1" + string(rune(0x00A0)) + "340 pts"
	h.Font.Draw(p.tampon, score, render.Width-12-h.Font.Advance(score),
		12+hauteur+2, h.Color("texte"))

	// Le panneau de choix n'est pas remaquetté : la vue `cartes` montre celui que
	// le jeu peint, sur une partie où le plancher de temps l'a ouvert. Une
	// maquette qui doublerait la mise en page livrée resterait juste jusqu'au
	// premier déplacement de colonne, puis montrerait un écran que personne ne
	// joue — c'est ce qui était arrivé au bandeau.
}

// main écrit la planche, ou sort en échec.
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run ouvre la fenêtre que le dessin exige, puis lance la planche.
//
// La partie, elle, se monte une fois par vue : voir la godoc de `planche`.
func run() error {
	if err := os.MkdirAll(sortie, 0o750); err != nil {
		return err
	}

	hud, err := render.LoadHUD(cohue.Assets, cohue.InterfaceManifest)
	if err != nil {
		return err
	}

	// Le manifeste de décor est relu ici plutôt que pris sur une partie : le
	// catalogue de formes n'en dépend pas, et en monter une pour l'obtenir
	// laisserait le terrain de toutes les vues accroché à un lieu qu'on jette.
	decor, err := level.LoadDecor(cohue.Assets, cohue.DecorManifest)
	if err != nil {
		return err
	}
	tuiles, err := sprite.LoadTiles(cohue.Assets, cohue.DecorDir, decor)
	if err != nil {
		return err
	}
	profils, err := game.LoadProfiles(cohue.Assets, cohue.CharacterManifest)
	if err != nil {
		return err
	}
	troupe, err := render.NewCast(cohue.Assets, cohue.CharacterDir, profils)
	if err != nil {
		return err
	}
	catalogue, err := game.LoadObjects(cohue.Assets, cohue.ObjectManifest)
	if err != nil {
		return err
	}
	objets, err := render.NewStage(cohue.Assets, cohue.ObjectDir, cohue.ObjectManifest, catalogue)
	if err != nil {
		return err
	}

	ebiten.SetWindowTitle("Cohue — planche")
	ebiten.SetWindowSize(render.Width, render.Height)
	return ebiten.RunGame(&planche{
		tampon:  ebiten.NewImage(render.Width, render.Height),
		agrandi: ebiten.NewImage(render.Width*echelle, render.Height*echelle),
		hud:     hud,
		tuiles:  tuiles,
		troupe:  troupe,
		objets:  objets,
	})
}

// Les deux sortes d'effet que les vues attendent, prises en variables parce
// qu'une table d'entrées ne peut pas prendre l'adresse d'une constante.
var (
	effetCaisse  = game.FxCrate
	effetSouffle = game.FxBlast
	effetChiffre = game.FxDamage
)

// poserUneBaudruche en fait apparaître une à deux tuiles du joueur.
//
// **À deux tuiles et non collée** : l'arme vise la plus proche à portée, donc il
// suffit qu'elle soit là pour qu'elle tombe, et l'écarter un peu laisse voir
// l'onde entière plutôt qu'à moitié sous le personnage. La vue attend ensuite
// l'explosion, qui vient quand la mèche a fini de brûler.
func poserUneBaudruche(partie *session.Session) error {
	profil := -1
	for i, p := range partie.Profiles.Enemies {
		if p.Key == "eclateur" {
			profil = i
		}
	}
	if profil < 0 {
		return errors.New("vue souffle : aucun profil « eclateur » au manifeste")
	}
	x, y := partie.World.Player()
	if _, pose := partie.World.SpawnEnemy(profil, x+2*game.One, y); !pose {
		return errors.New("vue souffle : le bassin refuse la Baudruche")
	}
	return nil
}

// semerDesAges pose une rangée de gemmes échelonnées sur toute leur durée de vie.
//
// Hors de portée du joueur, sinon la première passe de ramassage les retirerait
// avant qu'on les voie. Et sur une ligne plutôt qu'en tas : ce qu'il faut juger
// est l'écart entre deux âges voisins, qu'un chevauchement rendrait illisible.
//
// Les gemmes sont posées directement dans le bassin, ce que le jeu ne fait
// jamais — il passe par la mort d'une créature. C'est du même ordre que la horde
// retirée : la planche met de côté ce qui empêche d'arriver à l'état qu'elle
// montre, jamais l'état lui-même.
func semerDesAges(monde *game.World, u, v int) {
	// Le compteur est un `game.Tick` et non un entier, si bien que le calcul de
	// l'âge ne convertit rien : une conversion vers l'entier du compteur de
	// ticks est un débordement possible, et le contrôle de sécurité a raison de
	// le dire même quand les bornes l'excluent ici.
	const rangee game.Tick = 6

	vie := monde.GemLife()
	for i := range rangee {
		monde.Gems().Spawn(game.Gem{
			X: game.FromInt(u+2+int(i)) + game.One/2,
			Y: game.FromInt(v) + game.One/2,
			// Le dernier rang est à un tick de disparaître, pas au-delà : une
			// gemme déjà expirée serait retirée au premier pas et la rangée
			// perdrait l'extrémité qui compte le plus.
			Born: -(vie - 1) * i / (rangee - 1),
		})
	}
}

// degagerLaHorde retire toutes les créatures de la salle.
//
// Appelée avant chaque pas et après le dernier : le spawner rachète à chaque
// tick, et une salle vidée seulement au départ se serait repeuplée avant qu'on
// dessine.
func degagerLaHorde(monde *game.World) {
	horde := monde.Enemies()
	for horde.Len() > 0 {
		horde.RemoveAt(0)
	}
}

// attendreUnGrosChiffre joue jusqu'à ce qu'une déflagration ait porté.
//
// **Elle s'arrête sur l'événement et non sur un compte de pas**, ce que la
// doctrine exige d'un artefact — et les essais qui l'ont précédée montrent
// pourquoi : la mèche d'une demi-seconde et un chiffre d'un tiers ne se
// recouvrent qu'un instant, si bien qu'un compte trop court montre l'emprise
// encore allumée et un compte trop long des chiffres déjà éteints. Aucune valeur
// n'est bonne, et celle qui le paraîtrait cesserait de l'être au premier réglage
// de mèche.
//
// **Le déclenchement se retente à chaque pas**, pour la même raison : une lourde
// ne part que si une cible est à portée, et le pilote tourne à distance de la
// horde. Déclencher une fois puis attendre laissait la charge intacte et la
// planche sans chiffre — ce qui ne se voyait pas, l'emprise n'étant simplement
// jamais apparue.
//
// La borne est large : ce qu'elle garde est l'arrêt de la planche, jamais une
// propriété de la déflagration.
func attendreUnGrosChiffre(monde *game.World) {
	for tick := range 10 * game.TPS {
		effets := monde.Fxs()
		for i := range effets.Active() {
			if e := effets.At(i); e.Kind == game.FxDamage && e.Amount > 1 {
				return
			}
		}
		if monde.Blasts().Len() == 0 {
			monde.Trigger(0)
		}
		monde.Step(session.Pilot(game.Tick(tick)))
	}
}

// caisseGarnie rend la première caisse du lieu qui porte une arme.
//
// **Garnie et non la première venue**, parce qu'une caisse vide n'annonce rien :
// la vue qui juge l'icône montrerait une caisse nue, et celle qui juge l'appui ne
// dirait pas que l'annonce tient pendant qu'on pousse. Une caisse sur deux en
// porte, si bien que le semis livré en a toujours une.
//
// Elle échoue plutôt que de se rabattre sur `At(0)` : un repli rendrait la vue
// muette sans rien dire, et c'est le contrôle privé de son entrée qui doit
// échouer plutôt que passer.
func caisseGarnie(monde *game.World, vue string) (*game.Crate, error) {
	caisses := monde.Crates()
	for i := range caisses.Active() {
		if c := caisses.At(i); c.Weapon != 0 {
			return c, nil
		}
	}
	return nil, fmt.Errorf("vue %s : aucune des %d caisses du lieu ne porte d'arme",
		vue, caisses.Len())
}

// arrive dit si la scène que la vue attend est là.
//
// **La mort arrête toutes les vues jouées, quel que soit ce qu'elles attendent
// par ailleurs.** `World.Step` continue de tourner après elle — c'est l'écran qui
// fige, et la planche l'appelle directement —, si bien qu'un cadavre continue de
// tirer, de ramasser et de monter de niveau. La vue du choix montrait ainsi un
// niveau gagné trente secondes après la mort.
func (v vue) arrive(monde *game.World) bool {
	if !monde.Alive() {
		return true
	}
	switch {
	case v.jusquAuChoix:
		return monde.Choosing()
	case v.jusquAuxDegats:
		return monde.Health() <= monde.MaxHealth()*3/4
	case v.jusquAuDanger:
		return monde.InDanger()
	case v.jusquAuTir:
		x, y := monde.Player()
		joueur := game.Vec{X: x, Y: y}
		tirs := monde.Shots()
		for i := range tirs.Active() {
			p := tirs.At(i)
			if (game.Vec{X: p.X, Y: p.Y}).Sub(joueur).Len() >= game.FromInt(2) {
				return true
			}
		}
	case v.jusquAMiAppui:
		caisses := monde.Crates()
		for i := range caisses.Active() {
			if caisses.At(i).Press*2 <= monde.CratePress() {
				return true
			}
		}
	case v.jusquALepave:
		epaves := monde.Wrecks()
		for i := range epaves.Active() {
			if monde.WreckAge(epaves.At(i)) >= dureeDeRupture {
				return true
			}
		}
	case v.jusquAlEffet != nil:
		// **À mi-vie et non à l'émission.** Au premier tick, les huit éclats sont
		// encore au point de départ et l'onde à sa première image : la vue
		// montrerait l'instant où l'effet commence, c'est-à-dire rien. Ce qu'elle
		// doit donner à relire est la gerbe ouverte, et c'est un état de l'effet
		// — pas un compte de pas depuis sa naissance.
		effets := monde.Fxs()
		for i := range effets.Active() {
			if e := effets.At(i); e.Kind == *v.jusquAlEffet && e.Life*2 <= e.Total {
				return true
			}
		}
	}
	return false
}
