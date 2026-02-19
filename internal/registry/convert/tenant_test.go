package convert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func TestTenantConversion(t *testing.T) {
	tests := []struct {
		name string
		in   *v1alpha1.Tenant
	}{
		{
			name: "minimal",
			in: &v1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{Name: "ns-1"},
			},
		},
		{
			name: "full spec",
			in: &v1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{Name: "ns-full"},
				Spec: v1alpha1.TenantSpec{
					DisplayName: "Full Tenant",
					Comment:     "test comment",
					Description: "test description",
				},
			},
		},
		{
			name: "with labels and annotations",
			in: &v1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "ns-labels",
					Labels:      map[string]string{"env": "prod", "team": "platform"},
					Annotations: map[string]string{"note": "important"},
				},
			},
		},
		{
			name: "with UID, ResourceVersion, and CreationTimestamp",
			in: &v1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "ns-meta",
					UID:               types.UID("abc-123"),
					ResourceVersion:   "42",
					CreationTimestamp: metav1.NewTime(time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)),
				},
				Spec: v1alpha1.TenantSpec{DisplayName: "Meta NS"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto := TenantToProto(tt.in)
			require.NotNil(t, proto)

			got := TenantFromProto(proto)
			require.NotNil(t, got)

			// ObjectMeta fields
			require.Equal(t, tt.in.Name, got.Name)
			require.Equal(t, tt.in.Namespace, got.Namespace)
			require.Equal(t, tt.in.UID, got.UID)
			require.Equal(t, tt.in.ResourceVersion, got.ResourceVersion)
			require.Equal(t, tt.in.Labels, got.Labels)
			require.Equal(t, tt.in.Annotations, got.Annotations)
			if !tt.in.CreationTimestamp.IsZero() {
				require.True(t, tt.in.CreationTimestamp.Time.Equal(got.CreationTimestamp.Time),
					"CreationTimestamp mismatch: want %v, got %v", tt.in.CreationTimestamp.Time, got.CreationTimestamp.Time)
			}

			// Spec fields
			require.Equal(t, tt.in.Spec.DisplayName, got.Spec.DisplayName)
			require.Equal(t, tt.in.Spec.Comment, got.Spec.Comment)
			require.Equal(t, tt.in.Spec.Description, got.Spec.Description)

			// TypeMeta is set by FromProto
			require.Equal(t, "Tenant", got.Kind)
			require.Equal(t, v1alpha1.SchemeGroupVersion.String(), got.APIVersion)
		})
	}
}

func TestTenantNilSafety(t *testing.T) {
	require.Nil(t, TenantToProto(nil))
	require.Nil(t, TenantFromProto(nil))
}
