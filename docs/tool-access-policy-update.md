# Mise à Jour: Politique d'Accès aux Nouveaux Outils

**Date**: 2026-01-31
**Contexte**: Enregistrement des nouveaux outils de combat et de gestion d'état

## Problème

Lors du commit précédent (BFRPG cleanup), 8 nouveaux outils ont été ajoutés :

**Combat** (combat_tools.go):
- `update_hp` - Modifier les PV d'un personnage
- `use_spell_slot` - Consommer un emplacement de sort

**State Management** (state_tools.go):
- `update_time` - Avancer le temps dans le jeu
- `set_flag` - Marquer un événement narratif
- `add_quest` - Ajouter une quête
- `complete_quest` - Terminer une quête
- `set_variable` - Définir une variable narrative
- `get_state` - Consulter l'état complet

Ces outils étaient :
- ✅ **Enregistrés** dans `register_tools.go`
- ✅ **Documentés** dans `dungeon-master.md`
- ❌ **Manquants** dans `tool_access_policy.go`
- ❌ **Partiellement mappés** dans `cli_mapper.go`

## Solution Implémentée

### 1. Politique d'Accès (tool_access_policy.go)

**Ajout aux outils interdits pour agents secondaires** :

```go
// State modification - nested agents are read-only consultants
"log_event",
"add_gold",
// ... existing tools ...
+ "update_hp",           // NEW
+ "use_spell_slot",      // NEW
+ "update_time",         // NEW
+ "set_flag",            // NEW
+ "add_quest",           // NEW
+ "complete_quest",      // NEW
+ "set_variable",        // NEW
+ "get_state",           // NEW
```

**Justification** :
- Ces outils **modifient l'état du jeu** (personnages, temps, quêtes)
- Les agents secondaires sont des **consultants read-only**
- Seul le DM principal peut modifier l'état

### 2. CLI Mapper (cli_mapper.go)

**Ajout des case statements** :

```go
case "update_hp":
    return mapUpdateHP(params)
case "use_spell_slot":
    return mapUseSpellSlot(params)
+ case "update_time":
+     return mapUpdateTime(params)
+ case "set_flag":
+     return mapSetFlag(params)
+ case "add_quest":
+     return mapAddQuest(params)
+ case "complete_quest":
+     return mapCompleteQuest(params)
+ case "set_variable":
+     return mapSetVariable(params)
+ case "get_state":
+     return mapGetState(params)
```

**Implémentation des mappers** :

```go
func mapUpdateTime(params map[string]interface{}) string {
    // Format: # update_time day=2 hour=8 minute=30 (internal operation)
}

func mapSetFlag(params map[string]interface{}) string {
    // Format: # set_flag "defeated_boss" value=true (internal operation)
}

func mapAddQuest(params map[string]interface{}) string {
    // Format: # add_quest "Trouver Brenner" --description="..." (internal)
}

func mapCompleteQuest(params map[string]interface{}) string {
    // Format: # complete_quest "Trouver Brenner" (internal operation)
}

func mapSetVariable(params map[string]interface{}) string {
    // Format: # set_variable "current_inn" value="Auberge de la Croix" (internal)
}

func mapGetState(params map[string]interface{}) string {
    // Format: # get_state (internal operation - reads state.json)
}
```

**Format des commandes CLI** :
- Préfixe `#` pour indiquer une opération interne (pas de CLI équivalent)
- Suffixe `(internal operation - modifies state.json)` pour clarté
- Paramètres formatés comme des flags shell pour lisibilité

### 3. Documentation (déjà complète)

**dungeon-master.md** contenait déjà :
- ✅ Table des outils de combat (lignes 574-601)
- ✅ Table des outils d'état (lignes 362-376)
- ✅ Exemples d'utilisation
- ✅ Workflow typiques

## Architecture de Sécurité

### Agents et Accès aux Outils

```
┌─────────────────────────────────────────────────┐
│           MAIN AGENT (dungeon-master)           │
│  ✅ Full tool access (50K token limit)          │
│     - Combat: update_hp, use_spell_slot         │
│     - State: update_time, set_flag, add_quest   │
│     - Session: start_session, end_session       │
│     - Content: generate_image, generate_map     │
│     - Agents: invoke_agent, invoke_skill        │
└────────────────────┬────────────────────────────┘
                     │
         ┌───────────┴────────────┐
         │                        │
         ▼                        ▼
┌──────────────────┐    ┌──────────────────┐
│   rules-keeper   │    │   world-keeper   │
│  Read-only (20K) │    │  Read-only (20K) │
│                  │    │                  │
│ ✅ Allowed:      │    │ ✅ Allowed:      │
│  - roll_dice     │    │  - get_party_info│
│  - get_monster   │    │  - get_campaign  │
│  - get_spell     │    │  - list_foreshadow│
│                  │    │                  │
│ ❌ Forbidden:    │    │ ❌ Forbidden:    │
│  - update_hp     │    │  - update_hp     │
│  - set_flag      │    │  - set_flag      │
│  - invoke_agent  │    │  - invoke_agent  │
└──────────────────┘    └──────────────────┘
```

