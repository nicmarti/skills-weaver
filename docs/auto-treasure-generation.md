# Génération Automatique de Trésor Après Combat

**Date**: 2026-01-31
**Contexte**: Amélioration du workflow DM pour générer systématiquement du butin

## Motivation

Le Dungeon Master oubliait parfois de générer du trésor après les combats, créant une expérience frustrante pour les joueurs. Le tool `generate_treasure` existait déjà, mais n'était pas systématiquement utilisé.

Cette mise à jour ajoute des instructions claires et une section dédiée dans `dungeon-master.md` pour s'assurer que le DM génère TOUJOURS du butin après un combat.

## Changements Effectués

### 1. Nouvelle Section "Après le Combat" (dungeon-master.md)

Ajoutée après le "Workflow Combat Typique" (ligne 603), cette section comprend :

#### Workflow Post-Combat en 7 Étapes

```
1. Victoire des PJ
2. log_event {"event_type": "combat", "content": "Victoire contre 3 gobelins"}
3. add_xp {"amount": 150, "reason": "Combat gobelins"}
4. generate_treasure {"treasure_type": "R"}  ← OBLIGATOIRE
5. Décrire le butin narrativement
6. add_gold {"amount": <total_po>, "reason": "Butin gobelins"}
7. add_item pour chaque objet magique/utile trouvé
```

#### Tables de Référence Treasure Types

**Monstres courants** :

| Créature | CR | Treasure Type | Contenu typique |
|----------|----|--------------|-----------------|
| Gobelin | 1/4 | R | 3d6 pc, quelques pa |
| Orc | 1/2 | D | 2d6 pa, 3d6 pc |
| Ogre | 2 | C | 3d6 pp, gemmes possibles |
| Squelette/Zombie | 1/4 | none | Pas de trésor |
| Loup | 1/4 | none | Pas de trésor (animal) |

**Humanoïdes courants** :

| Créature | CR | Treasure Type | Contenu typique |
|----------|----|--------------|-----------------|
| Bandit | 1/8 | U | Quelques pc/pa, rien de spécial |
| Garde | 1/8 | B | 1d6 pa, équipement standard |
| Cultiste | 1/8 | U | Amulette, quelques pa |
| Voyou | 1/2 | U | 2d6 pa, objets volés |
| Noble | 1/8 | V | 5d6 po, bijoux précieux |
| Bandit Captain | 2 | V | Carte trésor, gemmes, or |
| Knight | 3 | V | Arme +1, armure fine, bourse |
| Mage | 6 | V | Parchemins, baguette, composants |

#### Cas Particuliers

**Animaux et morts-vivants** (treasure_type: "none") :
- Pas de `generate_treasure`
- Mais le DM peut improviser des composants narratifs :
  - "Vous récupérez une dent de loup (5 pa chez un alchimiste)"
  - "Le squelette portait un médaillon rouillé (1 po)"

**Groupes mixtes** :
- Génère pour le type le plus élevé
- Exemple : 3 gobelins + 1 chef gobelin → Type R (mais double la quantité)

**Boss importants** :
- Utilise leur treasure_type + improvise un objet narratif unique
- Exemple : Ogre (Type C) + "Grande hache de chef orcish (+1 dégât, valeur 50 po)"

#### Intégration Narrative

Le guide inclut des exemples de ce qu'il faut faire et éviter :

**INTERDIT** (Trop mécanique) :
```
Vous fouillez les corps. Vous trouvez 12 pc, 5 pa, 1 gemme de 10 po.
```

**CORRECT** (Narratif) :
```
En fouillant les gobelins morts, Marcus découvre une bourse de cuir
puant contenant quelques pièces de cuivre tachées de boue. Lyra
remarque qu'un des gobelins portait un collier grossier avec un
petit rubis mal taillé - probablement volé à un voyageur.

(12 pc, 5 pa, 1 rubis 10 po)
```

### 2. Correction Paramètre Tool (ligne 308)

