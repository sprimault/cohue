// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// La fenêtre et le tampon interne : ce qu'Ebitengine a besoin de savoir pour
// tourner, et rien de plus. La projection, la caméra et le tri en profondeur
// viendront à côté.

// Package render dessine ce que la simulation a calculé, et ne décide de rien.
//
// La frontière avec `internal/game` va dans un seul sens : ce paquet lit l'état
// du monde, et rien de ce qu'il produit n'y revient. C'est ce qui lui permet de
// calculer en flottants — interpolations, lissage de caméra — sans menacer le
// déterminisme de la run, et ce qui interdit à une notion d'écran de redescendre
// dans la simulation.
//
// **Ce paquet n'a pas de fichier de test, et n'en aura pas.** Importer
// Ebitengine initialise GLFW, qui panique sans `DISPLAY` : sur un runner sans
// écran, n'importe quel test du paquet tombe avant d'avoir commencé, y compris
// un test qui n'ouvrirait aucune fenêtre. La suite par défaut reste donc
// exécutable partout, et le rendu se juge à l'œil — c'est ce que la doctrine de
// test énonçait déjà, et l'absence de tests ici en est la conséquence et non un
// oubli. Ce qui doit être vérifié mécaniquement vit du côté simulation, qui
// n'importe pas Ebitengine.
package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Les dimensions du tampon interne, agrandi en entier vers la fenêtre.
//
// Un tampon de 480×270 ne montrerait que sept tuiles de large, bien trop serré
// pour voir la horde arriver ; 960×540 en donne une quinzaine et se multiplie
// par deux pour du 1080p, donc pixels carrés garantis.
const (
	Largeur = 960
	Hauteur = 540
)

// fond est la couleur d'un écran que rien ne remplit encore. Un gris sombre
// plutôt qu'un noir : un noir pur ne se distingue pas d'une fenêtre qui n'a pas
// fini de s'ouvrir.
var fond = color.RGBA{R: 24, G: 24, B: 28, A: 255}

// Screen est le jeu tel qu'Ebitengine le voit.
//
// Il ne porte encore aucun état : ce lot n'existe que pour lier la bibliothèque
// au binaire et vérifier que la matrice de compilation tient — Windows et
// WebAssembly sans cgo, Linux et macOS avec. Une dépendance qui changerait ses
// exigences de ce côté se verrait ici, et non à la première publication.
type Screen struct{}

// Update avance d'un pas de simulation.
func (s *Screen) Update() error {
	// à implémenter : étape 2
	return nil
}

// Draw peint le tampon interne.
func (s *Screen) Draw(ecran *ebiten.Image) {
	// à implémenter : étape 2
	ecran.Fill(fond)
}

// Layout fixe la taille du tampon interne, quelle que soit celle de la fenêtre.
//
// Ebitengine agrandit ensuite vers la fenêtre. Le rapport n'est pas forcément
// entier tant que la caméra n'existe pas ; c'est elle qui l'imposera, avec les
// pixels entiers que le pixel art exige.
func (s *Screen) Layout(largeurFenetre, hauteurFenetre int) (int, int) {
	return Largeur, Hauteur
}
