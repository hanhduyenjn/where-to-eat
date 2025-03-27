package port

import (
	"context"
	"wheretoeat/internal/core/domain"
)

type PlacesAPIAdapter interface {
	FetchPlaces(ctx context.Context, params domain.RequestParams) ([]interface{}, error)
}
