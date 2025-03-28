package port

import (
	"context"
)

type Uploader interface {
	Upload(ctx context.Context, src string, dst string) (string, error)
}