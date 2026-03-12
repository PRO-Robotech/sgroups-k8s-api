package convert

import (
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// HostToProto converts Host to protobuf.
func HostToProto(in *v1alpha1.Host) *sgroupsv1.Host {
	if in == nil {
		return nil
	}

	return &sgroupsv1.Host{
		Metadata: objectMetaToProto(in.ObjectMeta),
		Spec: &sgroupsv1.Host_Spec{
			DisplayName: in.Spec.DisplayName,
			Comment:     in.Spec.Comment,
			Description: in.Spec.Description,
		},
	}
}

// HostFromProto converts protobuf to Host.
func HostFromProto(in *sgroupsv1.Host) *v1alpha1.Host {
	if in == nil {
		return nil
	}
	out := &v1alpha1.Host{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindHost,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.HostSpec{
			DisplayName: in.GetSpec().GetDisplayName(),
			Comment:     in.GetSpec().GetComment(),
			Description: in.GetSpec().GetDescription(),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}

// HostFromProtoExt converts list/watch protobuf payload to Host.
func HostFromProtoExt(in *sgroupsv1.HostResp_HostExt) *v1alpha1.Host {
	if in == nil {
		return nil
	}
	out := &v1alpha1.Host{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindHost,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.HostSpec{
			DisplayName: in.GetSpec().GetDisplayName(),
			Comment:     in.GetSpec().GetComment(),
			Description: in.GetSpec().GetDescription(),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}
