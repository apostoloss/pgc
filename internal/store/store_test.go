package store

import (
	"testing"

	"my-solution/internal/catalog"
	"my-solution/internal/models"
)

func TestStore_AddListRemoveEdit(t *testing.T) {
	catalog.Initialize()
	store := NewMemoryStore()
	userID := "user-test"

	// Mock catalog with assets, using asset IDs
	chart := &models.Chart{
		AssetBase: models.AssetBase{
			ID:          "chart-1",
			Name:        "Revenue Q1",
			Description: "Shows quarterly revenue",
		},
		ChartType:  "bar",
		DataSource: "db-q1",
	}
	insight := &models.Insight{
		AssetBase: models.AssetBase{
			ID:          "insight-1",
			Name:        "Social Media Insight",
			Description: "40% engage 3+ hours",
		},
		Metric: "Engagement",
		Value:  "High",
	}
	audience := &models.Audience{
		AssetBase: models.AssetBase{
			ID:          "audience-1",
			Name:        "Gen Z Females",
			Description: "Females aged 18-24",
		},
		Segment: "Females 18-24",
		Size:    12000,
	}

	// Populate the (global) catalog for asset lookups.
	// Needs an importable or assignable catalog.Global map for joining.
	catalog.Global.AddAsset(chart.GetID(), chart)
	catalog.Global.AddAsset(insight.GetID(), insight)
	catalog.Global.AddAsset(audience.GetID(), audience)

	// Add Chart Favorite (by ID)
	err := store.AddFavorite(userID, chart.GetID(), chart.GetDescription())
	if err != nil {
		t.Fatalf("add chart: %v", err)
	}

	// Add Insight Favorite
	err = store.AddFavorite(userID, insight.GetID(), insight.GetDescription())
	if err != nil {
		t.Fatalf("add insight: %v", err)
	}

	// Add Audience Favorite
	err = store.AddFavorite(userID, audience.GetID(), audience.GetDescription())
	if err != nil {
		t.Fatalf("add audience: %v", err)
	}

	// List Favorites
	favs, err := store.ListFavorites(userID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(favs) != 3 {
		t.Errorf("expected 3 favorites, got %d", len(favs))
	}

	// Edit Description of chart
	newDesc := "Updated Description"
	if err := store.EditFavoriteDescription(userID, chart.GetID(), newDesc); err != nil {
		t.Fatalf("edit chart desc: %v", err)
	}
	favs, _ = store.ListFavorites(userID)
	found := false
	for _, a := range favs {
		// Check desc updated for chart
		if a.AssetID == chart.GetID() && a.Description != newDesc {
			t.Errorf("description not updated, got %v", a.Description)
		}
		if a.AssetID == insight.GetID() {
			found = true
		}
	}
	if !found {
		t.Errorf("insight asset not found after edit")
	}

	// Remove Audience Favorite
	if err := store.RemoveFavorite(userID, audience.GetID()); err != nil {
		t.Fatalf("remove audience: %v", err)
	}
	favs, _ = store.ListFavorites(userID)
	if len(favs) != 2 {
		t.Errorf("expected 2 favorites after remove, got %d", len(favs))
	}
}
