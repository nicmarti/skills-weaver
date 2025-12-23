# Plan : Générateur de Noms de Lieux

## Architecture Recommandée : Hybride

### 1. Skill Autonome `name-location-generator`
- Génère des noms rapidement selon faction et type
- CLI : `sw-location-names generate <type> --kingdom=<faction>`
- Utilisable directement par le dungeon-master

### 2. Intégration dans world-keeper
- Utilise la skill pour générer
- Vérifie cohérence (nom n'existe pas déjà)
- Documente automatiquement dans `geography.json`
- Commande : `/world-create-location <type> <kingdom>`

---

## Styles de Noms par Faction

### Valdorine (Maritime, Pragmatique)
**Style** : Maritime, commercial, cosmopolite

**Villes/Ports** :
- Préfixes : Cor-, Port-, Havre-, Mar-, Nav-
- Racines : -dova, -luna, -vel, -mar, -pel
- Suffixes : -ia, -aven, -bay, -point
- Exemples : Cordova, Port-de-Lune, Havre-d'Argent, Marvelia

**Villages** :
- Préfixes : San-, Belle-, Beau-, Clair-
- Suffixes : -mont, -val, -fort, -roc
- Exemples : Bellerive, Clairmont, Beaufort

**Régions** :
- Style : Descriptif maritime
- Exemples : Côte Occidentale, Golfe des Marchands, Îles d'Or

---

### Karvath (Militariste, Défensif)
**Style** : Fort, dur, martial, honneur

**Forteresses/Villes** :
- Préfixes : Fer-, Acier-, Roc-, Forte-, Garde-
- Racines : -lance, -marteau, -épée, -porte, -mur
- Suffixes : -garde, -fort, -heim, -burg
- Exemples : Fer-de-Lance, Porte-de-Fer, Forge-Noire, Rocburg

**Villages** :
- Préfixes : Haut-, Mont-, Val-
- Suffixes : -garde, -bourg, -stein, -wald
- Exemples : Hautegarde, Valbourg, Steinfeld

**Régions** :
- Style : Géographique militaire
- Exemples : Montagnes de Fer, Plaines du Bouclier, Défilé de l'Aigle

---

### Lumenciel (Théocratique, Hypocrite)
**Style** : Lumineux, pieux, céleste (mais cache corruption)

**Cités Religieuses** :
- Préfixes : Aurore-, Lumière-, Saint-, Céleste-, Divine-
- Racines : -sancta, -lumen, -ciel, -âme, -lumina
- Suffixes : -sainte, -bénie, -divine, -des-anges
- Exemples : Aurore-Sainte, Lumenciel, Saint-Aethel, Divinia

**Villages** :
- Préfixes : Vallon-, Pré-, Mont-
- Suffixes : -de-Prière, -du-Repos, -de-Foi, -Béni
- Exemples : Vallon-de-Prière, Pré-Béni, Mont-de-Foi

**Régions** :
- Style : Spirituel et naturel
- Exemples : Terres Saintes, Forêt de la Grâce, Val de Lumière

---

### Astrène (Décadent, Érudit)
**Style** : Noble ancien, mélancolique, érudit

**Cités Impériales** :
- Préfixes : Étoile-, Lune-, Astro-, Nyx-, Alba-
- Racines : -automne, -crépuscule, -aurore, -noctis
- Suffixes : -d'Automne, -Impériale, -Ancienne, -d'Éther
- Exemples : Étoile-d'Automne, Lune-Crépusculaire, Albastra

**Villages** :
- Préfixes : Val-, Ombre-, Brume-
- Suffixes : -ombre, -ancienne, -fanée, -oubliée
- Exemples : Valombre, Brume-Ancienne, Ombreuse

**Régions** :
- Style : Mélancolique et ancien
- Exemples : Terres du Sud, Val de l'Oubli, Plaines Fanées

---

## Types de Lieux

### Par Taille
1. **Continent** : Noms épiques (Le Grand Continent, Terres des Quatre Royaumes)
2. **Région** : Descriptifs géographiques (Montagnes de Fer, Côte Occidentale)
3. **Cité** : Noms complets selon faction
4. **Ville** : Noms moyens
5. **Village** : Noms simples
6. **Hameau** : Prénom + suffixe

### Par Type Géographique
- **Port** : Préfixes maritimes (Port-, Havre-, Mar-)
- **Forteresse** : Préfixes militaires (Fer-, Garde-, Roc-)
- **Monastère** : Préfixes religieux (Saint-, Vallon-)
- **Ruines** : Préfixes anciens (Vieux-, Ancienne-, Oublié-)

---

## Structure de Données : `data/location-names.json`

```json
{
  "valdorine": {
    "cities": {
      "prefixes": ["Cor", "Port", "Havre", "Mar", "Nav"],
      "roots": ["dova", "luna", "vel", "mar", "pel"],
      "suffixes": ["ia", "aven", "bay", "point"]
    },
    "villages": {
      "prefixes": ["San", "Belle", "Beau", "Clair"],
      "suffixes": ["mont", "val", "fort", "roc"]
    },
    "regions": {
      "templates": [
        "Côte {adjective}",
        "Golfe {noun-gen}",
        "Îles {adjective-pl}"
      ],
      "adjectives": ["Occidentale", "Dorée", "Argentée"],
      "nouns": ["Marchands", "Navigateurs", "Corsaires"]
    }
  },
  "karvath": {
    "cities": {
      "prefixes": ["Fer", "Acier", "Roc", "Forte", "Garde"],
      "roots": ["lance", "marteau", "épée", "porte", "mur"],
      "suffixes": ["garde", "fort", "heim", "burg"]
    },
    "villages": {
      "prefixes": ["Haut", "Mont", "Val"],
      "suffixes": ["garde", "bourg", "stein", "wald"]
    },
    "regions": {
      "templates": [
        "Montagnes {noun-gen}",
        "Plaines {noun-gen}",
        "Défilé {noun-gen}"
      ],
      "nouns": ["Fer", "Bouclier", "Aigle", "Acier"]
    }
  },
  "lumenciel": {
    "cities": {
      "prefixes": ["Aurore", "Lumière", "Saint", "Céleste", "Divine"],
      "roots": ["sancta", "lumen", "ciel", "âme", "lumina"],
      "suffixes": ["sainte", "bénie", "divine", "des-anges"]
    },
    "villages": {
      "prefixes": ["Vallon", "Pré", "Mont"],
      "suffixes": ["de-Prière", "du-Repos", "de-Foi", "Béni"]
    },
    "regions": {
      "templates": [
        "Terres {adjective-pl}",
        "Forêt {noun-gen}",
        "Val {noun-gen}"
      ],
      "adjectives": ["Saintes", "Bénies", "Divines"],
      "nouns": ["Grâce", "Lumière", "Pureté"]
    }
  },
  "astrene": {
    "cities": {
      "prefixes": ["Étoile", "Lune", "Astro", "Nyx", "Alba"],
      "roots": ["automne", "crépuscule", "aurore", "noctis", "stra"],
      "suffixes": ["d'Automne", "Impériale", "Ancienne", "d'Éther"]
    },
    "villages": {
      "prefixes": ["Val", "Ombre", "Brume"],
      "suffixes": ["ombre", "ancienne", "fanée", "oubliée"]
    },
    "regions": {
      "templates": [
        "Terres {noun-gen}",
        "Val {noun-gen}",
        "Plaines {adjective-pl}"
      ],
      "nouns": ["Sud", "Oubli", "Passé"],
      "adjectives": ["Fanées", "Anciennes", "Oubliées"]
    }
  },
  "neutral": {
    "ruins": {
      "prefixes": ["Vieux", "Ancienne", "Oublié", "Perdu"],
      "suffixes": ["Ruines", "Vestiges", "Décombres"]
    },
    "generic": {
      "geographical": [
        "Collines",
        "Montagnes",
        "Vallée",
        "Plaines",
        "Forêt",
        "Marais",
        "Désert"
      ],
      "adjectives": [
        "Sombre",
        "Brumeux",
        "Verdoyant",
        "Aride",
        "Gelé",
        "Brûlé"
      ]
    }
  }
}
```

---

## Implémentation Technique

### 1. Package Go : `internal/locations/`

```go
// internal/locations/locations.go
package locations

type LocationNames struct {
    Cities   NameParts `json:"cities"`
    Villages NameParts `json:"villages"`
    Regions  RegionTemplates `json:"regions"`
}

type NameParts struct {
    Prefixes []string `json:"prefixes"`
    Roots    []string `json:"roots"`
    Suffixes []string `json:"suffixes"`
}

type RegionTemplates struct {
    Templates  []string `json:"templates"`
    Adjectives []string `json:"adjectives"`
    Nouns      []string `json:"nouns"`
}

func GenerateCity(kingdom string) string
func GenerateVillage(kingdom string) string
func GenerateRegion(kingdom string) string
```

### 2. CLI : `cmd/location-names/main.go`

```bash
# Compiler
go build -o sw-location-names ./cmd/location-names

# Utilisation
sw-location-names city --kingdom=valdorine
sw-location-names village --kingdom=karvath
sw-location-names region --kingdom=lumenciel
sw-location-names port --kingdom=valdorine
sw-location-names fortress --kingdom=karvath

# Options
--count=N       # Générer N noms
--type=<type>   # city, village, region, port, fortress, ruins
```

### 3. Skill : `.claude/skills/name-location-generator/skill.md`

Invocation : `/name-location-generator`

### 4. Intégration world-keeper

Ajouter commandes :
```bash
/world-create-location <type> <kingdom>
```

Le world-keeper :
1. Utilise `sw-location-names` pour générer
2. Vérifie que le nom n'existe pas dans `geography.json`
3. Si existe, régénère
4. Documente dans `geography.json`
5. Retourne le nom au DM

---

## Exemples d'Utilisation

### Pour le Dungeon Master (impro rapide)
```bash
# Besoin rapide d'un nom de ville valdine
sw-location-names city --kingdom=valdorine
# Output: Marvelia

# Plusieurs options
sw-location-names village --kingdom=karvath --count=5
# Output: Hautegarde, Valbourg, Steinfeld, Rocheim, Eisenwald
```

### Pour le World-Keeper (documentation)
```bash
# Création avec documentation automatique
/world-create-location city valdorine
# World-Keeper génère "Portaven", vérifie qu'il n'existe pas, l'ajoute à geography.json
```

### Pour le Dungeon Master (via world-keeper)
```
DM: Les PJ veulent aller dans une ville valdine non encore nommée
DM: /world-create-location city valdorine
World-Keeper: ✓ Créé "Navelmaris" (Valdorine, city, Côte Occidentale)
DM: [Utilise ce nom dans narration]
```

---

## Cohérence Assurée

### Validation par world-keeper
1. **Unicité** : Vérifie que le nom n'existe pas déjà
2. **Style** : Respect du style de faction
3. **Géographie** : Cohérence avec région (port sur côte, forteresse en montagne)
4. **Documentation** : Ajout automatique dans `geography.json`

### Exemples de Cohérence
- ❌ "Port-de-Fer" pour Valdorine (style Karvath)
- ✅ "Port-de-Lune" pour Valdorine (style maritime)
- ❌ "Vallon-de-Prière" dans Karvath (style Lumenciel)
- ✅ "Hautegarde" dans Karvath (style militaire)

---

## Phases d'Implémentation

### Phase 1 : Données (1 tâche)
- [ ] Créer `data/location-names.json` avec vocabulaire complet

### Phase 2 : Package Go (2 tâches)
- [ ] Créer `internal/locations/locations.go`
- [ ] Créer `internal/locations/locations_test.go`

### Phase 3 : CLI (1 tâche)
- [ ] Créer `cmd/location-names/main.go`

### Phase 4 : Skill (1 tâche)
- [ ] Créer `.claude/skills/name-location-generator/skill.md`

### Phase 5 : Intégration (2 tâches)
- [ ] Mettre à jour `world-keeper.md` (commande `/world-create-location`)
- [ ] Mettre à jour `dungeon-master.md` (référence à la skill)

### Phase 6 : Makefile (1 tâche)
- [ ] Ajouter `sw-location-names` au Makefile

---

## Avantages de Cette Approche

✅ **Skill autonome** : Utilisable directement par le DM (improvisation rapide)
✅ **Intégration world-keeper** : Garantit cohérence et documentation
✅ **Styles distincts** : Chaque faction a son identité
✅ **Évolutif** : Facile d'ajouter nouveaux types ou factions
✅ **Cohérence** : world-keeper vérifie unicité et pertinence
✅ **Documentation** : Noms automatiquement ajoutés à `geography.json`

---

## Prochaines Étapes

1. **Validation du plan** : Accord sur l'architecture hybride ?
2. **Création des données** : `location-names.json` avec vocabulaire
3. **Implémentation** : Package Go + CLI
4. **Intégration** : world-keeper + dungeon-master
5. **Tests** : Vérifier cohérence et unicité
