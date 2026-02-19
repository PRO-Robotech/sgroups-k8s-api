package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// FromGRPC maps gRPC errors to Kubernetes API errors.
func FromGRPC(err error, resource schema.GroupResource, name string) error {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return apierrors.NewInternalError(err)
	}

	switch st.Code() {
	case codes.NotFound:
		return apierrors.NewNotFound(resource, name)
	case codes.AlreadyExists:
		return apierrors.NewAlreadyExists(resource, name)
	case codes.InvalidArgument:
		return apierrors.NewBadRequest(st.Message())
	case codes.FailedPrecondition, codes.Aborted:
		return apierrors.NewConflict(resource, name, fmt.Errorf("%s", st.Message()))
	case codes.PermissionDenied:
		return apierrors.NewForbidden(resource, name, fmt.Errorf("%s", st.Message()))
	case codes.Unauthenticated:
		return apierrors.NewUnauthorized(st.Message())
	case codes.DeadlineExceeded:
		return apierrors.NewTimeoutError(st.Message(), 0)
	case codes.Unavailable:
		return apierrors.NewServiceUnavailable(st.Message())
	default:
		return apierrors.NewInternalError(fmt.Errorf("%s", st.Message()))
	}
}
