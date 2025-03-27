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


// TextSearchAPI implements PlacesAPIAdapter for Text Search
type TextSearchAPI struct{}

func NewTextSearchAPI() *TextSearchAPI {
	return &TextSearchAPI{}
}

func (a *TextSearchAPI) FetchPlaces(ctx context.Context, params domain.RequestParams) ([]interface{}, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	url := "https://places.googleapis.com/v1/places:searchText"

	// Text Search uses a query string and optional location bias
	requestBody, _ := json.Marshal(map[string]interface{}{
		"textQuery": params.Query,
		"locationBias": map[string]interface{}{
			"circle": map[string]interface{}{
				"center": map[string]float64{
					"latitude":  params.Circle.Lat,
					"longitude": params.Circle.Lng,
				},
				"radius": params.Circle.Radius,
			},
		},
		"maxResultCount": domain.MaxResultsPerReq,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("error creating API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.id,places.displayName,places.types,places.viewport")

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