package port

import (
	"context"

	"wheretoeat/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
)

// Raw data from Google Places API are stored in SearchResultsRepository
type SearchResultsRepository interface {
	SaveSearchResults(ctx context.Context, category string, circle domain.Circle, places []interface{}) error
	AreaHasBeenScanned(ctx context.Context, category string, circle domain.Circle) (bool, error)
	GetNumPlaces(ctx context.Context, category string, circle domain.Circle) (int64, error) // avoid fetching area of same circle again
}

type PlacesRepository interface {
	SavePlace(ctx context.Context, place domain.Place) error
	SaveBatch(ctx context.Context, places []domain.Place, photos []domain.Photo, reviews []domain.Review, openingHours []struct {
		PlaceID string
		Type    string
		Periods string
	}, placeTypes []struct {
		PlaceID string
		Type    string
	}) error
	GetPhotos(ctx context.Context, limit, offset int) ([]domain.Photo, error)
	GetNearbyPlaces(ctx context.Context, category string, circle domain.Circle, searchString string) ([]domain.Place, error)
	UpdatePhotoURL(ctx context.Context, imgURL, placeID, photoURL string) error
}

type CategoriesRepository interface {
	GetCategoryTypes(ctx context.Context, category string) ([]string, error)
}

type AreasRepository interface {
	SaveArea(ctx context.Context, area bson.M) error
	GetAreaByPlaceID(ctx context.Context, placeID string) (bson.M, error)
}
