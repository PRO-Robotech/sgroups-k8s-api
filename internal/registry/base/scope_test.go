package base

import (
	"context"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
)

func TestRequireClusterScope(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		expectErr bool
	}{
		{
			name: "no namespace in context",
			ctx:  context.Background(),
		},
		{
			name: "empty namespace",
			ctx:  apirequest.WithNamespace(context.Background(), ""),
		},
		{
			name: "NamespaceNone",
			ctx:  apirequest.WithNamespace(context.Background(), metav1.NamespaceNone),
		},
		{
			name:      "namespace present",
			ctx:       apirequest.WithNamespace(context.Background(), "default"),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireClusterScope(tt.ctx)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !apierrors.IsBadRequest(err) {
					t.Fatalf("expected BadRequest, got %v", err)
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRequireNamespace(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantNS    string
		expectErr bool
	}{
		{
			name:      "no namespace in context",
			ctx:       context.Background(),
			expectErr: true,
		},
		{
			name:      "empty namespace",
			ctx:       apirequest.WithNamespace(context.Background(), ""),
			expectErr: true,
		},
		{
			name:   "valid namespace",
			ctx:    apirequest.WithNamespace(context.Background(), "default"),
			wantNS: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, err := RequireNamespace(tt.ctx)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !apierrors.IsBadRequest(err) {
					t.Fatalf("expected BadRequest, got %v", err)
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ns != tt.wantNS {
				t.Fatalf("got namespace %q, want %q", ns, tt.wantNS)
			}
		})
	}
}

func TestResolveNamespace(t *testing.T) {
	tests := []struct {
		name           string
		ctx            context.Context
		fieldNamespace string
		wantNS         string
		expectErr      bool
	}{
		{
			name:           "no request namespace uses field namespace",
			ctx:            context.Background(),
			fieldNamespace: "field-ns",
			wantNS:         "field-ns",
		},
		{
			name:           "request namespace used when present",
			ctx:            apirequest.WithNamespace(context.Background(), "request-ns"),
			fieldNamespace: "",
			wantNS:         "request-ns",
		},
		{
			name:           "request namespace overrides empty field",
			ctx:            apirequest.WithNamespace(context.Background(), "request-ns"),
			fieldNamespace: "request-ns",
			wantNS:         "request-ns",
		},
		{
			name:           "field namespace mismatch rejected",
			ctx:            apirequest.WithNamespace(context.Background(), "request-ns"),
			fieldNamespace: "other-ns",
			expectErr:      true,
		},
		{
			name:           "NamespaceAll allows any field namespace",
			ctx:            apirequest.WithNamespace(context.Background(), metav1.NamespaceAll),
			fieldNamespace: "any-ns",
			wantNS:         "any-ns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, err := ResolveNamespace(tt.ctx, tt.fieldNamespace)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ns != tt.wantNS {
				t.Fatalf("got namespace %q, want %q", ns, tt.wantNS)
			}
		})
	}
}

type fakeObject struct {
	metav1.ObjectMeta
}

func TestNormalizeObjectNamespace(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		objNS     string
		wantNS    string
		expectErr bool
	}{
		{
			name:      "no namespace in context",
			ctx:       context.Background(),
			objNS:     "",
			expectErr: true,
		},
		{
			name:   "empty object namespace gets set",
			ctx:    apirequest.WithNamespace(context.Background(), "default"),
			objNS:  "",
			wantNS: "default",
		},
		{
			name:   "matching namespace passes",
			ctx:    apirequest.WithNamespace(context.Background(), "default"),
			objNS:  "default",
			wantNS: "default",
		},
		{
			name:      "mismatched namespace rejected",
			ctx:       apirequest.WithNamespace(context.Background(), "default"),
			objNS:     "other",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &fakeObject{}
			obj.Namespace = tt.objNS
			ns, err := NormalizeObjectNamespace(tt.ctx, obj)
			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ns != tt.wantNS {
				t.Fatalf("got namespace %q, want %q", ns, tt.wantNS)
			}
			if obj.Namespace != tt.wantNS {
				t.Fatalf("object namespace not set: got %q, want %q", obj.Namespace, tt.wantNS)
			}
		})
	}
}
