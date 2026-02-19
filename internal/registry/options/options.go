package options

import (
	"context"
	"time"
)

// StorageOptions configures storage behavior.
type StorageOptions struct {
	Timeout time.Duration
}

// WithTimeout applies a default timeout when none is set on the context.
func (o StorageOptions) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if o.Timeout <= 0 {
		return ctx, nil
	}
	if _, ok := ctx.Deadline(); ok {
		return ctx, nil
	}

	return context.WithTimeout(ctx, o.Timeout)
}
