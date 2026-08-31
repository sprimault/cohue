# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Vérifie que chaque fichier versionné porte sa mention de licence.

    python entetes.py          -> contrôle, sortie non nulle si un fichier manque

La règle et la liste close des dispensés sont dans `docs/go.md`. Ce script en
est le pendant exécutable : une règle que rien ne contrôle se découvre violée
six mois plus tard, sur des fichiers que plus personne ne relit.

Le périmètre est ce que git publierait, pas ce qui traîne sur le disque : les
fichiers suivis **et** les fichiers non suivis que rien n'exclut. Les deux, parce
que l'index seul est vide tant qu'aucun commit n'existe, et qu'un contrôle qui
passe sur une liste vide ne contrôle rien.

Bibliothèque standard uniquement : ce contrôle doit tourner sur un dépôt
fraîchement cloné, avant toute installation.
"""

import subprocess
import sys
from pathlib import Path

MARQUEUR = "SPDX-License-Identifier: Apache-2.0"

# Les deux lignes vivent dans les toutes premières du fichier. Chercher plus
# loin laisserait passer une mention perdue au milieu d'un fichier, où elle
# n'aurait plus valeur d'en-tête.
LIGNES_ENTETE = 3

DISPENSES = {
    ".editorconfig", ".gitattributes", ".gitignore", ".golangci.yml", "Makefile",
    "go.mod", "go.sum", "LICENSE", "NOTICE", "THIRD-PARTY-NOTICES",
}

# Le Markdown s'adresse à un lecteur ; les binaires ne portent pas de
# commentaire et c'est le manifeste de leur lot qui porte la mention pour eux.
SUFFIXES_DISPENSES = {".md", ".png", ".jpg", ".wav", ".ogg", ".zip", ".gz"}

# Le JSON n'a pas de commentaires : la mention y est un `$comment` en première
# clé, et c'est `ressources.py --controle` qui l'exige sur les manifestes.
#
# La dispense s'arrête donc où s'arrête ce contrôle-là, c'est-à-dire à `assets/`.
# Portée à toute l'extension, elle laissait sans garde les JSON du reste du
# dépôt — ceux des données de test, écrits à la main et par personne d'autre. Un
# contrôle qui saute ce que personne d'autre ne regarde ne signale plus l'écart,
# il le certifie.
SUFFIXES_JSON = {".json"}
RACINE_RESSOURCES = "assets"


def versionnes(racine):
    """Rend ce que git publierait, relatif à la racine du dépôt.

    `--others --exclude-standard` ajoute aux fichiers suivis ceux que rien
    n'exclut : sans eux, le contrôle rendrait une liste vide sur un dépôt sans
    commit, et passerait au vert sans avoir rien lu.
    """
    sortie = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard"],
        cwd=racine, capture_output=True, text=True, check=True).stdout
    return sorted({Path(l) for l in sortie.splitlines() if l})


def porte_entete(chemin):
    """Dit si les premières lignes du fichier portent le marqueur SPDX."""
    with open(chemin, encoding="utf-8", errors="replace") as f:
        for _ in range(LIGNES_ENTETE):
            ligne = f.readline()
            if not ligne:
                return False
            if MARQUEUR in ligne:
                return True
    return False


def controler(racine):
    """Rend la liste des fichiers versionnés auxquels la mention manque."""
    manquants = []
    for relatif in versionnes(racine):
        if relatif.name in DISPENSES or relatif.suffix in SUFFIXES_DISPENSES:
            continue
        if relatif.suffix in SUFFIXES_JSON and relatif.parts[0] == RACINE_RESSOURCES:
            continue
        if not porte_entete(racine / relatif):
            manquants.append(relatif)
    return manquants


def main():
    """Contrôle le dépôt et sort en échec dès qu'une mention manque."""
    racine = Path(__file__).resolve().parent.parent
    manquants = controler(racine)
    if manquants:
        print(f"{len(manquants)} fichier(s) sans mention de licence :")
        for m in manquants:
            print(f"  {m.as_posix()}")
        raise SystemExit(1)
    print("mention de licence présente partout où elle est exigée")


if __name__ == "__main__":
    sys.exit(main())
