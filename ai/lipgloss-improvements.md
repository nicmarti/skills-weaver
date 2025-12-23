# Améliorations Lipgloss pour sw-dm

## Résumé

Ajout d'un système de rendu markdown simple pour styliser le texte du Dungeon Master dans sw-dm.

## Changements effectués

### 1. Nouveau fichier : `internal/ui/markdown.go`

Parser markdown simple avec support pour :
- `*texte*` → **gras**
- `_texte_` → *italique*

**Fonctions publiques** :
```go
RenderMarkdown(text string, baseStyle lipgloss.Style) string
RenderDMText(text string) string  // Raccourci avec DMStyle
```

**Architecture** :
- Utilise des regex pour détecter les patterns
- Applique les styles lipgloss segment par segment
- Traite bold et italic en deux passes séparées

### 2. Tests : `internal/ui/markdown_test.go`

Tests unitaires couvrant :
- Gras simple et multiple
- Italique
- Texte mixte (gras + italique)
- Cas limites (texte vide, sans markdown)

✅ Tous les tests passent

### 3. Modification : `cmd/dm/main.go`

Ajout du rendu markdown dans `OnTextChunk()` :
```go
func (to *TerminalOutput) OnTextChunk(text string) {
    rendered := ui.RenderDMText(text)
    fmt.Print(rendered)
}
```

### 4. Documentation : `internal/ui/README.md`

Documentation complète du package UI avec :
- Guide d'utilisation des styles
- Exemples de markdown
- Architecture du parser
- Limitations actuelles
- Propositions d'améliorations futures

### 5. Démo : `examples/markdown_demo.go`

Programme de démonstration montrant les capacités du parser.

## Utilisation

### Pour les MJ (Dungeon Masters)

Dans vos narratives, utilisez simplement :

```
Vous entrez dans *la crypte des ombres*. L'air est _glacial_.
Devant vous se trouve *un autel ancien*.
```

Le texte entre `*` apparaîtra en gras, le texte entre `_` en italique.

### Pour les développeurs

```go
import "dungeons/internal/ui"

// Style personnalisé
text := "Attention : *danger imminent*!"
styled := ui.RenderMarkdown(text, ui.ErrorStyle)

// Raccourci pour DM
narrative := "Le *dragon* vous observe _attentivement_"
fmt.Print(ui.RenderDMText(narrative))
```

## Limitations actuelles

1. **Pas d'imbrication** : `*gras _et italique_*` ne fonctionne pas correctement
2. **Pas d'échappement** : Impossible d'afficher un `*` littéral
3. **Syntaxe limitée** : Seulement bold et italic
4. **Problèmes de spacing** : Quelques artefacts d'espacement sur texte multilignes complexes

## Améliorations futures recommandées

### Court terme (rapide à implémenter)

1. **Support de `**bold**`** : Alternative aux `*bold*`
   - Regex : `\*\*([^*]+)\*\*`
   - Priorité haute pour compatibilité markdown standard

2. **Échappement** : Support de `\*` pour astérisques littéraux
   - Pre-processing avant regex
   - Restauration après rendu

3. **Code inline** : Support de `` `code` ``
   - Style mono-space avec fond gris
   - Utile pour commandes et valeurs numériques

### Moyen terme (effort modéré)

4. **Parser unifié** : Traiter tous les patterns en une seule passe
   - Évite les problèmes de spacing
   - Permet l'imbrication correcte
   - Architecture plus robuste

5. **Styles contextuels** : Préfixes pour contexte
   ```
   [PNJ] _Bienvenue, voyageur!_
   [Lieu] *La crypte sombre*
   [Combat] *Attaque critique!*
   ```

6. **Streaming markdown** : Buffer pour patterns incomplets
   - Important pour l'affichage streaming du DM
   - Évite les coupures en milieu de pattern

### Long terme (effort important)

7. **Intégration Glamour** : Markdown complet
   - Support de toute la spec markdown
   - Rendu de documents complets
   - Thèmes personnalisables
   - Nécessite dépendance supplémentaire

8. **Configuration utilisateur** : Fichier de config
   ```toml
   [ui.styles]
   dm_text_color = "255"
   enable_markdown = true
   markdown_bold = "*"
   markdown_italic = "_"
   ```

## Comparaison avec Glamour

| Feature | markdown.go (actuel) | Glamour |
|---------|---------------------|---------|
| Dépendances | Lipgloss uniquement | Lipgloss + Glamour + deps |
| Poids | ~200 LOC | ~10K LOC |
| Patterns supportés | 2 (bold, italic) | 30+ (markdown complet) |
| Performance | Très rapide | Rapide |
| Contrôle fine | Total | Limité aux thèmes |
| Streaming | Facile | Difficile |
| Apprentissage | Immédiat | Courbe d'apprentissage |

**Recommandation** : Garder markdown.go pour sw-dm (léger, rapide, suffisant).
Glamour serait utile si on veut rendre des docs markdown complets (ex: afficher un fichier README.md stylé).

## Capacités de Lipgloss (non utilisées)

Lipgloss offre bien plus que bold/italic :

### Styles inline
```go
.Underline(true)      // Souligné
.Strikethrough(true)  // Barré
.Faint(true)          // Atténué
.Blink(true)          // Clignotant (éviter!)
.Reverse(true)        // Inverse fg/bg
```

### Layout
```go
.Width(40)            // Largeur fixe
.Height(10)           // Hauteur fixe
.Align(lipgloss.Center) // Alignement
.Padding(1, 2)        // Espacement interne
.Margin(2, 4)         // Espacement externe
```

### Bordures
```go
.Border(lipgloss.RoundedBorder())
.BorderForeground(lipgloss.Color("63"))
.BorderTop(false)
```

### Transformations
```go
.Transform(func(s string) string {
    return strings.ToUpper(s)
})
```

### Couleurs
```go
// Couleurs adaptatives (light/dark)
lipgloss.AdaptiveColor{Light: "16", Dark: "255"}

// Couleurs fixes
lipgloss.Color("#FF5733")  // Hex
lipgloss.Color("63")       // ANSI 256
```

## Tests et validation

```bash
# Tests unitaires
make test

# Tests UI spécifiques
go test ./internal/ui/... -v

# Démo interactive
go run examples/markdown_demo.go

# Compiler sw-dm
make sw-dm
```

## Prochaines étapes suggérées

1. ✅ Tester avec une vraie session sw-dm
2. ⬜ Ajouter support de `**bold**` et `__italic__`
3. ⬜ Implémenter l'échappement `\*`
4. ⬜ Ajouter support de `` `code` ``
5. ⬜ Créer styles contextuels avec préfixes
6. ⬜ Documenter dans CLAUDE.md
7. ⬜ Ajouter exemples dans skills/dungeon-master

## Références

- [Lipgloss Repo](https://github.com/charmbracelet/lipgloss)
- [Lipgloss Tutorial](https://github.com/charmbracelet/lipgloss/blob/master/examples)
- [Glamour (alternative)](https://github.com/charmbracelet/glamour)
- [ANSI Escape Codes](https://en.wikipedia.org/wiki/ANSI_escape_code)
