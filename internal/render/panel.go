// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Le bandeau de la partie : la vie, l'expérience et le temps écoulé, composés à
// partir des primitives. Aucune dimension d'élément n'est écrite ici — une
// jauge suit son contenu, un libellé se pose à la mesure du texte.

package render

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/sprimault/cohue/internal/game"
)

// margeEcran est la distance entre un bord du tampon et ce qui s'y accroche.
//
// La marge du thème sépare un contenu de son cadre ; au bord de l'écran il n'y a
// pas de cadre, et la reprendre telle quelle collerait la jauge à l'arête. Douze
// pixels sont un retrait de mise en page, la seule valeur de ce fichier qui ne
// se dérive de rien.
const margeEcran = 12

// largeurJauge est la longueur d'une jauge du bandeau, en pixels.
//
// Plus longue que la vie n'a de points, et c'est ce qui la fixe : le
// remplissage est arrondi vers le bas, si bien qu'une jauge de quatre-vingts
// pixels pour cent points de vie laisserait un point sur cinq ne rien déplacer.
// Cent quarante-huit en donnent une et demie par point.
const largeurJauge = 148

// contenuEmplacement est le côté de ce qu'une case d'emplacement contient.
//
// Il fixe la case, et non l'inverse : `Slot` en dérive son côté en ajoutant la
// marge et le bord. **Vingt pixels, la taille que le manifeste des objets donne
// à une icône d'interface** — le jour où le générateur la change, ce nombre suit.
//
// **Une seule pour toutes les cases, et c'est une correction.** Il y en a eu deux
// — douze pour l'aimant, vingt pour une arme lourde —, et trois endroits en
// dépendaient : les deux tracés, plus la hauteur du bandeau, qui calculait avec
// la petite. Une case d'arme débordait donc du fond de huit pixels et son chiffre
// tombait sur le décor ; les cases se posaient par ailleurs au pas de la grande
// alors que la première occupait la petite, ce qui creusait un trou de douze
// pixels entre les deux premières. Une partie jouée l'a signalé, et la godoc de
// `Slot` l'annonçait sans le fermer : le côté « n'est pas un réglage, il se
// calcule ».
//
// L'aimant pose un aplat de cette taille faute d'icône dessinée — voir
// `emplacement`, qui dit ce qui manque.
const contenuEmplacement = 20

// toucheAimant est ce que le joueur presse pour déclencher sa charge.
//
// **Les chiffres appartiennent aux emplacements**, et ils les gardent toute la
// partie. C'est ce qui a fait passer le choix des cartes aux flèches : une carte
// mal choisie se rattrape au niveau suivant, un aimant déclenché à vide est perdu
// jusqu'à la prochaine apparition, et le coût n'est pas symétrique.
const toucheAimant = "1"

// touchesLourdes sont les chiffres des deux emplacements d'armes lourdes.
//
// Ils suivent celui de l'aimant, qui garde le sien : la conception veut que
// l'aimant ne partage jamais son emplacement, sans quoi il ne serait jamais
// gardé face au soin.
var touchesLourdes = [game.Slots]string{"2", "3"}

// toucheFiole est ce que le joueur presse pour boire.
//
// Elle suit les deux emplacements d'armes, et une seule case porte le stock : ce
// que la conception limite à deux ou trois est le nombre de sortes de
// consommables, pas le nombre de flacons.
const toucheFiole = "4"

// emplacementsLourds résout ce que les emplacements montrent.
//
// **L'icône se résout ici et non dans le bandeau**, qui ne connaît pas le
// catalogue : il pose ce qu'on lui donne. Une arme sans icône dessinée rendrait
// une case vide plutôt qu'un défaut — la conception veut qu'un dessin manquant se
// voie, jamais qu'il arrête le jeu.
func (s *Screen) emplacementsLourds() [game.Slots]Held {
	var tenues [game.Slots]Held
	for place := range tenues {
		arme, charges := s.monde.HeldHeavy(place)
		if charges <= 0 {
			continue
		}
		tenues[place] = Held{Icon: s.objets.Icon(arme.Key), Charges: charges}
	}
	return tenues
}

