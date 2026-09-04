# Les règles de Cohue

## Le but

Survivre. Une horde converge sur vous sans arrêt, elle ne s'épuise pas, et vous
finirez par tomber. Ce qui compte est **combien de temps**, et ce que vous aurez
construit avant.

Quand vous mourez, une touche relance immédiatement, sur le même lieu.

## Ce que vous contrôlez

**Le déplacement, et rien d'autre.** Les flèches ou le carré ZQSD — les deux
marchent, et la diagonale se prend en pressant deux touches.

**Le tir est automatique.** Votre arme vise en permanence la créature la plus
proche à sa portée, dans toutes les directions, et tire dès qu'elle est prête.
Vous ne visez pas et vous n'appuyez pas : là où vous allez décide de ce que vous
touchez.

C'est pour cela que le déplacement est tout le jeu. Reculer, contourner, choisir
quel côté de la salle traverser — chaque pas est une décision de combat.

**Le sol n'est pas uniforme.** Certaines cases — une flaque, un sol fissuré —
coûtent double à traverser : vous y avancez deux fois moins vite. Ce n'est pas
un défaut de réactivité, c'est un terrain, et il ralentit la horde autant que
vous : coincé, on peut mener ses poursuivants dans une flaque pour prendre de
l'avance.

## La vie

Vous avez **cent points de vie**, affichés en haut à gauche.

Une créature qui vous touche vous blesse **en continu tant qu'elle vous touche**,
pas d'un coup. Plusieurs créatures collées ne cumulent pas indéfiniment : les
dégâts subis par seconde sont **plafonnés à vingt**, quel que soit leur nombre.

Autrement dit, être encerclé est très dangereux mais laisse une fenêtre pour se
dégager — environ cinq secondes depuis la pleine vie. Traversez, ne restez pas.

**Sous trente points de vie, les bords de l'écran rougissent** et le restent
jusqu'à ce que vous vous soigniez. C'est le seul signal que vous verrez sans
quitter votre personnage des yeux — la jauge est en haut à gauche, et on ne la
regarde pas en fuyant. Trente points, c'est aussi ce que rend une fiole : quand
la vignette s'allume, en boire une ne gaspille rien.

## Les gemmes

Une créature abattue laisse une **gemme** au sol, à l'endroit exact où elle
meurt. C'est votre expérience, et la seule.

**Marchez dessus pour la prendre.** La portée de ramassage est courte — une
tuile —, donc il faut y aller.

**Une gemme s'efface.** Elle s'éteint progressivement et disparaît au bout de six
secondes. Sa couleur dit son âge : une gemme vive vient de tomber, une gemme
sombre va partir. Vous pouvez donc estimer ce qu'un tas vaut encore avant de
décider d'y retourner.

C'est le cœur de la tension du jeu : **ramasser oblige à revenir là où vous venez
de tuer, c'est-à-dire là où la horde converge.** Le trajet de collecte va contre
le trajet de fuite.

## L'aimant

Un **aimant** apparaît régulièrement dans la salle, toujours à bonne distance de
vous. Il ne se déclenche pas quand vous marchez dessus : vous le **ramassez et
vous le gardez**.

Quand vous le dépensez, **toutes les gemmes au sol convergent vers vous**, où
qu'elles soient, et cessent de s'effacer pendant leur trajet.

Trois choses à savoir :

- vous n'en tenez **qu'un seul** à la fois ;
- un nouvel aimant n'apparaît pas tant que vous en gardez un — c'est une raison
  de dépenser celui que vous avez ;
- déclenché sans charge, il ne se passe rien.

L'aimant est le recours contre l'effacement. Attendre augmente la récolte, mais
les gemmes les plus anciennes partiront avant : c'est le seul vrai arbitrage de
la collecte.

## Monter de niveau

Les gemmes ramassées font monter de niveau. **Le seuil du niveau suivant
augmente à chaque fois**, mais une gemme vaut toujours la même chose : une
créature qui rapporte davantage en laisse plusieurs, et c'est la quantité au sol
qui vous dit ce que vous allez gagner.

