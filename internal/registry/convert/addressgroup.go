package convert

import (
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func AddressGroupToProto(in *v1alpha1.AddressGroup) *sgroupsv1.AddressGroup {
	if in == nil {
		return nil
	}

	return &sgroupsv1.AddressGroup{
		Metadata: objectMetaToProto(in.ObjectMeta),
		Spec: &sgroupsv1.AddressGroup_Spec{
			DisplayName:   in.Spec.DisplayName,
			Comment:       in.Spec.Comment,
			Description:   in.Spec.Description,
			DefaultAction: actionToProto(in.Spec.DefaultAction),
			Logs:          in.Spec.Logs,
			Trace:         in.Spec.Trace,
		},
	}
}

//nolint:dupl // AddressGroupFromProto and AddressGroupFromProtoExt share structure but differ in input type
func AddressGroupFromProto(in *sgroupsv1.AddressGroup) *v1alpha1.AddressGroup {
	if in == nil {
		return nil
	}
	out := &v1alpha1.AddressGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindAddressGroup,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.AddressGroupSpec{
			DisplayName:   in.GetSpec().GetDisplayName(),
			Comment:       in.GetSpec().GetComment(),
			Description:   in.GetSpec().GetDescription(),
			DefaultAction: actionFromProto(in.GetSpec().GetDefaultAction()),
			Logs:          in.GetSpec().GetLogs(),
			Trace:         in.GetSpec().GetTrace(),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}

//nolint:dupl // AddressGroupFromProto and AddressGroupFromProtoExt share structure but differ in input type
func AddressGroupFromProtoExt(in *sgroupsv1.AddressGroupResp_AddressGroupExt) *v1alpha1.AddressGroup {
	if in == nil {
		return nil
	}
	out := &v1alpha1.AddressGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindAddressGroup,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.AddressGroupSpec{
			DisplayName:   in.GetSpec().GetDisplayName(),
			Comment:       in.GetSpec().GetComment(),
			Description:   in.GetSpec().GetDescription(),
			DefaultAction: actionFromProto(in.GetSpec().GetDefaultAction()),
			Logs:          in.GetSpec().GetLogs(),
			Trace:         in.GetSpec().GetTrace(),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}
