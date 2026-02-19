package generic

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"

	"sgroups.io/sgroups-k8s-api/internal/registry/base"
	"sgroups.io/sgroups-k8s-api/internal/registry/options"
)

// Storage implements the K8s REST storage interfaces using a generic Backend.
type Storage[T runtime.Object, TList runtime.Object] struct {
	rest.TableConvertor

	backend      Backend[T, TList]
	strategy     Strategy
	opts         options.StorageOptions
	newT         func() T
	newTList     func() TList
	singularName string
}

// NewStorage creates a generic Storage for the given backend.
func NewStorage[T runtime.Object, TList runtime.Object](
	backend Backend[T, TList],
	opts options.StorageOptions,
	strategy Strategy,
	newT func() T,
	newTList func() TList,
	singularName string,
) *Storage[T, TList] {
	return &Storage[T, TList]{
		TableConvertor: rest.NewDefaultTableConvertor(backend.Resource()),
		backend:        backend,
		strategy:       strategy,
		opts:           opts,
		newT:           newT,
		newTList:       newTList,
		singularName:   singularName,
	}
}

// NamespaceScoped delegates to the backend.
func (s *Storage[T, TList]) NamespaceScoped() bool {
	return s.backend.NamespaceScoped()
}

// New returns an empty resource instance.
func (s *Storage[T, TList]) New() runtime.Object {
	return s.newT()
}

// NewList returns an empty list instance.
func (s *Storage[T, TList]) NewList() runtime.Object {
	return s.newTList()
}

// ConvertListOptions returns metav1.ListOptions.
func (s *Storage[T, TList]) ConvertListOptions(_ context.Context, opts *metainternalversion.ListOptions) *metav1.ListOptions {
	if opts == nil {
		return &metav1.ListOptions{}
	}
	out := &metav1.ListOptions{
		ResourceVersion:      opts.ResourceVersion,
		ResourceVersionMatch: opts.ResourceVersionMatch,
		TimeoutSeconds:       opts.TimeoutSeconds,
		Limit:                opts.Limit,
		Continue:             opts.Continue,
		SendInitialEvents:    opts.SendInitialEvents,
		Watch:                opts.Watch,
	}
	if opts.LabelSelector != nil {
		out.LabelSelector = opts.LabelSelector.String()
	}
	if opts.FieldSelector != nil {
		out.FieldSelector = opts.FieldSelector.String()
	}

	return out
}

// Get fetches a single resource by name.
func (s *Storage[T, TList]) Get(ctx context.Context, name string, opts *metav1.GetOptions) (runtime.Object, error) {
	if err := base.ValidateGetOptions(opts); err != nil {
		return nil, err
	}
	ctx, cancel := s.opts.WithTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}
	if err := base.RequireName(name); err != nil {
		return nil, err
	}

	sel := base.Selection{Name: name}
	if s.backend.NamespaceScoped() {
		ns, err := base.RequireNamespace(ctx)
		if err != nil {
			return nil, err
		}
		sel.Namespace = ns
	} else {
		if err := base.RequireClusterScope(ctx); err != nil {
			return nil, err
		}
	}

	list, err := s.backend.List(ctx, sel)
	if err != nil {
		return nil, err
	}
	items, err := meta.ExtractList(list)
	if err != nil {
		return nil, apierrors.NewInternalError(err)
	}
	if len(items) == 0 {
		return nil, apierrors.NewNotFound(s.backend.Resource(), name)
	}

	return items[0], nil
}

// List lists resources matching the options.
func (s *Storage[T, TList]) List(ctx context.Context, opts *metainternalversion.ListOptions) (runtime.Object, error) {
	v1Opts := s.ConvertListOptions(ctx, opts)
	ctx, cancel := s.opts.WithTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}
	sel, err := base.SelectionFromListOptions(ctx, v1Opts, s.backend.NamespaceScoped())
	if err != nil {
		return nil, err
	}

	return s.backend.List(ctx, sel)
}

