// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"strings"
	"testing"
	"testing/fstest"
)

// piece est une structure de format minimale, avec le commentaire embarqué que
// tout ce qui se décode doit porter.
type piece struct {
	Commentable
	Nom string `json:"nom"`
}

// fichier monte un système de fichiers d'un seul fichier, nommé `p.json`.
func fichier(contenu string) fstest.MapFS {
	return fstest.MapFS{"p.json": &fstest.MapFile{Data: []byte(contenu)}}
}

// TestDecodeRefuseUneCleInconnue vérifie que la faute de frappe qui ajoute une
// clé ne se charge pas en silence, et que le message nomme la coupable.
//
// Ces trois cas sont éprouvés par le chargeur de lieux, qui les rencontre en
// vrai. Ils le sont aussi ici parce qu'une garantie dont la seule preuve vit
// dans un autre paquet disparaît le jour où ce paquet change, sans que rien ne
// le dise.
func TestDecodeRefuseUneCleInconnue(t *testing.T) {
	_, err := Decode[piece](fichier(`{"nom": "salle", "nmo": "salle"}`), "p.json")
	if err == nil {
		t.Fatal("une clé inconnue s'est chargée en silence")
	}
	if !strings.Contains(err.Error(), "nmo") {
		t.Errorf("le message ne nomme pas la clé fautive : %v", err)
	}
}

// TestDecodeAccepteUnCommentaire vérifie que `$comment` passe là où une
// structure embarque Commentable.
func TestDecodeAccepteUnCommentaire(t *testing.T) {
	p, err := Decode[piece](fichier(`{"$comment": "pourquoi", "nom": "salle"}`), "p.json")
	if err != nil {
		t.Fatalf("un commentaire a fait échouer le décodage : %v", err)
	}
	if p.Nom != "salle" {
		t.Errorf("nom « %s », attendu « salle »", p.Nom)
	}
}

// TestDecodeNommeLeFichierAbsent vérifie que l'erreur dit lequel manque.
//
// Sans le chemin, un lieu qui cite trois pièces rend trois fois le même message
// et il faut les ouvrir une par une pour savoir laquelle est absente.
func TestDecodeNommeLeFichierAbsent(t *testing.T) {
	_, err := Decode[piece](fstest.MapFS{}, "absent.json")
	if err == nil {
		t.Fatal("un fichier absent s'est décodé")
	}
	if !strings.Contains(err.Error(), "absent.json") {
		t.Errorf("le message ne nomme pas le fichier : %v", err)
	}
}
