package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"my-solution/internal/catalog"
	"my-solution/internal/models"
	"my-solution/internal/store"

	"github.com/gorilla/mux"
)

// Helper to setup router for integration tests
func setupIntegrationRouter() (*mux.Router, store.Store) {
	s := store.NewMemoryStore()
	api := &API{
		Store: s,
	}
	r := mux.NewRouter()
	api.RegisterHandlers(r)
	return r, s
}

// Helper to execute request
func executeRequest(r *mux.Router, method, path string, payload interface{}) *httptest.ResponseRecorder {
	var body []byte
	if payload != nil {
		body, _ = json.Marshal(payload)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	return res
}

func TestIntegration_HappyPath_Lifecycle(t *testing.T) {
	// Initialize and add assets to the catalog
	catalog.Initialize()
	chart := &models.Chart{
		AssetBase:  models.AssetBase{ID: "chart-2024", Name: "Revenue 2024", Description: "Q1-Q4 Revenue"},
		ChartType:  "bar",
		DataSource: "Sales DB",
	}
	insight := &models.Insight{
		AssetBase: models.AssetBase{ID: "insight-growth", Name: "Growth Insight", Description: "Yearly growth"},
		Metric:    "Growth",
		Value:     "15%",
	}
	catalog.Global.AddAsset(chart.GetID(), chart)
	catalog.Global.AddAsset(insight.GetID(), insight)

	r, _ := setupIntegrationRouter()
	userID := "happy_user"

	// 1. Add Chart by assetID
	chartPayload := map[string]interface{}{
		"assetID":     chart.GetID(),
		"description": "Custom: Q1-Q4 Revenue",
	}
	res := executeRequest(r, "POST", "/users/"+userID+"/favorites", chartPayload)
	if res.Code != http.StatusCreated {
		t.Fatalf("Failed to add chart: %d", res.Code)
	}
	// var resp map[string]string
	// json.NewDecoder(res.Body).Decode(&resp)
	chartID := chartPayload["assetID"].(string)

	// 2. Add Insight by assetID
	insightPayload := map[string]interface{}{
		"assetID":     insight.GetID(),
		"description": "Custom: Growth",
	}
	res = executeRequest(r, "POST", "/users/"+userID+"/favorites", insightPayload)
	if res.Code != http.StatusCreated {
		t.Fatalf("Failed to add insight: %d", res.Code)
	}

	// 3. List Favorites
	res = executeRequest(r, "GET", "/users/"+userID+"/favorites", nil)
	if res.Code != http.StatusOK {
		t.Fatalf("Failed to list favorites: %d", res.Code)
	}
	var assets []map[string]interface{}
	json.NewDecoder(res.Body).Decode(&assets)
	if len(assets) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(assets))
	}

	// 4. Edit Chart Description
	editPayload := map[string]string{"description": "Updated Revenue"}
	res = executeRequest(r, "PATCH", "/users/"+userID+"/favorites/"+chartID, editPayload)
	if res.Code != http.StatusNoContent {
		t.Fatalf("Failed to edit chart: %d", res.Code)
	}

	// 5. Verify Edit
	res = executeRequest(r, "GET", "/users/"+userID+"/favorites", nil)
	json.NewDecoder(res.Body).Decode(&assets)
	found := false
	for _, a := range assets {
		if a["assetId"] == chartID {
			if a["description"] != "Updated Revenue" {
				t.Errorf("Description not updated. Got %s", a["Description"])
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Chart not found after edit got: %s", assets)
	}

	// 6. Remove Chart
	res = executeRequest(r, "DELETE", "/users/"+userID+"/favorites/"+chartID, nil)
	if res.Code != http.StatusNoContent {
		t.Fatalf("Failed to remove chart: %d", res.Code)
	}

	// 7. Verify Removal
	res = executeRequest(r, "GET", "/users/"+userID+"/favorites", nil)
	json.NewDecoder(res.Body).Decode(&assets)
	if len(assets) != 1 {
		t.Errorf("Expected 1 asset after removal, got %d", len(assets))
	}
}

func TestIntegration_ValidationAndErrors(t *testing.T) {
	r, _ := setupIntegrationRouter()
	userID := "validation_user"

	// 1. Invalid Payload (Missing Required Fields)
	invalidChart := map[string]interface{}{
		"type": "chart",
		"chart": map[string]interface{}{
			"name": "No Type Chart",
			// Missing chartType
		},
	}
	res := executeRequest(r, "POST", "/users/"+userID+"/favorites", invalidChart)
	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing chartType, got %d", res.Code)
	}

	// 2. Duplicate Names
	chart1 := map[string]interface{}{
		"type": "chart",
		"chart": map[string]interface{}{
			"name":      "Unique Name",
			"chartType": "bar",
		},
	}
	executeRequest(r, "POST", "/users/"+userID+"/favorites", chart1)

	chart2 := map[string]interface{}{
		"type": "chart",
		"chart": map[string]interface{}{
			"name":      "unique name", // Case-insensitive duplicate
			"chartType": "line",
		},
	}
	res = executeRequest(r, "POST", "/users/"+userID+"/favorites", chart2)
	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for duplicate name, got %d", res.Code)
	}

	// 3. Malformed JSON
	req := httptest.NewRequest("POST", "/users/"+userID+"/favorites", bytes.NewReader([]byte(`{"type": "chart", "chart": {`))) // Incomplete JSON
	resRec := httptest.NewRecorder()
	r.ServeHTTP(resRec, req)
	if resRec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for malformed JSON, got %d", resRec.Code)
	}

	// 4. Unknown Asset Type
	unknownType := map[string]interface{}{
		"type": "alien_tech",
	}
	res = executeRequest(r, "POST", "/users/"+userID+"/favorites", unknownType)
	if res.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for unknown asset type, got %d", res.Code)
	}

	// 5. Edit Non-Existent Asset
	editPayload := map[string]string{"description": "Ghost update"}
	res = executeRequest(r, "PATCH", "/users/"+userID+"/favorites/non-existent-uuid", editPayload)
	if res.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for editing non-existent asset, got %d", res.Code)
	}
}

