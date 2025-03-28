package port

import (
	"context"
	"wheretoeat/internal/core/domain"
)

type PlacesAPIPort interface {
	FetchPlaces(ctx context.Context, params domain.RequestParams) ([]interface{}, error)
}
