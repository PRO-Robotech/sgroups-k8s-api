package base

import (
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateListOptions(t *testing.T) {
	tests := []struct {
		name      string
		options   *metav1.ListOptions
		expectErr bool
	}{
		{
			name:    "nil options",
			options: nil,
		},
		{
			name:    "empty options",
			options: &metav1.ListOptions{},
		},
		{
			name:    "resourceVersion 0 allowed",
			options: &metav1.ListOptions{ResourceVersion: "0"},
		},
		{
			name:    "continue silently ignored",
			options: &metav1.ListOptions{Continue: "token"},
		},
		{
			name:    "limit silently ignored",
			options: &metav1.ListOptions{Limit: 10},
		},
		{
			name:    "specific resourceVersion allowed",
			options: &metav1.ListOptions{ResourceVersion: "42"},
		},
		{
			name:      "resourceVersionMatch not supported",
			options:   &metav1.ListOptions{ResourceVersionMatch: metav1.ResourceVersionMatchExact},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateListOptions(tt.options)
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

func TestValidateWatchOptions(t *testing.T) {
	tests := []struct {
		name      string
		options   *metav1.ListOptions
		expectErr bool
	}{
		{
			name:    "nil options",
			options: nil,
		},
		{
			name:    "empty options",
			options: &metav1.ListOptions{},
		},
		{
			name:    "resourceVersion allowed for watch",
			options: &metav1.ListOptions{ResourceVersion: "42"},
		},
		{
			name:    "continue silently ignored",
			options: &metav1.ListOptions{Continue: "token"},
		},
		{
			name:    "limit silently ignored",
			options: &metav1.ListOptions{Limit: 10},
		},
		{
			name:      "resourceVersionMatch not supported",
			options:   &metav1.ListOptions{ResourceVersionMatch: metav1.ResourceVersionMatchExact},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWatchOptions(tt.options)
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

func TestValidateGetOptions(t *testing.T) {
	tests := []struct {
		name      string
		options   *metav1.GetOptions
		expectErr bool
	}{
		{
			name:    "nil options",
			options: nil,
		},
		{
			name:    "empty options",
			options: &metav1.GetOptions{},
		},
		{
			name:    "resourceVersion 0 allowed",
			options: &metav1.GetOptions{ResourceVersion: "0"},
		},
		{
			name:    "specific resourceVersion allowed",
			options: &metav1.GetOptions{ResourceVersion: "42"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGetOptions(tt.options)
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
