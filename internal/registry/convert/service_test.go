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

func TestServiceConversion(t *testing.T) {
	tests := []struct {
		name string
		in   *v1alpha1.Service
	}{
		{
			name: "minimal",
			in: &v1alpha1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "svc-1", Namespace: "default"},
			},
		},
		{
			name: "full spec with transports",
			in: &v1alpha1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "svc-full", Namespace: "prod"},
				Spec: v1alpha1.ServiceSpec{
					DisplayName: "Web Service",
					Comment:     "HTTP service",
					Description: "Production web service",
					Transports: []v1alpha1.ServiceTransport{
						{
							Protocol: v1alpha1.ProtocolTCP,
							IPv:      v1alpha1.IpAddrFamilyIPv4,
							Entries: []v1alpha1.ServiceTransportEntry{
								{
									Description: "HTTP traffic",
									Comment:     "standard HTTP",
									Ports:       "80,443",
								},
							},
						},
						{
							Protocol: v1alpha1.ProtocolUDP,
							IPv:      v1alpha1.IpAddrFamilyIPv6,
							Entries: []v1alpha1.ServiceTransportEntry{
								{
									Description: "DNS traffic",
									Ports:       "53",
								},
							},
						},
						{
							Protocol: v1alpha1.ProtocolICMP,
							Entries: []v1alpha1.ServiceTransportEntry{
								{
									Description: "Ping",
									Types:       []uint32{8, 0},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "with labels annotations uid rv ts",
			in: &v1alpha1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "svc-meta",
					Namespace:         "staging",
					UID:               types.UID("uid-svc-1"),
					ResourceVersion:   "42",
					CreationTimestamp: metav1.NewTime(time.Date(2025, 4, 11, 8, 30, 0, 0, time.UTC)),
					Labels:            map[string]string{"env": "staging"},
					Annotations:       map[string]string{"note": "x"},
				},
				Spec: v1alpha1.ServiceSpec{
					DisplayName: "Meta service",
					Transports: []v1alpha1.ServiceTransport{
						{
							Protocol: v1alpha1.ProtocolTCP,
							Entries: []v1alpha1.ServiceTransportEntry{
								{Ports: "8080"},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := ServiceToProto(tt.in)
			require.NotNil(t, pb)

			got := ServiceFromProto(pb)
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
			require.Equal(t, len(tt.in.Spec.Transports), len(got.Spec.Transports))
			for i, tr := range tt.in.Spec.Transports {
				require.Equal(t, tr.Protocol, got.Spec.Transports[i].Protocol)
				require.Equal(t, tr.IPv, got.Spec.Transports[i].IPv)
				require.Equal(t, len(tr.Entries), len(got.Spec.Transports[i].Entries))
				for j, e := range tr.Entries {
					require.Equal(t, e.Description, got.Spec.Transports[i].Entries[j].Description)
					require.Equal(t, e.Comment, got.Spec.Transports[i].Entries[j].Comment)
					require.Equal(t, e.Ports, got.Spec.Transports[i].Entries[j].Ports)
					require.Equal(t, e.Types, got.Spec.Transports[i].Entries[j].Types)
				}
			}

			require.Equal(t, v1alpha1.KindService, got.Kind)
			require.Equal(t, v1alpha1.SchemeGroupVersion.String(), got.APIVersion)
		})
	}
}

func TestServiceFromProtoExt(t *testing.T) {
	ts := time.Date(2025, 3, 10, 8, 0, 0, 0, time.UTC)

	ext := &sgroupsv1.ServiceResp_ServiceExt{
		Metadata: &commonpb.Metadata{
			Uid:               "ext-uid",
			Name:              "svc-ext",
			Namespace:         "default",
			ResourceVersion:   "19",
			CreationTimestamp: timestamppb.New(ts),
			Labels:            map[string]string{"env": "test"},
		},
		Spec: &sgroupsv1.Service_Spec{
			DisplayName: "Ext service",
			Comment:     "ext comment",
			Description: "ext description",
			Transports: []*commonpb.Transport{
				{
					Protocol: commonpb.Transport_TCP,
					Ipv:      commonpb.IpAddrFamily_IPV4,
					Entries: []*commonpb.Transport_Entry{
						{
							Description: "HTTP",
							Ports:       "80",
						},
					},
				},
			},
		},
	}

	got := ServiceFromProtoExt(ext)
	require.NotNil(t, got)

	require.Equal(t, "svc-ext", got.Name)
	require.Equal(t, "default", got.Namespace)
	require.Equal(t, types.UID("ext-uid"), got.UID)
	require.Equal(t, "19", got.ResourceVersion)
	require.Equal(t, map[string]string{"env": "test"}, got.Labels)
	require.Equal(t, "Ext service", got.Spec.DisplayName)
	require.Equal(t, "ext comment", got.Spec.Comment)
	require.Equal(t, "ext description", got.Spec.Description)
	require.Equal(t, 1, len(got.Spec.Transports))
	require.Equal(t, v1alpha1.ProtocolTCP, got.Spec.Transports[0].Protocol)
	require.Equal(t, v1alpha1.IpAddrFamilyIPv4, got.Spec.Transports[0].IPv)
	require.Equal(t, 1, len(got.Spec.Transports[0].Entries))
	require.Equal(t, "HTTP", got.Spec.Transports[0].Entries[0].Description)
	require.Equal(t, "80", got.Spec.Transports[0].Entries[0].Ports)
	require.True(t, ts.Equal(got.CreationTimestamp.Time))
}

func TestServiceNilSafety(t *testing.T) {
	require.Nil(t, ServiceToProto(nil))
	require.Nil(t, ServiceFromProto(nil))
	require.Nil(t, ServiceFromProtoExt(nil))
}
