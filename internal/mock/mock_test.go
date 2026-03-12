package mock

import (
	"context"
	"testing"

	commonpb "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

func TestNamespaceUpsertList(t *testing.T) {
	mb := New()
	_, err := mb.UpsertNamespaces(context.Background(), &sgroupsv1.NamespaceReq_Upsert{
		Namespaces: []*sgroupsv1.Namespace{
			{
				Metadata: &commonpb.Metadata{Name: "default"},
				Spec:     &sgroupsv1.Namespace_Spec{DisplayName: "Default"},
			},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	list, err := mb.ListNamespaces(context.Background(), &sgroupsv1.NamespaceReq_List{
		Selectors: []*sgroupsv1.NamespaceReq_Selector{{}},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list.Namespaces) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(list.Namespaces))
	}
	if list.Namespaces[0].Metadata.Uid == "" {
		t.Fatalf("expected uid to be set")
	}
}

func TestAddressGroupUpsertList(t *testing.T) {
	mb := New()
	_, err := mb.UpsertAddressGroups(context.Background(), &sgroupsv1.AddressGroupReq_Upsert{
		AddressGroups: []*sgroupsv1.AddressGroup{
			{
				Metadata: &commonpb.Metadata{Name: "ag-1", Namespace: "default"},
				Spec:     &sgroupsv1.AddressGroup_Spec{DefaultAction: commonpb.Action_ALLOW},
			},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	list, err := mb.ListAddressGroups(context.Background(), &sgroupsv1.AddressGroupReq_List{
		Selectors: []*commonpb.ResSelector{{}},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list.AddressGroups) != 1 {
		t.Fatalf("expected 1 address group, got %d", len(list.AddressGroups))
	}
	if list.AddressGroups[0].Metadata.Uid == "" {
		t.Fatalf("expected uid to be set")
	}
}

func TestNetworkUpsertList(t *testing.T) {
	mb := New()
	_, err := mb.UpsertNetworks(context.Background(), &sgroupsv1.NetworkReq_Upsert{
		Networks: []*sgroupsv1.Network{
			{
				Metadata: &commonpb.Metadata{Name: "nw-1", Namespace: "default"},
				Spec:     &sgroupsv1.Network_Spec{Cidr: "10.0.0.0/24"},
			},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	list, err := mb.ListNetworks(context.Background(), &sgroupsv1.NetworkReq_List{
		Selectors: []*commonpb.ResSelector{{}},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list.Networks) != 1 {
		t.Fatalf("expected 1 network, got %d", len(list.Networks))
	}
	if list.Networks[0].Metadata.Uid == "" {
		t.Fatalf("expected uid to be set")
	}
}

func TestHostUpsertList(t *testing.T) {
	mb := New()
	_, err := mb.UpsertHosts(context.Background(), &sgroupsv1.HostReq_Upsert{
		Hosts: []*sgroupsv1.Host{
			{
				Metadata: &commonpb.Metadata{Name: "host-1", Namespace: "default"},
				Spec:     &sgroupsv1.Host_Spec{DisplayName: "Host 1"},
			},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	list, err := mb.ListHosts(context.Background(), &sgroupsv1.HostReq_List{
		Selectors: []*commonpb.ResSelector{{}},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list.Hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(list.Hosts))
	}
	if list.Hosts[0].Metadata.Uid == "" {
		t.Fatalf("expected uid to be set")
	}
}

func TestHostBindingUpsertList(t *testing.T) {
	mb := New()
	_, err := mb.UpsertHostBindings(context.Background(), &sgroupsv1.HostBindingReq_Upsert{
		HostBindings: []*sgroupsv1.HostBinding{
			{
				Metadata: &commonpb.Metadata{Name: "hb-1", Namespace: "default"},
				Spec: &sgroupsv1.HostBinding_Spec{
					DisplayName:  "Binding 1",
					AddressGroup: &commonpb.ResourceIdentifier{Name: "ag-1", Namespace: "default"},
					Host:         &commonpb.ResourceIdentifier{Name: "host-1", Namespace: "default"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	list, err := mb.ListHostBindings(context.Background(), &sgroupsv1.HostBindingReq_List{
		Selectors: []*sgroupsv1.HostBindingReq_Selectors{{}},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list.HostBindings) != 1 {
		t.Fatalf("expected 1 host binding, got %d", len(list.HostBindings))
	}
	if list.HostBindings[0].Metadata.Uid == "" {
		t.Fatalf("expected uid to be set")
	}
}
