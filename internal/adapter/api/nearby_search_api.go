package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"wheretoeat/internal/core/domain"
)

// NearbySearchAPI implements PlacesAPIAdapter for Nearby Search
type NearbySearchAPI struct{}

func NewNearbySearchAPI() *NearbySearchAPI {
	return &NearbySearchAPI{}
}

func (a *NearbySearchAPI) FetchPlaces(ctx context.Context, params domain.RequestParams) ([]interface{}, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	url := "https://places.googleapis.com/v1/places:searchNearby"

	requestBody, _ := json.Marshal(map[string]interface{}{
		"includedPrimaryTypes": params.Types,
		"maxResultCount":       domain.MaxResultsPerReq,
		"locationRestriction": map[string]interface{}{
			"circle": map[string]interface{}{
				"center": map[string]float64{
					"latitude":  params.Circle.Lat,
					"longitude": params.Circle.Lng,
				},
				"radius": params.Circle.Radius,
			},
		},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "*")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making API request: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding API response: %w", err)
	}

	places, ok := result["places"].([]interface{})
	if !ok {
		return nil, nil // No places found
	}

	return places, nil
}