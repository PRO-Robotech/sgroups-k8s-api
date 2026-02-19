package testutil

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"sgroups.io/sgroups-k8s-api/internal/backend"
	"sgroups.io/sgroups-k8s-api/internal/grpcclient"
	"sgroups.io/sgroups-k8s-api/internal/mock"
	"sgroups.io/sgroups-k8s-api/pkg/client"
)

const bufSize = 1024 * 1024

// NewBufconnClient starts an in-memory gRPC server and returns a client bound to it.
func NewBufconnClient(t *testing.T, b backend.Backend) (*client.Client, func()) {
	t.Helper()

	lis := bufconn.Listen(bufSize)
	grpcServer := grpc.NewServer()
	mock.RegisterServices(grpcServer, b)

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	dialer := func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := client.Dial(
		"passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(grpcclient.UserMetadataUnaryInterceptor()),
		grpc.WithChainStreamInterceptor(grpcclient.UserMetadataStreamInterceptor()),
	)
	if err != nil {
		grpcServer.Stop()
		_ = lis.Close()
		t.Fatalf("dial bufconn: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		grpcServer.Stop()
		_ = lis.Close()
	}

	return conn, cleanup
}
