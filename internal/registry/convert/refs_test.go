package convert

import (
	"testing"

	common "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	"github.com/stretchr/testify/require"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func TestResourceRefsFromProto(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		require.Nil(t, ResourceRefsFromProto(nil))
	})

	t.Run("empty input", func(t *testing.T) {
		require.Nil(t, ResourceRefsFromProto([]*common.ResourceRef{}))
	})

	t.Run("skips nil entries", func(t *testing.T) {
		got := ResourceRefsFromProto([]*common.ResourceRef{nil, nil})
		require.Empty(t, got)
	})

	t.Run("converts refs", func(t *testing.T) {
		refs := []*common.ResourceRef{
			{Name: "web-01", Namespace: "prod", ResType: "Host"},
			{Name: "corp-lan", Namespace: "prod", ResType: "Network"},
		}
		got := ResourceRefsFromProto(refs)
		require.Equal(t, []v1alpha1.ResourceRef{
			{Name: "web-01", Namespace: "prod", Kind: "Host"},
			{Name: "corp-lan", Namespace: "prod", Kind: "Network"},
		}, got)
	})

	t.Run("mixed nil and valid", func(t *testing.T) {
		refs := []*common.ResourceRef{
			nil,
			{Name: "svc-1", Namespace: "default", ResType: "Service"},
			nil,
		}
		got := ResourceRefsFromProto(refs)
		require.Len(t, got, 1)
		require.Equal(t, "svc-1", got[0].Name)
		require.Equal(t, "Service", got[0].Kind)
	})
}
