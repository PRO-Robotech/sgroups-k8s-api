package base

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidateListOptions rejects list/watch options we do not support.
// Note: limit and continue are silently ignored — our gRPC backend does not
// support pagination and always returns the full result set.
func ValidateListOptions(options *metav1.ListOptions) error {
	if options == nil {
		return nil
	}
	if options.ResourceVersionMatch != "" {
		return apierrors.NewBadRequest(fmt.Sprintf("resourceVersionMatch %q is not supported", options.ResourceVersionMatch))
	}

	return nil
}

// ValidateWatchOptions rejects watch-specific unsupported options.
// Unlike ValidateListOptions, it allows ResourceVersion for List+Watch resumption
// and accepts resourceVersionMatch=NotOlderThan which K8s automatically injects.
func ValidateWatchOptions(options *metav1.ListOptions) error {
	if options == nil {
		return nil
	}
	switch options.ResourceVersionMatch {
	case "", metav1.ResourceVersionMatchNotOlderThan:
		// empty and NotOlderThan are fine for watch
	default:
		return apierrors.NewBadRequest(fmt.Sprintf("resourceVersionMatch %q is not supported for watch", options.ResourceVersionMatch))
	}

	return nil
}

// ValidateGetOptions rejects unsupported get options.
func ValidateGetOptions(options *metav1.GetOptions) error {
	if options == nil {
		return nil
	}

	return nil
}
