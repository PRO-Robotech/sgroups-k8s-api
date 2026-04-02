package convert

import (
	common "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// ResourceRefsFromProto converts proto ResourceRef slice to K8s ResourceRef slice.
func ResourceRefsFromProto(refs []*common.ResourceRef) []v1alpha1.ResourceRef {
	if len(refs) == 0 {
		return nil
	}
	out := make([]v1alpha1.ResourceRef, 0, len(refs))
	for _, r := range refs {
		if r == nil {
			continue
		}
		out = append(out, v1alpha1.ResourceRef{
			Name:      r.GetName(),
			Namespace: r.GetNamespace(),
			Kind:      r.GetResType(),
		})
	}

	return out
}
