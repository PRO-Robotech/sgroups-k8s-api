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

type NetworkBackend interface {
	UpsertNetworks(ctx context.Context, req *sgroupsv1.NetworkReq_Upsert) (*sgroupsv1.NetworkResp_Upsert, error)
	DeleteNetworks(ctx context.Context, req *sgroupsv1.NetworkReq_Delete) error
	ListNetworks(ctx context.Context, req *sgroupsv1.NetworkReq_List) (*sgroupsv1.NetworkResp_List, error)
	WatchNetworks(ctx context.Context, req *sgroupsv1.NetworkReq_Watch) (WatchStream[*sgroupsv1.NetworkResp_Watch], error)
}

type HostBackend interface {
	UpsertHosts(ctx context.Context, req *sgroupsv1.HostReq_Upsert) (*sgroupsv1.HostResp_Upsert, error)
	DeleteHosts(ctx context.Context, req *sgroupsv1.HostReq_Delete) error
	ListHosts(ctx context.Context, req *sgroupsv1.HostReq_List) (*sgroupsv1.HostResp_List, error)
	WatchHosts(ctx context.Context, req *sgroupsv1.HostReq_Watch) (WatchStream[*sgroupsv1.HostResp_Watch], error)
}

type HostBindingBackend interface {
	UpsertHostBindings(ctx context.Context, req *sgroupsv1.HostBindingReq_Upsert) (*sgroupsv1.HostBindingResp_Upsert, error)
	DeleteHostBindings(ctx context.Context, req *sgroupsv1.HostBindingReq_Delete) error
	ListHostBindings(ctx context.Context, req *sgroupsv1.HostBindingReq_List) (*sgroupsv1.HostBindingResp_List, error)
	WatchHostBindings(ctx context.Context, req *sgroupsv1.HostBindingReq_Watch) (WatchStream[*sgroupsv1.HostBindingResp_Watch], error)
}

type NetworkBindingBackend interface {
	UpsertNetworkBindings(ctx context.Context, req *sgroupsv1.NetworkBindingReq_Upsert) (*sgroupsv1.NetworkBindingResp_Upsert, error)
	DeleteNetworkBindings(ctx context.Context, req *sgroupsv1.NetworkBindingReq_Delete) error
	ListNetworkBindings(ctx context.Context, req *sgroupsv1.NetworkBindingReq_List) (*sgroupsv1.NetworkBindingResp_List, error)
	WatchNetworkBindings(ctx context.Context, req *sgroupsv1.NetworkBindingReq_Watch) (WatchStream[*sgroupsv1.NetworkBindingResp_Watch], error)
}

type Backend struct {
	Namespaces      NamespaceBackend
	AddressGroups   AddressGroupBackend
	Networks        NetworkBackend
	Hosts           HostBackend
	HostBindings    HostBindingBackend
	NetworkBindings NetworkBindingBackend
}
