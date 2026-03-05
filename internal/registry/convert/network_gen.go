package convert

import (
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// NetworkToProto converts Network to protobuf.
func NetworkToProto(in *v1alpha1.Network) *sgroupsv1.Network {
	if in == nil {
		return nil
	}

	return &sgroupsv1.Network{
		Metadata: objectMetaToProto(in.ObjectMeta),
		Spec: &sgroupsv1.Network_Spec{
			DisplayName: in.Spec.DisplayName,
			Comment:     in.Spec.Comment,
			Description: in.Spec.Description,
			Cidr:        in.Spec.CIDR,
		},
	}
}

// NetworkFromProto converts protobuf to Network.
func NetworkFromProto(in *sgroupsv1.Network) *v1alpha1.Network {
	if in == nil {
		return nil
	}
	out := &v1alpha1.Network{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindNetwork,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.NetworkSpec{
			DisplayName: in.GetSpec().GetDisplayName(),
			Comment:     in.GetSpec().GetComment(),
			Description: in.GetSpec().GetDescription(),
			CIDR:        in.GetSpec().GetCidr(),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}

// NetworkFromProtoExt converts list/watch protobuf payload to Network.
func NetworkFromProtoExt(in *sgroupsv1.NetworkResp_NetworkExt) *v1alpha1.Network {
	if in == nil {
		return nil
	}
	out := &v1alpha1.Network{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindNetwork,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.NetworkSpec{
			DisplayName: in.GetSpec().GetDisplayName(),
			Comment:     in.GetSpec().GetComment(),
			Description: in.GetSpec().GetDescription(),
			CIDR:        in.GetSpec().GetCidr(),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}
