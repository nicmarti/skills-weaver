# Package UI - Terminal Styling avec Lipgloss

Ce package fournit des utilitaires de style pour le terminal en utilisant [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss).

## Fonctionnalités

### 1. Styles prédéfinis (styles.go)

Styles adaptatifs qui fonctionnent sur fonds clairs et foncés :

```go
// Couleurs adaptatives
Purple, Gold, Gray, Green, Red, Text, Cyan

// Styles disponibles
LogoStyle           // Logo ASCII art
TitleStyle          // Titres principaux avec bordure double
SubtitleStyle       // Sous-titres en italique
PromptStyle         // Prompt utilisateur (vert, gras)
DMStyle             // Texte narratif du Dungeon Master
ToolStyle           // Messages d'outils (violet)
ErrorStyle          // Messages d'erreur (rouge, gras)
InfoBoxStyle        // Boîtes d'information avec bordure arrondie
AdventureTitleStyle // Nom d'aventure (or, gras)
MenuItemStyle       // Items de menu
MenuSelectedStyle   // Item sélectionné (vert, gras)
```

### 2. Rendu Markdown (markdown.go & markdown_v2.go)

Parse et applique des styles inline sur du texte avec notation markdown.

#### Version 2 (markdown_v2.go) - Recommandée ✅

Syntaxe markdown standard avec support de l'imbrication :

- `**texte**` → **gras**
- `*texte*` → *italique*
- `*texte avec **gras** imbriqué*` → Combinaison

**Utilisation** :

```go
import "dungeons/internal/ui"

// Rendu avec style personnalisé
text := "Vous voyez **un dragon** qui s'approche!"
styled := ui.RenderMarkdownV2(text, ui.DMStyle)

// Raccourci pour texte du DM
styled := ui.RenderDMTextV2("Le **dragon rouge** attaque!")
```

**Exemples** :

```go
// Gras simple
ui.RenderDMTextV2("Vous voyez **un dragon rouge** qui s'approche!")

// Italique
ui.RenderDMTextV2("Le gardien murmure *« Vous ne passerez pas... »*")

// Imbrication (gras dans italique)
ui.RenderDMTextV2("*Vous entrez dans la **nuit noire**.*")

// Dialogue avec personnage
ui.RenderDMTextV2("**Gareth** *(regardant Elara)* : — *\"On doit partir.\"*")
```

#### Version 1 (markdown.go) - Legacy

Syntaxe non-standard (conservée pour compatibilité) :

- `*texte*` → **gras** (différent de markdown standard!)
- `_texte_` → *italique*

```go
// V1 - Ancienne syntaxe
ui.RenderDMText("Vous voyez *un dragon* qui s'approche!")
```

**Note** : La V2 est maintenant utilisée par défaut dans sw-dm. Utilisez la V1 uniquement si vous avez du contenu existant avec l'ancienne syntaxe.

### 3. Fonctions d'affichage (screen.go)

```go
// Efface l'écran
ui.ClearScreen()

// Affiche la bannière avec le nom du modèle
ui.ShowBanner("Claude Haiku 4.5")

// Affiche les infos d'aventure dans une boîte stylée
ui.ShowAdventureInfo(
    "La Crypte des Ombres",  // nom
    "Entrée de la crypte",   // localisation
    1250,                    // or
    "Aldric, Lyra, Thorin",  // groupe
    "Le groupe entre...",    // dernière action
)

// Affiche un séparateur visuel
ui.ShowSeparator()
```

## Architecture du parser Markdown

Le parser traite le texte en deux passes :

1. **Passe 1 : Bold** - Détecte `*texte*` et applique `Bold(true)`
2. **Passe 2 : Italic** - Détecte `_texte_` et applique `Italic(true)`

Chaque pattern est traité avec une regex et les styles sont appliqués segment par segment.

### Limitations actuelles

- Pas de support pour les patterns imbriqués (`*gras _et italique_*`)
- Pas de support pour l'échappement (`\*` pour un astérisque littéral)
- Pas de support pour d'autres éléments markdown (liens, listes, etc.)

## Améliorations futures possibles

### 1. Support complet Markdown avec Glamour

Pour un rendu markdown complet, utilisez [charmbracelet/glamour](https://github.com/charmbracelet/glamour) :

```go
import "github.com/charmbracelet/glamour"

func RenderFullMarkdown(text string) (string, error) {
    r, _ := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
        glamour.WithWordWrap(80),
    )
    return r.Render(text)
}
```

**Avantages** :
- Support complet de Markdown (liens, listes, tableaux, code blocks)
- Thèmes personnalisables
- Rendu de fichiers markdown complets

**Inconvénients** :
- Plus lourd (dépendance supplémentaire)
- Moins de contrôle fin sur les styles individuels

### 2. Parser markdown amélioré

Extensions possibles pour `markdown.go` :

```go
// Patterns supplémentaires
**texte**     → gras (alternative)
__texte__     → italique (alternative)
~~texte~~     → barré
`texte`       → code inline
\*            → échappement
[texte](url)  → liens (affichage uniquement)

// Patterns imbriqués
*gras _et italique_*  → combinaison de styles
```

### 3. Streaming markdown

Pour le streaming du DM (texte qui arrive par chunks), gérer les patterns incomplets :

```go
type MarkdownStreamer struct {
    buffer string
    style  lipgloss.Style
}

func (ms *MarkdownStreamer) Write(chunk string) {
    // Accumule les chunks
    // Parse et rend seulement les patterns complets
    // Garde les patterns incomplets en buffer
}
```

### 4. Styles contextuels additionnels

```go
// Dialogues de PNJ
NPCDialogStyle = lipgloss.NewStyle().
    Foreground(Cyan).
    Italic(true)

// Descriptions de lieux
LocationStyle = lipgloss.NewStyle().
    Foreground(Purple).
    Bold(true)

// Actions de combat
CombatStyle = lipgloss.NewStyle().
    Foreground(Red).
    Bold(true)
```

Utilisation avec prefixes :

```go
"[PNJ] _Bienvenue, voyageur!_"  → style PNJ + italique
"[Lieu] *La crypte sombre*"     → style lieu + gras
"[Combat] *Attaque!*"           → style combat + gras
```

## Tests

```bash
# Lancer les tests du package UI
go test ./internal/ui/... -v

# Démo interactive
go run examples/markdown_demo.go
```

## Références

- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Glamour Documentation](https://github.com/charmbracelet/glamour)
- [Bubble Tea Framework](https://github.com/charmbracelet/bubbletea)
- [Charm CLI Tools](https://github.com/charmbracelet)
