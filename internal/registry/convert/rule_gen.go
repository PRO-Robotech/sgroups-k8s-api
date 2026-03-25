package convert

import (
	common "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// RuleToProto converts Rule to protobuf.
func RuleToProto(in *v1alpha1.Rule) *sgroupsv1.Rule {
	if in == nil {
		return nil
	}

	return &sgroupsv1.Rule{
		Metadata: objectMetaToProto(in.ObjectMeta),
		Spec: &sgroupsv1.Rule_Spec{
			DisplayName: in.Spec.DisplayName,
			Comment:     in.Spec.Comment,
			Description: in.Spec.Description,
			Action:      actionToProto(in.Spec.Action),
			Session:     sessionToProto(in.Spec.Session),
			Endpoints:   endpointsToProto(in.Spec.Endpoints),
			Transport:   transportToProto(in.Spec.Transport),
		},
	}
}

// RuleFromProto converts protobuf to Rule.
func RuleFromProto(in *sgroupsv1.Rule) *v1alpha1.Rule {
	if in == nil {
		return nil
	}
	out := &v1alpha1.Rule{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindRule,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.RuleSpec{
			DisplayName: in.GetSpec().GetDisplayName(),
			Comment:     in.GetSpec().GetComment(),
			Description: in.GetSpec().GetDescription(),
			Action:      actionFromProto(in.GetSpec().GetAction()),
			Session:     sessionFromProto(in.GetSpec().GetSession()),
			Endpoints:   endpointsFromProto(in.GetSpec().GetEndpoints()),
			Transport:   transportFromProto(in.GetSpec().GetTransport()),
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}

// --- Session converters ---

func sessionToProto(in *v1alpha1.RuleSession) *common.Session {
	if in == nil {
		return nil
	}

	return &common.Session{
		Traffic: trafficToProto(in.Traffic),
	}
}

func sessionFromProto(in *common.Session) *v1alpha1.RuleSession {
	if in == nil {
		return nil
	}

	return &v1alpha1.RuleSession{
		Traffic: trafficFromProto(in.GetTraffic()),
	}
}

func trafficToProto(t v1alpha1.Traffic) common.Session_Traffic {
	switch t {
	case v1alpha1.TrafficIngress:
		return common.Session_INGRESS
	case v1alpha1.TrafficEgress:
		return common.Session_EGRESS
	default:
		return common.Session_BOTH
	}
}

func trafficFromProto(t common.Session_Traffic) v1alpha1.Traffic {
	switch t {
	case common.Session_INGRESS:
		return v1alpha1.TrafficIngress
	case common.Session_EGRESS:
		return v1alpha1.TrafficEgress
	default:
		return v1alpha1.TrafficBoth
	}
}

// --- Endpoints converters ---

func endpointsToProto(in *v1alpha1.RuleEndpoints) *common.Endpoints {
	if in == nil {
		return nil
	}

	return &common.Endpoints{
		Local:  localEndpointToProto(in.Local),
		Remote: remoteEndpointToProto(in.Remote),
	}
}

func endpointsFromProto(in *common.Endpoints) *v1alpha1.RuleEndpoints {
	if in == nil {
		return nil
	}

	return &v1alpha1.RuleEndpoints{
		Local:  localEndpointFromProto(in.GetLocal()),
		Remote: remoteEndpointFromProto(in.GetRemote()),
	}
}

func localEndpointToProto(in *v1alpha1.RuleEndpoint) *common.Endpoints_Local {
	if in == nil {
		return nil
	}

	return &common.Endpoints_Local{
		Name:      in.Name,
		Namespace: in.Namespace,
		Type:      endpointTypeToProto(in.Type),
		Labels:    copyStringMap(in.Labels),
	}
}

func localEndpointFromProto(in *common.Endpoints_Local) *v1alpha1.RuleEndpoint {
	if in == nil {
		return nil
	}

	return &v1alpha1.RuleEndpoint{
		Name:      in.GetName(),
		Namespace: in.GetNamespace(),
		Type:      endpointTypeFromProto(in.GetType()),
		Labels:    copyStringMap(in.GetLabels()),
	}
}

func remoteEndpointToProto(in *v1alpha1.RuleEndpoint) *common.Endpoints_Remote {
	if in == nil {
		return nil
	}

	return &common.Endpoints_Remote{
		Name:      in.Name,
		Namespace: in.Namespace,
		Type:      endpointTypeToProto(in.Type),
		Value:     in.Value,
		Labels:    copyStringMap(in.Labels),
	}
}

func remoteEndpointFromProto(in *common.Endpoints_Remote) *v1alpha1.RuleEndpoint {
	if in == nil {
		return nil
	}

	return &v1alpha1.RuleEndpoint{
		Name:      in.GetName(),
		Namespace: in.GetNamespace(),
		Type:      endpointTypeFromProto(in.GetType()),
		Value:     in.GetValue(),
		Labels:    copyStringMap(in.GetLabels()),
	}
}

func endpointTypeToProto(t v1alpha1.EndpointType) common.Endpoints_Type {
	switch t {
	case v1alpha1.EndpointTypeAddressGroup:
		return common.Endpoints_ADDRESS_GROUP
	case v1alpha1.EndpointTypeService:
		return common.Endpoints_SERVICE
	case v1alpha1.EndpointTypeFQDN:
		return common.Endpoints_FQDN
	case v1alpha1.EndpointTypeCIDR:
		return common.Endpoints_CIDR
	default:
		return common.Endpoints_UNKNOWN
	}
}

func endpointTypeFromProto(t common.Endpoints_Type) v1alpha1.EndpointType {
	switch t {
	case common.Endpoints_ADDRESS_GROUP:
		return v1alpha1.EndpointTypeAddressGroup
	case common.Endpoints_SERVICE:
		return v1alpha1.EndpointTypeService
	case common.Endpoints_FQDN:
		return v1alpha1.EndpointTypeFQDN
	case common.Endpoints_CIDR:
		return v1alpha1.EndpointTypeCIDR
	default:
		return v1alpha1.EndpointTypeUnknown
	}
}

// --- Transport converters ---

func transportToProto(in *v1alpha1.RuleTransport) *common.Transport {
	if in == nil {
		return nil
	}
	entries := make([]*common.Transport_Entry, 0, len(in.Entries))
	for _, e := range in.Entries {
		entries = append(entries, transportEntryToProto(e))
	}

	return &common.Transport{
		Protocol: protocolToProto(in.Protocol),
		Ipv:      ipAddrFamilyToProto(in.IPv),
		Entries:  entries,
	}
}

func transportFromProto(in *common.Transport) *v1alpha1.RuleTransport {
	if in == nil {
		return nil
	}
	entries := make([]v1alpha1.TransportEntry, 0, len(in.GetEntries()))
	for _, e := range in.GetEntries() {
		entries = append(entries, transportEntryFromProto(e))
	}

	return &v1alpha1.RuleTransport{
		Protocol: protocolFromProto(in.GetProtocol()),
		IPv:      ipAddrFamilyFromProto(in.GetIpv()),
		Entries:  entries,
	}
}

func transportEntryToProto(in v1alpha1.TransportEntry) *common.Transport_Entry {
	return &common.Transport_Entry{
		Description: in.Description,
		Comment:     in.Comment,
		Ports:       in.Ports,
		Types:       in.Types,
	}
}

func transportEntryFromProto(in *common.Transport_Entry) v1alpha1.TransportEntry {
	if in == nil {
		return v1alpha1.TransportEntry{}
	}

	return v1alpha1.TransportEntry{
		Description: in.GetDescription(),
		Comment:     in.GetComment(),
		Ports:       in.GetPorts(),
		Types:       in.GetTypes(),
	}
}

func protocolToProto(p v1alpha1.Protocol) common.Transport_Protocol {
	switch p {
	case v1alpha1.ProtocolUDP:
		return common.Transport_UDP
	case v1alpha1.ProtocolICMP:
		return common.Transport_ICMP
	default:
		return common.Transport_TCP
	}
}

func protocolFromProto(p common.Transport_Protocol) v1alpha1.Protocol {
	switch p {
	case common.Transport_UDP:
		return v1alpha1.ProtocolUDP
	case common.Transport_ICMP:
		return v1alpha1.ProtocolICMP
	default:
		return v1alpha1.ProtocolTCP
	}
}

func ipAddrFamilyToProto(f v1alpha1.IpAddrFamily) common.IpAddrFamily {
	switch f {
	case v1alpha1.IpAddrFamilyIPv4:
		return common.IpAddrFamily_IPV4
	case v1alpha1.IpAddrFamilyIPv6:
		return common.IpAddrFamily_IPV6
	default:
		return common.IpAddrFamily_IPV_UNDEF
	}
}

func ipAddrFamilyFromProto(f common.IpAddrFamily) v1alpha1.IpAddrFamily {
	switch f {
	case common.IpAddrFamily_IPV4:
		return v1alpha1.IpAddrFamilyIPv4
	case common.IpAddrFamily_IPV6:
		return v1alpha1.IpAddrFamilyIPv6
	default:
		return v1alpha1.IpAddrFamilyUndef
	}
}
