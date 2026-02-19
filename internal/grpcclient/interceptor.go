package grpcclient

import (
	"context"

	"google.golang.org/grpc"
)

// UserMetadataUnaryInterceptor returns a grpc.UnaryClientInterceptor that injects
// Kubernetes user info into every outgoing unary RPC call.
func UserMetadataUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(InjectUserMetadata(ctx), method, req, reply, cc, opts...)
	}
}

// UserMetadataStreamInterceptor returns a grpc.StreamClientInterceptor that injects
// Kubernetes user info into every outgoing streaming RPC call.
func UserMetadataStreamInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
		method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(InjectUserMetadata(ctx), desc, cc, method, opts...)
	}
}
