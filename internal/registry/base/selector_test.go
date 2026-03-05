package base

import (
	"reflect"
	"testing"
)

func TestLabelsToMap(t *testing.T) {
	tests := []struct {
		name      string
		selector  string
		want      map[string]string
		expectErr bool
	}{
		{
			name:     "empty",
			selector: "",
			want:     nil,
		},
		{
			name:     "equals",
			selector: "env=prod,app=web",
			want: map[string]string{
				"env": "prod",
				"app": "web",
			},
		},
		{
			name:     "in single value",
			selector: "env in (prod)",
			want: map[string]string{
				"env": "prod",
			},
		},
		{
			name:      "in multiple values",
			selector:  "env in (prod,staging)",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LabelsToMap(tt.selector)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldsToNameNamespace(t *testing.T) {
	tests := []struct {
		name      string
		selector  string
		wantName  string
		wantNS    string
		expectErr bool
	}{
		{
			name:     "empty",
			selector: "",
		},
		{
			name:     "name",
			selector: "metadata.name=foo",
			wantName: "foo",
		},
		{
			name:     "name and namespace",
			selector: "metadata.name=foo,metadata.namespace=bar",
			wantName: "foo",
			wantNS:   "bar",
		},
		{
			name:      "unsupported field",
			selector:  "spec.foo=bar",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotNS, err := FieldsToNameNamespace(tt.selector)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error")
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotName != tt.wantName || gotNS != tt.wantNS {
				t.Fatalf("got name=%q ns=%q, want name=%q ns=%q", gotName, gotNS, tt.wantName, tt.wantNS)
			}
		})
	}
}
