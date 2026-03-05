package base

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Selection represents parsed list/watch selectors.
type Selection struct {
	Name      string
	Namespace string
	Labels    map[string]string
}

// SelectionFromListOptions parses and validates list options into a Selection.
func SelectionFromListOptions(ctx context.Context, options *metav1.ListOptions, namespaced bool) (Selection, error) {
	return selectionFromOptions(ctx, options, namespaced, ValidateListOptions)
}

// SelectionFromWatchOptions parses and validates watch options into a Selection.
func SelectionFromWatchOptions(ctx context.Context, options *metav1.ListOptions, namespaced bool) (Selection, error) {
	return selectionFromOptions(ctx, options, namespaced, ValidateWatchOptions)
}

func selectionFromOptions(
	ctx context.Context,
	options *metav1.ListOptions,
	namespaced bool,
	validate func(*metav1.ListOptions) error,
) (Selection, error) {
	if err := validate(options); err != nil {
		return Selection{}, err
	}

	name, fieldNamespace, err := FieldsToNameNamespace(options.FieldSelector)
	if err != nil {
		return Selection{}, err
	}

	labelsMap, err := LabelsToMap(options.LabelSelector)
	if err != nil {
		return Selection{}, err
	}

	if !namespaced {
		if err := RequireClusterScope(ctx); err != nil {
			return Selection{}, err
		}
		if fieldNamespace != "" {
			return Selection{}, errors.NewBadRequest("field selector metadata.namespace is not supported for cluster-scoped resources")
		}

		return Selection{Name: name, Labels: labelsMap}, nil
	}

	namespace, err := ResolveNamespace(ctx, fieldNamespace)
	if err != nil {
		return Selection{}, err
	}

	return Selection{Name: name, Namespace: namespace, Labels: labelsMap}, nil
}
