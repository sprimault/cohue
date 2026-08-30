# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Fabrique les bruitages du jeu par synthèse, sans aucune dépendance.

    python outils/sons.py                 -> tout, dans ./assets/sons
    python outils/sons.py tir impact      -> quelques sons
    python outils/sons.py --liste

Même principe que le décor et les créatures : rien n'est enregistré, tout est
calculé. Un bruitage est une enveloppe appliquée à un oscillateur, plus un peu
de bruit — c'est le procédé de sfxr, et il couvre exactement le registre d'un
survivor : tirs, impacts, ramassages, explosions.

Seule la bibliothèque standard est utilisée : `wave` écrit le fichier, `math` et
`random` produisent les échantillons. La graine de chaque son est fixée, donc
deux exécutions donnent des fichiers identiques au bit près.
"""

import argparse
import json
import math
import random
import struct
import wave
from pathlib import Path

from manifestes import ecrire_manifeste

TAUX = 22050          # suffisant pour des bruitages, et deux fois plus léger que 44 kHz
AMPLITUDE = 0.72      # marge avant saturation, le mixage du jeu ajoutera des voix

# Gain par son, en fraction de l'amplitude. Il ne s'agit pas d'un réglage de
# confort : en tir automatique, le tir de base part plusieurs fois par seconde
# pendant quinze minutes. Au même niveau que les autres, il recouvre la musique
# et sature l'oreille. Les sons rares ont donc le droit d'être forts, les sons
# répétés doivent rester sous la nappe.
#
# Le moteur applique en plus son propre volume par catégorie ; ces valeurs
# fixent le rapport entre les sons, pas le volume absolu.
GAINS = {
    "tir": 0.28,               # plusieurs fois par seconde, en continu
    "impact": 0.32,            # aussi fréquent que le tir
    "mort_ennemi": 0.42,       # cent fois par run
    "gemme": 0.40,             # par volées, la gamme fait déjà le relief
    "interface_choix": 0.55,
    "caisse_appui": 0.50,
    "tir_lourd": 0.80,         # rare et voulu spectaculaire
    "caisse_rupture": 0.70,
    "ramassage_arme": 0.85,
    "soin": 0.80,
    "telegraphe": 0.90,        # doit percer la horde, c'est un avertissement
    "degat_joueur": 0.95,      # doit couper l'attention
    "explosion": 1.0,
    "montee_niveau": 0.95,
    "porte_ouverte": 0.85,
    "mort_joueur": 1.0,
}


# --- Formes d'onde ---------------------------------------------------------

def _carre(phase, rapport=0.5):
    return 1.0 if (phase % 1.0) < rapport else -1.0


def _scie(phase):
    return 2.0 * (phase % 1.0) - 1.0


def _sinus(phase):
    return math.sin(2 * math.pi * phase)


def _triangle(phase):
    p = phase % 1.0
    return 4 * p - 1 if p < 0.5 else 3 - 4 * p


ONDES = {"carre": _carre, "scie": _scie, "sinus": _sinus, "triangle": _triangle}


# --- Enveloppes ------------------------------------------------------------

def _enveloppe(avancement, attaque, maintien, chute):
    """Amplitude à un instant donné, en fractions de la durée totale."""
    if avancement < attaque:
        return avancement / attaque if attaque else 1.0
    if avancement < attaque + maintien:
        return 1.0
    reste = 1.0 - attaque - maintien
    if reste <= 0:
        return 0.0
    return max(0.0, 1.0 - (avancement - attaque - maintien) / reste)


def _synthese(duree, frequence_debut, frequence_fin=None, onde="carre",
              attaque=0.01, maintien=0.1, chute=0.89, bruit=0.0, rapport=0.5,
              vibrato=0.0, vibrato_hz=0.0, graine=0):
    """Rend une liste d'échantillons entre -1 et 1.

    La glissade de fréquence est ce qui distingue un tir d'un impact : montante
    elle sonne comme un gain, descendante comme un choc.
    """
    alea = random.Random(graine)
    forme = ONDES[onde]
    total = int(duree * TAUX)
    frequence_fin = frequence_debut if frequence_fin is None else frequence_fin

    echantillons = []
    phase = 0.0
    for i in range(total):
        avancement = i / total
        frequence = frequence_debut + (frequence_fin - frequence_debut) * avancement
        if vibrato:
            frequence *= 1.0 + vibrato * math.sin(2 * math.pi * vibrato_hz * i / TAUX)
        phase += frequence / TAUX

        valeur = forme(phase, rapport) if onde == "carre" else forme(phase)
        if bruit:
            valeur = valeur * (1 - bruit) + alea.uniform(-1, 1) * bruit

        echantillons.append(valeur * _enveloppe(avancement, attaque, maintien, chute))
    return echantillons


def _mixer(*pistes):
    """Superpose des pistes de longueurs quelconques, puis normalise."""
    longueur = max(len(p) for p in pistes)
    melange = [0.0] * longueur
    for piste in pistes:
        for i, valeur in enumerate(piste):
            melange[i] += valeur
    crete = max(abs(v) for v in melange) or 1.0
    return [v / crete for v in melange]


def _ecrire(chemin, echantillons, gain=1.0):
    with wave.open(str(chemin), "wb") as sortie:
        sortie.setnchannels(1)
        sortie.setsampwidth(2)
        sortie.setframerate(TAUX)
        sortie.writeframes(b"".join(
            struct.pack("<h", int(max(-1.0, min(1.0, v)) * AMPLITUDE * gain * 32767))
            for v in echantillons))


# --- Le catalogue ----------------------------------------------------------
# Un son par événement du jeu. Les valeurs ont été réglées à l'oreille : elles
# n'ont pas d'autre justification que de sonner juste à côté des autres.

def tir():
    """Tir de base : très court et mat, il part plusieurs fois par seconde.

    Sa hauteur est volontairement haute et sa durée minimale : un son bref et
    aigu se superpose à une nappe sans la masquer, là où un son grave et
    traînant occupe la même place qu'elle.
    """
    return _synthese(0.055, 1040, 380, onde="carre", rapport=0.28,
                     attaque=0.003, maintien=0.02, bruit=0.10, graine=1)


def tir_lourd():
    """Arme lourde : plus grave et plus long, pour qu'on l'entende par-dessus."""
    return _mixer(
        _synthese(0.22, 340, 70, onde="scie", attaque=0.01, maintien=0.12, graine=2),
        _synthese(0.22, 120, 40, onde="carre", bruit=0.5, maintien=0.2, graine=3),
    )


def impact():
    """Projectile qui touche : bref, mat, sans hauteur définie."""
    return _synthese(0.07, 200, 90, onde="carre", bruit=0.7,
                     attaque=0.002, maintien=0.03, graine=4)


def degat_joueur():
    """Coup reçu : descendant et un peu long, il doit couper l'attention."""
    return _mixer(
        _synthese(0.30, 420, 90, onde="scie", attaque=0.005, maintien=0.1, graine=5),
        _synthese(0.30, 90, 50, onde="carre", bruit=0.4, maintien=0.15, graine=6),
    )


