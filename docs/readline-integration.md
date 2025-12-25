# Intégration readline dans sw-dm

## Problème Résolu

**Avant** : Utilisation de `bufio.Scanner` qui ne gère pas les séquences d'échappement ANSI
- ❌ Touches fléchées affichent `^[[D`, `^[[C`, etc.
- ❌ Pas d'édition de ligne (Home, End, Delete)
- ❌ Pas d'historique des commandes
- ❌ Expérience utilisateur frustrante

**Après** : Utilisation de `github.com/chzyer/readline`
- ✅ Édition de ligne complète (←, →, Home, End, Ctrl+A, Ctrl+E)
- ✅ Historique des commandes (↑, ↓)
- ✅ Gestion propre de Ctrl+C et Ctrl+D
- ✅ Suppression de caractères (Backspace, Delete)
- ✅ Historique persistant entre sessions
- ✅ Expérience utilisateur professionnelle

## Fonctionnalités

### Édition de Ligne

| Touche | Action |
|--------|--------|
| ← → | Déplacement caractère par caractère |
| Ctrl+A / Home | Début de ligne |
| Ctrl+E / End | Fin de ligne |
| Ctrl+W | Supprimer mot précédent |
| Backspace | Supprimer caractère précédent |
| Delete | Supprimer caractère suivant |

### Historique

| Touche | Action |
|--------|--------|
| ↑ | Commande précédente |
| ↓ | Commande suivante |
| Ctrl+R | Recherche dans l'historique |

### Contrôle

| Touche | Action |
|--------|--------|
| Ctrl+C | Interruption (avec ligne vide = quitter) |
| Ctrl+D | EOF (quitter proprement) |
| Ctrl+L | Effacer l'écran |

## Implémentation

### Configuration readline

```go
rl, err := readline.NewEx(&readline.Config{
    Prompt:          ui.PromptStyle.Render("> "),
    HistoryFile:     "/tmp/sw-dm-history.txt",
    InterruptPrompt: "^C",
    EOFPrompt:       "exit",
})
if err != nil {
    fmt.Fprintf(os.Stderr, "Error initializing readline: %v\n", err)
    os.Exit(1)
}
defer rl.Close()
```

### Boucle REPL

```go
for {
    line, err := rl.Readline()
    if err == readline.ErrInterrupt {
        // Ctrl+C pressed
        if len(line) == 0 {
            fmt.Println("Au revoir !")
            break
        }
        continue
    } else if err == io.EOF {
        // Ctrl+D pressed
        fmt.Println("Au revoir !")
        break
    }

    input := strings.TrimSpace(line)
    if input == "" {
        fmt.Println("Message vide détecté")
        continue
    }

    // Process input...
}
```

## Fichier d'Historique

**Localisation** : `/tmp/sw-dm-history.txt`

**Format** : Texte brut, une commande par ligne

**Comportement** :
- ✅ Historique persistant entre sessions
- ✅ Automatiquement trié (plus récents en bas)
- ✅ Pas de doublons consécutifs
- ✅ Limite de taille (500 entrées par défaut)

**Note** : Pour effacer l'historique : `rm /tmp/sw-dm-history.txt`

## Tests

### Tests de Validation d'Input (8 cas)
- Empty string, spaces, tabs, whitespace mixte
- Input valide avec/sans whitespace
- Commandes exit/quit

### Tests d'Historique (4 cas)
- Filtrage des entrées vides
- Trim du whitespace
- Préservation de l'ordre

### Tests de Commandes Exit (6 cas)
- Normalisation case-insensitive
- Variantes : exit, quit, EXIT, QUIT, Exit, Quit

### Tests de Sanitization (5 cas)
- Input normal, avec espaces, avec tabs
- Espaces uniquement
- Espaces multiples internes (préservés)

### Tests de Caractères de Contrôle (4 cas)
- Séquences ANSI (arrows, home, end)
- Vérification qu'elles n'apparaissent pas dans l'input final

**Total** : 27 tests ✅ Tous passent

## Migration depuis bufio.Scanner

### Changements de Code

**Avant** :
```go
scanner := bufio.NewScanner(os.Stdin)
for {
    if !scanner.Scan() {
        break
    }
    input := strings.TrimSpace(scanner.Text())
    // ...
}
```

**Après** :
```go
rl, err := readline.NewEx(&readline.Config{
    Prompt: "> ",
    HistoryFile: "/tmp/sw-dm-history.txt",
})
if err != nil {
    return err
}
defer rl.Close()

for {
    line, err := rl.Readline()
    if err == readline.ErrInterrupt || err == io.EOF {
        break
    }
    input := strings.TrimSpace(line)
    // ...
}
```

### Avantages

1. **Expérience Utilisateur** : Édition de ligne professionnelle
2. **Productivité** : Historique des commandes
3. **Robustesse** : Gestion propre des signaux (Ctrl+C, Ctrl+D)
4. **Cross-platform** : Fonctionne sur Linux, macOS, Windows
5. **Zéro dépendance système** : Pure Go

## Dépendance

**Package** : `github.com/chzyer/readline v1.5.1`

**Installation** :
```bash
go get github.com/chzyer/readline
```

**Licence** : MIT

**Stars GitHub** : ~5000+

**Maintenance** : Active (dernière release 2023)

## Références

- [readline GitHub](https://github.com/chzyer/readline)
- [Documentation readline](https://pkg.go.dev/github.com/chzyer/readline)
- [GNU readline (inspiration)](https://tiswww.case.edu/php/chet/readline/rltop.html)
