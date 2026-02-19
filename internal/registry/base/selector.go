package base

import (
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// LabelsToMap converts a label selector string into an equality-only map.
func LabelsToMap(selector string) (map[string]string, error) {
	if strings.TrimSpace(selector) == "" {
		return nil, nil //nolint:nilnil // empty selector = nil map, nil error is the valid contract
	}
	sel, err := labels.Parse(selector)
	if err != nil {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("invalid label selector: %v", err))
	}
	reqs, selectable := sel.Requirements()
	if !selectable {
		return nil, apierrors.NewBadRequest("unsupported label selector")
	}

	out := make(map[string]string, len(reqs))
	for _, req := range reqs {
		switch req.Operator() {
		case selection.Equals, selection.In:
			values := req.Values().List()
			if len(values) != 1 {
				return nil, apierrors.NewBadRequest("label selector must have exactly one value per key")
			}
			out[req.Key()] = values[0]
		default:
			return nil, apierrors.NewBadRequest(fmt.Sprintf("label selector operator %q is not supported", req.Operator()))
		}
	}

	return out, nil
}

// FieldsToNameNamespace converts field selectors into name/namespace.
// Only metadata.name and metadata.namespace are supported.
func FieldsToNameNamespace(selector string) (string, string, error) {
	if strings.TrimSpace(selector) == "" {
		return "", "", nil
	}
	sel, err := fields.ParseSelector(selector)
	if err != nil {
		return "", "", apierrors.NewBadRequest(fmt.Sprintf("invalid field selector: %v", err))
	}

	name, hasName := sel.RequiresExactMatch("metadata.name")
	namespace, hasNamespace := sel.RequiresExactMatch("metadata.namespace")

	var allowed fields.Selector
	switch {
	case hasName && hasNamespace:
		allowed = fields.AndSelectors(
			fields.OneTermEqualSelector("metadata.name", name),
			fields.OneTermEqualSelector("metadata.namespace", namespace),
		)
	case hasName:
		allowed = fields.OneTermEqualSelector("metadata.name", name)
	case hasNamespace:
		allowed = fields.OneTermEqualSelector("metadata.namespace", namespace)
	default:
		allowed = fields.Everything()
	}

	if sel.String() != allowed.String() {
		return "", "", apierrors.NewBadRequest("field selector only supports metadata.name and metadata.namespace")
	}

	return name, namespace, nil
}
