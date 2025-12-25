package world

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadGeography loads geography data from data/world/geography.json
func LoadGeography(dataDir string) (*Geography, error) {
	path := filepath.Join(dataDir, "world", "geography.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading geography file: %w", err)
	}

	var geo Geography
	if err := json.Unmarshal(data, &geo); err != nil {
		return nil, fmt.Errorf("parsing geography JSON: %w", err)
	}

	return &geo, nil
}

// LoadFactions loads faction data from data/world/factions.json
func LoadFactions(dataDir string) (*Factions, error) {
	path := filepath.Join(dataDir, "world", "factions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading factions file: %w", err)
	}

	var factions Factions
	if err := json.Unmarshal(data, &factions); err != nil {
		return nil, fmt.Errorf("parsing factions JSON: %w", err)
	}

	return &factions, nil
}

// LoadLocationNames loads location naming conventions from data/location-names.json
func LoadLocationNames(dataDir string) (*LocationNames, error) {
	path := filepath.Join(dataDir, "location-names.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading location-names file: %w", err)
	}

	var names LocationNames
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, fmt.Errorf("parsing location-names JSON: %w", err)
	}

	return &names, nil
}

// GetLocationByName searches for a location by name across all regions.
// Returns the location and its parent region if found.
func (g *Geography) GetLocationByName(name string) (*Location, *Region, error) {
	nameLower := strings.ToLower(name)

	for _, continent := range g.Continents {
		for i := range continent.Regions {
			region := &continent.Regions[i]
			for j := range region.Cities {
				loc := &region.Cities[j]
				if strings.ToLower(loc.Name) == nameLower {
					return loc, region, nil
				}
			}
		}
	}

	return nil, nil, fmt.Errorf("location '%s' not found", name)
}

// GetLocationsByKingdom returns all locations belonging to a kingdom.
func (g *Geography) GetLocationsByKingdom(kingdomID string) []Location {
	var locations []Location
	kingdomLower := strings.ToLower(kingdomID)

	for _, continent := range g.Continents {
		for _, region := range continent.Regions {
			if strings.ToLower(region.Kingdom) == kingdomLower {
				locations = append(locations, region.Cities...)
			}
		}
	}

	return locations
}

// GetRegionsByKingdom returns all regions belonging to a kingdom.
func (g *Geography) GetRegionsByKingdom(kingdomID string) []Region {
	var regions []Region
	kingdomLower := strings.ToLower(kingdomID)

	for _, continent := range g.Continents {
		for _, region := range continent.Regions {
			if strings.ToLower(region.Kingdom) == kingdomLower {
				regions = append(regions, region)
			}
		}
	}

	return regions
}

// GetTradeRoutesByLocation returns trade routes passing through a location.
func (g *Geography) GetTradeRoutesByLocation(locationName string) []TradeRoute {
	var routes []TradeRoute
	nameLower := strings.ToLower(locationName)

	for _, route := range g.TradeRoutes {
		for _, loc := range route.Locations {
			if strings.ToLower(loc) == nameLower {
				routes = append(routes, route)
				break
			}
		}
	}

	return routes
}

// GetAllLocations returns all locations from all regions.
func (g *Geography) GetAllLocations() []Location {
	var locations []Location

	for _, continent := range g.Continents {
		for _, region := range continent.Regions {
			locations = append(locations, region.Cities...)
		}
	}

	return locations
}

// GetAllRegions returns all regions from all continents.
func (g *Geography) GetAllRegions() []Region {
	var regions []Region

	for _, continent := range g.Continents {
		regions = append(regions, continent.Regions...)
	}

	return regions
}

// GetKingdomByID returns a kingdom by its ID from factions data.
func (f *Factions) GetKingdomByID(kingdomID string) (*Kingdom, error) {
	idLower := strings.ToLower(kingdomID)

	for i := range f.Kingdoms {
		kingdom := &f.Kingdoms[i]
		if strings.ToLower(kingdom.ID) == idLower {
			return kingdom, nil
		}
	}

	return nil, fmt.Errorf("kingdom '%s' not found", kingdomID)
}

// GetAllKingdoms returns all kingdoms.
func (f *Factions) GetAllKingdoms() []Kingdom {
	return f.Kingdoms
}

// GetKingdomByLocation returns the kingdom that controls a location.
func (f *Factions) GetKingdomByLocation(location *Location) (*Kingdom, error) {
	return f.GetKingdomByID(location.Kingdom)
}
