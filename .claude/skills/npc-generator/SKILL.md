---
name: npc-generator
description: Génère des PNJ complets pour BFRPG avec apparence, personnalité, motivations et secrets. Utilise le générateur de noms. Parfait pour peupler le monde de jeu avec des personnages mémorables.
allowed-tools: Bash
---

# NPC Generator - Générateur de PNJ pour BFRPG

Skill pour générer des personnages non-joueurs complets avec description physique, personnalité, motivations et secrets.

## Utilisation Rapide

```bash
# Compiler si nécessaire
go build -o sw-npc ./cmd/npc

# Générer un PNJ complet
./sw-npc generate

# Générer plusieurs PNJ rapidement
./sw-npc quick --count=5
```

## Commandes Disponibles

### Génération Complète

```bash
./sw-npc generate [options]

# Options:
#   --race=<race>          Race (human, dwarf, elf, halfling)
#   --gender=<m|f>         Sexe
#   --occupation=<type>    Type d'occupation
#   --attitude=<type>      Attitude envers les PJ
#   --format=<md|json|short>  Format de sortie
```

### Génération Rapide

```bash
./sw-npc quick [options]

# Mêmes options + --count=N pour plusieurs PNJ
```

## Types d'Occupation

| Type | Description | Exemples |
|------|-------------|----------|
| `commoner` | Gens du peuple | fermier, boulanger, serveur, mendiant |
| `skilled` | Artisans qualifiés | marchand, apothicaire, musicien, scribe |
| `authority` | Figures d'autorité | garde, sergent, noble, magistrat |
| `underworld` | Monde criminel | voleur, espion, contrebandier, assassin |
| `religious` | Religieux | prêtre, moine, pèlerin, inquisiteur |
| `adventurer` | Aventuriers | chasseur de primes, explorateur, mercenaire |

## Attitudes

| Attitude | Description |
|----------|-------------|
| `positive` | Amical, serviable, accueillant |
| `neutral` | Professionnel, indifférent, prudent |
| `negative` | Méfiant, hostile, moqueur |

## Exemples

### PNJ Complet

```bash
./sw-npc generate --race=dwarf --gender=m --occupation=skilled
```

Résultat:
```markdown
## Thorin Ironfoot

**Nain Homme** - forgeron

### Apparence
Petit trapu, de stature trapu. Cheveux bruns tressés, yeux noisette...

### Personnalité
- **Trait principal** : travailleur
- **Trait secondaire** : traditionnel
- **Qualité** : loyal envers ses amis
- **Défaut** : est têtu

### Comportement
- **Voix** : grave et profonde, parle lentement
- **Tic** : se gratte la barbe en réfléchissant
- **Attitude** : professionnel et distant

### Secrets (MJ seulement)
- **Objectif** : amasser une fortune
- **Peur** : l'échec
- **Secret** : a des dettes importantes
```

### Liste Rapide de PNJ

```bash
./sw-npc quick --occupation=commoner --count=5
```

Résultat:
```
Aldric Ironhand - humain homme, fermier (calme, indifférent)
Rose Greenhill - halfelin femme, serveur (jovial, curieux mais réservé)
Legolas Moonwhisper - elfe homme, berger (distant, poli mais pressé)
...
```

### Export JSON

```bash
./sw-npc generate --format=json
```

## Intégration avec Adventure Manager

Pour logger les rencontres de PNJ :

```bash
# Générer un PNJ
./sw-npc generate --occupation=authority --attitude=positive

# Logger dans l'aventure
./sw-adventure log "Mon Aventure" npc "Rencontre avec le capitaine Aldric"
```

## Structure des Données

Le générateur utilise deux fichiers de données :

- `data/names.json` - Dictionnaires de noms par race
- `data/npc-traits.json` - Traits d'apparence, personnalité, motivations

### Traits Générés

**Apparence** :
- Corpulence, taille
- Couleur et style de cheveux
- Couleur des yeux, teint de peau
- Trait facial distinctif
- Signe particulier

**Personnalité** :
- Trait principal (amical, distant, courageux...)
- Trait secondaire (superstitieux, romantique...)
- Qualité principale
- Défaut principal

**Comportement** :
- Ton de voix
- Manière de parler
- Tic ou habitude

**Motivations (pour le MJ)** :
- Objectif de vie
- Peur principale
- Secret caché

## Conseils d'Utilisation

### Pour un PNJ récurrent
```bash
./sw-npc generate --format=md
```
Sauvegardez la description complète pour référence future.

### Pour une foule de figurants
```bash
./sw-npc quick --count=10 --occupation=commoner
```
Descriptions courtes pour des PNJ de passage.

### Pour un antagoniste
```bash
./sw-npc generate --attitude=negative --occupation=underworld
```
Un PNJ avec des motivations hostiles.

### Pour un allié potentiel
```bash
./sw-npc generate --attitude=positive --occupation=adventurer
```
Un PNJ qui pourrait aider le groupe.

## Races et Ajustements

Le générateur ajuste automatiquement l'apparence selon la race :

| Race | Ajustements |
|------|-------------|
| Nain | Petit, trapu/musclé/robuste |
| Elfe | Grand, mince/svelte/élancé |
| Halfelin | Très petit |
| Humain | Variable |