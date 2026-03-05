package generic

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"

	"sgroups.io/sgroups-k8s-api/internal/registry/base"
)

// Backend abstracts gRPC operations for a single resource type.
// All methods return K8s API errors (map gRPC errors internally via regerrors.FromGRPC).
// User metadata injection is handled by the gRPC interceptor — backends must NOT call InjectUserMetadata.
type Backend[T runtime.Object, TList runtime.Object] interface {
	// List returns a list of resources matching the selection.
	// For Get operations, sel will contain Name (and Namespace for namespaced resources).
	List(ctx context.Context, sel base.Selection) (TList, error)

	// Upsert creates or updates a resource and returns the result.
	Upsert(ctx context.Context, obj T) (T, error)

	// Delete removes a resource by name and namespace.
	// For cluster-scoped resources, namespace is empty.
	Delete(ctx context.Context, name, namespace string) error

	// Watch opens a watch stream for resources matching the selection.
	Watch(ctx context.Context, sel base.Selection, resourceVersion string) (watch.Interface, error)

	// Resource returns the GroupResource for error messages.
	Resource() schema.GroupResource

	// NamespaceScoped reports whether this resource is namespace-scoped.
	NamespaceScoped() bool
}
