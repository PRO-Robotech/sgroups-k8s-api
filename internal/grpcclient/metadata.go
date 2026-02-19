package grpcclient

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
)

// InjectUserMetadata adds Kubernetes user info into gRPC outgoing metadata.
func InjectUserMetadata(ctx context.Context) context.Context {
	user, ok := apirequest.UserFrom(ctx)
	if !ok || user == nil {
		return ctx
	}

	outgoing, _ := metadata.FromOutgoingContext(ctx)
	md := outgoing.Copy()

	if name := user.GetName(); name != "" {
		md.Set("x-k8s-user", name)
	}
	if uid := user.GetUID(); uid != "" {
		md.Set("x-k8s-uid", uid)
	}
	for _, group := range user.GetGroups() {
		if group == "" {
			continue
		}
		md.Append("x-k8s-groups", group)
	}
	for key, values := range user.GetExtra() {
		if key == "" {
			continue
		}
		header := "x-k8s-extra-" + sanitizeHeaderKey(key)
		for _, v := range values {
			if v == "" {
				continue
			}
			md.Append(header, v)
		}
	}

	return metadata.NewOutgoingContext(ctx, md)
}

func sanitizeHeaderKey(key string) string {
	key = strings.ToLower(key)

	return strings.ReplaceAll(key, "_", "-")
}
