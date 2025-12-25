// Package world provides programmatic access to world-keeper data files.
package world

// Geography contains the complete geographic structure of the world.
type Geography struct {
	Continents  []Continent  `json:"continents"`
	TradeRoutes []TradeRoute `json:"trade_routes"`
	Frontiers   []Frontier   `json:"frontiers"`
	Notes       []string     `json:"notes"`
}

// Continent represents a continent with its regions.
type Continent struct {
	Name    string   `json:"name"`
	Regions []Region `json:"regions"`
}

// Region represents a geographic region with cities.
type Region struct {
	Name        string     `json:"name"`
	Kingdom     string     `json:"kingdom"`
	Description string     `json:"description"`
	Cities      []Location `json:"cities"`
}

// Location represents a city, town, village, or settlement.
type Location struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"` // "port majeur", "village", "forteresse capitale", etc.
	Kingdom      string            `json:"kingdom"`
	Population   string            `json:"population"`
	Description  string            `json:"description"`
	KeyLocations []string          `json:"key_locations"` // POIs within this location
	Distances    map[string]string `json:"distances"`     // Distances to other locations
	Notes        string            `json:"notes,omitempty"`
}

// TradeRoute represents a trade route between locations.
type TradeRoute struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"` // "maritime" or "terrestrial"
	Locations []string `json:"locations"`
	Goods     []string `json:"goods"`
	Notes     string   `json:"notes,omitempty"`
}

// Frontier represents a border between kingdoms.
type Frontier struct {
	Kingdoms []string `json:"kingdoms"`
	Status   string   `json:"status"` // "neutral", "contested", "hostile", etc.
	Notes    string   `json:"notes,omitempty"`
}

// Factions contains all kingdom/faction data.
type Factions struct {
	Kingdoms     []Kingdom     `json:"kingdoms"`
	Organizations []Organization `json:"organizations,omitempty"`
}

// Kingdom represents a political kingdom/faction.
type Kingdom struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Motto           string            `json:"motto"`
	Capital         string            `json:"capital"`
	Government      string            `json:"government"`
	Population      int               `json:"population"`
	Territory       string            `json:"territory"`
	Symbol          string            `json:"symbol"`
	Colors          []string          `json:"colors"`
	Ruler           Ruler             `json:"ruler"`
	Motivation      string            `json:"motivation"`
	Strengths       []string          `json:"strengths"`
	Weaknesses      []string          `json:"weaknesses"`
	Values          []string          `json:"values"`
	Religion        string            `json:"religion,omitempty"`
	DominantClass   []string          `json:"dominant_class,omitempty"`
	Language        string            `json:"language,omitempty"`
	Relations       map[string]string `json:"relations,omitempty"`
	Military        Military          `json:"military,omitempty"`
	ValuesDisplayed []string          `json:"values_displayed,omitempty"` // For Lumenciel hypocrisy
	ValuesReal      []string          `json:"values_real,omitempty"`
}

// Ruler represents the ruler of a kingdom.
type Ruler struct {
	Name        string `json:"name"`
	Age         int    `json:"age"`
	Personality string `json:"personality"`
	Legitimacy  string `json:"legitimacy"`
}

// Military represents military strength data.
type Military struct {
	ProfessionalSoldiers int      `json:"professional_soldiers"`
	Conscripts           int      `json:"conscripts"`
	Navy                 string   `json:"navy"`
	SpecialUnits         []string `json:"special_units,omitempty"`
}

// Organization represents a secret organization or guild.
type Organization struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "guild", "secret", "religious", etc.
	Founded     string   `json:"founded,omitempty"`
	Headquarters string   `json:"headquarters,omitempty"`
	Activities  []string `json:"activities,omitempty"`
	Members     []string `json:"members,omitempty"`
	Influence   []string `json:"influence,omitempty"`
	Notes       string   `json:"notes,omitempty"`
}

// LocationNames contains naming conventions for all kingdoms.
type LocationNames struct {
	Kingdoms map[string]KingdomNaming `json:"valdorine,karvath,lumenciel,astrene"`
	Neutral  NeutralNaming            `json:"neutral,omitempty"`
}

// KingdomNaming contains naming patterns for a specific kingdom.
type KingdomNaming struct {
	Cities   NamingPattern `json:"cities"`
	Towns    NamingPattern `json:"towns"`
	Villages NamingPattern `json:"villages"`
	Regions  RegionNaming  `json:"regions,omitempty"`
}

// NamingPattern defines prefixes, roots, and suffixes for name generation.
type NamingPattern struct {
	Prefixes []string `json:"prefixes"`
	Roots    []string `json:"roots,omitempty"`
	Suffixes []string `json:"suffixes"`
	Names    []string `json:"names,omitempty"` // For regions (direct names)
}

// RegionNaming defines region naming templates.
type RegionNaming struct {
	Prefixes  []string `json:"prefixes,omitempty"`
	Suffixes  []string `json:"suffixes,omitempty"`
	Templates []string `json:"templates,omitempty"` // e.g., "Côte de {name}"
}

// NeutralNaming contains neutral location naming patterns.
type NeutralNaming struct {
	Ruins   NamingPattern `json:"ruins,omitempty"`
	Generic NamingPattern `json:"generic,omitempty"`
	Special NamingPattern `json:"special,omitempty"`
}
