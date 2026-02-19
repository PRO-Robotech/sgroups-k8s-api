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
