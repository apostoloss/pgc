package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"my-solution/internal/models"
)

// Catalog holds all available assets in the system.
// This represents the "huge list of assets" that all users have access to.
type Catalog struct {
	mu     sync.RWMutex
	assets map[string]models.Asset
}

// Global is the singleton instance of the catalog.
var Global *Catalog

// Initialize creates and loads the global catalog.
func Initialize() {
	Global = &Catalog{
		assets: make(map[string]models.Asset),
	}
}

// AddAsset inserts an asset for testing purposes.
func (c *Catalog) AddAsset(id string, asset models.Asset) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.assets[id] = asset
}

// LoadFromFile loads assets from a JSON seed file.
func (c *Catalog) LoadFromFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No seed file, start with empty catalog
			return nil
		}
		return fmt.Errorf("failed to open catalog file: %w", err)
	}
	defer file.Close()

	// Structure matches scripts/seed_assets.json
	type SeedData struct {
		Charts    []models.Chart    `json:"charts"`
		Insights  []models.Insight  `json:"insights"`
		Audiences []models.Audience `json:"audiences"`
	}

	var data SeedData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("failed to parse catalog file: %w", err)
	}

	// Load charts
	for _, chart := range data.Charts {
		c.assets[chart.ID] = chart
	}

	// Load insights
	for _, insight := range data.Insights {
		c.assets[insight.ID] = insight
	}

	// Load audiences
	for _, audience := range data.Audiences {
		c.assets[audience.ID] = audience
	}

	return nil
}

// Get retrieves an asset by ID.
func (c *Catalog) Get(id string) (models.Asset, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	asset, ok := c.assets[id]
	return asset, ok
}

// List returns all available assets.
func (c *Catalog) List() []models.Asset {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]models.Asset, 0, len(c.assets))
	for _, asset := range c.assets {
		result = append(result, asset)
	}
	return result
}

// Count returns the total number of assets in the catalog.
func (c *Catalog) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.assets)
}