def mort_ennemi():
    """Mort ordinaire : courte, grave, sans emphase — elle sonne cent fois."""
    return _synthese(0.16, 260, 60, onde="carre", bruit=0.45,
                     attaque=0.005, maintien=0.05, graine=7)


def explosion():
    """Baudruche : bruit large, longue chute, une basse pour le corps."""
    return _mixer(
        _synthese(0.55, 160, 30, onde="carre", bruit=0.85, maintien=0.12, graine=8),
        _synthese(0.55, 70, 25, onde="sinus", attaque=0.01, maintien=0.08, graine=9),
    )


def telegraphe():
    """Avertissement avant explosion ou charge : montant, régulier, inquiétant."""
    return _synthese(0.35, 300, 700, onde="triangle", attaque=0.02, maintien=0.3,
                     vibrato=0.06, vibrato_hz=18, graine=10)


def gemme(indice=0):
    """Ramassage : la hauteur monte avec la volée, et retombe après un silence.

    C'est le son le plus joué du jeu et le moment de plaisir maximal du genre :
    huit degrés d'une gamme suffisent à rendre une pluie de gemmes agréable là
    où une hauteur unique deviendrait vite pénible.
    """
    degres = (0, 2, 4, 5, 7, 9, 11, 12)
    hauteur = 660 * (2 ** (degres[indice % len(degres)] / 12))
    return _synthese(0.07, hauteur, hauteur * 1.5, onde="triangle",
                     attaque=0.005, maintien=0.02, graine=11 + indice)


def ramassage_arme():
    """Trouvaille : deux notes montantes, plus longues qu'une gemme."""
    return _mixer(
        _synthese(0.10, 520, 520, onde="triangle", maintien=0.3, graine=12),
        [0.0] * int(0.08 * TAUX) + _synthese(0.16, 780, 900, onde="triangle",
                                             maintien=0.3, graine=13),
    )


def soin():
    """Fiole : montée douce et tenue, à l'opposé du registre sec des tirs."""
    return _synthese(0.40, 440, 880, onde="sinus", attaque=0.05, maintien=0.25,
                     graine=14)


def montee_niveau():
    """Trois notes : c'est la récompense, elle a le droit d'être longue."""
    pistes = []
    for rang, hauteur in enumerate((523, 659, 784)):
        silence = [0.0] * int(rang * 0.09 * TAUX)
        pistes.append(silence + _synthese(0.26, hauteur, hauteur, onde="triangle",
                                          attaque=0.01, maintien=0.35,
                                          graine=15 + rang))
    return _mixer(*pistes)


