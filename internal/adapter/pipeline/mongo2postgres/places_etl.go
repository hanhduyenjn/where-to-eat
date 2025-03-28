package pipeline

import (
	"context"
	"log"
	"encoding/json"


	"wheretoeat/internal/adapter/repository/mongodb"
	"wheretoeat/internal/adapter/repository/postgres"
	"wheretoeat/internal/core/domain"
)

const batchSize = 1000 // Adjust based on performance testing

type PlacesETLService struct {
	mongoRepo *mongodb.PlacesRepo
	pgRepo    *postgres.PlacesRepo
}

func NewPlacesETLService(mongoRepo *mongodb.PlacesRepo, pgRepo *postgres.PlacesRepo) *PlacesETLService {
	return &PlacesETLService{mongoRepo: mongoRepo, pgRepo: pgRepo}
}

func (s *PlacesETLService) SearchResultsToPostgres(ctx context.Context) error {
	// Extract raw data from MongoDB
	rawResponses, err := s.mongoRepo.GetRawPlaces(ctx)
	if err != nil {
		return err
	}

	// Prepare batches
	var placesBatch []domain.Place
	var photosBatch []domain.Photo
	var reviewsBatch []domain.Review
	var openingHoursBatch []struct {
		PlaceID string
		Type    string
		Periods string
	}
	var placeTypesBatch []struct {
		PlaceID string
		Type    string
	}

	for _, raw := range rawResponses {
		for _, place := range raw.Places {
			// Enrich place with parent-level fields
			lat := raw.Lat
			lng := raw.Lng
			if place.Location != nil {
				lat = place.Location.Lat
				lng = place.Location.Lng
			}
			place.Category = raw.Category
			place.Lat = lat
			place.Lng = lng
			if place.DisplayName != nil {
				place.Name = place.DisplayName.Text
			}
			log.Printf("Processing place user_rating: %v", place.UserRatingCount)

			placesBatch = append(placesBatch, place)

			// Collect related data
			for _, photo := range place.Photos {
				photo.PlaceID = place.ID
				photosBatch = append(photosBatch, photo)
			}

			if place.OpeningHours != nil {
				periodsJSON, err := json.Marshal(place.OpeningHours.Periods)
				if err != nil {
					log.Printf("Failed to serialize opening hours periods for place %s: %v", place.ID, err)
					continue
				}
				openingHoursBatch = append(openingHoursBatch, struct {
					PlaceID string
					Type    string
					Periods string
				}{place.ID, "regular", string(periodsJSON)})
			}
			if place.CurrentOpeningHours != nil {
				periodsJSON, err := json.Marshal(place.CurrentOpeningHours.Periods)
				if err != nil {
					log.Printf("Failed to serialize current opening hours periods for place %s: %v", place.ID, err)
					continue
				}
				openingHoursBatch = append(openingHoursBatch, struct {
					PlaceID string
					Type    string
					Periods string
				}{place.ID, "current", string(periodsJSON)})
			}
			for _, t := range place.Types {
				placeTypesBatch = append(placeTypesBatch, struct {
					PlaceID string
					Type    string
				}{place.ID, t})
			}

			// Process batch if it reaches the size limit
			if len(placesBatch) >= batchSize {
				if err := s.processBatch(ctx, placesBatch, photosBatch, reviewsBatch, openingHoursBatch, placeTypesBatch); err != nil {
					log.Printf("Failed to process batch: %v", err)
				}
				placesBatch = nil
				photosBatch = nil
				reviewsBatch = nil
				openingHoursBatch = nil
				placeTypesBatch = nil
			}
		}
	}

	// Process any remaining items
	if len(placesBatch) > 0 {
		if err := s.processBatch(ctx, placesBatch, photosBatch, reviewsBatch, openingHoursBatch, placeTypesBatch); err != nil {
			log.Printf("Failed to process final batch: %v", err)
		}
	}

	return nil
}

func (s *PlacesETLService) processBatch(ctx context.Context, places []domain.Place, photos []domain.Photo, reviews []domain.Review, openingHours []struct {
	PlaceID string
	Type    string
	Periods string
}, placeTypes []struct {
	PlaceID string
	Type    string
}) error {
	err := s.pgRepo.SaveBatch(ctx, places, photos, reviews, openingHours, placeTypes)
	if err != nil {
		return err
	}
	log.Printf("Processed batch of %d places", len(places))
	return nil
}