// Held est ce qu'un emplacement d'arme lourde donne à voir.
//
// **Une icône déjà résolue et non un nom.** Le bandeau ne connaît pas le
// catalogue, donc il ne peut pas savoir quelle icône va où : il pose celle qu'on
// lui donne. C'est ce qui lui permet de ne charger aucune image tout en en
// posant une.
type Held struct {
	// Icon est le dessin de face de l'arme, nul quand l'emplacement est vide.
	Icon *ebiten.Image
	// Charges est le nombre de tirs que l'emplacement porte.
	//
	// **Un compte et non une proportion**, depuis qu'une arme du même type
	// s'ajoute à celle qu'on tient : il n'y a plus de maximum, donc plus de
	// dénominateur. Le champ `Max` qui accompagnait celui-ci est parti avec les
	// pastilles qu'il servait à compter.
	Charges int
}

// Readings est ce que le bandeau montre d'une partie.
//
// Des nombres et non le monde : la planche de relecture compose le bandeau sans
// partie derrière, et une signature qui exigerait un `game.World` l'aurait
// obligée à en monter une pour juger une mise en page.
type Readings struct {
	// Health et MaxHealth sont la vie restante et celle du profil.
	Health, MaxHealth int
	// Level est le niveau atteint, le premier valant un.
	Level int
	// Experience et Threshold sont les gemmes acquises vers le niveau suivant,
	// et ce qu'il coûte.
	Experience, Threshold int
	// Elapsed est l'âge de la partie.
	Elapsed game.Tick
	// Charged dit si le joueur tient un aimant.
	Charged bool
	// Heavies sont les emplacements d'armes lourdes, dans l'ordre des touches.
	Heavies [game.Slots]Held
	// Vials est la case des fioles : leur nombre et leur dessin.
	//
	// La même structure qu'un emplacement d'arme, parce que la case dit la même
	// chose — ce qu'elle tient, et combien. Ce qui les sépare est la touche, qui
	// boit ici et déclenche là.
	Vials Held
	// Mark est l'accusé d'un repère, vide quand il n'y a rien à confirmer.
	Mark string
	// Objective est l'avancement vers l'ouverture de la porte, vide quand le
	// lieu n'a pas de sortie.
	//
	// Une chaîne déjà mise en forme plutôt que deux entiers : le bandeau ne sait
	// pas ce qu'une porte demande — un compteur d'abattus aujourd'hui, un temps
	// ou une élite quand la conception les apportera —, et il n'a pas à
	// l'apprendre pour poser une ligne de texte.
	Objective string
}

