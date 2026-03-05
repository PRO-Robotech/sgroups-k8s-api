package mock

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"sgroups.io/sgroups-k8s-api/internal/backend"

	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

// MockNamespaceServer implements the SGroupsNamespaceAPI for tests and local runs.
type MockNamespaceServer struct {
	sgroupsv1.UnimplementedSGroupsNamespaceAPIServer
	backend backend.NamespaceBackend
}

// MockAddressGroupServer implements the SGroupsAddressGroupsAPI for tests and local runs.
type MockAddressGroupServer struct {
	sgroupsv1.UnimplementedSGroupsAddressGroupsAPIServer
	backend backend.AddressGroupBackend
}

// Run starts the mock gRPC server.
func Run(ctx context.Context, addr string, b backend.Backend) error {
	lis, err := net.Listen("tcp", addr) //nolint:noctx // mock server for local dev, no need for ListenConfig
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	)
	RegisterServices(grpcServer, b)
	reflection.Register(grpcServer)

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	return grpcServer.Serve(lis)
}

// RegisterServices registers gRPC services backed by the provided backend.
func RegisterServices(grpcServer *grpc.Server, b backend.Backend) {
	sgroupsv1.RegisterSGroupsNamespaceAPIServer(grpcServer, &MockNamespaceServer{backend: b.Namespaces})
	sgroupsv1.RegisterSGroupsAddressGroupsAPIServer(grpcServer, &MockAddressGroupServer{backend: b.AddressGroups})
}

func (s *MockNamespaceServer) Upsert(ctx context.Context, req *sgroupsv1.NamespaceReq_Upsert) (*sgroupsv1.NamespaceResp_Upsert, error) {
	if s.backend == nil {
		return nil, status.Error(codes.Unavailable, "namespace backend is not configured")
	}
	if req == nil || len(req.Namespaces) == 0 {
		return nil, status.Error(codes.InvalidArgument, "namespaces are required")
	}

	return s.backend.UpsertNamespaces(ctx, req)
}

func (s *MockNamespaceServer) Delete(ctx context.Context, req *sgroupsv1.NamespaceReq_Delete) (*emptypb.Empty, error) {
	if s.backend == nil {
		return nil, status.Error(codes.Unavailable, "namespace backend is not configured")
	}
	if req == nil || len(req.Namespaces) == 0 {
		return nil, status.Error(codes.InvalidArgument, "namespaces are required")
	}
	if err := s.backend.DeleteNamespaces(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *MockNamespaceServer) List(ctx context.Context, req *sgroupsv1.NamespaceReq_List) (*sgroupsv1.NamespaceResp_List, error) {
	if s.backend == nil {
		return nil, status.Error(codes.Unavailable, "namespace backend is not configured")
	}
	if req == nil || len(req.Selectors) == 0 {
		return nil, status.Error(codes.InvalidArgument, "selectors are required")
	}

	return s.backend.ListNamespaces(ctx, req)
}

func (s *MockNamespaceServer) Watch(req *sgroupsv1.NamespaceReq_Watch, stream grpc.ServerStreamingServer[sgroupsv1.NamespaceResp_Watch]) error {
	if s.backend == nil {
		return status.Error(codes.Unavailable, "namespace backend is not configured")
	}
	if req == nil || len(req.Selectors) == 0 {
		return status.Error(codes.InvalidArgument, "selectors are required")
	}
	ws, err := s.backend.WatchNamespaces(stream.Context(), req)
	if err != nil {
		return err
	}
	defer func() {
		if ws.Close != nil {
			ws.Close()
		}
	}()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case evt, ok := <-ws.C:
			if !ok {
				return nil
			}
			if err := stream.Send(evt); err != nil {
				return err
			}
		}
	}
}

func (s *MockAddressGroupServer) Upsert(ctx context.Context, req *sgroupsv1.AddressGroupReq_Upsert) (*sgroupsv1.AddressGroupResp_Upsert, error) {
	if s.backend == nil {
		return nil, status.Error(codes.Unavailable, "address group backend is not configured")
	}
	if req == nil || len(req.AddressGroups) == 0 {
		return nil, status.Error(codes.InvalidArgument, "address groups are required")
	}

	return s.backend.UpsertAddressGroups(ctx, req)
}

func (s *MockAddressGroupServer) Delete(ctx context.Context, req *sgroupsv1.AddressGroupReq_Delete) (*emptypb.Empty, error) {
	if s.backend == nil {
		return nil, status.Error(codes.Unavailable, "address group backend is not configured")
	}
	if req == nil || len(req.AddressGroups) == 0 {
		return nil, status.Error(codes.InvalidArgument, "address groups are required")
	}
	if err := s.backend.DeleteAddressGroups(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *MockAddressGroupServer) List(ctx context.Context, req *sgroupsv1.AddressGroupReq_List) (*sgroupsv1.AddressGroupResp_List, error) {
	if s.backend == nil {
		return nil, status.Error(codes.Unavailable, "address group backend is not configured")
	}
	if req == nil || len(req.Selectors) == 0 {
		return nil, status.Error(codes.InvalidArgument, "selectors are required")
	}

	return s.backend.ListAddressGroups(ctx, req)
}

func (s *MockAddressGroupServer) Watch(req *sgroupsv1.AddressGroupReq_Watch, stream grpc.ServerStreamingServer[sgroupsv1.AddressGroupResp_Watch]) error {
	if s.backend == nil {
		return status.Error(codes.Unavailable, "address group backend is not configured")
	}
	if req == nil || len(req.Selectors) == 0 {
		return status.Error(codes.InvalidArgument, "selectors are required")
	}
	ws, err := s.backend.WatchAddressGroups(stream.Context(), req)
	if err != nil {
		return err
	}
	defer func() {
		if ws.Close != nil {
			ws.Close()
		}
	}()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case evt, ok := <-ws.C:
			if !ok {
				return nil
			}
			if err := stream.Send(evt); err != nil {
				return err
			}
		}
	}
}
