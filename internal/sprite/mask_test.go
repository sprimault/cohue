// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Ce que le masque garde : l'étendue serrée qu'il relève, et le fait que deux
// boîtes qui se croisent ne suffisent pas à faire un recouvrement.

package sprite

import (
	"image"
	"image/color"
	"testing"

	"github.com/sprimault/cohue"
	"github.com/sprimault/cohue/internal/game"
)

// pave rend une image transparente où les rectangles donnés sont opaques.
func pave(largeur, hauteur int, pleins ...image.Rectangle) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, largeur, hauteur))
	for _, plein := range pleins {
		for y := plein.Min.Y; y < plein.Max.Y; y++ {
			for x := plein.Min.X; x < plein.Max.X; x++ {
				img.SetNRGBA(x, y, color.NRGBA{R: 200, G: 200, B: 200, A: 255})
			}
		}
	}
	return img
}

// TestLEtendueSerreeIgnoreLeVide relève la forme et non la bande qui la porte.
//
// C'est la mesure du défaut qu'on corrige : une bande de personnage fait 64×64
// et le joueur en occupe 19×45, soit un cinquième de la surface.
func TestLEtendueSerreeIgnoreLeVide(t *testing.T) {
	m := NewMask(pave(64, 64, image.Rect(22, 10, 41, 55)))

	if got, veut := m.Bounds(), image.Rect(22, 10, 41, 55); got != veut {
		t.Errorf("étendue serrée %v, attendue %v", got, veut)
	}
}

// TestUnDessinVideNaPasDEtendue garde le cas d'une pose entièrement transparente.
//
// Sans lui, l'étendue vaudrait le rectangle dégénéré que la boucle laisse
// derrière elle, et `Overlaps` en tirerait un croisement.
func TestUnDessinVideNaPasDEtendue(t *testing.T) {
	m := NewMask(pave(32, 32))

	if !m.Bounds().Empty() {
		t.Errorf("étendue %v pour un dessin vide, attendue vide", m.Bounds())
	}
	if m.Overlaps(0, 0, NewMask(pave(32, 32, image.Rect(0, 0, 32, 32))), 0, 0) {
		t.Error("un dessin vide recouvre quelque chose")
	}
}

// TestDeuxBoitesQuiSeCroisentNeSeRecouvrentPas est le défaut que le masque ferme.
//
// Deux bandes de soixante-quatre pixels décalées de trente se croisent
// largement ; les deux figures qu'elles portent, chacune sur sa moitié, ne se
// touchent pas. C'est le cas que le test de boîtes déclarait recouvert quatre
// fois sur cinq.
func TestDeuxBoitesQuiSeCroisentNeSeRecouvrentPas(t *testing.T) {
	gauche := NewMask(pave(64, 64, image.Rect(0, 20, 20, 60)))
	droite := NewMask(pave(64, 64, image.Rect(44, 20, 64, 60)))

	boiteG := image.Rect(0, 0, 64, 64)
	boiteD := image.Rect(30, 0, 94, 64)
	if !boiteG.Overlaps(boiteD) {
		t.Fatal("les deux boîtes ne se croisent pas, le cas ne dit plus rien")
	}
	if gauche.Overlaps(0, 0, droite, 30, 0) {
		t.Error("recouvrement annoncé là où aucun pixel n'est partagé")
	}

	// Les mêmes bandes au même décalage, les figures échangées : celle de
	// droite est devant, celle de gauche la rattrape à partir de 44.
	if !droite.Overlaps(0, 0, gauche, 30, 0) {
		t.Error("recouvrement manqué là où les figures se croisent")
	}
}