// Watch starts watching resources matching the options.
// When sendInitialEvents=true (WatchList protocol), it synthesizes initial
// ADDED events from a List snapshot followed by a BOOKMARK before forwarding
// live events from the backend watch.
func (s *Storage[T, TList]) Watch(ctx context.Context, opts *metainternalversion.ListOptions) (watch.Interface, error) {
	v1Opts := s.ConvertListOptions(ctx, opts)
	sel, err := base.SelectionFromWatchOptions(ctx, v1Opts, s.backend.NamespaceScoped())
	if err != nil {
		return nil, err
	}

	sendInitialEvents := v1Opts.SendInitialEvents != nil && *v1Opts.SendInitialEvents
	if !sendInitialEvents {
		return s.backend.Watch(ctx, sel, v1Opts.ResourceVersion)
	}

	// WatchList protocol: synthesize initial events from List + live Watch.
	// Start Watch FIRST so we don't miss events between List and Watch.
	innerWatch, err := s.backend.Watch(ctx, sel, v1Opts.ResourceVersion)
	if err != nil {
		return nil, err
	}

	listObj, err := s.backend.List(ctx, sel)
	if err != nil {
		innerWatch.Stop()
		return nil, err
	}

	items, err := meta.ExtractList(listObj)
	if err != nil {
		innerWatch.Stop()
		return nil, apierrors.NewInternalError(err)
	}

	listAccessor, _ := meta.ListAccessor(listObj)
	listRV := ""
	if listAccessor != nil {
		listRV = listAccessor.GetResourceVersion()
	}

	return base.NewWatchListWatch(innerWatch, items, func() runtime.Object {
		return s.newT()
	}, listRV), nil
}

func (s *Storage[T, TList]) Create(
	ctx context.Context,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	opts *metav1.CreateOptions,
) (runtime.Object, error) {
	ctx, cancel := s.opts.WithTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}

	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, apierrors.NewInternalError(err)
	}
	if err := base.RequireName(accessor.GetName()); err != nil {
		return nil, err
	}

	if s.backend.NamespaceScoped() {
		if _, err := base.NormalizeObjectNamespace(ctx, accessor); err != nil {
			return nil, err
		}
	} else {
		if err := base.RequireClusterScopeObject(ctx, accessor); err != nil {
			return nil, err
		}
	}

	if createValidation != nil {
		if err := createValidation(ctx, obj); err != nil {
			return nil, err
		}
	}
	if s.strategy != nil {
		if err := s.strategy.ValidateCreate(ctx, obj); err != nil {
			return nil, err
		}
	}

	typed, ok := obj.(T)
	if !ok {
		return nil, apierrors.NewBadRequest("invalid object type")
	}

	return s.backend.Upsert(ctx, typed)
}

//nolint:gocyclo,gocognit // K8s REST Update requires validation, fetch-merge, and save in one method
func (s *Storage[T, TList]) Update(
	ctx context.Context,
	name string,
	objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc,
	updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	opts *metav1.UpdateOptions,
) (runtime.Object, bool, error) {
	ctx, cancel := s.opts.WithTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}

	var oldObj runtime.Object
	existing, err := s.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) || !forceAllowCreate {
			return nil, false, err
		}
		oldObj = s.newT()
	} else {
		oldObj = existing
	}

	newObj, err := objInfo.UpdatedObject(ctx, oldObj)
	if err != nil {
		return nil, false, err
	}

	accessor, err := meta.Accessor(newObj)
	if err != nil {
		return nil, false, apierrors.NewInternalError(err)
	}

	// Preserve immutable fields from the existing object (standard K8s PrepareForUpdate pattern).
	if existing != nil {
		oldAccessor, err := meta.Accessor(existing)
		if err == nil {
			if accessor.GetUID() == "" {
				accessor.SetUID(oldAccessor.GetUID())
			}
			ct := accessor.GetCreationTimestamp()
			if ct.IsZero() {
				accessor.SetCreationTimestamp(oldAccessor.GetCreationTimestamp())
			}
		}
	}

	if accessor.GetName() == "" {
		accessor.SetName(name)
	}
	if accessor.GetName() != name {
		return nil, false, apierrors.NewBadRequest("metadata.name does not match request name")
	}
	if err := base.RequireName(name); err != nil {
		return nil, false, err
	}

	if s.backend.NamespaceScoped() {
		if _, err := base.NormalizeObjectNamespace(ctx, accessor); err != nil {
			return nil, false, err
		}
	} else {
		if err := base.RequireClusterScopeObject(ctx, accessor); err != nil {
			return nil, false, err
		}
	}

	if oldObj != nil && updateValidation != nil {
		if err := updateValidation(ctx, newObj, oldObj); err != nil {
			return nil, false, err
		}
	}
	if s.strategy != nil {
		if err := s.strategy.ValidateUpdate(ctx, newObj, oldObj); err != nil {
			return nil, false, err
		}
	}

	typed, ok := newObj.(T)
	if !ok {
		return nil, false, apierrors.NewBadRequest("invalid object type")
	}
	result, err := s.backend.Upsert(ctx, typed)
	if err != nil {
		return nil, false, err
	}

	return result, existing == nil, nil
}

