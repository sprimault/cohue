# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Confronte THIRD-PARTY-NOTICES à ce qui est réellement lié au binaire.

Le fichier accompagne chaque archive publiée, et la plupart des licences
permissives exigent que leur notice suive toute copie du logiciel. Une notice
manquante est donc un manquement légal, et une notice de trop désigne un
composant que le binaire n'embarque pas — ce qui est faux, simplement.

Le périmètre est ce que `go list -deps` rapporte, et non le contenu de `go.mod` :
une dépendance indirecte dont aucun paquet n'est lié n'a rien à déclarer. Il
dépend aussi de la cible — `jezek/xgb` n'entre que dans le binaire Linux —, d'où
l'union sur les cibles publiées. `js/wasm` en est exclu : il se compile pour
détecter l'arrivée de cgo, mais ne se publie pas, donc son archive n'existe pas.

Deux sens, comme toujours : un module lié sans notice, une notice sans module
lié. Et les versions, parce qu'une notice qui reste sur une version antérieure
cesse d'être le texte qui accompagne le binaire.
"""

import argparse
import re
import subprocess
import sys
from pathlib import Path

NOTICES = Path("THIRD-PARTY-NOTICES")

# Les cibles dont une archive est publiée. `js/wasm` n'y est pas : il se compile
# sans se publier, et une notice ne suit que ce qu'on distribue.
CIBLES = (
    ("windows", "amd64", "0"),
    ("linux", "amd64", "1"),
    ("linux", "arm64", "1"),
    ("darwin", "amd64", "1"),
    ("darwin", "arm64", "1"),
)

# Une ligne de déclaration : chemin de module, version, tiret cadratin, licence.
DECLARATION = re.compile(r"^(\S+) (v\S+) — (\S+)$")

MODULE_LOCAL = "github.com/sprimault/cohue"


def lies():
    """Rend les modules effectivement liés, en union sur les cibles publiées."""
    trouves = {}
    for goos, goarch, cgo in CIBLES:
        sortie = subprocess.run(
            ["go", "list", "-deps", "-f",
             "{{with .Module}}{{.Path}} {{.Version}}{{end}}", "./cmd/cohue"],
            capture_output=True, text=True,
            env={**_env(), "GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": cgo},
        )
        if sortie.returncode != 0:
            raise SystemExit(f"go list a échoué pour {goos}/{goarch} :\n{sortie.stderr}")
        for ligne in sortie.stdout.splitlines():
            if not ligne.strip() or ligne.startswith(MODULE_LOCAL):
                continue
            chemin, version = ligne.split()
            trouves[chemin] = version
    return trouves


def _env():
    """Rend l'environnement courant, pour ne pas perdre GOCACHE ni GOMODCACHE."""
    import os
    return dict(os.environ)


def declares(chemin):
    """Rend les modules que le fichier de notices déclare.

    Un fichier sans aucune déclaration est une erreur et non un contrôle sans
    objet : ce serait la façon la plus simple de faire taire ce script.
    """
    trouves = {}
    for ligne in chemin.read_text(encoding="utf-8").splitlines():
        if m := DECLARATION.match(ligne):
            trouves[m.group(1)] = m.group(2)
    if not trouves:
        raise SystemExit(f"{chemin} : aucune déclaration « module version — licence »")
    return trouves


def controler(chemin):
    """Rend la liste des écarts entre les notices et ce qui est lié."""
    reels, notes = lies(), declares(chemin)
    defauts = []
    for module, version in sorted(reels.items()):
        if module not in notes:
            defauts.append(f"« {module} » est lié au binaire et n'a pas de notice")
        elif notes[module] != version:
            defauts.append(f"« {module} » : notice en {notes[module]}, lié en {version}")
    for module in sorted(notes):
        if module not in reels:
            defauts.append(f"« {module} » a une notice sans être lié à aucune cible publiée")
    return defauts, len(reels)


def main():
    """Contrôle les notices et sort en échec au premier écart."""
    analyseur = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    analyseur.add_argument("--fichier", type=Path, default=NOTICES)
    options = analyseur.parse_args()

    defauts, combien = controler(options.fichier)
    for d in defauts:
        print(f"  {d}", file=sys.stderr)
    if defauts:
        print(f"{len(defauts)} ecart(s) entre {options.fichier} et ce qui est lie",
              file=sys.stderr)
        return 1
    print(f"{options.fichier} : {combien} module(s) lié(s), tous déclarés à la bonne version")
    return 0


if __name__ == "__main__":
    sys.exit(main())
