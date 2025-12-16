package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"my-solution/internal/catalog"
	"my-solution/internal/models"
	"my-solution/internal/store"

	"github.com/gorilla/mux"
)

func setupRouter() (*mux.Router, store.Store) {
	s := store.NewMemoryStore()
	api := &API{Store: s}
	r := mux.NewRouter()
	api.RegisterHandlers(r)
	return r, s
}

func TestAddAndListFavoritesHandler(t *testing.T) {
	// Initialize catalog and add asset
	catalog.Initialize()
	chart := &models.Chart{
		AssetBase:  models.AssetBase{ID: "chart1-id", Name: "Chart1", Description: "Desc1"},
		ChartType:  "bar",
		DataSource: "source",
	}
	catalog.Global.AddAsset(chart.GetID(), chart)

	r, _ := setupRouter()
	userID := "user123"

	// Add favorite using assetID reference
	body := map[string]interface{}{
		"assetID":     chart.GetID(),
		"description": "Desc1",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/users/"+userID+"/favorites", bytes.NewReader(b))
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body: %s", res.Code, res.Body.String())
	}

	// List favorites
	req = httptest.NewRequest("GET", "/users/"+userID+"/favorites", nil)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}

	var assets []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&assets); err != nil {
		t.Fatalf("bad assets json: %v", err)
	}
	if len(assets) != 1 {
		t.Errorf("expected 1 asset, got %d", len(assets))
	}
}

func TestRemoveAndEditFavoriteHandler(t *testing.T) {
	catalog.Initialize()
	chart := &models.Chart{
		AssetBase: models.AssetBase{ID: "edit-chart", Name: "n", Description: "d"},
		ChartType: "t",
	}
	catalog.Global.AddAsset(chart.GetID(), chart)

	r, _ := setupRouter()
	userID := "userA"

	// Add favorite via API
	favBody := map[string]interface{}{
		"assetID":     chart.GetID(),
		"description": "d",
	}
	b, _ := json.Marshal(favBody)
	req := httptest.NewRequest("POST", "/users/"+userID+"/favorites", bytes.NewReader(b))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("add favorite: %d %s", res.Code, res.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(res.Body).Decode(&resp)
	id := chart.GetID()

	// Remove
	req = httptest.NewRequest("DELETE", "/users/"+userID+"/favorites/"+id, nil)
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Errorf("delete: expected 204 got %d", res.Code)
	}

	// Add again for edit test
	req = httptest.NewRequest("POST", "/users/"+userID+"/favorites", bytes.NewReader(b))
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("add favorite (2nd time): %d %s", res.Code, res.Body.String())
	}
	json.NewDecoder(res.Body).Decode(&resp)
	// id = resp["id"]

	edit := map[string]string{"description": "newdesc"}
	b, _ = json.Marshal(edit)

	req = httptest.NewRequest("PATCH", "/users/"+userID+"/favorites/"+id, bytes.NewReader(b))
	res = httptest.NewRecorder()

	r.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Errorf("patch: expected 204 got %d", res.Code)
	}
}

func TestAddDuplicateFavoriteHandler(t *testing.T) {
	catalog.Initialize()
	chart := &models.Chart{
		AssetBase: models.AssetBase{ID: "dup-chart", Name: "UniqueName", Description: "desc"},
		ChartType: "bar",
	}
	catalog.Global.AddAsset(chart.GetID(), chart)

	r, _ := setupRouter()
	userID := "userDup"

	// POST payload refers by assetID
	body := map[string]interface{}{
		"assetID":     chart.GetID(),
		"description": "desc",
	}
	b, _ := json.Marshal(body)

	// First add
	req := httptest.NewRequest("POST", "/users/"+userID+"/favorites", bytes.NewReader(b))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("first add failed: %d", res.Code)
	}

	// Second add (duplicate, same assetID and description)
	req = httptest.NewRequest("POST", "/users/"+userID+"/favorites", bytes.NewReader(b))
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusConflict {
		t.Errorf("expected 400 for duplicate, got %d", res.Code)
	}
}

func TestEditNonExistentFavoriteHandler(t *testing.T) {
	r, _ := setupRouter()
	userID := "userNonExistent"
	assetID := "fake-id"

	edit := map[string]string{"description": "newdesc"}
	b, _ := json.Marshal(edit)

	req := httptest.NewRequest("PATCH", "/users/"+userID+"/favorites/"+assetID, bytes.NewReader(b))
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent asset, got %d", res.Code)
	}
}

func TestAddInvalidFavoriteHandler(t *testing.T) {
	r, _ := setupRouter()
	userID := "userInvalid"

	// Missing name
	body := map[string]interface{}{
		"type": "chart",
		"chart": map[string]interface{}{
			"chartType": "bar",
		},
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users/"+userID+"/favorites", bytes.NewReader(b))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing name, got %d", res.Code)
	}
}
