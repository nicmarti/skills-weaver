package charactersheet

import (
	"bytes"
	"fmt"
	"html/template"
)

// TemplateManager handles HTML templates
type TemplateManager struct {
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{}
}

// TemplateData holds all data for template rendering
type TemplateData struct {
	Character     *CharacterDisplay
	Biography     *Biography
	Equipment     *EquipmentSummary
	RaceName      string
	ClassName     string
	Gold          int
	ClassBanner   string
	GeneratedAt   string
	AbilityScores []AbilityScore
}

// CharacterDisplay wraps character for template display
type CharacterDisplay struct {
	Name          string
	Level         int
	XP            int
	HitPoints     int
	MaxHitPoints  int
	ArmorClass    int
	Appearance    *AppearanceDisplay
}

// AppearanceDisplay wraps appearance for template
type AppearanceDisplay struct {
	ReferenceImage string
}

// AbilityScore for template rendering
type AbilityScore struct {
	Name           string
	Score          int
	ModifierString string
}

// RenderTemplate renders the Dark Fantasy template
func (m *TemplateManager) RenderTemplate(sheet *Sheet) (string, error) {
	// Prepare template data
	data := m.prepareTemplateData(sheet)

	// Parse and execute template
	tmpl, err := template.New("sheet").Parse(darkFantasyTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// prepareTemplateData prepares data for template rendering
func (m *TemplateManager) prepareTemplateData(sheet *Sheet) *TemplateData {
	// Convert character to display format
	charDisplay := &CharacterDisplay{
		Name:         sheet.Character.Name,
		Level:        sheet.Character.Level,
		XP:           sheet.Character.XP,
		HitPoints:    sheet.Character.HitPoints,
		MaxHitPoints: sheet.Character.MaxHitPoints,
		ArmorClass:   sheet.Character.ArmorClass,
	}

	// Add appearance if available
	if sheet.Character.Appearance != nil {
		charDisplay.Appearance = &AppearanceDisplay{
			ReferenceImage: sheet.Character.Appearance.ReferenceImage,
		}
	}

	// Prepare ability scores
	abilities := []AbilityScore{
		{Name: "FOR", Score: sheet.Character.Abilities.Strength, ModifierString: formatModifier(sheet.Character.Modifiers.Strength)},
		{Name: "DEX", Score: sheet.Character.Abilities.Dexterity, ModifierString: formatModifier(sheet.Character.Modifiers.Dexterity)},
		{Name: "CON", Score: sheet.Character.Abilities.Constitution, ModifierString: formatModifier(sheet.Character.Modifiers.Constitution)},
		{Name: "INT", Score: sheet.Character.Abilities.Intelligence, ModifierString: formatModifier(sheet.Character.Modifiers.Intelligence)},
		{Name: "SAG", Score: sheet.Character.Abilities.Wisdom, ModifierString: formatModifier(sheet.Character.Modifiers.Wisdom)},
		{Name: "CHA", Score: sheet.Character.Abilities.Charisma, ModifierString: formatModifier(sheet.Character.Modifiers.Charisma)},
	}

	return &TemplateData{
		Character:     charDisplay,
		Biography:     sheet.Biography,
		Equipment:     sheet.Equipment,
		RaceName:      sheet.RaceName,
		ClassName:     sheet.ClassName,
		Gold:          sheet.Gold,
		ClassBanner:   sheet.ClassBanner,
		GeneratedAt:   sheet.GeneratedAt,
		AbilityScores: abilities,
	}
}

// formatModifier formats ability modifier with sign
func formatModifier(mod int) string {
	if mod >= 0 {
		return fmt.Sprintf("+%d", mod)
	}
	return fmt.Sprintf("%d", mod)
}

// darkFantasyTemplate is the Dark Fantasy (Baldur's Gate style) HTML template
const darkFantasyTemplate = `<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Character.Name}} - Fiche de Personnage</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600;700&display=swap" rel="stylesheet">
  <style>
    body {
      background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
      font-family: 'Inter', sans-serif;
      min-height: 100vh;
    }

    .card-gradient {
      background: linear-gradient(to bottom, rgba(42,42,42,0.95) 0%, rgba(31,31,31,0.95) 100%);
      box-shadow: 0 8px 24px rgba(0,0,0,0.6);
      border: 2px solid #555555;
    }

    .gold-accent {
      color: #d4af37;
    }

    .stat-glow {
      text-shadow: 0 0 10px rgba(212, 175, 55, 0.5);
    }

    .hp-red {
      color: #ff4444;
      text-shadow: 0 0 8px rgba(255, 68, 68, 0.6);
    }

    .ac-blue {
      color: #4a90e2;
      text-shadow: 0 0 8px rgba(74, 144, 226, 0.6);
    }

    @media print {
      body { background: white; }
      .card-gradient { background: white; border: 1px solid #333; }
    }
  </style>
</head>
<body class="p-8">

  {{if .ClassBanner}}
  <div class="mb-4">
    <img src="{{.ClassBanner}}" alt="Class Banner" class="w-full h-32 object-cover rounded-lg opacity-60" />
  </div>
  {{end}}

  <!-- Header Card -->
  <div class="card-gradient rounded-xl p-8 mb-6">
    <div class="flex items-start justify-between">
      <div class="flex items-center space-x-6">
        {{if .Character.Appearance}}
        <img src="{{.Character.Appearance.ReferenceImage}}"
             alt="{{.Character.Name}}"
             class="w-32 h-32 rounded-full border-4 border-yellow-600 shadow-2xl" />
        {{else}}
        <div class="w-32 h-32 rounded-full border-4 border-yellow-600 bg-gray-800 flex items-center justify-center">
          <span class="text-gray-500">Portrait</span>
        </div>
        {{end}}

        <div>
          <h1 class="text-5xl font-bold text-white stat-glow mb-2">{{.Character.Name}}</h1>
          <p class="text-xl gold-accent tracking-wide uppercase">{{.RaceName}} {{.ClassName}}</p>
          <p class="text-gray-400 mt-1">Niveau {{.Character.Level}}</p>
          <div class="mt-3 inline-block bg-yellow-600 px-4 py-1 rounded-md">
            <span class="text-sm font-bold text-black">XP: {{.Character.XP}}</span>
          </div>
        </div>
      </div>

      <div class="text-right">
        <div class="bg-gray-800 px-4 py-2 rounded-lg mb-2">
          <span class="text-gray-400 text-sm">OR</span>
          <span class="text-yellow-500 font-bold text-lg ml-2">{{.Gold}} po</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Ability Scores Grid -->
  <div class="grid grid-cols-6 gap-4 mb-6">
    {{range .AbilityScores}}
    <div class="card-gradient rounded-lg p-4 text-center">
      <h3 class="text-xs text-gray-400 uppercase tracking-wider mb-2">{{.Name}}</h3>
      <p class="text-4xl font-bold text-white mb-1">{{.Score}}</p>
      <div class="inline-block bg-yellow-600 px-3 py-1 rounded">
        <span class="text-sm font-bold text-black">{{.ModifierString}}</span>
      </div>
    </div>
    {{end}}
  </div>

  <!-- Combat Stats -->
  <div class="grid grid-cols-2 gap-6 mb-6">
    <div class="card-gradient rounded-lg p-6 border-red-900">
      <h3 class="text-sm text-gray-400 uppercase tracking-wider mb-3">Points de Vie</h3>
      <p class="text-5xl font-bold hp-red">{{.Character.HitPoints}} / {{.Character.MaxHitPoints}}</p>
    </div>

    <div class="card-gradient rounded-lg p-6 border-blue-900">
      <h3 class="text-sm text-gray-400 uppercase tracking-wider mb-3">Classe d'Armure</h3>
      <p class="text-5xl font-bold ac-blue">{{.Character.ArmorClass}}</p>
    </div>
  </div>

  <!-- Equipment Section -->
  <div class="card-gradient rounded-xl p-8 mb-6">
    <h2 class="text-2xl font-bold gold-accent mb-4 uppercase tracking-wide">Équipement</h2>
    <hr class="border-gray-700 mb-6" />

    {{if .Equipment.Empty}}
      <p class="text-gray-500 italic">Aucun équipement personnel défini.</p>
      <p class="text-gray-600 text-sm mt-2">Utilisez --adventure pour voir l'inventaire partagé.</p>
    {{else}}
      {{if .Equipment.Items}}
      <div class="mb-6">
        <h3 class="text-lg text-gray-300 mb-3">Personnel</h3>
        {{range .Equipment.Items}}
        <div class="flex items-center justify-between py-2 border-b border-gray-800">
          <div class="flex items-center space-x-3">
            {{if .Icon}}
            <img src="{{.Icon}}" alt="icon" class="w-8 h-8" />
            {{end}}
            <span class="text-gray-200">{{.Name}}</span>
          </div>
          <span class="text-gray-500 text-sm">{{.Details}}</span>
        </div>
        {{end}}
      </div>
      {{end}}

      {{if .Equipment.SharedItems}}
      <div>
        <h3 class="text-lg text-gray-300 mb-3">Inventaire Partagé</h3>
        <p class="text-yellow-500 mb-3">Or du groupe: {{.Equipment.SharedGold}} po</p>
        {{range .Equipment.SharedItems}}
        <div class="flex items-center justify-between py-2 border-b border-gray-800">
          <span class="text-gray-200">{{.Name}}</span>
          <span class="text-gray-500 text-sm">×{{.Quantity}}</span>
        </div>
        {{end}}
      </div>
      {{end}}
    {{end}}
  </div>

  {{if .Biography}}
  <div class="card-gradient rounded-xl p-8 mb-6">
    <h2 class="text-2xl font-bold gold-accent mb-4 uppercase tracking-wide">Biographie</h2>
    <hr class="border-gray-700 mb-6" />

    <div class="space-y-4 text-gray-300">
      <div>
        <h3 class="text-sm gold-accent uppercase mb-1">Origine</h3>
        <p>{{.Biography.Origin}}</p>
      </div>

      <div>
        <h3 class="text-sm gold-accent uppercase mb-1">Passé</h3>
        <p>{{.Biography.Background}}</p>
      </div>

      <div>
        <h3 class="text-sm gold-accent uppercase mb-1">Motivation</h3>
        <p>{{.Biography.Motivation}}</p>
      </div>

      {{if .Biography.Personality}}
      <div>
        <h3 class="text-sm gold-accent uppercase mb-1">Personnalité</h3>
        <p>{{.Biography.Personality}}</p>
      </div>
      {{end}}

      {{if .Biography.Bonds}}
      <div>
        <h3 class="text-sm gold-accent uppercase mb-2">Relations</h3>
        {{range .Biography.Bonds}}
        <div class="pl-4 border-l-2 border-gray-700 mb-2">
          <p class="font-semibold">{{.Name}} <span class="text-xs text-gray-500">({{.Type}})</span></p>
          <p class="text-sm text-gray-400">{{.Description}}</p>
        </div>
        {{end}}
      </div>
      {{end}}

      {{if .Biography.Secrets}}
      <div>
        <h3 class="text-sm gold-accent uppercase mb-2">Secrets</h3>
        {{range .Biography.Secrets}}
        <p class="text-sm text-gray-400 italic">• {{.}}</p>
        {{end}}
      </div>
      {{end}}
    </div>
  </div>
  {{end}}

  <!-- Footer -->
  <div class="text-center text-gray-600 text-sm mt-8 pb-4">
    <p>Fiche générée le {{.GeneratedAt}} • SkillsWeaver v1.0</p>
  </div>

</body>
</html>`