// Panel pose les trois lectures : la vie, l'expérience et le temps écoulé.
//
// **Les libellés sont posés à plat, sans contour.** Le contour est fait pour un
// chiffre de dégâts isolé, à qui il ajoute une épaisseur ; sur du texte à cette
// taille il ferme les contre-formes — un zéro devient un carré plein, « 62 / 100 »
// se lit « 620/1000 ». La relecture le montre en une image.
//
// **Ce que le contour aurait couvert, le fond du bandeau le couvre.** La
// question était restée ouverte — un libellé clair sur un sol clair —, et on
// l'avait renvoyée aux sprites faute de pouvoir la juger sur une planche où le
// décor n'a que trois gris. Une partie jouée a tranché en une minute : le texte
// se perdait. Le bandeau a donc son fond, ce que le manifeste prévoyait depuis le
// début sans que personne ne le lise ici.
func (h *HUD) Panel(dst *ebiten.Image, r Readings) {
	x, y := margeEcran, margeEcran
	ecart := 2 * h.Margin()

	// **Le bandeau a son propre fond, comme le panneau des cartes.** Sans lui, la
	// vie et le niveau se posaient à même le décor : le texte disparaissait sur
	// le sol clair, et une jauge à moitié vide se confondait avec la case sous
	// elle. `bandeau_fond` était déclaré au manifeste et lu nulle part ici, ce que
	// personne n'avait vu tant qu'on jugeait la mise en page sur une planche
	// plutôt qu'en jouant.
	h.Band(dst, 0, hauteurBandeau(h))

	h.Gauge(dst, x, y, largeurJauge, part(r.Health, r.MaxHealth), h.Color("jauge_vie"))
	h.libelle(dst, fmt.Sprintf("%d / %d", r.Health, r.MaxHealth),
		x+largeurJauge+ecart, y, h.Color("texte"))

	y += h.Font.Height()
	h.Gauge(dst, x, y, largeurJauge, part(r.Experience, r.Threshold),
		h.Color("jauge_experience"))

	// **Le niveau en pleine teinte, comme la vie.** Il était atténué, ce qui
	// range un texte au second plan : c'est ce qu'on fait d'une phrase
	// d'explication sur une carte, pas d'une des trois lectures que le bandeau
	// existe pour donner.
	h.libelle(dst, fmt.Sprintf("Niveau %d", r.Level),
		x+largeurJauge+ecart, y, h.Color("texte"))

	// Le minuteur s'aligne sur le bord droit par mesure, et non à distance fixe :
	// il s'allonge d'un caractère au passage de la dixième minute.
	temps := minuteur(r.Elapsed)
	h.libelle(dst, temps, Width-margeEcran-h.Font.Advance(temps), margeEcran,
		h.Color("texte"))

	// L'accusé se pose sous le minuteur, dans la teinte atténuée : il confirme
	// une pose, il n'est pas une des lectures que le bandeau existe pour donner,
	// et l'œil qui vérifie l'heure d'un repère est déjà dans ce coin-là.
	if r.Mark != "" {
		h.libelle(dst, r.Mark, Width-margeEcran-h.Font.Advance(r.Mark),
			margeEcran+h.Font.Height(), h.Color("texte_attenue"))
	}

	// **L'objectif se pose sous le minuteur, en pleine teinte.** C'est une des
	// lectures qui décident — savoir qu'il reste trois créatures avant de pouvoir
	// partir change ce qu'on fait de la minute suivante —, pas un accusé de
	// réception. Il descend d'une ligne quand un repère occupe la sienne, plutôt
	// que de la lui disputer.
	if r.Objective != "" {
		ligne := margeEcran + h.Font.Height()
		if r.Mark != "" {
			ligne += h.Font.Height()
		}
		h.libelle(dst, r.Objective, Width-margeEcran-h.Font.Advance(r.Objective),
			ligne, h.Color("texte"))
	}

	bas := y + h.Font.Height() + h.Margin()
	h.emplacement(dst, margeEcran, bas, r.Charged)
	h.lourdes(dst, margeEcran, bas, r)
	h.fioles(dst, margeEcran, bas, r)
}

// fioles pose la case des consommables, après les deux emplacements d'armes.
//
// **Sa place est fixe et ne se tasse pas sur les cases vides.** Les touches
// portent des chiffres, et une case qui glisserait quand une arme s'épuise
// changerait ce qu'il faut presser au moment où le joueur a le moins de temps
// pour le relire.
//
// **Elle n'existe que tenue, comme une arme lourde et non comme l'aimant.** Une
// fiole ne revient pas d'elle-même : elle sort des caisses, si bien qu'une case
// vide n'annoncerait rien qui vienne, à la différence de l'aimant que sa période
// ramène.
func (h *HUD) fioles(dst *ebiten.Image, x, y int, r Readings) {
	if r.Vials.Charges <= 0 {
		return
	}

	cote := h.SlotSide(contenuEmplacement)
	gauche := x + (game.Slots+1)*(cote+h.Margin())
	h.Slot(dst, gauche, y, contenuEmplacement, toucheFiole)

	bord := (cote - contenuEmplacement) / 2
	if r.Vials.Icon != nil {
		h.Image(dst, r.Vials.Icon, gauche+bord, y+bord)
	}
	h.compteDansLaCase(dst, gauche+h.Border()+1, y+cote-h.Border(), r.Vials.Charges)
}

