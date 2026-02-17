# TODO - SkillsWeaver

Liste des observations, bugs et idées d'amélioration.

---

## Bugs et Corrections

### Priorité Haute

- [ ] **Linter warnings `interface{}` → `any`**
  - Fichiers concernés : `combat_tools.go`, `character_tools.go`, `state_tools.go`, `cli_mapper.go`
  - Go 1.18+ recommande `any` au lieu de `interface{}`
  - Impact : Aucun fonctionnel, mais modernisation du code

- [ ] **Paramètre inutilisé dans `cli_mapper.go:106`**
  - Fonction `mapGetInventory(params)` n'utilise pas `params`
  - Solution : Soit utiliser le paramètre, soit le remplacer par `_`

### Priorité Moyenne

- [ ] **PV temporaires non gérés par `update_hp`**
  - Le champ `TempHitPoints` existe dans `Character` mais n'est pas utilisé
  - Comportement D&D 5e : Les PV temporaires absorbent les dégâts en premier
  - Fichier : `internal/dmtools/combat_tools.go`

---

## Tests Manquants

### Nouveaux Packages

- [ ] **`internal/i18n/dnd_terms.go`** - Aucun test
  ```go
  // Tests suggérés :
  - TestTranslate_CombatAbilities
  - TestTranslate_DamageTypes
  - TestTranslate_Conditions
  - TestTranslate_UnknownTerm_ReturnsOriginal
  ```

- [ ] **`internal/dmtools/combat_tools.go`** - Aucun test
  ```go
  // Tests suggérés :
  - TestUpdateHP_Damage
  - TestUpdateHP_Healing
  - TestUpdateHP_ClampToZero
  - TestUpdateHP_ClampToMax
  - TestUpdateHP_CharacterNotFound
  - TestUseSpellSlot_Success
  - TestUseSpellSlot_NoSlotsAvailable
  - TestUseSpellSlot_NotACaster
  ```

### Packages Existants Sans Tests

- [ ] `internal/equipment/` - Pas de fichiers de test
- [ ] `internal/spell/` - Pas de fichiers de test
- [ ] `internal/treasure/` - Pas de fichiers de test
- [ ] `internal/web/` - Pas de fichiers de test
- [ ] `internal/charactersheet/` - Pas de fichiers de test
- [ ] `internal/names/` - Pas de fichiers de test

---

## Nouvelles Fonctionnalités

### Combat (Priorité Haute)

- [ ] **Tool `rest`** - Gestion des repos
  ```
  Paramètres :
  - type: "short" | "long"
  - character_name: optionnel (tous si omis)

  Repos court :
  - Dépenser dés de vie pour récupérer PV
  - Récupérer capacités "1/repos court" (Second souffle, etc.)
  - Warlock : récupérer emplacements de sorts

  Repos long :
  - Récupérer tous les PV
  - Récupérer tous les emplacements de sorts
  - Récupérer la moitié des dés de vie (minimum 1)
  - Récupérer capacités "1/repos long"
  ```
  - Note : `Character.RestoreSpellSlots()` existe déjà

- [ ] **Tool `add_temp_hp`** - PV temporaires
  ```
  Paramètres :
  - character_name: string
  - amount: int
  - source: string (optionnel, ex: "Armure du mage")

  Comportement :
  - Ne se cumulent pas (garde le plus haut)
  - Disparaissent après repos long
  ```

- [ ] **Tool `apply_condition`** - Appliquer/retirer conditions
  ```
  Paramètres :
  - character_name: string
  - condition: string (prone, stunned, poisoned, etc.)
  - action: "add" | "remove"
  - duration: optionnel (tours ou "until_end_of_turn")

  Nécessite : Ajouter champ Conditions []string au Character
  ```

- [ ] **Tool `roll_initiative`** - Initiative de combat
  ```
  Paramètres :
  - include_monsters: bool
  - monsters: []string (noms des monstres)

  Retourne : Ordre d'initiative trié
  ```

### Magie (Priorité Moyenne)

- [ ] **Tool `cast_spell`** - Lancer un sort complet
  ```
  Combine :
  1. Vérification que le sort est connu/préparé
  2. use_spell_slot (si pas cantrip)
  3. Jets d'attaque ou dégâts automatiques
  4. Application des effets (update_hp pour soins/dégâts)

  Paramètres :
  - caster_name: string
  - spell_name: string
  - target: string | []string
  - spell_level: int (pour upcast)
  ```

