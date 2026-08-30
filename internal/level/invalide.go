// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package level

import (
	"fmt"
	"strings"
)

// Invalide porte tout ce qui manque à un lieu, et non le premier manquement.
//
// C'est ici, et nulle part ailleurs, que « listés en une fois » devient vrai :
// le décodage s'arrête au premier écart parce que la bibliothèque standard ne
// sait pas faire autrement, mais les manquements de validation arrivent par
// grappes — une pièce absente, un ancrage qui manque, des dimensions qui ne
// s'accordent pas. C'est là que l'aller-retour coûte à qui met au point un
// niveau, et donc là qu'il faut résister au premier `return`.
type Invalide struct {
	Chemin  string
	Manques []string
}

// Error énumère les manquements, un par ligne.
func (e *Invalide) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s : %d manquement(s)", e.Chemin, len(e.Manques))
	for _, m := range e.Manques {
		b.WriteString("\n  " + m)
	}
	return b.String()
}

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
	}
	if len(jeu.Palette) == 0 {
		dire("%s.palette : le jeu de pièces n'associe aucun caractère à une tuile", jeu.ID)
	}

	for i, piece := range pieces {
		ou := fmt.Sprintf("pieces[%d] « %s »", i, lieu.Placements[i].RoomID)
		if piece.Format != FormatRoom {
			dire("%s.version_format : %d, ce binaire lit la %d", ou, piece.Format, FormatRoom)
		}
		if piece.Set != lieu.SetID {
			dire("%s.jeu : « %s », alors que le lieu emploie « %s »", ou, piece.Set, lieu.SetID)
		}
		manques = append(manques, validerGrille(ou, piece, jeu)...)
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
