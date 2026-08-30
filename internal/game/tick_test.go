// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package game

import (
	"errors"
	"testing"
)

// TestTicksFromMs vérifie la conversion des durées que portent les manifestes.
func TestTicksFromMs(t *testing.T) {
	cas := map[int]Tick{
		17:  1,  // le premier qui atteint un pas
		40:  2,  // 2,4 pas : on descend
		50:  3,  // 3,0 pas
		80:  5,  // la cadence d'attaque la plus rapide du manifeste
		100: 6,  // la marche
		330: 20, // l'appui sur une caisse : 19,8 pas, on monte
	}
	for ms, attendu := range cas {
		obtenu, err := TicksFromMs(ms)
		if err != nil {
			t.Errorf("%d ms : %v", ms, err)
			continue
		}
		if obtenu != attendu {
			t.Errorf("%d ms = %d ticks, attendu %d", ms, obtenu, attendu)
		}
	}
}

// TestTicksFromMsRefuseSousLePas vérifie qu'une durée trop courte est refusée et
// non relevée à un tick.
//
// La relever produirait un fichier qui ment : huit millisecondes auraient l'air
// de valoir du 125 Hz, et le cycle tournerait à 60 sans que rien ne le signale.
func TestTicksFromMsRefuseSousLePas(t *testing.T) {
	for _, ms := range []int{0, 1, 8, 16} {
		if _, err := TicksFromMs(ms); !errors.Is(err, ErrDurationTooShort) {
			t.Errorf("%d ms accepté, ou refusé pour une autre raison : %v", ms, err)
		}
	}
}

// TestTicksFromMsToutesLesDureesLivrees vérifie que le manifeste actuel passe.
//
// Le générateur refuse déjà d'écrire une durée sous le pas ; ce test dit que le
// chargeur et lui s'accordent, ce qu'aucun des deux ne peut affirmer seul.
func TestTicksFromMsToutesLesDureesLivrees(t *testing.T) {
	// Les durées présentes dans assets/, en millisecondes.
	for _, ms := range []int{40, 60, 80, 100, 120, 140, 200, 330} {
		if _, err := TicksFromMs(ms); err != nil {
			t.Errorf("durée livrée refusée : %d ms — %v", ms, err)
		}
	}
}

// TestSeconds vérifie la conversion pour l'affichage.
func TestSeconds(t *testing.T) {
	if obtenu := Tick(TPS * 90).Seconds(); obtenu != 90 {
		t.Errorf("90 secondes de ticks rendent %v secondes", obtenu)
	}
}
