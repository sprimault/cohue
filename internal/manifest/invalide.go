// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"fmt"
	"strings"
)

// Invalide porte tout ce qui manque à un fichier, et non le premier manquement.
//
// C'est ici, et nulle part ailleurs, que « listés en une fois » devient vrai :
// le décodage s'arrête au premier écart parce que la bibliothèque standard ne
// sait pas faire autrement, mais les manquements de validation arrivent par
// grappes — une pièce absente, un ancrage qui manque, des dimensions qui ne
// s'accordent pas. C'est là que l'aller-retour coûte à qui met au point un
// niveau, et donc là qu'il faut résister au premier `return`.
//
// Le type vit ici et non chez le chargeur de lieux parce qu'un manifeste de
// profils doit la même chose à qui l'écrit. La validation, elle, reste chez le
// consommateur : lui seul sait ce que son format exige.
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
