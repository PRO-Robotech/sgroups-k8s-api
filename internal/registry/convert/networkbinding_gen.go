package convert

import (
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// NetworkBindingToProto converts NetworkBinding to protobuf.
func NetworkBindingToProto(in *v1alpha1.NetworkBinding) *sgroupsv1.NetworkBinding {
	if in == nil {
		return nil
	}

	return &sgroupsv1.NetworkBinding{
		Metadata: objectMetaToProto(in.ObjectMeta),
		Spec: &sgroupsv1.NetworkBinding_Spec{
			DisplayName:  in.Spec.DisplayName,
			Comment:      in.Spec.Comment,
			Description:  in.Spec.Description,
			AddressGroup: resourceIdentifierToProto(in.Spec.AddressGroup),
			Network:      resourceIdentifierToProto(in.Spec.Network),
		},
	}
}

// NetworkBindingFromProto converts protobuf to NetworkBinding.
func NetworkBindingFromProto(in *sgroupsv1.NetworkBinding) *v1alpha1.NetworkBinding {
	if in == nil {
		return nil
	}
	out := &v1alpha1.NetworkBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindNetworkBinding,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.NetworkBindingSpec{
			DisplayName:  in.GetSpec().GetDisplayName(),
			Comment:      in.GetSpec().GetComment(),
			Description:  in.GetSpec().GetDescription(),
			AddressGroup: resourceIdentifierFromProto(in.GetSpec().GetAddressGroup()),
			Network:      resourceIdentifierFromProto(in.GetSpec().GetNetwork()),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}
