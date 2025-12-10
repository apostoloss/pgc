package version

import "os"

// Version is the current runtime version of the application.
var Version = "0.1"

// Get returns the runtime version (via env or set string)
func Get() string {
	if v := os.Getenv("FAVORITES_API_VERSION"); v != "" {
		return v
	}
	return Version
}
