package convert

import (
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func TenantToProto(in *v1alpha1.Tenant) *sgroupsv1.Namespace {
	if in == nil {
		return nil
	}

	return &sgroupsv1.Namespace{
		Metadata: objectMetaToProto(in.ObjectMeta),
		Spec: &sgroupsv1.Namespace_Spec{
			DisplayName: in.Spec.DisplayName,
			Comment:     in.Spec.Comment,
			Description: in.Spec.Description,
		},
	}
}

func TenantFromProto(in *sgroupsv1.Namespace) *v1alpha1.Tenant {
	if in == nil {
		return nil
	}
	out := &v1alpha1.Tenant{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindTenant,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.TenantSpec{
			DisplayName: in.GetSpec().GetDisplayName(),
			Comment:     in.GetSpec().GetComment(),
			Description: in.GetSpec().GetDescription(),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}