- [ ] **Tool `prepare_spells`** - Préparer les sorts du jour
  ```
  Pour classes qui préparent (Clerc, Druide, Paladin, Magicien) :
  - Nombre max = niveau + mod caractéristique
  - Valider contre sorts connus/liste de classe
  ```

- [ ] **Gestion de la concentration**
  ```
  - Tracker quel personnage concentre sur quel sort
  - Rappel automatique lors de dégâts (JdS CON DD 10 ou dégâts/2)
  - Alerte si nouveau sort de concentration lancé

  Nécessite : Ajouter champ ConcentratingOn string au Character
  ```

### Traduction (Priorité Basse)

- [ ] **Étendre `internal/i18n/`**
  - Ajouter traductions des noms de sorts
  - Ajouter traductions des noms de monstres
  - Ajouter traductions des objets magiques
  - Créer fonction `TranslateSpell(spellID string) string`

- [ ] **Intégration automatique i18n dans les tools**
  - Les tools pourraient utiliser `i18n.Translate()` dans leurs réponses
  - Ex: `get_monster` pourrait traduire les types de dégâts automatiquement

---

## Améliorations UX

### Interface Web (`sw-web`)

- [ ] **Affichage des PV en temps réel**
  - Après `update_hp`, envoyer un événement SSE pour mettre à jour l'affichage
  - Barre de vie visuelle pour chaque personnage

- [ ] **Panneau de combat**
  - Tracker d'initiative visuel
  - Boutons rapides pour actions communes
  - Affichage des conditions actives

- [ ] **Historique des jets de dés**
  - Panneau latéral avec les derniers jets
  - Filtrable par type (attaque, dégâts, sauvegarde)

### CLI (`sw-dm`)

- [ ] **Commande `/combat`** - Mode combat dédié
  - Affiche l'ordre d'initiative
  - Raccourcis pour attaques et dégâts
  - Suivi automatique des tours

- [ ] **Commande `/status`** - État rapide du groupe
  - PV de tous les personnages
  - Conditions actives
  - Emplacements de sorts restants

- [ ] **Auto-complétion des noms de personnages**
  - Dans les paramètres `character_name` des tools

---

## Refactoring

### Code Quality

- [ ] **Extraire interface `Combatant`**
  ```go
  type Combatant interface {
      GetName() string
      GetHP() int
      GetMaxHP() int
      GetAC() int
      TakeDamage(amount int) int
      Heal(amount int) int
  }
  ```
  - Permettrait de traiter PJ et monstres de manière uniforme

- [ ] **Centraliser la gestion des personnages**
  - Actuellement dispersée entre `adventure.GetCharacters()` et chargement direct
  - Créer un `CharacterManager` avec cache

- [ ] **Standardiser les réponses des tools**
  ```go
  type ToolResponse struct {
      Success bool
      Display string        // Pour affichage utilisateur
      Data    interface{}   // Données structurées
      Error   string        // Si Success = false
  }
  ```

### Documentation

- [ ] **Documenter l'API des tools**
  - Fichier `docs/tools-api.md` avec tous les tools
  - Paramètres, retours, exemples

- [ ] **Guide de contribution**
  - Comment ajouter un nouveau tool
  - Conventions de code
  - Process de test

---

## Performance

- [ ] **Cache des personnages chargés**
  - Éviter de recharger les JSON à chaque appel de tool
  - Invalider le cache après modification

- [ ] **Lazy loading des données de référence**
  - Monstres, sorts, équipement chargés à la demande
  - Actuellement tout est chargé au démarrage

---

## Notes de Session

### 2025-01-31 - Implémentation Combat Tools

**Réalisé** :
- Créé `internal/i18n/dnd_terms.go` avec traductions FR
- Créé `internal/dmtools/combat_tools.go` avec `update_hp` et `use_spell_slot`
- Mis à jour `dungeon-master.md` avec documentation combat et terminologie FR

**Observations** :
- Le struct `Character` a `SpellSlots` comme champ, pas méthode
- `Character.UseSpellSlot(level)` et `Character.RestoreSpellSlots()` existent déjà
- Bonne séparation des responsabilités dans le code existant

---

## Priorités Suggérées

1. **Court terme** (prochaine session)
   - Corriger les linter warnings
   - Ajouter tests pour `combat_tools.go`
   - Implémenter tool `rest`

2. **Moyen terme** (semaine prochaine)
   - Tool `apply_condition`
   - Tool `roll_initiative`
   - Gestion PV temporaires

3. **Long terme** (backlog)
   - Mode combat dédié
   - Intégration i18n automatique
   - Interface web améliorée
