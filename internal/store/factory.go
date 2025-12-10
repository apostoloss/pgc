package store

// TODO: Factory disabled until FileStore is refactored
/*
import (
	"os"
)

// NewStore creates a new store based on environment configuration.
// If DATA_FILE is set, it uses FileStore, otherwise MemoryStore.
func NewStore() (Store, error) {
	dataFile := os.Getenv("DATA_FILE")

	if dataFile != "" {
		// Use file-based storage
		return NewFileStore(dataFile)
	}

	// Default to in-memory storage
	return NewMemoryStore(), nil
}
*/
