package models

import "time"

// Favorite represents a user's reference to a favorited asset.
// It stores only the asset ID and user-specific metadata, not the full asset.
type Favorite struct {
	AssetID     string    `json:"assetId"`
	Description string    `json:"description"` // User's personal note
	CreatedAt   time.Time `json:"createdAt"`
}

// FavoriteWithAsset combines the favorite reference with the full asset data.
// This is what we return to API clients when listing favorites.
type FavoriteWithAsset struct {
	AssetID     string    `json:"assetId"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	Asset       Asset     `json:"asset"` // Full asset from catalog
}