def caisse_appui():
    """Bois qui travaille pendant le délai de contact : rêche, sans hauteur."""
    return _synthese(0.12, 150, 120, onde="carre", bruit=0.8,
                     attaque=0.02, maintien=0.05, graine=18)


def caisse_rupture():
    """Rupture : le bois cède, éclats plus aigus par-dessus."""
    return _mixer(
        _synthese(0.28, 220, 60, onde="carre", bruit=0.85, maintien=0.06, graine=19),
        _synthese(0.20, 900, 300, onde="scie", bruit=0.6, maintien=0.04, graine=20),
    )


def porte_ouverte():
    """Sortie franchie : grave, ample, une seule fois par lieu."""
    return _mixer(
        _synthese(0.60, 110, 220, onde="sinus", attaque=0.05, maintien=0.35, graine=21),
        _synthese(0.45, 330, 440, onde="triangle", attaque=0.1, maintien=0.3, graine=22),
    )


def mort_joueur():
    """Fin de partie : descendante et longue, elle laisse la place à la relance."""
    return _mixer(
        _synthese(0.90, 330, 55, onde="scie", attaque=0.02, maintien=0.2, graine=23),
        _synthese(0.90, 110, 40, onde="sinus", attaque=0.05, maintien=0.25, graine=24),
    )


def interface_choix():
    """Carte d'amélioration retenue : net, court, sans emphase."""
    return _synthese(0.06, 700, 900, onde="carre", rapport=0.25,
                     attaque=0.003, maintien=0.02, graine=25)


CATALOGUE = {
    "tir": tir,
    "tir_lourd": tir_lourd,
    "impact": impact,
    "degat_joueur": degat_joueur,
    "mort_ennemi": mort_ennemi,
    "explosion": explosion,
    "telegraphe": telegraphe,
    "ramassage_arme": ramassage_arme,
    "soin": soin,
    "montee_niveau": montee_niveau,
    "caisse_appui": caisse_appui,
    "caisse_rupture": caisse_rupture,
    "porte_ouverte": porte_ouverte,
    "mort_joueur": mort_joueur,
    "interface_choix": interface_choix,
}

# Catégorie de mixage : le joueur doit pouvoir baisser les effets sans toucher à
# la musique, et l'interface reste audible quand tout le reste est baissé.
CATEGORIES = {
    "interface_choix": "interface",
    "montee_niveau": "interface",
    "porte_ouverte": "interface",
}

# Les gemmes forment une gamme : le moteur joue le degré suivant à chaque
# ramassage rapproché, et repart du premier après un silence.
DEGRES_GEMME = 8


def main():
    analyseur = argparse.ArgumentParser(description=__doc__)
    analyseur.add_argument("sons", nargs="*", help="noms à générer, vide pour tout")
    analyseur.add_argument("--sortie", type=Path, default=Path("assets/sons"))
    analyseur.add_argument("--liste", action="store_true")
    options = analyseur.parse_args()

    if options.liste:
        print("\n".join(sorted(CATALOGUE) + [f"gemme_0 … gemme_{DEGRES_GEMME - 1}"]))
        return

    noms = options.sons or sorted(CATALOGUE)
    inconnus = [n for n in noms if n not in CATALOGUE]
    if inconnus:
        analyseur.error(f"son inconnu : {', '.join(inconnus)}")

    options.sortie.mkdir(parents=True, exist_ok=True)
    manifeste = {}

    for nom in noms:
        echantillons = CATALOGUE[nom]()
        gain = GAINS.get(nom, 0.8)
        _ecrire(options.sortie / f"{nom}.wav", echantillons, gain)
        duree = len(echantillons) / TAUX
        manifeste[nom] = {"duree_ms": round(duree * 1000), "taux": TAUX, "gain": gain,
                          "categorie": CATEGORIES.get(nom, "effets"), "boucle": False}
        print(f"{nom:18} {duree * 1000:5.0f} ms   gain {gain:.2f}")

    if not options.sons:
        for indice in range(DEGRES_GEMME):
            echantillons = gemme(indice)
            _ecrire(options.sortie / f"gemme_{indice}.wav", echantillons, GAINS["gemme"])
            manifeste[f"gemme_{indice}"] = {"duree_ms": round(len(echantillons) / TAUX * 1000),
                                            "taux": TAUX, "degre": indice,
                                            "gain": GAINS["gemme"],
                                            "categorie": "effets", "boucle": False}
        print(f"{'gemme':18} {DEGRES_GEMME} degrés")

        ecrire_manifeste(options.sortie / "manifeste.json", "sons.py",
                         {"version_format": 1, "sons": manifeste})

    print(f"\n{len(manifeste)} sons dans {options.sortie}")


if __name__ == "__main__":
    main()
