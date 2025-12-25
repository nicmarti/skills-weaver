# CLI Command Logging in sw-dm

## Vue d'ensemble

Depuis la dernière mise à jour, `sw-dm` log automatiquement les commandes CLI équivalentes pour chaque tool appelé par l'agent. Cela permet de :

1. **Reproduire facilement** : Copier-coller la commande pour rejouer l'opération
2. **Déboguer** : Comprendre exactement quels paramètres ont été utilisés
3. **Tester** : Affiner les outils en rejouant les commandes avec différents paramètres
4. **Documenter** : Voir toutes les opérations effectuées pendant une session

## Format du Log

⚠️ **Log Rotation** : Depuis la dernière mise à jour, les logs sont créés par session pour éviter des fichiers trop gros :
- `sw-dm-session-1.log`, `sw-dm-session-2.log`, etc.
- L'ancien `sw-dm.log` monolithique est automatiquement archivé s'il fait > 1MB
- Voir [docs/log-rotation.md](log-rotation.md) pour plus de détails

Le log est écrit dans `data/adventures/<nom-aventure>/sw-dm-session-N.log`. Pour chaque tool call, vous verrez :

```
[2025-12-25 18:55:17] TOOL CALL: generate_npc (ID: toolu_019mdxsi5zFptF9nq7a8q75N)
  Parameters:
  {
    "attitude": "friendly",
    "context": "Aubergiste expérimentée à L'Étoile de Garde, Valbourg",
    "gender": "f",
    "occupation": "aubergiste",
    "race": "human"
  }
  Equivalent CLI:
  ./sw-npc generate --race=human --gender=f --occupation=aubergiste --attitude=friendly

[2025-12-25 18:55:17] TOOL RESULT: generate_npc (ID: toolu_019mdxsi5zFptF9nq7a8q75N)
  Result:
  {
    "display": "Marta Ironwood - humain femme, aubergiste (amical, serviable) [ID: npc_003, saved to session_5]",
    ...
  }
```

## Exemples par Tool

### generate_map (le plus utile !)

```
[2025-12-25 19:30:45] TOOL CALL: generate_map (ID: toolu_01Abc...)
  Parameters:
  {
    "type": "city",
    "name": "Port-Sombre",
    "kingdom": "valdorine",
    "features": ["Port", "Docks", "Marché aux poissons"],
    "generate_image": true
  }
  Equivalent CLI:
  ./sw-map generate city "Port-Sombre" --kingdom=valdorine --features="Port,Docks,Marché aux poissons" --generate-image
```

**Usage** : Copiez la commande CLI pour régénérer la carte avec des ajustements :

```bash
# Régénérer la même carte
./sw-map generate city "Port-Sombre" --kingdom=valdorine --features="Port,Docks,Marché aux poissons" --generate-image

# Tester sans image (plus rapide)
./sw-map generate city "Port-Sombre" --kingdom=valdorine --features="Port,Docks,Marché aux poissons"

# Ajouter plus de features
./sw-map generate city "Port-Sombre" --kingdom=valdorine --features="Port,Docks,Marché aux poissons,Taverne,Forteresse" --generate-image
```

### generate_npc

```
[2025-12-25 19:31:12] TOOL CALL: generate_npc (ID: toolu_01Def...)
  Parameters:
  {
    "race": "dwarf",
    "gender": "m",
    "occupation": "forgeron",
    "attitude": "neutral"
  }
  Equivalent CLI:
  ./sw-npc generate --race=dwarf --gender=m --occupation=forgeron --attitude=neutral
```

### generate_treasure

```
[2025-12-25 19:32:05] TOOL CALL: generate_treasure (ID: toolu_01Ghi...)
  Parameters:
  {
    "treasure_type": "r"
  }
  Equivalent CLI:
  ./sw-treasure generate R
```

### generate_image

```
[2025-12-25 19:33:20] TOOL CALL: generate_image (ID: toolu_01Jkl...)
  Parameters:
  {
    "type": "scene",
    "description": "Combat contre des gobelins dans une crypte sombre",
    "scene_type": "battle"
  }
  Equivalent CLI:
  ./sw-image scene "Combat contre des gobelins dans une crypte sombre" --type=battle
```

### roll_dice

```
[2025-12-25 19:34:10] TOOL CALL: roll_dice (ID: toolu_01Mno...)
  Parameters:
  {
    "notation": "2d6+3"
  }
  Equivalent CLI:
  ./sw-dice roll 2d6+3
```

### get_monster

```
[2025-12-25 19:35:00] TOOL CALL: get_monster (ID: toolu_01Pqr...)
  Parameters:
  {
    "monster_id": "goblin"
  }
  Equivalent CLI:
  ./sw-monster show goblin
```

## Tools Sans Équivalent CLI

Certains tools sont purement internes et n'ont pas de commande CLI équivalente. Dans ce cas, la ligne "Equivalent CLI:" n'est pas affichée :

- `log_event` : Enregistre dans le journal (interne à l'aventure)
- `add_gold` : Modifie l'inventaire (interne à l'aventure)
- `get_inventory` : Consulte l'inventaire (interne à l'aventure)
- `update_npc_importance` : Met à jour les NPCs (interne à l'aventure)
- `get_npc_history` : Consulte l'historique des NPCs (interne à l'aventure)

## Workflow Recommandé

### 1. Session de Jeu Normale

Jouez normalement avec `./sw-dm`. Le log capture automatiquement toutes les commandes.

### 2. Review Post-Session

Ouvrez le log pour voir toutes les opérations :

```bash
tail -n 500 data/adventures/la-crypte-des-ombres/sw-dm.log
```

### 3. Repérer une Commande Intéressante

Cherchez "Equivalent CLI:" pour trouver les commandes à rejouer :

```bash
grep "Equivalent CLI:" data/adventures/la-crypte-des-ombres/sw-dm.log
```

### 4. Rejouer et Améliorer

Copiez la commande et testez des variations :

```bash
# Commande originale
./sw-map generate city "Valbourg" --kingdom=karvath

# Test avec features
./sw-map generate city "Valbourg" --kingdom=karvath --features="Caserne,Mur d'enceinte,Place d'armes"

# Test avec génération d'image
./sw-map generate city "Valbourg" --kingdom=karvath --features="Caserne,Mur d'enceinte,Place d'armes" --generate-image
```

### 5. Itérer

Utilisez le feedback pour affiner les prompts et améliorer les outils.

## Grep Patterns Utiles

```bash
# Voir toutes les maps générées
grep -A 10 "generate_map" data/adventures/*/sw-dm.log

# Voir toutes les commandes CLI
grep "Equivalent CLI:" data/adventures/*/sw-dm.log

# Voir les NPCs générés avec leurs commandes
grep -B 5 "generate_npc" data/adventures/*/sw-dm.log | grep "Equivalent CLI:"

# Extraire juste les commandes CLI (sans le préfixe)
grep "Equivalent CLI:" data/adventures/*/sw-dm.log | sed 's/.*Equivalent CLI://'
```

## Avantages

✅ **Debugging facile** : Voir exactement ce qui a été appelé
✅ **Reproduction 1:1** : Copier-coller la commande pour rejouer
✅ **Testing itératif** : Ajuster les paramètres et comparer
✅ **Documentation automatique** : Historique complet des opérations
✅ **Amélioration des outils** : Identifier les patterns d'usage

## Implémentation Technique

Le système est implémenté en deux parties :

1. **cli_mapper.go** : Convertit les tool calls en commandes CLI
2. **logger.go** : Écrit les commandes dans le log

Voir `internal/agent/cli_mapper.go` pour ajouter le support de nouveaux tools.
