package mock

import (
	"context"
	"testing"
	"time"

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

func TestRuleUpsertList(t *testing.T) {
	mb := New()
	_, err := mb.UpsertRules(context.Background(), &sgroupsv1.RuleReq_Upsert{
		Rules: []*sgroupsv1.Rule{
			{
				Metadata: &commonpb.Metadata{Name: "rule-1", Namespace: "default"},
				Spec: &sgroupsv1.Rule_Spec{
					DisplayName: "Allow ingress from AG",
					Action:      commonpb.Action_ALLOW,
					Session:     &commonpb.Session{Traffic: commonpb.Session_INGRESS},
					Endpoints: &commonpb.Endpoints{
						Local: &commonpb.Endpoints_Local{
							Name:      "ag-web",
							Namespace: "default",
							Type:      commonpb.Endpoints_ADDRESS_GROUP,
						},
						Remote: &commonpb.Endpoints_Remote{
							Name:      "ag-db",
							Namespace: "default",
							Type:      commonpb.Endpoints_ADDRESS_GROUP,
						},
					},
					Transport: &commonpb.Transport{
						Protocol: commonpb.Transport_TCP,
						Entries: []*commonpb.Transport_Entry{
							{Ports: "443", Description: "HTTPS"},
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	list, err := mb.ListRules(context.Background(), &sgroupsv1.RuleReq_List{
		Selectors: []*sgroupsv1.RuleReq_Selectors{{}},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(list.Rules))
	}
	r := list.Rules[0]
	if r.Metadata.Uid == "" {
		t.Fatalf("expected uid to be set")
	}
	if r.Spec.Action != commonpb.Action_ALLOW {
		t.Fatalf("expected action ALLOW, got %v", r.Spec.Action)
	}
	if r.Spec.Session.Traffic != commonpb.Session_INGRESS {
		t.Fatalf("expected traffic INGRESS, got %v", r.Spec.Session.Traffic)
	}
	if r.Spec.Endpoints.Local.Type != commonpb.Endpoints_ADDRESS_GROUP {
		t.Fatalf("expected local type ADDRESS_GROUP, got %v", r.Spec.Endpoints.Local.Type)
	}
	if r.Spec.Transport.Protocol != commonpb.Transport_TCP {
		t.Fatalf("expected protocol TCP, got %v", r.Spec.Transport.Protocol)
	}
	if len(r.Spec.Transport.Entries) != 1 || r.Spec.Transport.Entries[0].Ports != "443" {
		t.Fatalf("expected transport entry with ports 443")
	}
}

func TestRuleDelete(t *testing.T) {
	mb := New()
	resp, err := mb.UpsertRules(context.Background(), &sgroupsv1.RuleReq_Upsert{
		Rules: []*sgroupsv1.Rule{
			{
				Metadata: &commonpb.Metadata{Name: "rule-del", Namespace: "default"},
				Spec:     &sgroupsv1.Rule_Spec{Action: commonpb.Action_DENY},
			},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	err = mb.DeleteRules(context.Background(), &sgroupsv1.RuleReq_Delete{
		Rules: []*sgroupsv1.RuleReq_Delete_Rule{
			{Metadata: &commonpb.MetadataScope{Name: "rule-del", Namespace: "default"}},
		},
	})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	list, err := mb.ListRules(context.Background(), &sgroupsv1.RuleReq_List{
		Selectors: []*sgroupsv1.RuleReq_Selectors{{}},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list.Rules) != 0 {
		t.Fatalf("expected 0 rules after delete, got %d", len(list.Rules))
	}
	_ = resp
}

func TestRuleWatch(t *testing.T) {
	mb := New()

	// Start watch before any upserts
	ws, err := mb.WatchRules(context.Background(), &sgroupsv1.RuleReq_Watch{
		Selectors: []*sgroupsv1.RuleReq_Selectors{{}},
	})
	if err != nil {
		t.Fatalf("watch failed: %v", err)
	}
	defer ws.Close()

	// Receive initial snapshot (empty)
	select {
	case evt := <-ws.C:
		if evt.Type != commonpb.WatchEventType_ADDED {
			t.Fatalf("expected ADDED snapshot, got %v", evt.Type)
		}
		if len(evt.Rules) != 0 {
			t.Fatalf("expected 0 rules in snapshot, got %d", len(evt.Rules))
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for initial snapshot")
	}

	// Upsert a rule
	_, err = mb.UpsertRules(context.Background(), &sgroupsv1.RuleReq_Upsert{
		Rules: []*sgroupsv1.Rule{
			{
				Metadata: &commonpb.Metadata{Name: "rule-watch", Namespace: "default"},
				Spec: &sgroupsv1.Rule_Spec{
					Action:  commonpb.Action_ALLOW,
					Session: &commonpb.Session{Traffic: commonpb.Session_BOTH},
					Transport: &commonpb.Transport{
						Protocol: commonpb.Transport_TCP,
						Entries:  []*commonpb.Transport_Entry{{Ports: "8080"}},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	// Receive ADDED event
	select {
	case evt := <-ws.C:
		if evt.Type != commonpb.WatchEventType_ADDED {
			t.Fatalf("expected ADDED, got %v", evt.Type)
		}
		if len(evt.Rules) != 1 {
			t.Fatalf("expected 1 rule in event, got %d", len(evt.Rules))
		}
		if evt.Rules[0].Metadata.Name != "rule-watch" {
			t.Fatalf("expected rule-watch, got %s", evt.Rules[0].Metadata.Name)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for ADDED event")
	}

	// Update the same rule (upsert with uid)
	list, _ := mb.ListRules(context.Background(), &sgroupsv1.RuleReq_List{
		Selectors: []*sgroupsv1.RuleReq_Selectors{{}},
	})
	uid := list.Rules[0].Metadata.Uid
	_, err = mb.UpsertRules(context.Background(), &sgroupsv1.RuleReq_Upsert{
		Rules: []*sgroupsv1.Rule{
			{
				Metadata: &commonpb.Metadata{Uid: uid, Name: "rule-watch", Namespace: "default"},
				Spec:     &sgroupsv1.Rule_Spec{Action: commonpb.Action_DENY},
			},
		},
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Receive MODIFIED event
	select {
	case evt := <-ws.C:
		if evt.Type != commonpb.WatchEventType_MODIFIED {
			t.Fatalf("expected MODIFIED, got %v", evt.Type)
		}
		if evt.Rules[0].Spec.Action != commonpb.Action_DENY {
			t.Fatalf("expected DENY after update, got %v", evt.Rules[0].Spec.Action)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for MODIFIED event")
	}

	// Delete
	err = mb.DeleteRules(context.Background(), &sgroupsv1.RuleReq_Delete{
		Rules: []*sgroupsv1.RuleReq_Delete_Rule{
			{Metadata: &commonpb.MetadataScope{Uid: uid}},
		},
	})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Receive DELETED event
	select {
	case evt := <-ws.C:
		if evt.Type != commonpb.WatchEventType_DELETED {
			t.Fatalf("expected DELETED, got %v", evt.Type)
		}
		if len(evt.Rules) != 1 {
			t.Fatalf("expected 1 rule in delete event, got %d", len(evt.Rules))
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for DELETED event")
	}
}

func TestRuleWatchWithSelector(t *testing.T) {
	mb := New()

	// Watch only namespace "prod"
	ws, err := mb.WatchRules(context.Background(), &sgroupsv1.RuleReq_Watch{
		Selectors: []*sgroupsv1.RuleReq_Selectors{{
			FieldSelector: &sgroupsv1.RuleReq_Selectors_FieldSelector{
				Namespace: "prod",
			},
		}},
	})
	if err != nil {
		t.Fatalf("watch failed: %v", err)
	}
	defer ws.Close()

	// Drain initial snapshot
	<-ws.C

	// Upsert rule in "dev" namespace — should NOT appear
	_, _ = mb.UpsertRules(context.Background(), &sgroupsv1.RuleReq_Upsert{
		Rules: []*sgroupsv1.Rule{
			{
				Metadata: &commonpb.Metadata{Name: "rule-dev", Namespace: "dev"},
				Spec:     &sgroupsv1.Rule_Spec{Action: commonpb.Action_ALLOW},
			},
		},
	})

	// Upsert rule in "prod" namespace — should appear
	_, _ = mb.UpsertRules(context.Background(), &sgroupsv1.RuleReq_Upsert{
		Rules: []*sgroupsv1.Rule{
			{
				Metadata: &commonpb.Metadata{Name: "rule-prod", Namespace: "prod"},
				Spec:     &sgroupsv1.Rule_Spec{Action: commonpb.Action_DENY},
			},
		},
	})

	// We should receive the prod event (dev may or may not arrive as empty filtered)
	received := false
	for i := 0; i < 3; i++ {
		select {
		case evt := <-ws.C:
			for _, r := range evt.Rules {
				if r.Metadata.Name == "rule-prod" {
					received = true
				}
				if r.Metadata.Name == "rule-dev" {
					t.Fatal("should not receive rule-dev with namespace=prod selector")
				}
			}
		case <-time.After(500 * time.Millisecond):
			break
		}
		if received {
			break
		}
	}
	if !received {
		t.Fatal("never received rule-prod event")
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
