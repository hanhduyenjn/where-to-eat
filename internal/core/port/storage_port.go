package port

import (
	"context"
)

type Storage interface {
	Upload(ctx context.Context, src string, dst string) (string, error)
}