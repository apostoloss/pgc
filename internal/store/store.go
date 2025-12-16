package store

import (
	"errors"
	"sync"
	"time"

	"my-solution/internal/catalog"
	"my-solution/internal/models"
)

var (
	ErrAssetNotFound = errors.New("asset not found")
	ErrDuplicateName = errors.New("asset with this name already exists")
)

// Store defines the interface for managing user favorites.
// Favorites store only references to assets (by ID) plus user metadata,
// not the full asset objects themselves.
type Store interface {
	// AddFavorite adds an asset to user's favorites by asset ID
	AddFavorite(userID, assetID, description string) error

	// ListFavorites returns user's favorites with full asset data joined from catalog
	ListFavorites(userID string) ([]models.FavoriteWithAsset, error)

	// RemoveFavorite removes an asset from user's favorites
	RemoveFavorite(userID, assetID string) error

	// EditFavoriteDescription updates the user's custom description for a favorite
	EditFavoriteDescription(userID, assetID, desc string) error
}

// MemoryStore manages user favorites in-memory with concurrency safety.
// It stores only favorite references (asset IDs + metadata), not full asset copies.
type MemoryStore struct {
	mu    sync.Mutex
	users map[string][]models.Favorite // userID -> array of favorite references
}

// NewMemoryStore initializes and returns a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users: make(map[string][]models.Favorite),
	}
}

// AddFavorite adds a favorite reference by asset ID.
func (s *MemoryStore) AddFavorite(userID, assetID, description string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already favorited
	for _, fav := range s.users[userID] {
		if fav.AssetID == assetID {
			return errors.New("asset already favorited")
		}
	}

	// Add new favorite reference
	s.users[userID] = append(s.users[userID], models.Favorite{
		AssetID:     assetID,
		Description: description,
		CreatedAt:   time.Now(),
	})
	return nil
}

// ListFavorites returns user's favorites with full asset data from catalog.
func (s *MemoryStore) ListFavorites(userID string) ([]models.FavoriteWithAsset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	favorites := s.users[userID]
	result := make([]models.FavoriteWithAsset, 0, len(favorites))

	for _, fav := range favorites {
		// Look up asset in catalog
		asset, ok := catalog.Global.Get(fav.AssetID)
		if !ok {
			// Asset no longer exists in catalog, skip it
			continue
		}

		result = append(result, models.FavoriteWithAsset{
			AssetID:     fav.AssetID,
			Description: fav.Description,
			CreatedAt:   fav.CreatedAt,
			Asset:       asset,
		})
	}

	return result, nil
}

// RemoveFavorite removes an asset from a user's favorites by asset ID.
func (s *MemoryStore) RemoveFavorite(userID, assetID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	favorites := s.users[userID]
	for i, fav := range favorites {
		if fav.AssetID == assetID {
			// Remove by swapping with last element and truncating
			s.users[userID] = append(favorites[:i], favorites[i+1:]...)
			return nil
		}
	}
	return nil // Not found, but not an error
}

// EditFavoriteDescription edits the user's custom description for a favorite.
func (s *MemoryStore) EditFavoriteDescription(userID, assetID, desc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	favorites := s.users[userID]
	for i, fav := range favorites {
		if fav.AssetID == assetID {
			s.users[userID][i].Description = desc
			return nil
		}
	}

	return ErrAssetNotFound
}
