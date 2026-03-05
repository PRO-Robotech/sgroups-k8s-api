package convert

import (
	"testing"
	"time"

	commonpb "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func TestNetworkConversion(t *testing.T) {
	tests := []struct {
		name string
		in   *v1alpha1.Network
	}{
		{
			name: "minimal",
			in: &v1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{Name: "nw-1", Namespace: "default"},
			},
		},
		{
			name: "full spec",
			in: &v1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{Name: "nw-full", Namespace: "prod"},
				Spec: v1alpha1.NetworkSpec{
					DisplayName: "Net Full",
					Comment:     "comment",
					Description: "description",
					CIDR:        "10.10.0.0/16",
				},
			},
		},
		{
			name: "with labels annotations uid rv ts",
			in: &v1alpha1.Network{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "nw-meta",
					Namespace:         "staging",
					UID:               types.UID("uid-nw-1"),
					ResourceVersion:   "23",
					CreationTimestamp: metav1.NewTime(time.Date(2025, 4, 11, 8, 30, 0, 0, time.UTC)),
					Labels:            map[string]string{"env": "staging"},
					Annotations:       map[string]string{"note": "x"},
				},
				Spec: v1alpha1.NetworkSpec{
					DisplayName: "Meta network",
					CIDR:        "192.168.0.0/24",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NetworkToProto(tt.in)
			require.NotNil(t, pb)

			got := NetworkFromProto(pb)
			require.NotNil(t, got)

			require.Equal(t, tt.in.Name, got.Name)
			require.Equal(t, tt.in.Namespace, got.Namespace)
			require.Equal(t, tt.in.UID, got.UID)
			require.Equal(t, tt.in.ResourceVersion, got.ResourceVersion)
			require.Equal(t, tt.in.Labels, got.Labels)
			require.Equal(t, tt.in.Annotations, got.Annotations)
			if !tt.in.CreationTimestamp.IsZero() {
				require.True(t, tt.in.CreationTimestamp.Time.Equal(got.CreationTimestamp.Time))
			}

			require.Equal(t, tt.in.Spec.DisplayName, got.Spec.DisplayName)
			require.Equal(t, tt.in.Spec.Comment, got.Spec.Comment)
			require.Equal(t, tt.in.Spec.Description, got.Spec.Description)
			require.Equal(t, tt.in.Spec.CIDR, got.Spec.CIDR)

			require.Equal(t, v1alpha1.KindNetwork, got.Kind)
			require.Equal(t, v1alpha1.SchemeGroupVersion.String(), got.APIVersion)
		})
	}
}

func TestNetworkFromProtoExt(t *testing.T) {
	ts := time.Date(2025, 3, 10, 8, 0, 0, 0, time.UTC)

	ext := &sgroupsv1.NetworkResp_NetworkExt{
		Metadata: &commonpb.Metadata{
			Uid:               "ext-uid",
			Name:              "nw-ext",
			Namespace:         "default",
			ResourceVersion:   "19",
			CreationTimestamp: timestamppb.New(ts),
			Labels:            map[string]string{"env": "test"},
		},
		Spec: &sgroupsv1.Network_Spec{
			DisplayName: "Ext network",
			Comment:     "ext comment",
			Description: "ext description",
			Cidr:        "10.1.0.0/16",
		},
	}

	got := NetworkFromProtoExt(ext)
	require.NotNil(t, got)

	require.Equal(t, "nw-ext", got.Name)
	require.Equal(t, "default", got.Namespace)
	require.Equal(t, types.UID("ext-uid"), got.UID)
	require.Equal(t, "19", got.ResourceVersion)
	require.Equal(t, map[string]string{"env": "test"}, got.Labels)
	require.Equal(t, "Ext network", got.Spec.DisplayName)
	require.Equal(t, "ext comment", got.Spec.Comment)
	require.Equal(t, "ext description", got.Spec.Description)
	require.Equal(t, "10.1.0.0/16", got.Spec.CIDR)
	require.True(t, ts.Equal(got.CreationTimestamp.Time))
}

func TestNetworkNilSafety(t *testing.T) {
	require.Nil(t, NetworkToProto(nil))
	require.Nil(t, NetworkFromProto(nil))
	require.Nil(t, NetworkFromProtoExt(nil))
}