// TestDeuxEtenduesQuiSeCroisentNeSeRecouvrentPas garde le second niveau du test.
//
// Le premier — l'étendue serrée — ne suffit pas : une forme creuse a une étendue
// pleine. Un personnage qui lève un bras et un pilier étroit peuvent partager
// leur rectangle sans partager un pixel, et c'est ce cas-là qui exige de
// descendre au bit. Sans lui, la boîte serrée passerait pour la correction
// entière, ce qu'elle n'est pas : la mesure indépendante a montré qu'elle
// n'enlève que la moitié des faux déclenchements.
func TestDeuxEtenduesQuiSeCroisentNeSeRecouvrentPas(t *testing.T) {
	// Une équerre : une barre à gauche, et un pixel isolé en bas à droite qui
	// étire l'étendue sur toute l'image sans rien remplir entre les deux.
	equerre := NewMask(pave(16, 16, image.Rect(0, 0, 2, 16), image.Rect(15, 15, 16, 16)))
	barre := NewMask(pave(16, 16, image.Rect(6, 0, 8, 10)))

	if !equerre.Bounds().Overlaps(barre.Bounds()) {
		t.Fatal("les deux étendues ne se croisent pas, le cas ne dit plus rien")
	}
	if equerre.Overlaps(0, 0, barre, 0, 0) {
		t.Error("recouvrement annoncé entre deux formes qui ne partagent aucun pixel")
	}

	// Six pixels vers la gauche, et la barre tombe sur le montant de l'équerre.
	if !equerre.Overlaps(0, 0, barre, -6, 0) {
		t.Error("recouvrement manqué là où la barre couvre le montant")
	}
}

// TestUnSeulPixelPartageSuffit garde la borne basse.
//
// Une silhouette se révèle dès qu'on cache quelque chose du personnage, et la
// conception ne pose aucun seuil de surface — un pixel est un pixel.
func TestUnSeulPixelPartageSuffit(t *testing.T) {
	a := NewMask(pave(10, 10, image.Rect(0, 0, 5, 5)))
	b := NewMask(pave(10, 10, image.Rect(0, 0, 5, 5)))

	// b posé de sorte que son coin haut-gauche tombe sur le dernier pixel de a.
	if !a.Overlaps(0, 0, b, 4, 4) {
		t.Error("un pixel commun n'a pas suffi")
	}
	if a.Overlaps(0, 0, b, 5, 5) {
		t.Error("recouvrement annoncé pour deux formes adjacentes")
	}
}

// TestLaFormeSuitCeQueFlattenGarde épingle le seuil commun aux deux.
//
// Un masque plus sévère que l'aplat révélerait des silhouettes dont le liseré
// déborde de ce qu'on a testé, et rien ne le dirait qu'un bord fondu.
func TestLaFormeSuitCeQueFlattenGarde(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 0})
	img.SetNRGBA(1, 0, color.NRGBA{R: 255, A: 1})
	img.SetNRGBA(2, 0, color.NRGBA{R: 255, A: 255})

	m := NewMask(img)
	aplat := Flatten(img)
	for x := range 3 {
		_, _, _, a := aplat.At(x, 0).RGBA()
		if got := m.opaque(x, 0); got != (a != 0) {
			t.Errorf("pixel %d : masque %v, aplat opaque %v", x, got, a != 0)
		}
	}
}

// TestLaFormeDuJoueurNoccupePasSaBande épingle la mesure qui justifie le masque.
//
// Le joueur occupe 19×45 d'une bande de 64×64, soit un sixième de sa surface.
// C'est de cet écart que venaient les 82 % de faux déclenchements de la
// silhouette, et c'est lui qu'il faut revoir le jour où le générateur change la
// taille des figurines : un personnage qui remplirait sa bande rendrait le
// masque inutile, et un plus petit encore le rendrait plus décisif.
func TestLaFormeDuJoueurNoccupePasSaBande(t *testing.T) {
	profils, err := game.LoadProfiles(cohue.Assets, manifesteLivre)
	if err != nil {
		t.Fatal(err)
	}
	feuille, err := Load(cohue.Assets, racineLivree, profils.Player.Figure)
	if err != nil {
		t.Fatal(err)
	}
	img, ok := feuille.Frame("repos", "S", 0, 0)
	if !ok {
		t.Fatal("le joueur livré n'a pas de pose de repos face à l'écran")
	}

	serree := NewMask(img).Bounds()
	bande := img.Bounds()
	part := float64(serree.Dx()*serree.Dy()) / float64(bande.Dx()*bande.Dy())
	t.Logf("bande %dx%d, forme %dx%d, %.1f %% de la boîte",
		bande.Dx(), bande.Dy(), serree.Dx(), serree.Dy(), 100*part)

	if part > 0.5 {
		t.Errorf("la forme occupe %.1f %% de sa bande : le masque ne sert plus à grand-chose",
			100*part)
	}
	if serree.Empty() {
		t.Error("la pose de repos du joueur est vide")
	}
}
