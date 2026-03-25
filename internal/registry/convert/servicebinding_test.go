package convert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func TestServiceBindingConversion(t *testing.T) {
	tests := []struct {
		name string
		in   *v1alpha1.ServiceBinding
	}{
		{
			name: "minimal",
			in: &v1alpha1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "sb-1", Namespace: "default"},
			},
		},
		{
			name: "full spec",
			in: &v1alpha1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "sb-full", Namespace: "prod"},
				Spec: v1alpha1.ServiceBindingSpec{
					DisplayName: "SB Full",
					Comment:     "comment",
					Description: "description",
					AddressGroup: v1alpha1.ResourceIdentifier{
						Name:      "web-servers",
						Namespace: "prod",
					},
					Service: v1alpha1.ResourceIdentifier{
						Name:      "web-service",
						Namespace: "prod",
					},
				},
			},
		},
		{
			name: "with labels annotations uid rv ts",
			in: &v1alpha1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "sb-meta",
					Namespace:         "staging",
					UID:               types.UID("uid-sb-1"),
					ResourceVersion:   "55",
					CreationTimestamp: metav1.NewTime(time.Date(2025, 4, 11, 8, 30, 0, 0, time.UTC)),
					Labels:            map[string]string{"env": "staging"},
					Annotations:       map[string]string{"note": "x"},
				},
				Spec: v1alpha1.ServiceBindingSpec{
					DisplayName: "Meta binding",
					AddressGroup: v1alpha1.ResourceIdentifier{
						Name:      "ag-1",
						Namespace: "staging",
					},
					Service: v1alpha1.ResourceIdentifier{
						Name:      "svc-1",
						Namespace: "staging",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := ServiceBindingToProto(tt.in)
			require.NotNil(t, pb)

			got := ServiceBindingFromProto(pb)
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
			require.Equal(t, tt.in.Spec.AddressGroup.Name, got.Spec.AddressGroup.Name)
			require.Equal(t, tt.in.Spec.AddressGroup.Namespace, got.Spec.AddressGroup.Namespace)
			require.Equal(t, tt.in.Spec.Service.Name, got.Spec.Service.Name)
			require.Equal(t, tt.in.Spec.Service.Namespace, got.Spec.Service.Namespace)

			require.Equal(t, v1alpha1.KindServiceBinding, got.Kind)
			require.Equal(t, v1alpha1.SchemeGroupVersion.String(), got.APIVersion)
		})
	}
}

func TestServiceBindingNilSafety(t *testing.T) {
	require.Nil(t, ServiceBindingToProto(nil))
	require.Nil(t, ServiceBindingFromProto(nil))
}
