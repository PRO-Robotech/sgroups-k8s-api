package addressgroup

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func TestValidateCreate(t *testing.T) {
	s := &addressGroupStrategy{}
	ctx := context.Background()

	tests := []struct {
		name    string
		action  v1alpha1.Action
		wantErr bool
	}{
		{"allow", v1alpha1.ActionAllow, false},
		{"deny", v1alpha1.ActionDeny, false},
		{"empty", "", true},
		{"unknown", v1alpha1.ActionUnknown, true},
		{"invalid string", v1alpha1.Action("INVALID"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &v1alpha1.AddressGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec:       v1alpha1.AddressGroupSpec{DefaultAction: tt.action},
			}
			err := s.ValidateCreate(ctx, obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreate(%q) error = %v, wantErr %v", tt.action, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCreate_WrongType(t *testing.T) {
	s := &addressGroupStrategy{}
	err := s.ValidateCreate(context.Background(), &unstructured.Unstructured{})
	if err == nil {
		t.Fatal("expected error for non-AddressGroup object")
	}
}

func TestValidateUpdate(t *testing.T) {
	s := &addressGroupStrategy{}
	ctx := context.Background()

	old := &v1alpha1.AddressGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec:       v1alpha1.AddressGroupSpec{DefaultAction: v1alpha1.ActionAllow},
	}

	tests := []struct {
		name    string
		action  v1alpha1.Action
		wantErr bool
	}{
		{"allow to deny", v1alpha1.ActionDeny, false},
		{"allow to allow", v1alpha1.ActionAllow, false},
		{"allow to empty", "", true},
		{"allow to unknown", v1alpha1.ActionUnknown, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &v1alpha1.AddressGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec:       v1alpha1.AddressGroupSpec{DefaultAction: tt.action},
			}
			err := s.ValidateUpdate(ctx, obj, old)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate(%q) error = %v, wantErr %v", tt.action, err, tt.wantErr)
			}
		})
	}
}

func TestValidateUpdate_WrongType(t *testing.T) {
	s := &addressGroupStrategy{}
	old := &v1alpha1.AddressGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec:       v1alpha1.AddressGroupSpec{DefaultAction: v1alpha1.ActionAllow},
	}
	err := s.ValidateUpdate(context.Background(), &unstructured.Unstructured{}, old)
	if err == nil {
		t.Fatal("expected error for non-AddressGroup object")
	}
}
