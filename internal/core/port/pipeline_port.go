package port

import (
	"context"
)

type PlacesETLServicePort interface {
	SearchResultsToPostgres(ctx context.Context) error
}
