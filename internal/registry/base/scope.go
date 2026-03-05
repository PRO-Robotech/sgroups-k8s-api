package base

import (
	"context"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
)

// RequireClusterScope ensures the request namespace is empty for cluster-scoped resources.
func RequireClusterScope(ctx context.Context) error {
	if ns, ok := apirequest.NamespaceFrom(ctx); ok && ns != metav1.NamespaceNone {
		return apierrors.NewBadRequest("namespace must be empty for cluster-scoped resource")
	}

	return nil
}

// RequireNamespace ensures the request namespace is present and is a valid RFC 1123 DNS label.
func RequireNamespace(ctx context.Context) (string, error) {
	ns, _ := apirequest.NamespaceFrom(ctx)
	if ns == metav1.NamespaceNone {
		return "", apierrors.NewBadRequest("namespace is required")
	}
	if err := validateDNS1123Label("namespace", ns); err != nil {
		return "", err
	}

	return ns, nil
}

// ResolveNamespace returns the effective namespace for list/watch requests.
// If the request is namespaced, it must match the field selector namespace.
func ResolveNamespace(ctx context.Context, fieldNamespace string) (string, error) {
	requestNamespace, _ := apirequest.NamespaceFrom(ctx)
	if requestNamespace != metav1.NamespaceAll {
		if fieldNamespace != "" && fieldNamespace != requestNamespace {
			return "", apierrors.NewBadRequest("field selector namespace does not match request namespace")
		}

		return requestNamespace, nil
	}

	return fieldNamespace, nil
}

// RequireName ensures name is present and is a valid RFC 1123 DNS label.
func RequireName(name string) error {
	if name == "" {
		return apierrors.NewBadRequest("metadata.name is required")
	}

	return validateDNS1123Label("metadata.name", name)
}

// validateDNS1123Label validates that value is a valid RFC 1123 DNS label.
func validateDNS1123Label(field, value string) error {
	if errs := validation.IsDNS1123Label(value); len(errs) > 0 {
		return apierrors.NewBadRequest(field + " is invalid: " + strings.Join(errs, "; "))
	}

	return nil
}

// RequireClusterScopeObject ensures both request context and object metadata have no namespace.
func RequireClusterScopeObject(ctx context.Context, obj metav1.Object) error {
	if err := RequireClusterScope(ctx); err != nil {
		return err
	}
	if obj.GetNamespace() != "" {
		return apierrors.NewBadRequest("metadata.namespace must be empty for cluster-scoped resource")
	}

	return nil
}

// NormalizeObjectNamespace enforces request namespace on the object metadata.
func NormalizeObjectNamespace(ctx context.Context, obj metav1.Object) (string, error) {
	ns, err := RequireNamespace(ctx)
	if err != nil {
		return "", err
	}
	if obj.GetNamespace() == "" {
		obj.SetNamespace(ns)
	}
	if obj.GetNamespace() != ns {
		return "", apierrors.NewBadRequest("metadata.namespace does not match request namespace")
	}

	return ns, nil
}
