package convert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func TestHostBindingConversion(t *testing.T) {
	tests := []struct {
		name string
		in   *v1alpha1.HostBinding
	}{
		{
			name: "minimal",
			in: &v1alpha1.HostBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "hb-1", Namespace: "default"},
			},
		},
		{
			name: "full spec",
			in: &v1alpha1.HostBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "hb-full", Namespace: "prod"},
				Spec: v1alpha1.HostBindingSpec{
					DisplayName:  "Binding Full",
					Comment:      "comment",
					Description:  "description",
					AddressGroup: v1alpha1.ResourceIdentifier{Name: "ag-1", Namespace: "prod"},
					Host:         v1alpha1.ResourceIdentifier{Name: "host-1", Namespace: "prod"},
				},
			},
		},
		{
			name: "with labels annotations uid rv ts",
			in: &v1alpha1.HostBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "hb-meta",
					Namespace:         "staging",
					UID:               types.UID("uid-hb-1"),
					ResourceVersion:   "42",
					CreationTimestamp: metav1.NewTime(time.Date(2025, 4, 11, 8, 30, 0, 0, time.UTC)),
					Labels:            map[string]string{"env": "staging"},
					Annotations:       map[string]string{"note": "x"},
				},
				Spec: v1alpha1.HostBindingSpec{
					DisplayName:  "Meta binding",
					AddressGroup: v1alpha1.ResourceIdentifier{Name: "ag-2", Namespace: "staging"},
					Host:         v1alpha1.ResourceIdentifier{Name: "host-2", Namespace: "staging"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := HostBindingToProto(tt.in)
			require.NotNil(t, pb)

			got := HostBindingFromProto(pb)
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
			require.Equal(t, tt.in.Spec.AddressGroup, got.Spec.AddressGroup)
			require.Equal(t, tt.in.Spec.Host, got.Spec.Host)

			require.Equal(t, v1alpha1.KindHostBinding, got.Kind)
			require.Equal(t, v1alpha1.SchemeGroupVersion.String(), got.APIVersion)
		})
	}
}

func TestHostBindingNilSafety(t *testing.T) {
	require.Nil(t, HostBindingToProto(nil))
	require.Nil(t, HostBindingFromProto(nil))
}
