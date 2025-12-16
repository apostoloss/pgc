package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"my-solution/internal/catalog"
	"my-solution/internal/store"

	"github.com/gorilla/mux"
)

// API represents the API server with its store backend.
type API struct {
	Store store.Store
}

// RegisterHandlers sets up all API routes on the provided router.
func (api *API) RegisterHandlers(r *mux.Router) {
	// Browse available assets (catalog)
	r.HandleFunc("/assets", api.listAssetsHandler).Methods("GET")

	// List favorites for a user
	r.HandleFunc("/users/{id}/favorites", api.listFavoritesHandler).Methods("GET")
	// Add asset to user favorites
	r.HandleFunc("/users/{id}/favorites", api.addFavoriteHandler).Methods("POST")
	// Remove asset from user favorites
	r.HandleFunc("/users/{id}/favorites/{assetID}", api.removeFavoriteHandler).Methods("DELETE")
	// Edit favorite description
	r.HandleFunc("/users/{id}/favorites/{assetID}", api.editFavoriteHandler).Methods("PATCH")
}

// listAssetsHandler returns the catalog of all available assets.
// @Summary List all available assets
// @Description Get a list of all assets in the catalog
// @Tags assets
// @Produce json
// @Success 200 {array} object
// @Router /assets [get]
func (api *API) listAssetsHandler(w http.ResponseWriter, r *http.Request) {
	assets := catalog.Global.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assets)
}

// listFavoritesHandler retrieves all favorites for a user.
// @Summary List user's favorites
// @Description Get all favorites for a specific user
// @Tags favorites
// @Param id path string true "User ID"
// @Produce json
// @Success 200 {array} object
// @Failure 500 {string} string "Internal server error"
// @Router /users/{id}/favorites [get]
func (api *API) listFavoritesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	favorites, err := api.Store.ListFavorites(userID)
	if err != nil {
		http.Error(w, "failed to list favorites", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorites)
}

// AddFavoriteRequest defines the body for adding a new favorite.
type AddFavoriteRequest struct {
	AssetID     string `json:"assetId"`
	Description string `json:"description"`
}

// addFavoriteHandler adds an asset reference to the user's favorites.
// @Summary Add a favorite
// @Description Add an asset to user's favorites
// @Tags favorites
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body api.AddFavoriteRequest true "Asset ID and optional description"
// @Success 201 {string} string "Created"
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "Asset not found"
// @Failure 409 {string} string "Asset already favorited"
// @Failure 500 {string} string "Internal server error"
// @Router /users/{id}/favorites [post]
func (api *API) addFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var req AddFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.AssetID == "" {
		http.Error(w, "assetId is required", http.StatusBadRequest)
		return
	}

	// Validate asset exists in catalog
	if _, ok := catalog.Global.Get(req.AssetID); !ok {
		http.Error(w, "asset not found in catalog", http.StatusNotFound)
		return
	}

	// Add favorite directly
	err := api.Store.AddFavorite(userID, req.AssetID, req.Description)
	if err != nil {
		if err.Error() == "asset already favorited" {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "failed to add favorite", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// removeFavoriteHandler deletes an asset from the user's favorites.
// @Summary Remove a favorite
// @Description Remove an asset from user's favorites
// @Tags favorites
// @Param id path string true "User ID"
// @Param assetID path string true "Asset ID"
// @Success 204 {string} string "No Content"
// @Failure 500 {string} string "Internal server error"
// @Router /users/{id}/favorites/{assetID} [delete]
func (api *API) removeFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	assetID := vars["assetID"]

	// Remove favorite directly
	if err := api.Store.RemoveFavorite(userID, assetID); err != nil {
		http.Error(w, "failed to remove favorite", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// EditFavoriteRequest defines the body for editing a favorite description.
type EditFavoriteRequest struct {
	Description string `json:"description"`
}

// editFavoriteHandler edits the description of a user's asset.
// @Summary Edit favorite description
// @Description Update the description of an existing favorite
// @Tags favorites
// @Accept json
// @Param id path string true "User ID"
// @Param assetID path string true "Asset ID"
// @Param request body api.EditFavoriteRequest true "New description"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "Bad request"
// @Failure 404 {string} string "Favorite not found"
// @Failure 500 {string} string "Internal server error"
// @Router /users/{id}/favorites/{assetID} [patch]
func (api *API) editFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	assetID := vars["assetID"]

	var req EditFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Update description directly
	if err := api.Store.EditFavoriteDescription(userID, assetID, req.Description); err != nil {
		if errors.Is(err, store.ErrAssetNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update description", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
