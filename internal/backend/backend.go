package backend

import (
	"context"

	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

type WatchStream[T any] struct {
	C     <-chan T
	Close func()
}

type NamespaceBackend interface {
	UpsertNamespaces(ctx context.Context, req *sgroupsv1.NamespaceReq_Upsert) (*sgroupsv1.NamespaceResp_Upsert, error)
	DeleteNamespaces(ctx context.Context, req *sgroupsv1.NamespaceReq_Delete) error
	ListNamespaces(ctx context.Context, req *sgroupsv1.NamespaceReq_List) (*sgroupsv1.NamespaceResp_List, error)
	WatchNamespaces(ctx context.Context, req *sgroupsv1.NamespaceReq_Watch) (WatchStream[*sgroupsv1.NamespaceResp_Watch], error)
}

type AddressGroupBackend interface {
	UpsertAddressGroups(ctx context.Context, req *sgroupsv1.AddressGroupReq_Upsert) (*sgroupsv1.AddressGroupResp_Upsert, error)
	DeleteAddressGroups(ctx context.Context, req *sgroupsv1.AddressGroupReq_Delete) error
	ListAddressGroups(ctx context.Context, req *sgroupsv1.AddressGroupReq_List) (*sgroupsv1.AddressGroupResp_List, error)
	WatchAddressGroups(ctx context.Context, req *sgroupsv1.AddressGroupReq_Watch) (WatchStream[*sgroupsv1.AddressGroupResp_Watch], error)
}

type Backend struct {
	Namespaces    NamespaceBackend
	AddressGroups AddressGroupBackend
}