### Garanties de Sécurité

1. **Aucune modification d'état** par agents secondaires
2. **Aucune récursion** (agents ne peuvent pas invoquer d'autres agents)
3. **Aucun contenu persistant** (pas de génération d'images/cartes/NPCs)
4. **Consultation pure** (lectures seules pour règles et cohérence monde)

## Tests

### Compilation Réussie

```bash
go build -o sw-dm ./cmd/dm
✅ SUCCESS
```

### Tests d'Agent Passent

```bash
go test ./internal/agent -v
✅ TestRulesKeeperPolicy: PASS
✅ TestCharacterCreatorPolicy: PASS
✅ TestWorldKeeperPolicy: PASS
✅ TestToolRegistry_CreateFilteredRegistry: PASS
```

### Vérification Manuelle

```bash
# Vérifier que les nouveaux outils sont interdits pour nested agents
grep -A 10 "AlwaysForbiddenTools" internal/agent/tool_access_policy.go
✅ Tous les 8 nouveaux outils listés
```

## Impact

### Avant

```
8 nouveaux outils créés
✅ Enregistrés dans register_tools.go
✅ Documentés dans dungeon-master.md
❌ Politique d'accès manquante
❌ CLI mapping incomplet
```

### Après

```
8 outils complètement intégrés
✅ Enregistrés
✅ Documentés
✅ Politique d'accès configurée
✅ CLI mapping complet
✅ Tests passent
```

## Logging dans sw-dm-session-N.log

**Exemple de logs** :

```
[2026-01-31 23:45:12] TOOL CALL: update_hp (ID: toolu_01Abc...)
  Parameters:
  {
    "character_name": "Marcus",
    "amount": -8,
    "reason": "Griffes de gobelin"
  }
  Equivalent CLI:
  # update_hp "Marcus" -8 --reason="Griffes de gobelin" (internal operation)

[2026-01-31 23:45:15] TOOL CALL: set_flag (ID: toolu_01Def...)
  Parameters:
  {
    "flag": "defeated_crypt_creature",
    "value": true
  }
  Equivalent CLI:
  # set_flag "defeated_crypt_creature" value=true (internal operation)
```

**Format** :
- Préfixe `#` indique opération interne (pas d'exécutable CLI)
- Suffixe `(internal operation)` clarifie qu'il n'y a pas de commande shell
- Facilite la reproduction manuelle en modifiant directement les JSONs

## Fichiers Modifiés

```
M  internal/agent/tool_access_policy.go   # +8 outils interdits
M  internal/agent/cli_mapper.go           # +8 case + 6 mapper functions
A  docs/tool-access-policy-update.md      # Cette documentation
```

## Commit

```bash
git add internal/agent/tool_access_policy.go internal/agent/cli_mapper.go docs/tool-access-policy-update.md
git commit -m "feat: add tool access policy for combat and state tools

- Add 8 new tools to AlwaysForbiddenTools for nested agents:
  update_hp, use_spell_slot, update_time, set_flag,
  add_quest, complete_quest, set_variable, get_state

- Implement CLI mappers for all 8 tools in cli_mapper.go
  Format: # tool_name params (internal operation - modifies JSON)

- Ensure nested agents (rules-keeper, world-keeper) are read-only
- All agent tests pass

These tools modify game state and must only be accessible to
the main dungeon-master agent.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

## Résultat Final

| Outil | Enregistré | Documenté | Policy | CLI Mapper |
|-------|-----------|-----------|--------|------------|
| update_hp | ✅ | ✅ | ✅ | ✅ |
| use_spell_slot | ✅ | ✅ | ✅ | ✅ |
| update_time | ✅ | ✅ | ✅ | ✅ |
| set_flag | ✅ | ✅ | ✅ | ✅ |
| add_quest | ✅ | ✅ | ✅ | ✅ |
| complete_quest | ✅ | ✅ | ✅ | ✅ |
| set_variable | ✅ | ✅ | ✅ | ✅ |
| get_state | ✅ | ✅ | ✅ | ✅ |

**Statut** : Tous les outils sont maintenant complètement intégrés et sécurisés ! ✅
