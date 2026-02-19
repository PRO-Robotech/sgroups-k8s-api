package generic

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

// Strategy provides optional per-resource custom validation.
// Nil strategy means no custom validation (only base name/namespace checks apply).
type Strategy interface {
	ValidateCreate(ctx context.Context, obj runtime.Object) error
	ValidateUpdate(ctx context.Context, obj, old runtime.Object) error
}
