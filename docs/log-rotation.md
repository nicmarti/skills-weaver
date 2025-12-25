# Log Rotation in sw-dm

## Problème Résolu

Les fichiers `sw-dm.log` pouvaient devenir très gros (plusieurs MB), dépassant les limites de tokens par fichier et rendant difficile la lecture et l'analyse.

## Solution : Log Par Session

Depuis la dernière mise à jour, `sw-dm` crée automatiquement un fichier de log séparé pour chaque session de jeu :

```
data/adventures/la-crypte-des-ombres/
├── sw-dm-session-1.log
├── sw-dm-session-2.log
├── sw-dm-session-3.log
├── sw-dm-session-4.log
├── sw-dm-session-5.log
└── sw-dm-session-6.log
```

## Comment Ça Marche

### 1. Détection Automatique de Session

Au démarrage de `sw-dm`, le système :
1. Lit `sessions.json` pour trouver le dernier ID de session
2. Calcule le numéro de la session en cours (dernier ID + 1)
3. Crée un fichier `sw-dm-session-N.log` pour cette session

**Exemple** :
```json
// sessions.json
{
  "sessions": [
    {"id": 1, "status": "completed"},
    {"id": 2, "status": "completed"},
    {"id": 3, "status": "completed"}
  ]
}
```
→ Prochain log : `sw-dm-session-4.log`

### 2. Archivage Automatique de l'Ancien sw-dm.log

Si un ancien fichier monolithique `sw-dm.log` existe ET fait plus de 1 MB, il est automatiquement archivé :

```
sw-dm.log → sw-dm.log.archived-20251225-193312
```

Ceci ne se produit qu'une fois lors de la première utilisation du nouveau système.

### 3. Fallback

Si le système ne peut pas déterminer le numéro de session (erreur de lecture de `sessions.json`), il crée un fichier avec timestamp :

```
sw-dm-20251225-193312.log
```

## Avantages

✅ **Fichiers plus petits** : Chaque session a son propre fichier de taille raisonnable
✅ **Organisation logique** : Aligné avec la structure des journaux (journal-session-N.json)
✅ **Meilleure performance** : Moins de tokens à charger par fichier
✅ **Analyse simplifiée** : Facile de retrouver les logs d'une session spécifique
✅ **Rétrocompatibilité** : L'ancien sw-dm.log est automatiquement archivé

## Script d'Extraction

Le script `extract-cli-commands.sh` a été mis à jour pour chercher dans tous les fichiers de log :

```bash
# Cherche dans sw-dm-session-*.log ET sw-dm.log
./scripts/extract-cli-commands.sh la-crypte-des-ombres

# Output:
# Adventure: la-crypte-des-ombres (sw-dm-session-1.log)
# ...
# Adventure: la-crypte-des-ombres (sw-dm-session-2.log)
# ...
```

## Format de Nommage

| Fichier | Utilisation |
|---------|-------------|
| `sw-dm-session-1.log` | Logs de la session 1 |
| `sw-dm-session-2.log` | Logs de la session 2 |
| `sw-dm-session-N.log` | Logs de la session N |
| `sw-dm.log.archived-YYYYMMDD-HHMMSS` | Ancien fichier monolithique archivé |
| `sw-dm-YYYYMMDD-HHMMSS.log` | Fallback si détection session échoue |

## Migration

### Pour les Aventures Existantes

**Rien à faire !** Le système gère automatiquement la migration :

1. Au prochain lancement de `sw-dm`, l'ancien `sw-dm.log` sera archivé (si > 1MB)
2. Un nouveau fichier `sw-dm-session-N.log` sera créé pour la session en cours
3. Les anciens logs restent accessibles dans le fichier archivé

### Pour Analyser les Anciens Logs

Les anciens logs archivés contiennent toujours les commandes CLI. Vous pouvez :

```bash
# Rechercher dans les fichiers archivés
grep "Equivalent CLI:" data/adventures/*/sw-dm.log.archived-*

# Extraire manuellement
cat data/adventures/la-crypte-des-ombres/sw-dm.log.archived-20251225-193312 | grep "Equivalent CLI:"
```

