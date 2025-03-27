package port

import (
	"context"

	"wheretoeat/internal/core/domain"

	"go.mongodb.org/mongo-driver/bson"
)

type PlacesRepository interface {
	SavePlaces(ctx context.Context, category string, circle domain.Circle, places []interface{}) error
	CircleExists(ctx context.Context, category string, circle domain.Circle) (bool, error)
	GetNumPlaces(ctx context.Context, category string, circle domain.Circle) (int64, error)
	GetPlaces(ctx context.Context, limit int, offset int) ([]domain.Place, error)
}

type CategoriesRepository interface {
	GetCategoryTypes(ctx context.Context, category string) ([]string, error)
}

type AreasRepository interface {
	SaveArea(ctx context.Context, area bson.M) error
	GetAreaByPlaceID(ctx context.Context, placeID string) (bson.M, error)
}
