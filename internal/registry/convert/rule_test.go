package convert

import (
	"testing"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func TestRuleConversion(t *testing.T) {
	in := &v1alpha1.Rule{}
	in.Name = "test-rule"
	in.Namespace = "default"
	in.Spec = v1alpha1.RuleSpec{
		DisplayName: "Test Rule",
		Comment:     "test comment",
		Description: "test description",
		Action:      v1alpha1.ActionAllow,
		Session:     &v1alpha1.RuleSession{Traffic: v1alpha1.TrafficIngress},
		Endpoints: &v1alpha1.RuleEndpoints{
			Local: &v1alpha1.RuleEndpoint{
				Name:      "ag-web",
				Namespace: "default",
				Type:      v1alpha1.EndpointTypeAddressGroup,
				Labels:    map[string]string{"env": "prod"},
			},
			Remote: &v1alpha1.RuleEndpoint{
				Name:      "external",
				Namespace: "default",
				Type:      v1alpha1.EndpointTypeCIDR,
				Value:     "10.0.0.0/8",
			},
		},
		Transport: &v1alpha1.RuleTransport{
			Protocol: v1alpha1.ProtocolTCP,
			IPv:      v1alpha1.IpAddrFamilyIPv4,
			Entries: []v1alpha1.TransportEntry{
				{Description: "HTTPS", Ports: "443"},
				{Description: "HTTP", Ports: "80"},
			},
		},
	}

	proto := RuleToProto(in)
	if proto == nil {
		t.Fatal("RuleToProto returned nil")
	}
	if proto.Metadata.Name != "test-rule" {
		t.Fatalf("expected name test-rule, got %s", proto.Metadata.Name)
	}

	out := RuleFromProto(proto)
	if out == nil {
		t.Fatal("RuleFromProto returned nil")
	}

	// Verify round-trip preserves fields
	if out.Spec.DisplayName != in.Spec.DisplayName {
		t.Fatalf("displayName mismatch: %s vs %s", out.Spec.DisplayName, in.Spec.DisplayName)
	}
	if out.Spec.Action != in.Spec.Action {
		t.Fatalf("action mismatch: %s vs %s", out.Spec.Action, in.Spec.Action)
	}
	if out.Spec.Session.Traffic != in.Spec.Session.Traffic {
		t.Fatalf("traffic mismatch: %s vs %s", out.Spec.Session.Traffic, in.Spec.Session.Traffic)
	}
	if out.Spec.Endpoints.Local.Type != in.Spec.Endpoints.Local.Type {
		t.Fatalf("local endpoint type mismatch: %s vs %s", out.Spec.Endpoints.Local.Type, in.Spec.Endpoints.Local.Type)
	}
	if out.Spec.Endpoints.Local.Labels["env"] != "prod" {
		t.Fatalf("local endpoint labels not preserved")
	}
	if out.Spec.Endpoints.Remote.Value != "10.0.0.0/8" {
		t.Fatalf("remote endpoint value mismatch: %s", out.Spec.Endpoints.Remote.Value)
	}
	if out.Spec.Transport.Protocol != in.Spec.Transport.Protocol {
		t.Fatalf("protocol mismatch: %s vs %s", out.Spec.Transport.Protocol, in.Spec.Transport.Protocol)
	}
	if out.Spec.Transport.IPv != in.Spec.Transport.IPv {
		t.Fatalf("IPv mismatch: %s vs %s", out.Spec.Transport.IPv, in.Spec.Transport.IPv)
	}
	if len(out.Spec.Transport.Entries) != 2 {
		t.Fatalf("expected 2 transport entries, got %d", len(out.Spec.Transport.Entries))
	}
	if out.Spec.Transport.Entries[0].Ports != "443" {
		t.Fatalf("entry ports mismatch: %s", out.Spec.Transport.Entries[0].Ports)
	}
}

func TestRuleConversionNil(t *testing.T) {
	if RuleToProto(nil) != nil {
		t.Fatal("RuleToProto(nil) should return nil")
	}
	if RuleFromProto(nil) != nil {
		t.Fatal("RuleFromProto(nil) should return nil")
	}
}
