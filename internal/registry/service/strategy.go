package service

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

type serviceStrategy struct{}

func (s *serviceStrategy) ValidateCreate(_ context.Context, obj runtime.Object) error {
	svc, ok := obj.(*v1alpha1.Service)
	if !ok {
		return apierrors.NewBadRequest("invalid object type")
	}

	return validateServiceSpec(&svc.Spec)
}

func (s *serviceStrategy) ValidateUpdate(_ context.Context, obj, _ runtime.Object) error {
	svc, ok := obj.(*v1alpha1.Service)
	if !ok {
		return apierrors.NewBadRequest("invalid object type")
	}

	return validateServiceSpec(&svc.Spec)
}

func validateServiceSpec(spec *v1alpha1.ServiceSpec) error {
	for i, t := range spec.Transports {
		if err := t.Protocol.Validate(); err != nil {
			return apierrors.NewBadRequest(fmt.Sprintf("spec.transports[%d].protocol: %v", i, err))
		}
		if err := t.IPv.Validate(); err != nil {
			return apierrors.NewBadRequest(fmt.Sprintf("spec.transports[%d].IPv: %v", i, err))
		}
	}

	return nil
}