// Delete deletes a single resource by name.
func (s *Storage[T, TList]) Delete(
	ctx context.Context,
	name string,
	deleteValidation rest.ValidateObjectFunc,
	opts *metav1.DeleteOptions,
) (runtime.Object, bool, error) {
	if err := base.RequireName(name); err != nil {
		return nil, false, err
	}
	ctx, cancel := s.opts.WithTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}

	var namespace string
	if s.backend.NamespaceScoped() {
		ns, err := base.RequireNamespace(ctx)
		if err != nil {
			return nil, false, err
		}
		namespace = ns
	} else {
		if err := base.RequireClusterScope(ctx); err != nil {
			return nil, false, err
		}
	}

	if deleteValidation != nil {
		if err := deleteValidation(ctx, s.newT()); err != nil {
			return nil, false, err
		}
	}

	if err := s.backend.Delete(ctx, name, namespace); err != nil {
		return nil, false, err
	}

	return &metav1.Status{Status: metav1.StatusSuccess}, true, nil
}

// DeleteCollection deletes all resources matching the list options.
func (s *Storage[T, TList]) DeleteCollection(
	ctx context.Context,
	deleteValidation rest.ValidateObjectFunc,
	opts *metav1.DeleteOptions,
	listOptions *metainternalversion.ListOptions,
) (runtime.Object, error) {
	ctx, cancel := s.opts.WithTimeout(ctx)
	if cancel != nil {
		defer cancel()
	}
	listObj, err := s.List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	items, err := meta.ExtractList(listObj)
	if err != nil {
		return nil, apierrors.NewInternalError(err)
	}
	for _, item := range items {
		accessor, err := meta.Accessor(item)
		if err != nil {
			return nil, apierrors.NewInternalError(err)
		}
		if _, _, err := s.Delete(ctx, accessor.GetName(), deleteValidation, opts); err != nil {
			return nil, err
		}
	}

	return &metav1.Status{Status: metav1.StatusSuccess}, nil
}

// GetSingularName returns the singular name of the resource (required by k8s 1.35+).
func (s *Storage[T, TList]) GetSingularName() string {
	return s.singularName
}

// Destroy cleans up resources. No-op for gRPC-backed storage.
func (s *Storage[T, TList]) Destroy() {}

// Compile-time interface checks.
var _ rest.Storage = (*Storage[runtime.Object, runtime.Object])(nil)
var _ rest.Getter = (*Storage[runtime.Object, runtime.Object])(nil)
var _ rest.Lister = (*Storage[runtime.Object, runtime.Object])(nil)
var _ rest.Watcher = (*Storage[runtime.Object, runtime.Object])(nil)
var _ rest.Creater = (*Storage[runtime.Object, runtime.Object])(nil) //nolint:misspell // K8s API naming convention
var _ rest.Updater = (*Storage[runtime.Object, runtime.Object])(nil)
var _ rest.GracefulDeleter = (*Storage[runtime.Object, runtime.Object])(nil)
var _ rest.CollectionDeleter = (*Storage[runtime.Object, runtime.Object])(nil)
var _ rest.Scoper = (*Storage[runtime.Object, runtime.Object])(nil)
var _ rest.SingularNameProvider = (*Storage[runtime.Object, runtime.Object])(nil)
