package rule

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

type ruleStrategy struct{}

func (s *ruleStrategy) ValidateCreate(_ context.Context, obj runtime.Object) error {
	r, ok := obj.(*v1alpha1.Rule)
	if !ok {
		return apierrors.NewBadRequest("invalid object type")
	}

	return validateRuleSpec(&r.Spec)
}

func (s *ruleStrategy) ValidateUpdate(_ context.Context, obj, _ runtime.Object) error {
	r, ok := obj.(*v1alpha1.Rule)
	if !ok {
		return apierrors.NewBadRequest("invalid object type")
	}

	return validateRuleSpec(&r.Spec)
}

func validateRuleSpec(spec *v1alpha1.RuleSpec) error {
	if err := spec.Action.Validate(); err != nil {
		return apierrors.NewBadRequest(fmt.Sprintf("spec.action: %v", err))
	}
	if spec.Session != nil {
		if err := spec.Session.Traffic.Validate(); err != nil {
			return apierrors.NewBadRequest(fmt.Sprintf("spec.session.traffic: %v", err))
		}
	}
	if spec.Endpoints != nil {
		if spec.Endpoints.Local != nil {
			if err := spec.Endpoints.Local.Type.Validate(); err != nil {
				return apierrors.NewBadRequest(fmt.Sprintf("spec.endpoints.local.type: %v", err))
			}
		}
		if spec.Endpoints.Remote != nil {
			if err := spec.Endpoints.Remote.Type.Validate(); err != nil {
				return apierrors.NewBadRequest(fmt.Sprintf("spec.endpoints.remote.type: %v", err))
			}
		}
	}
	if spec.Transport != nil {
		if err := spec.Transport.Protocol.Validate(); err != nil {
			return apierrors.NewBadRequest(fmt.Sprintf("spec.transport.protocol: %v", err))
		}
		if err := spec.Transport.IPv.Validate(); err != nil {
			return apierrors.NewBadRequest(fmt.Sprintf("spec.transport.IPv: %v", err))
		}
	}

	return nil
}