## Maintenance

### Nettoyer les Vieux Logs

Si vous souhaitez supprimer les anciens logs pour économiser de l'espace :

```bash
# Supprimer les logs archivés (plus vieux que l'ancien sw-dm.log)
rm data/adventures/*/sw-dm.log.archived-*

# Supprimer les logs de sessions spécifiques (attention : perte de données !)
rm data/adventures/la-crypte-des-ombres/sw-dm-session-1.log
rm data/adventures/la-crypte-des-ombres/sw-dm-session-2.log
```

⚠️ **Attention** : Une fois supprimés, ces logs ne peuvent pas être récupérés.

### Taille Typique par Session

Une session de 1-2 heures génère typiquement :
- **Sans log rotation** : Cumul dans un seul fichier (peut atteindre plusieurs MB)
- **Avec log rotation** : ~50-200 KB par session selon l'activité

## Implémentation Technique

### Fichiers Modifiés

1. **internal/agent/logger.go** :
   - `getCurrentSessionNumber()` : Détecte le numéro de session depuis sessions.json
   - `archiveOldLogIfNeeded()` : Archive sw-dm.log si > 1MB
   - `NewLogger()` : Crée le fichier de log approprié

2. **internal/agent/logger_test.go** :
   - Tests pour la détection de session
   - Tests pour l'archivage automatique

3. **scripts/extract-cli-commands.sh** :
   - Cherche dans `sw-dm-session-*.log` en plus de `sw-dm.log`
   - Affiche le nom du fichier de log traité

### Code de Détection de Session

```go
func getCurrentSessionNumber(adventurePath string) (int, error) {
    sessionsPath := filepath.Join(adventurePath, "sessions.json")

    // Si sessions.json n'existe pas, c'est la session 1
    if _, err := os.Stat(sessionsPath); os.IsNotExist(err) {
        return 1, nil
    }

    // Charge sessions.json et trouve le max ID
    // Retourne maxID + 1
}
```

## FAQ

**Q: Que se passe-t-il si je lance sw-dm plusieurs fois pendant la même session ?**
R: Le fichier `sw-dm-session-N.log` sera complété (mode append). Tous les lancements pendant la session écrivent dans le même fichier.

**Q: Comment savoir quelle session correspond à quel log ?**
R: Le numéro de session dans le nom du fichier correspond à l'ID dans `sessions.json`. Chaque log commence aussi par "Session N log started".

**Q: Puis-je désactiver la rotation ?**
R: Non, c'est activé par défaut pour éviter les fichiers trop gros. Mais vous pouvez toujours consulter les anciens logs archivés.

**Q: Les anciens outils continuent-ils de fonctionner ?**
R: Oui ! `extract-cli-commands.sh` cherche automatiquement dans tous les fichiers de log.

## Exemple Complet

```bash
# Session 1-5 déjà complétées
$ ls data/adventures/la-crypte-des-ombres/sw-dm*.log
sw-dm.log  # 3.2 MB (ancien fichier monolithique)

# Lancer sw-dm pour session 6
$ ./sw-dm
# Sélectionner "La Crypte des Ombres"
Archived old log to: data/adventures/la-crypte-des-ombres/sw-dm.log.archived-20251225-193312
Session 6 log started

# Après la session
$ ls data/adventures/la-crypte-des-ombres/sw-dm*.log
sw-dm.log.archived-20251225-193312  # 3.2 MB (archivé)
sw-dm-session-6.log                  # 150 KB (nouvelle session)

# Extraire les commandes
$ ./scripts/extract-cli-commands.sh la-crypte-des-ombres
=== CLI Commands Extractor ===

Adventure: la-crypte-des-ombres (sw-dm-session-6.log)

generate_map
  ./sw-map generate city "Cordova" --kingdom=valdorine

generate_npc
  ./sw-npc generate --race=human --gender=f --occupation=aubergiste --attitude=friendly

...
```
