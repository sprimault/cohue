# Sécurité

English: [SECURITY.md](SECURITY.md)

## Versions suivies

La dernière version publiée. En `0.x`, il n'y a pas de branche de maintenance :
un correctif sort dans la version suivante.

## Signaler une faille

Par un avis de sécurité privé sur le dépôt GitHub, jamais par une issue
publique. Réponse sous quelques jours.

## Ce qui est dans le périmètre

Le jeu charge des fichiers écrits par des tiers : niveaux, campagnes. C'est la
seule surface d'attaque réelle, et c'est celle qui compte.

- Un niveau qui fait planter le jeu, boucler indéfiniment ou consommer toute la
  mémoire au chargement.
- Un chemin de fichier construit depuis un niveau qui sortirait du dossier prévu.
- Tout ce qui ferait exécuter du code depuis un contenu partagé — **rien dans le
  format ne le permet, et c'est un invariant** : un niveau ne contient que des
  identifiants et des positions, aucun binaire, aucun script.

## Ce qui n'y est pas

Modifier ses propres fichiers de sauvegarde pour tricher hors ligne. Le jeu est
solo : il n'y a rien à protéger contre son propriétaire.
