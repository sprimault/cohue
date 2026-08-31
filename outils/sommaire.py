# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

"""Contrôle le sommaire par question de docs/go.md.

Le sommaire indexe le document par le problème qu'on cherche à résoudre, là où
les titres le décrivent par ce qu'ils affirment. Ce n'est donc pas une copie de
la structure mais une information qu'elle ne contient pas — et c'est ce qui le
rend légitime au regard de la règle des deux descriptions.

Reste que ses ancres, elles, sont bien une seconde description : une ancre vers
un titre renommé ne produit aucune erreur en Markdown, elle ne fait rien. Le
lecteur clique, la page ne bouge pas, et il conclut que le document est mal tenu
plutôt que le sommaire périmé. Un sommaire que rien ne vérifie n'est donc pas
neutre : il égare, ce qui est pire que son absence.

Deux sens, et le second est celui qui se produira :

- une entrée qui pointe vers un titre absent ;
- une section de premier rang qu'aucune entrée n'indexe — quelqu'un ajoute une
  règle sans la référencer, et elle devient invisible à qui cherche par le
  problème.

Les sous-sections sont indexables mais pas exigées : toutes ne méritent pas une
question, et une liste d'exemptions serait la troisième description de trop.
"""

import argparse
import re
import sys
import unicodedata
from pathlib import Path

# Le document contrôlé. Un seul aujourd'hui ; le jour où il y en aura deux, ce
# sera un argument de ligne de commande et non une seconde constante.
DOCUMENT = Path("docs/go.md")

# Le sommaire vit entre ces deux repères. Des commentaires HTML plutôt qu'un
# titre : ils ne s'affichent pas et ne peuvent pas devenir eux-mêmes une entrée.
DEBUT = "<!-- sommaire -->"
FIN = "<!-- fin du sommaire -->"

TITRE = re.compile(r"^(#{2,3}) (.+)$")
LIEN = re.compile(r"\]\(#([^)]+)\)")


def ancre(titre):
    """Rend l'ancre que GitHub fabrique pour un titre.

    Minuscules, ponctuation retirée, espaces en tirets. Les accents restent —
    c'est ce qui distingue cette règle de la plupart des générateurs, et ce qui
    fait qu'on ne peut pas la deviner sans l'écrire.
    """
    sans_ponctuation = "".join(
        c for c in titre.lower()
        if c.isalnum() or c in " -" or unicodedata.combining(c)
    )
    return sans_ponctuation.strip().replace(" ", "-")


def sections(lignes):
    """Rend les ancres du document, et celles de premier rang à part."""
    toutes, premier_rang = {}, {}
    dans_sommaire = False
    for ligne in lignes:
        if ligne.startswith(DEBUT):
            dans_sommaire = True
        elif ligne.startswith(FIN):
            dans_sommaire = False
        if dans_sommaire:
            continue
        if m := TITRE.match(ligne):
            niveau, titre = m.group(1), m.group(2)
            toutes[ancre(titre)] = titre
            if niveau == "##":
                premier_rang[ancre(titre)] = titre
    return toutes, premier_rang


def entrees(lignes):
    """Rend les ancres citées par le sommaire, ou lève s'il est absent.

    Un sommaire absent est un défaut et non un contrôle sans objet : c'est
    exactement la façon dont ce contrôle se désarmerait sans qu'on le touche.
    """
    dedans, citees = False, []
    for ligne in lignes:
        if ligne.startswith(DEBUT):
            dedans = True
            continue
        if ligne.startswith(FIN):
            return citees
        if dedans:
            citees.extend(LIEN.findall(ligne))
    raise SystemExit(f"{DOCUMENT} : sommaire absent, repères « {DEBUT} » et « {FIN} » attendus")


def controler(chemin):
    """Rend la liste des défauts du sommaire."""
    lignes = chemin.read_text(encoding="utf-8").splitlines()
    toutes, premier_rang = sections(lignes)
    citees = entrees(lignes)

    defauts = []
    for a in citees:
        if a not in toutes:
            defauts.append(f"l'entrée « #{a} » ne pointe sur aucun titre")
    for a, titre in premier_rang.items():
        if a not in citees:
            defauts.append(f"« {titre} » n'est indexée par aucune entrée")
    if len(citees) != len(set(citees)):
        vues = set()
        for a in citees:
            if a in vues:
                defauts.append(f"l'entrée « #{a} » est citée deux fois")
            vues.add(a)
    return defauts


def main():
    """Contrôle le sommaire et sort en échec au premier défaut."""
    analyseur = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    analyseur.add_argument("--document", type=Path, default=DOCUMENT,
                           help="le document à contrôler")
    options = analyseur.parse_args()

    defauts = controler(options.document)
    for d in defauts:
        print(f"  {d}", file=sys.stderr)
    if defauts:
        print(f"{len(defauts)} defaut(s) de sommaire dans {options.document}", file=sys.stderr)
        return 1
    print(f"sommaire de {options.document} : chaque entrée pointe, chaque section est indexée")
    return 0


if __name__ == "__main__":
    sys.exit(main())