**Et le temps compte aussi.** Un niveau vous est donné toutes les quarante-cinq
secondes, même si vous n'avez rien ramassé. Il ne retire rien de ce que vous avez
déjà : les gemmes acquises comptent pour le niveau suivant.

## Le choix

À chaque montée, **le jeu se fige et trois cartes s'affichent**. Les flèches
gauche et droite déplacent votre choix, qu'un liseré ambre désigne, et Entrée le
valide. Rien ne se décide tant que vous n'avez pas validé : le jeu attend.

Vous gardez la même arme du début à la fin : ce sont les cartes qui la
transforment.

| Carte | Ce qu'elle fait |
|---|---|
| **Cadence** | l'arme tire plus souvent |
| **Portée** | les tirs vont plus loin |
| **Souffle** | vous rend trente points de vie |

Chaque axe se prend **six fois au plus**. Quand vous l'avez épuisé, il disparaît
du menu et il faut basculer sur l'autre — c'est un moment de jeu, pas une
impasse. Quand les deux sont épuisés, seul le Souffle reste offert : les places
ne sont jamais vides.

## Les créatures

Elles vous poursuivent toutes, en contournant les obstacles. Elles se poussent
les unes les autres pour ne pas se superposer, si bien qu'une foule dense se
déforme au lieu de s'empiler.

Deux d'entre elles ne se contentent pas de foncer. L'**Arpenteur** dérive sur le
côté et vous contourne au lieu d'arriver de face. Le **Vigile** est le seul dont
le corps vous arrête : toutes les autres se traversent, lui vous bloque — pris
entre lui et un mur, vous tirez forcément dessus, puisque votre arme vise le plus
proche.

Le **Molosse** n'arrive jamais seul : il vient par trois, tous du même côté. Il
s'immobilise un instant avant de foncer, puis file en ligne droite sans plus
corriger sa trajectoire — un mur ou un pilier l'arrête, et il reste un moment
sans rien faire.

| Nom | Vitesse | Résistance | Dégâts au contact |
|---|---|---|---|
| **Badaud** | lente | 3 touches | 6 par seconde |
| **Molosse** | **très rapide** | 2 touches | 8 par seconde |
| **Arpenteur** | moyenne | 4 touches | 7 par seconde |
| **Buse** | lente | 5 touches | 4 par seconde |
| **Vigile** | **très lente** | **12 touches** | **10 par seconde** |
| **Baudruche** | moyenne | 4 touches | 5 par seconde |
| **Secouriste** | moyenne | 3 touches | 4 par seconde |

La résistance se compte en **touches de votre arme de base**, pas en points : une
arme qui grossit toute la partie rendrait un chiffre absolu illisible.

Aujourd'hui, le lieu de démonstration ne convoque que le **Badaud**. Les six
autres existent et attendent ce qui les rendra reconnaissables à l'écran : une
horde de sept sortes qu'aucune couleur ne distingue rend la difficulté illisible,
et on ne saurait pas dire ce qui a rendu une minute plus dure.

## L'écran

En haut à gauche, la jauge de **vie** en rouge, celle d'**expérience** en bleu, et
le niveau atteint. En haut à droite, le **temps écoulé**.

Le personnage est jaune, la horde rouge, les projectiles clairs, les gemmes
vertes.

Le sol se lit à sa teinte : gris pour une case libre, **bleuté pour une case
qui ralentit**, sombre pour un mur qu'on ne traverse pas.

## Les touches

| Touche | |
|---|---|
| Flèches, ou Z Q S D | se déplacer |
| 1, pavé numérique compris | dépenser l'aimant |
| Espace, en cours de partie | poser un repère dans le journal |
| Flèches ← → | désigner son choix, pendant une montée de niveau |
| Entrée, pavé numérique compris, ou Espace | valider le choix, et relancer après la mort |
