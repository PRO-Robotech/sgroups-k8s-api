package addressgroup

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

type addressGroupStrategy struct{}

func (s *addressGroupStrategy) ValidateCreate(_ context.Context, obj runtime.Object) error {
	ag, ok := obj.(*v1alpha1.AddressGroup)
	if !ok {
		return apierrors.NewBadRequest("invalid object type")
	}

	return validateAddressGroupSpec(&ag.Spec)
}

func (s *addressGroupStrategy) ValidateUpdate(_ context.Context, obj, _ runtime.Object) error {
	ag, ok := obj.(*v1alpha1.AddressGroup)
	if !ok {
		return apierrors.NewBadRequest("invalid object type")
	}

	return validateAddressGroupSpec(&ag.Spec)
}

func validateAddressGroupSpec(spec *v1alpha1.AddressGroupSpec) error {
	switch spec.DefaultAction {
	case v1alpha1.ActionAllow, v1alpha1.ActionDeny:
		return nil
	default:
		return apierrors.NewBadRequest(
			fmt.Sprintf("spec.defaultAction must be %q or %q",
				v1alpha1.ActionAllow, v1alpha1.ActionDeny))
	}
}