func TestIntegration_UserIsolation(t *testing.T) {

	catalog.Initialize()
	chart := &models.Chart{
		AssetBase:  models.AssetBase{ID: "chart-2024", Name: "Revenue 2024", Description: "Q1-Q4 Revenue"},
		ChartType:  "bar",
		DataSource: "Sales DB",
	}

	catalog.Global.AddAsset(chart.GetID(), chart)

	r, s := setupIntegrationRouter()
	userA := "user_a"
	userB := "user_b"

	// 1. Add Chart by assetID
	chartPayload := map[string]interface{}{
		"assetID":     chart.GetID(),
		"description": "Desc A",
	}
	res := executeRequest(r, "POST", "/users/"+userA+"/favorites", chartPayload)
	if res.Code != http.StatusCreated {
		t.Fatalf("Failed to add chart: %d", res.Code)
	}

	// After POST...
	resList := executeRequest(r, "GET", "/users/"+userA+"/favorites", nil)
	var listed []map[string]interface{}
	json.NewDecoder(resList.Body).Decode(&listed)
	if len(listed) == 0 {
		t.Fatalf("Expected at least one favorite for user %s immediately after addition", userA)
	}
	storeList, _ := s.ListFavorites(userA)
	if len(storeList) == 0 {
		t.Fatalf("Expected at least one favorite in store for user %s immediately after addition", userA)
	}
	idA := storeList[0].AssetID
	listedID := listed[0]["assetId"].(string)
	if idA != listedID {
		t.Fatalf("Mismatched IDs between store and API listing: store %s vs api %s", idA, listedID)
	}

	// 2. Verify User B has no assets
	resB := executeRequest(r, "GET", "/users/"+userB+"/favorites", nil)
	var assetsB []map[string]interface{}
	json.NewDecoder(resB.Body).Decode(&assetsB)
	if len(assetsB) != 0 {
		t.Errorf("Expected User B to have 0 assets, got %d", len(assetsB))
	}

	// 3. User B tries to edit User A's asset
	editPayload := map[string]string{"description": "Hacked Description"}
	res = executeRequest(r, "PATCH", "/users/"+userB+"/favorites/"+idA, editPayload)
	if res.Code != http.StatusNotFound {
		t.Errorf("Expected 404 when User B tries to edit User A's asset, got %d", res.Code)
	}

	// 4. Verify User A's asset is unchanged (store verification)
	assetsA, err := s.ListFavorites(userA)
	if err != nil {
		t.Fatalf("list favorites: %v", err)
	}
	if len(assetsA) == 0 {
		t.Fatalf("User A's assets missing after edit attempt")
	}
	if assetsA[0].Description != "Desc A" {
		t.Error("User A's asset was modified by User B!")
	}

	// 4b. Verify User A's asset via API
	resA := executeRequest(r, "GET", "/users/"+userA+"/favorites", nil)
	var apiAssetsA []map[string]interface{}
	json.NewDecoder(resA.Body).Decode(&apiAssetsA)
	if len(apiAssetsA) == 0 {
		t.Fatalf("User A's assets missing via API after edit attempt")
	}
	// Compare descriptions
	if desc, ok := apiAssetsA[0]["description"].(string); !ok || desc != "Desc A" {
		t.Errorf("User A's API asset was modified! Got description: %v", apiAssetsA[0]["description"])
	}
}

func TestIntegration_EdgeCases(t *testing.T) {

	catalog.Initialize()
	chart := &models.Chart{
		AssetBase:  models.AssetBase{ID: "chart-2024", Name: "Revenue 2024", Description: "Q1-Q4 Revenue"},
		ChartType:  "bar",
		DataSource: "Sales DB",
	}

	catalog.Global.AddAsset(chart.GetID(), chart)

	r, _ := setupIntegrationRouter()
	userID := "edge_user"

	// 1. Empty List
	res := executeRequest(r, "GET", "/users/"+userID+"/favorites", nil)
	if res.Code != http.StatusOK {
		t.Errorf("Expected 200 for empty list, got %d", res.Code)
	}
	var assets []map[string]interface{}
	json.NewDecoder(res.Body).Decode(&assets)
	if assets == nil || len(assets) != 0 {
		t.Error("Expected empty slice [], got nil or non-empty")
	}

	// 2. Large Payload
	longDesc := strings.Repeat("A", 10000)
	largeChart := map[string]interface{}{
		"assetID":     chart.GetID(),
		"description": longDesc,
	}

	res = executeRequest(r, "POST", "/users/"+userID+"/favorites", largeChart)
	if res.Code != http.StatusCreated {
		t.Errorf("Failed to add large payload: %d", res.Code)
	}
}
