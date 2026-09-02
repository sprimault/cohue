// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce qui empêche un lieu de se charger, accumulé plutôt que rendu au premier
// manquement : qui met au point un niveau veut la liste, pas un aller-retour par
// défaut.

package level

import "fmt"

// valider rend tout ce qui empêche un lieu de se charger.
//
// Un lieu invalide fait échouer son chargement entier plutôt que d'être chargé à
// moitié : une pièce manquante laisserait un trou dans la carte, et le champ de
// flux y enverrait les ennemis tourner en rond.
func valider(nom string, lieu *Level, jeu *Set, pieces []*Room) []string {
	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	if lieu.ID == "" {
		dire("identifiant : le lieu n'en a pas")
	} else if lieu.ID != nom {
		// Le cas concret : quelqu'un duplique un lieu pour en faire une
		// variante, renomme le dossier et oublie l'identifiant. Sans ce refus,
		// la copie se charge en se croyant l'original.
		dire("identifiant : « %s », alors que le dossier se nomme « %s »", lieu.ID, nom)
	}
	if lieu.SetID == "" {
		dire("jeu_pieces : le lieu ne dit pas de quel jeu il tire ses pièces")
	} else if jeu.ID != lieu.SetID {
		// **Ce contrôle est né du nom fixe.** Le jeu de pièces s'appelait
		// autrefois du nom de son identifiant, si bien que le chemin le
		// vérifiait : un fichier mal nommé ne se chargeait pas. Le nom étant
		// désormais fixe, plus rien ne rapprochait les deux, et un jeu déposé
		// dans le mauvais lieu se serait chargé en silence avec la mauvaise
		// palette — donc en changeant le sens de tous les caractères.
		dire("jeu_pieces : le lieu emploie « %s », le jeu du dossier se nomme « %s »",
			lieu.SetID, jeu.ID)
	}
	if len(jeu.Palette) == 0 {
		dire("%s.palette : le jeu de pièces n'associe aucun caractère à une tuile", jeu.ID)
	}

	// Les poses ne sont comptées que si toutes tiennent sur l'assiette qu'elles
	// annoncent : une taille nulle ou une origine négative ferait sortir le
	// compte de couverture de sa propre table.
	assiettes := true
	for i, piece := range pieces {
		ou := fmt.Sprintf("pieces[%d] « %s »", i, lieu.Placements[i].RoomID)
		if piece.Format != FormatRoom {
			dire("%s.version_format : %d, ce binaire lit la %d", ou, piece.Format, FormatRoom)
		}
		if piece.Set != lieu.SetID {
			dire("%s.jeu : « %s », alors que le lieu emploie « %s »", ou, piece.Set, lieu.SetID)
		}
		if pose := lieu.Placements[i]; pose.U < 0 || pose.V < 0 {
			// La cuisson ignore ce qui tombe hors de la grille, et l'origine
			// d'un lieu est son coin (0, 0) : une pose négative perdrait ses
			// premières lignes sans que rien ne l'écrive.
			dire("%s : posée en (%d, %d), un lieu commence à son coin", ou, pose.U, pose.V)
			assiettes = false
		}
		if piece.Size[0] <= 0 || piece.Size[1] <= 0 {
			assiettes = false
		}
		manques = append(manques, validerGrille(ou, piece, jeu)...)
	}
	if assiettes {
		manques = append(manques, validerCouverture(lieu, pieces)...)
	}
	return manques
}

// validerCouverture exige que chaque case du lieu soit posée par une pièce, et
// par une seule.
//
// **Les deux moitiés sont muettes autrement.** Une case que personne ne pose
// garde le coût d'une grille neuve, qui est celui d'un sol ordinaire : le trou
// se traverse, ne se dessine pas, et ne se remarque qu'au moment où une créature
// y flotte. Et deux pièces qui se recouvrent se départagent aujourd'hui par
// l'ordre des poses, la dernière écrivant par-dessus la première — un ordre que
// rien n'annonce et dont aucun auteur n'a idée. Les refuser rend cet ordre sans
// effet, ce qui vaut mieux que de le documenter.
//
// Elle compte sur les tailles déclarées et non sur les grilles : c'est la taille
// qui dit ce qu'une pièce occupe dans le lieu, et une grille qui la dément est
// déjà refusée pour cette raison-là.
func validerCouverture(lieu *Level, pieces []*Room) []string {
	var largeur, hauteur int
	for i, pose := range lieu.Placements {
		largeur = max(largeur, pose.U+pieces[i].Size[0])
		hauteur = max(hauteur, pose.V+pieces[i].Size[1])
	}

	poses := make([]int, largeur*hauteur)
	for i, pose := range lieu.Placements {
		for v := range pieces[i].Size[1] {
			for u := range pieces[i].Size[0] {
				poses[(pose.V+v)*largeur+pose.U+u]++
			}
		}
	}

	var trous, doublons int
	var premierTrou, premierDoublon [2]int
	for i, n := range poses {
		switch u, v := i%largeur, i/largeur; {
		case n == 0:
			if trous == 0 {
				premierTrou = [2]int{u, v}
			}
			trous++
		case n > 1:
			if doublons == 0 {
				premierDoublon = [2]int{u, v}
			}
			doublons++
		}
	}

	// Le compte plutôt que la liste, et la première case plutôt qu'un
	// échantillon : un trou d'un bloc en produit un millier, et mille lignes
	// diraient moins que « mille cases, à partir de celle-ci ».
	var manques []string
	if trous > 0 {
		manques = append(manques, fmt.Sprintf(
			"pieces : %d case(s) qu'aucune pièce ne pose, la première en (%d, %d)",
			trous, premierTrou[0], premierTrou[1]))
	}
	if doublons > 0 {
		manques = append(manques, fmt.Sprintf(
			"pieces : %d case(s) posées deux fois ou plus, la première en (%d, %d)",
			doublons, premierDoublon[0], premierDoublon[1]))
	}
	return manques
}

// validerGrille vérifie qu'une grille tient ce que sa pièce annonce.
//
// Le désaccord de dimensions ne disparaît pas parce que la grille et le
// descripteur vivent dans le même fichier : il s'y déplace, et c'est ici qu'il
// se voit d'un coup au lieu de produire une carte trouée.
func validerGrille(ou string, piece *Room, jeu *Set) []string {
	var manques []string
	dire := func(format string, args ...any) {
		manques = append(manques, fmt.Sprintf(format, args...))
	}

	largeur, hauteur := piece.Size[0], piece.Size[1]
	if largeur <= 0 || hauteur <= 0 {
		dire("%s.taille : %v, une pièce a au moins une case", ou, piece.Size)
		return manques
	}
	if len(piece.Rows) != hauteur {
		dire("%s.grille : %d ligne(s) pour une taille qui en annonce %d",
			ou, len(piece.Rows), hauteur)
	}

	for v, ligne := range piece.Rows {
		jetons := []rune(ligne)
		if len(jetons) != largeur {
			dire("%s.grille[%d] : %d caractère(s) pour une largeur de %d",
				ou, v, len(jetons), largeur)
		}
		for u, jeton := range jetons {
			if _, connu := jeu.Palette[string(jeton)]; !connu {
				dire("%s.grille[%d][%d] : « %c » absent de la palette de « %s »",
					ou, v, u, jeton, jeu.ID)
			}
		}
	}
	return manques
}
