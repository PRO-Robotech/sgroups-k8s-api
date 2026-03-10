package convert

import (
	commonpb "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// HostBindingToProto converts HostBinding to protobuf.
func HostBindingToProto(in *v1alpha1.HostBinding) *sgroupsv1.HostBinding {
	if in == nil {
		return nil
	}

	return &sgroupsv1.HostBinding{
		Metadata: objectMetaToProto(in.ObjectMeta),
		Spec: &sgroupsv1.HostBinding_Spec{
			DisplayName:  in.Spec.DisplayName,
			Comment:      in.Spec.Comment,
			Description:  in.Spec.Description,
			AddressGroup: resourceIdentifierToProto(in.Spec.AddressGroup),
			Host:         resourceIdentifierToProto(in.Spec.Host),
		},
	}
}

// HostBindingFromProto converts protobuf to HostBinding.
func HostBindingFromProto(in *sgroupsv1.HostBinding) *v1alpha1.HostBinding {
	if in == nil {
		return nil
	}
	out := &v1alpha1.HostBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindHostBinding,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.HostBindingSpec{
			DisplayName:  in.GetSpec().GetDisplayName(),
			Comment:      in.GetSpec().GetComment(),
			Description:  in.GetSpec().GetDescription(),
			AddressGroup: resourceIdentifierFromProto(in.GetSpec().GetAddressGroup()),
			Host:         resourceIdentifierFromProto(in.GetSpec().GetHost()),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}

func resourceIdentifierToProto(in v1alpha1.ResourceIdentifier) *commonpb.ResourceIdentifier {
	return &commonpb.ResourceIdentifier{
		Name:      in.Name,
		Namespace: in.Namespace,
	}
}

func resourceIdentifierFromProto(in *commonpb.ResourceIdentifier) v1alpha1.ResourceIdentifier {
	if in == nil {
		return v1alpha1.ResourceIdentifier{}
	}

	return v1alpha1.ResourceIdentifier{
		Name:      in.GetName(),
		Namespace: in.GetNamespace(),
	}
}
