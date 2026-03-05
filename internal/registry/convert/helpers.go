package convert

import (
	"maps"

	common "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func actionToProto(action v1alpha1.Action) common.Action {
	switch action {
	case v1alpha1.ActionAllow:
		return common.Action_ALLOW
	case v1alpha1.ActionDeny:
		return common.Action_DENY
	default:
		return common.Action_UNKNOWN
	}
}

func actionFromProto(action common.Action) v1alpha1.Action {
	switch action {
	case common.Action_ALLOW:
		return v1alpha1.ActionAllow
	case common.Action_DENY:
		return v1alpha1.ActionDeny
	default:
		return v1alpha1.ActionUnknown
	}
}

func objectMetaToProto(meta metav1.ObjectMeta) *common.Metadata {
	out := &common.Metadata{
		Uid:         string(meta.UID),
		Name:        meta.Name,
		Namespace:   meta.Namespace,
		Labels:      copyStringMap(meta.Labels),
		Annotations: copyStringMap(meta.Annotations),
	}
	if !meta.CreationTimestamp.IsZero() {
		out.CreationTimestamp = timestamppb.New(meta.CreationTimestamp.Time)
	}
	if meta.ResourceVersion != "" {
		out.ResourceVersion = meta.ResourceVersion
	}

	return out
}

func objectMetaFromProto(out *metav1.ObjectMeta, meta *common.Metadata) {
	if out == nil || meta == nil {
		return
	}
	out.UID = types.UID(meta.GetUid())
	out.Name = meta.GetName()
	out.Namespace = meta.GetNamespace()
	out.Labels = copyStringMap(meta.GetLabels())
	out.Annotations = copyStringMap(meta.GetAnnotations())
	if meta.GetCreationTimestamp() != nil {
		out.CreationTimestamp = metav1.NewTime(meta.GetCreationTimestamp().AsTime())
	}
	if meta.GetResourceVersion() != "" {
		out.ResourceVersion = meta.GetResourceVersion()
	}
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	maps.Copy(out, in)

	return out
}
