package port

import (
	"context"
)

type FetchAreasPort interface {
	FetchAreas(ctx context.Context, query string) error
}

type FetchPlacesPort interface {
	FetchPlaces(ctx context.Context, minLat, maxLat, minLng, maxLng float64, category string) error
}

type FetchImagesPort interface {
	FetchImages(ctx context.Context) error
}
