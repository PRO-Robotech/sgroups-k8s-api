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

func TestAddressGroupConversion(t *testing.T) {
	tests := []struct {
		name       string
		in         *v1alpha1.AddressGroup
		wantAction v1alpha1.Action // if set, overrides in.Spec.DefaultAction for assertion
	}{
		{
			name: "minimal (empty action normalizes to Unknown)",
			in: &v1alpha1.AddressGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "ag-1", Namespace: "default"},
			},
			wantAction: v1alpha1.ActionUnknown,
		},
		{
			name: "full spec with Allow action",
			in: &v1alpha1.AddressGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "ag-full", Namespace: "prod"},
				Spec: v1alpha1.AddressGroupSpec{
					DisplayName:   "Full AG",
					Comment:       "test comment",
					Description:   "test description",
					DefaultAction: v1alpha1.ActionAllow,
					Logs:          true,
					Trace:         true,
				},
			},
		},
		{
			name: "Deny action",
			in: &v1alpha1.AddressGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "ag-deny", Namespace: "default"},
				Spec: v1alpha1.AddressGroupSpec{
					DefaultAction: v1alpha1.ActionDeny,
				},
			},
		},
		{
			name: "Unknown action",
			in: &v1alpha1.AddressGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "ag-unknown", Namespace: "default"},
				Spec: v1alpha1.AddressGroupSpec{
					DefaultAction: v1alpha1.ActionUnknown,
				},
			},
		},
		{
			name: "with labels, annotations, UID, RV, CreationTimestamp",
			in: &v1alpha1.AddressGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "ag-meta",
					Namespace:         "staging",
					UID:               types.UID("uid-456"),
					ResourceVersion:   "99",
					CreationTimestamp: metav1.NewTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
					Labels:            map[string]string{"app": "web"},
					Annotations:       map[string]string{"desc": "test"},
				},
				Spec: v1alpha1.AddressGroupSpec{
					DisplayName:   "Meta AG",
					DefaultAction: v1alpha1.ActionAllow,
					Logs:          true,
				},
			},
		},
		{
			name: "Logs false, Trace true",
			in: &v1alpha1.AddressGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "ag-trace", Namespace: "default"},
				Spec: v1alpha1.AddressGroupSpec{
					DefaultAction: v1alpha1.ActionDeny,
					Logs:          false,
					Trace:         true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto := AddressGroupToProto(tt.in)
			require.NotNil(t, proto)

			got := AddressGroupFromProto(proto)
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
			expectedAction := tt.in.Spec.DefaultAction
			if tt.wantAction != "" {
				expectedAction = tt.wantAction
			}
			require.Equal(t, expectedAction, got.Spec.DefaultAction)
			require.Equal(t, tt.in.Spec.Logs, got.Spec.Logs)
			require.Equal(t, tt.in.Spec.Trace, got.Spec.Trace)

			// TypeMeta
			require.Equal(t, "AddressGroup", got.Kind)
			require.Equal(t, v1alpha1.SchemeGroupVersion.String(), got.APIVersion)
		})
	}
}

func TestAddressGroupFromProtoExt(t *testing.T) {
	ts := time.Date(2025, 3, 10, 8, 0, 0, 0, time.UTC)

	ext := &sgroupsv1.AddressGroupResp_List_AddressGroupExt{
		Metadata: &commonpb.Metadata{
			Uid:               "ext-uid",
			Name:              "ag-ext",
			Namespace:         "default",
			ResourceVersion:   "7",
			CreationTimestamp: timestamppb.New(ts),
			Labels:            map[string]string{"env": "test"},
		},
		Spec: &sgroupsv1.AddressGroup_Spec{
			DisplayName:   "Ext AG",
			Comment:       "ext comment",
			Description:   "ext desc",
			DefaultAction: commonpb.Action_DENY,
			Logs:          true,
			Trace:         false,
		},
	}

	got := AddressGroupFromProtoExt(ext)
	require.NotNil(t, got)

	// The Ext variant should produce the same result as FromProto
	// when given the same Metadata+Spec.
	ag := &sgroupsv1.AddressGroup{
		Metadata: ext.Metadata,
		Spec:     ext.Spec,
	}
	fromProto := AddressGroupFromProto(ag)
	require.NotNil(t, fromProto)

	require.Equal(t, fromProto.Name, got.Name)
	require.Equal(t, fromProto.Namespace, got.Namespace)
	require.Equal(t, fromProto.UID, got.UID)
	require.Equal(t, fromProto.ResourceVersion, got.ResourceVersion)
	require.Equal(t, fromProto.Labels, got.Labels)
	require.Equal(t, fromProto.Spec, got.Spec)
	require.True(t, fromProto.CreationTimestamp.Time.Equal(got.CreationTimestamp.Time))
}

func TestAddressGroupNilSafety(t *testing.T) {
	require.Nil(t, AddressGroupToProto(nil))
	require.Nil(t, AddressGroupFromProto(nil))
	require.Nil(t, AddressGroupFromProtoExt(nil))
}

func TestActionConversion(t *testing.T) {
	tests := []struct {
		name   string
		action v1alpha1.Action
		proto  commonpb.Action
	}{
		{"Allow", v1alpha1.ActionAllow, commonpb.Action_ALLOW},
		{"Deny", v1alpha1.ActionDeny, commonpb.Action_DENY},
		{"Unknown", v1alpha1.ActionUnknown, commonpb.Action_UNKNOWN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := actionToProto(tt.action)
			require.Equal(t, tt.proto, got)

			back := actionFromProto(got)
			require.Equal(t, tt.action, back)
		})
	}
}

func TestObjectMetaConversion(t *testing.T) {
	ts := time.Date(2025, 7, 20, 15, 30, 0, 0, time.UTC)

	meta := metav1.ObjectMeta{
		Name:              "test-obj",
		Namespace:         "my-ns",
		UID:               types.UID("uid-789"),
		ResourceVersion:   "100",
		CreationTimestamp: metav1.NewTime(ts),
		Labels:            map[string]string{"a": "1", "b": "2"},
		Annotations:       map[string]string{"x": "y"},
	}

	proto := objectMetaToProto(meta)
	require.NotNil(t, proto)
	require.Equal(t, "test-obj", proto.Name)
	require.Equal(t, "my-ns", proto.Namespace)
	require.Equal(t, "uid-789", proto.Uid)
	require.Equal(t, "100", proto.ResourceVersion)
	require.Equal(t, map[string]string{"a": "1", "b": "2"}, proto.Labels)
	require.Equal(t, map[string]string{"x": "y"}, proto.Annotations)
	require.True(t, ts.Equal(proto.CreationTimestamp.AsTime()))

	var out metav1.ObjectMeta
	objectMetaFromProto(&out, proto)
	require.Equal(t, meta.Name, out.Name)
	require.Equal(t, meta.Namespace, out.Namespace)
	require.Equal(t, meta.UID, out.UID)
	require.Equal(t, meta.ResourceVersion, out.ResourceVersion)
	require.Equal(t, meta.Labels, out.Labels)
	require.Equal(t, meta.Annotations, out.Annotations)
	require.True(t, ts.Equal(out.CreationTimestamp.Time))
}

func TestObjectMetaNilSafety(t *testing.T) {
	// nil Metadata should not panic
	var out metav1.ObjectMeta
	objectMetaFromProto(&out, nil)
	require.Empty(t, out.Name)

	// nil out should not panic
	objectMetaFromProto(nil, &commonpb.Metadata{Name: "x"})
}