// lourdes pose les emplacements d'armes lourdes à la suite de celui de l'aimant.
//
// **Une case n'existe que tenue**, à l'inverse de celle de l'aimant qui reste
// toujours là : la conception veut qu'à l'épuisement l'interface disparaisse,
// sans message. L'aimant, lui, revient de lui-même toutes les trente secondes, si
// bien qu'une case vide y annonce ce qui va venir ; une arme lourde ne revient
// que si le joueur en trouve une.
//
// La conséquence à assumer est qu'un emplacement vide ne montre pas sa touche.
// C'est ce que « pas de message » veut dire : rien n'invite à presser une touche
// qui ne ferait rien.
func (h *HUD) lourdes(dst *ebiten.Image, x, y int, r Readings) {
	cote := h.SlotSide(contenuEmplacement)
	for place, tenue := range r.Heavies {
		if tenue.Charges <= 0 {
			continue
		}
		gauche := x + (place+1)*(cote+h.Margin())
		h.Slot(dst, gauche, y, contenuEmplacement, touchesLourdes[place])

		bord := (cote - contenuEmplacement) / 2
		if tenue.Icon != nil {
			h.Image(dst, tenue.Icon, gauche+bord, y+bord)
		}
		h.compteDansLaCase(dst, gauche+h.Border()+1, y+cote-h.Border(), tenue.Charges)
	}
}

// compteDansLaCase pose ce qu'un emplacement tient, au coin bas gauche de son
// icône — des tirs pour une arme lourde, des flacons pour une fiole.
//
// **Un compte a remplacé les pastilles, et la conception a suivi plutôt que
// cédé.** Elle voulait une rangée de marques « parce qu'une proportion se lit en
// vision périphérique, un 3/5 demande de regarder » — et une proportion suppose
// un maximum. Une arme du même type s'ajoutant désormais à celle qu'on tient, il
// n'y a plus de dénominateur : ce qui reste est un absolu, et une rangée qui
// s'allongerait sans borne déborderait de sa case au troisième ramassage.
//
// **Dans le coin de la case et non dans celui de l'icône.** Un chiffre fait neuf
// pixels de haut pour une marge de cinq : il déborde donc forcément, et le poser
// au ras du contenu le mettait en plein sur le corps de l'arme. Descendu dans la
// marge, il n'en mord plus que le bas — la région que les quatre dessins
// laissent la plus vide.
//
// **Contouré, parce qu'il se pose malgré tout sur une partie du dessin.** Un
// chiffre clair sur un corps clair ne se lirait pas, et c'est la règle que le
// chapitre 2 applique à tout texte posé sur autre chose que son propre fond.
func (h *HUD) compteDansLaCase(dst *ebiten.Image, x, bas, tenu int) {
	h.Font.DrawOutlined(dst, strconv.Itoa(tenu), x, bas-h.Font.Height(),
		h.Color("texte"), h.Color("cadre_fond"))
}

// emplacement pose la case de l'aimant sous les jauges.
//
// **La case est toujours là, pleine ou vide.** Un emplacement qui n'apparaîtrait
// qu'une fois chargé apprendrait au joueur l'existence de l'objet au moment où il
// le tient déjà, et une case vide est ce qui fait chercher l'aimant dans la
// salle. Ce que la charge change est ce qu'il y a dedans, pas la case.
//
// La touche s'écrit dessous et non le nom, comme la conception l'exige d'un
// emplacement : ce que le joueur cherche sous la case en jouant est ce qu'il doit
// presser, et l'icône dira de quoi il s'agit quand elle existera.
func (h *HUD) emplacement(dst *ebiten.Image, x, y int, chargee bool) {
	cote := h.Slot(dst, x, y, contenuEmplacement, toucheAimant)
	if !chargee {
		return
	}

	// Un aplat centré tient lieu d'icône, dans la teinte de l'objet au sol : sans
	// elle, rien ne dirait que la case et ce qu'on vient de ramasser sont la même
	// chose.
	//
	// **Il devait tomber avec les sprites, et il reste — parce que l'aimant n'a
	// pas d'icône.** Le sprite du monde ne peut pas en tenir lieu : une icône
	// d'emplacement se dessine de face, quand celui-là est en isométrie et porte
	// son ombre au sol. Ce qui manque est donc un dessin, `aimant_icone` de vingt
	// pixels comme ceux des armes, et il vient du générateur. Jusque-là cette
	// teinte est une seconde description de la couleur du sprite, tenue à la
	// main, et c'est ce que la ligne suivante coûte.
	bord := (cote - contenuEmplacement) / 2
	h.Rect(dst, x+bord, y+bord, contenuEmplacement, contenuEmplacement, teinteAimant)
}

