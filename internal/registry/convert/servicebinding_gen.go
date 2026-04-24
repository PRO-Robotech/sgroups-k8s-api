//nolint:dupl // parallel scaffold for NetworkBinding and ServiceBinding intentionally mirrors structure
package convert

import (
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// ServiceBindingToProto converts ServiceBinding to protobuf.
func ServiceBindingToProto(in *v1alpha1.ServiceBinding) *sgroupsv1.ServiceBinding {
	if in == nil {
		return nil
	}

	return &sgroupsv1.ServiceBinding{
		Metadata: objectMetaToProto(in.ObjectMeta),
		Spec: &sgroupsv1.ServiceBinding_Spec{
			DisplayName:  in.Spec.DisplayName,
			Comment:      in.Spec.Comment,
			Description:  in.Spec.Description,
			AddressGroup: resourceIdentifierToProto(in.Spec.AddressGroup),
			Service:      resourceIdentifierToProto(in.Spec.Service),
		},
	}
}

// ServiceBindingFromProto converts protobuf to ServiceBinding.
func ServiceBindingFromProto(in *sgroupsv1.ServiceBinding) *v1alpha1.ServiceBinding {
	if in == nil {
		return nil
	}
	out := &v1alpha1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindServiceBinding,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.ServiceBindingSpec{
			DisplayName:  in.GetSpec().GetDisplayName(),
			Comment:      in.GetSpec().GetComment(),
			Description:  in.GetSpec().GetDescription(),
			AddressGroup: resourceIdentifierFromProto(in.GetSpec().GetAddressGroup()),
			Service:      resourceIdentifierFromProto(in.GetSpec().GetService()),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}
