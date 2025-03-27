package fetch

import (
	"context"
	"log"
	"time"

	"wheretoeat/internal/core/domain"
	"wheretoeat/internal/core/port"

	"go.mongodb.org/mongo-driver/bson"
)

type FetchAreasService struct {
	areasRepo  port.AreasRepository
	apiAdapter port.PlacesAPIAdapter
}

func NewFetchAreasService(areasRepo port.AreasRepository, apiAdapter port.PlacesAPIAdapter) *FetchAreasService {
	return &FetchAreasService{
		areasRepo:  areasRepo,
		apiAdapter: apiAdapter,
	}
}

func (s *FetchAreasService) FetchAreas(ctx context.Context, query string) error {
	// Define request parameters for Text Search without location bias
	params := domain.RequestParams{
		Query: query,
	}

	// Fetch places using the Text Search API
	placesRaw, err := s.apiAdapter.FetchPlaces(ctx, params)
	if err != nil {
		log.Printf("Failed to fetch areas for query '%s': %v", query, err)
		return err
	}

	if len(placesRaw) == 0 {	
		log.Printf("No areas found for query '%s'", query)
		return nil
	}

	place, ok := placesRaw[0].(map[string]interface{})
	if !ok {
		log.Printf("Invalid place format for query '%s'", query)
		return nil
	}

	typesRaw, ok := place["types"].([]interface{})
	if !ok || len(typesRaw) == 0 {
		log.Printf("Place %s has no valid types for query '%s'", query, query)
		return nil
	}

	firstType, ok := typesRaw[0].(string)
	if !ok || (firstType != "administrative_area_level_2" && firstType != "sublocality_level_1") {
		log.Printf("Place %s is not a ward nor a district for query '%s'", query, query)
		return nil
	}

	placeID, _ := place["id"].(string)
	nameRaw, _ := place["displayName"].(map[string]interface{})
	name, _ := nameRaw["text"].(string)
	viewportRaw, ok := place["viewport"].(map[string]interface{})
	if !ok || viewportRaw == nil {
		log.Printf("No viewport data for place %s", name)
		return nil
	}
	lowRaw, _ := viewportRaw["low"].(map[string]interface{})
	highRaw, _ := viewportRaw["high"].(map[string]interface{})
	minLat, _ := lowRaw["latitude"].(float64)
	minLng, _ := lowRaw["longitude"].(float64)
	maxLat, _ := highRaw["latitude"].(float64)
	maxLng, _ := highRaw["longitude"].(float64)

	area := bson.M{
		"placeID":   placeID,
		"name":      name,
		"minLat":    minLat,
		"maxLat":    maxLat,
		"minLng":    minLng,
		"maxLng":    maxLng,
		"query":     query, // Store the query instead of centerLat/centerLng/radius
		"timestamp": time.Now(),
	}

	err = s.areasRepo.SaveArea(ctx, area)
	if err != nil {
		log.Printf("Failed to upsert area for place %s: %v", name, err)
		return err
	}
	log.Printf("Upserted area for %s: minLat=%f, maxLat=%f, minLng=%f, maxLng=%f",
		name, minLat, maxLat, minLng, maxLng)

	return nil
}