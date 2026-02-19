package errors

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestFromGRPC(t *testing.T) {
	gr := schema.GroupResource{Group: "sgroups.io", Resource: "namespaces"}

	tests := []struct {
		name       string
		err        error
		wantNil    bool
		wantReason func(*apierrors.StatusError) bool
	}{
		{
			name:    "nil error",
			err:     nil,
			wantNil: true,
		},
		{
			name: "NotFound",
			err:  status.Error(codes.NotFound, "not found"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsNotFound(se)
			},
		},
		{
			name: "AlreadyExists",
			err:  status.Error(codes.AlreadyExists, "exists"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsAlreadyExists(se)
			},
		},
		{
			name: "InvalidArgument",
			err:  status.Error(codes.InvalidArgument, "bad input"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsBadRequest(se)
			},
		},
		{
			name: "FailedPrecondition",
			err:  status.Error(codes.FailedPrecondition, "conflict"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsConflict(se)
			},
		},
		{
			name: "Aborted",
			err:  status.Error(codes.Aborted, "aborted"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsConflict(se)
			},
		},
		{
			name: "PermissionDenied",
			err:  status.Error(codes.PermissionDenied, "forbidden"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsForbidden(se)
			},
		},
		{
			name: "Unauthenticated",
			err:  status.Error(codes.Unauthenticated, "unauth"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsUnauthorized(se)
			},
		},
		{
			name: "DeadlineExceeded",
			err:  status.Error(codes.DeadlineExceeded, "timeout"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsTimeout(se)
			},
		},
		{
			name: "Unavailable",
			err:  status.Error(codes.Unavailable, "unavailable"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsServiceUnavailable(se)
			},
		},
		{
			name: "Unknown code maps to InternalError",
			err:  status.Error(codes.DataLoss, "data loss"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsInternalError(se)
			},
		},
		{
			name: "non-gRPC error maps to InternalError",
			err:  errors.New("plain error"),
			wantReason: func(se *apierrors.StatusError) bool {
				return apierrors.IsInternalError(se)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromGRPC(tt.err, gr, "test-obj")
			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}

				return
			}
			if got == nil {
				t.Fatal("expected non-nil error")
			}
			var se *apierrors.StatusError
			if !errors.As(got, &se) {
				t.Fatalf("expected *apierrors.StatusError, got %T", got)
			}
			if !tt.wantReason(se) {
				t.Fatalf("unexpected error reason: %v", se)
			}
		})
	}
}