Corrigé l'exemple du tool `generate_treasure` dans les "Tools Fréquents" :

```diff
- | `generate_treasure` | Butin | `{"type": "C"}` |
+ | `generate_treasure` | Butin | `{"treasure_type": "R"}` |
```

Le paramètre correct est `treasure_type`, pas `type`.

## Fichier Modifié

```
M  core_agents/agents/dungeon-master.md
   - Ligne 308: Correction paramètre tool generate_treasure
   - Ligne 603-701: Nouvelle section "### Après le Combat" (98 lignes)
```

## Exemple de Workflow Complet

### Avant (Sans Instructions Claires)

Le DM pouvait oublier le trésor :

```
DM: "Vous triomphez des trois gobelins !"
[PJ attendaient le butin... rien ne se passait]
```

### Après (Avec Instructions)

Le DM suit le workflow :

```
DM: "Vous triomphez des trois gobelins !"
1. log_event {"event_type": "combat", "content": "Victoire contre 3 gobelins"}
2. add_xp {"amount": 150, "reason": "Combat gobelins"}
3. generate_treasure {"treasure_type": "R"}
   → Résultat: 12 pc, 5 pa, 1 gemme (rubis) 10 po
4. Narration: "En fouillant les gobelins morts, Marcus découvre..."
5. add_gold {"amount": 10, "reason": "Butin gobelins"}
6. add_item {"item": "Gemme (rubis, 10 po)", "quantity": 1}
```

## Tests

Tous les tests passent après les modifications :

```bash
go build -o sw-dm ./cmd/dm
✅ SUCCESS

go test ./internal/agent/... -v
✅ PASS (80 tests)
```

## Impact sur les Sessions de Jeu

### Avantages

1. **Cohérence** : Le butin est TOUJOURS généré après un combat
2. **Équilibrage** : Les treasure types sont basés sur D&D 5e officiel
3. **Immersion** : Guide pour intégrer narrativement le butin
4. **Référence Rapide** : Tables de treasure_types par créature

### Pas de Breaking Changes

- Les aventures existantes continuent de fonctionner normalement
- C'est une amélioration du prompt du DM, pas du code
- Les tools existants (`generate_treasure`, `add_gold`, `add_item`) restent inchangés

## Utilisation par le DM

Le DM consultera maintenant cette section après chaque victoire :

1. **Identifier la créature** → `get_monster("goblin")` pour voir treasure_type
2. **Générer le trésor** → `generate_treasure {"treasure_type": "R"}`
3. **Narrer le résultat** → Description immersive du butin trouvé
4. **Ajouter à l'inventaire** → `add_gold` + `add_item` si nécessaire

## Treasure Types par Fichier

### data/5e/monsters.json

| Monstre | Treasure Type |
|---------|--------------|
| Gobelin | R |
| Orc | D |
| Ogre | C |
| Squelette | none |
| Zombie | none |
| Loup | none |
| Loup Sanguinaire | none |
| Ours | none |

### data/5e/humanoids.json

| Humanoïde | Treasure Type |
|-----------|--------------|
| Garde | B |
| Bandit | U |
| Cultiste | U |
| Acolyte | U |
| Noble | V |
| Voyou | U |
| Espion | U |
| Prêtre | U |
| Bandit Captain | V |
| Knight | V |
| Vétéran | U |
| Mage | V |
| Assassin | V |

## Conclusion

Cette amélioration transforme un outil existant sous-utilisé en un workflow obligatoire et bien documenté. Le DM ne pourra plus oublier le butin car :

1. ✅ **Instructions claires** : "OBLIGATOIRE" en gras
2. ✅ **Workflow structuré** : 7 étapes numérotées
3. ✅ **Tables de référence** : Treasure types par créature
4. ✅ **Exemples concrets** : Narrative vs mécanique
5. ✅ **Cas particuliers** : Animaux, boss, groupes mixtes

Le système de génération de trésor D&D 5e est maintenant pleinement intégré dans l'expérience de jeu SkillsWeaver.