// libelle pose un texte du bandeau, aligné sur la jauge qu'il commente.
//
// Le pixel de décalage vers le haut n'est pas un ajustement à l'œil : la jauge
// est haute de six pixels et la ligne de texte de neuf, et les centrer l'une sur
// l'autre demande la moitié de leur différence.
func (h *HUD) libelle(dst *ebiten.Image, texte string, x, y int, teinte color.RGBA) {
	h.Font.Draw(dst, texte, x, y-(h.Font.Height()-h.theme.GaugeHeight)/2, teinte)
}

// part rend la fraction d'une jauge, zéro quand son maximum n'en est pas un.
//
// Le garde-fou n'est pas de la défensive : sans lui, un maximum nul produirait
// un NaN, et le bornage de `Gauge` ne l'arrête pas — toute comparaison avec un
// NaN étant fausse, la jauge afficherait n'importe quelle longueur au lieu de
// rien.
func part(valeur, maximum int) float64 {
	if maximum <= 0 {
		return 0
	}
	return float64(valeur) / float64(maximum)
}

// minuteur rend l'âge d'une partie en minutes et secondes.
//
// **C'est le cas où `Tick.Seconds` est légitime.** Sa godoc interdit de décider
// avec — un seuil, une cadence, un plancher se comptent en ticks —, pas de
// montrer avec. Sans cette phrase, l'avertissement se relit comme une
// interdiction générale, et quelqu'un réinventerait une conversion à côté.
func minuteur(ecoule game.Tick) string {
	secondes := int(ecoule.Seconds())
	return fmt.Sprintf("%02d:%02d", secondes/60, secondes%60)
}

// peindreBandeau pose le bandeau de la partie en cours.
//
// Après les entités et avant le voile de fin : par-dessus le monde, parce
// qu'une créature qui passerait devant la jauge de vie la rendrait illisible au
// pire moment ; sous le voile que `peindreFin` pose, parce que le niveau atteint
// et le temps tenu font partie de ce que le joueur relit à la fin d'une partie —
// c'est pour cela que cet écran-là n'a que deux lignes.
func (s *Screen) peindreBandeau(ecran *ebiten.Image) {
	if s.hud == nil {
		return
	}
	s.hud.Panel(ecran, Readings{
		Health:     s.monde.Health(),
		MaxHealth:  s.monde.MaxHealth(),
		Level:      s.monde.Level(),
		Experience: s.monde.Experience(),
		Threshold:  s.monde.Threshold(),
		Elapsed:    s.monde.Tick(),
		Charged:    s.monde.Charged(),
		Heavies:    s.emplacementsLourds(),
		Vials: Held{
			Icon:    s.objets.Icon(s.monde.VialKey()),
			Charges: s.monde.VialStock(),
		},
		Mark:      s.marque(),
		Objective: s.objectif(),
	})
}

// objectif met en forme l'avancement vers la porte, vide sans sortie.
//
// **Le compte s'arrête à l'objectif au lieu de le dépasser.** Ce que la ligne
// dit est « puis-je partir », pas « combien ai-je tué » : un « 34 / 20 » ferait
// lire un dépassement là où il n'y a qu'une porte ouverte, et le bandeau porte
// déjà le niveau pour qui veut mesurer sa récolte.
func (s *Screen) objectif() string {
	sortie := s.monde.Exit()
	if sortie == nil {
		return ""
	}
	if s.monde.DoorOpen() {
		return "Porte ouverte"
	}
	return fmt.Sprintf("Porte %d / %d", s.monde.Kills(), sortie.Kills)
}

// hauteurBandeau rend ce que le bandeau occupe, mesuré plutôt qu'écrit.
//
// Il descend jusqu'au libellé de touche posé sous l'emplacement, plus une marge :
// une hauteur en dur se serait démentie au premier changement de police ou de
// hauteur de jauge, et c'est exactement ce que ce fichier s'interdit.
//
// **Le côté vient de `SlotSide` et non d'une formule recopiée.** Elle l'était, et
// c'est ce qui a laissé le bandeau se démentir : le calcul restait juste, mais il
// s'appliquait à un contenu qui n'était plus celui des cases posées.
func hauteurBandeau(h *HUD) int {
	haut := margeEcran + 2*h.Font.Height() + h.Margin()
	return haut + h.SlotSide(contenuEmplacement) + h.Font.Height() + h.Margin()
}
