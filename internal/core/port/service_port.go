package port

import (
	"context"
	"wheretoeat/internal/core/domain"
)

type GetPlacesServicePort interface {
	GetNearbyPlaces(ctx context.Context, lat, lng, radius float64, category string, searchString string) ([]domain.Place, error)
}