# Crédits

Une ligne par ressource tierce, avec sa licence, tenue **dès son introduction**
et jamais après : retrouver la provenance d'un fichier six mois plus tard est
impossible. Le fichier de licence d'origine est conservé à côté des fichiers
concernés quand la licence l'exige.

## Décor et personnages

Le décor est généré par `outils/decor_iso.py`. La matière de ses revêtements
vient de textures originales, versionnées dans `outils/textures/`, que le
générateur applique ; quelques formes sont dessinées en entier, et le générateur
n'en écrit alors que l'entrée de manifeste. Aucune source tierce : tout suit la
licence du dépôt, dont le texte est conservé dans `assets/decors/LICENSE.txt`
pour les formes et dans `outils/textures/LICENSE.txt` pour les matières.

Les personnages sont des dessins originaux, tenus à la main et non régénérés,
sous licence Apache 2.0 comme le reste du dépôt ; le texte en est conservé dans
le `LICENSE.txt` du dossier de chaque profil, sous `assets/personnages/`.

## Police

**Pixel Operator 8**, de Jayvee Enaguas (HarvettFox96), version 2018.10.04-1,
sous licence CC0 1.0 — <https://creativecommons.org/licenses/zero/1.0/>. Le
texte de la licence est conservé dans `assets/polices/LICENSE.txt`.

Le fichier `.ttf` est la source, jamais ce que le jeu lit : `outils/interface.py`
en rastérise une planche de glyphes que le moteur charge par son manifeste. Seule la graisse normale est retenue — le gras n'a pas d'usage tant
que le coup critique n'existe pas.

## Bruitages

Générés par `outils/sons.py`, par synthèse. Aucune source tierce.

## Musique

À compléter. Vérifier la licence piste par piste : les banques libres mélangent
CC0 et CC-BY, et seule la seconde impose l'attribution. Une piste sous licence
entrera ici avec son auteur, son URL et sa licence, dans le commit qui
l'introduit.
