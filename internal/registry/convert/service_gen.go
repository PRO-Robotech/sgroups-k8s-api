package convert

import (
	common "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// ServiceToProto converts Service to protobuf.
func ServiceToProto(in *v1alpha1.Service) *sgroupsv1.Service {
	if in == nil {
		return nil
	}

	transports := make([]*common.Transport, 0, len(in.Spec.Transports))
	for _, t := range in.Spec.Transports {
		transports = append(transports, serviceTransportToProto(t))
	}

	return &sgroupsv1.Service{
		Metadata: objectMetaToProto(in.ObjectMeta),
		Spec: &sgroupsv1.Service_Spec{
			DisplayName: in.Spec.DisplayName,
			Comment:     in.Spec.Comment,
			Description: in.Spec.Description,
			Transports:  transports,
		},
	}
}

// ServiceFromProto converts protobuf to Service.
func ServiceFromProto(in *sgroupsv1.Service) *v1alpha1.Service {
	if in == nil {
		return nil
	}

	transports := make([]v1alpha1.ServiceTransport, 0, len(in.GetSpec().GetTransports()))
	for _, t := range in.GetSpec().GetTransports() {
		transports = append(transports, serviceTransportFromProto(t))
	}

	out := &v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindService,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.ServiceSpec{
			DisplayName: in.GetSpec().GetDisplayName(),
			Comment:     in.GetSpec().GetComment(),
			Description: in.GetSpec().GetDescription(),
			Transports:  transports,
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}

// ServiceFromProtoExt converts list/watch protobuf payload to Service.
func ServiceFromProtoExt(in *sgroupsv1.ServiceResp_ServiceExt) *v1alpha1.Service {
	if in == nil {
		return nil
	}

	transports := make([]v1alpha1.ServiceTransport, 0, len(in.GetSpec().GetTransports()))
	for _, t := range in.GetSpec().GetTransports() {
		transports = append(transports, serviceTransportFromProto(t))
	}

	out := &v1alpha1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindService,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Spec: v1alpha1.ServiceSpec{
			DisplayName: in.GetSpec().GetDisplayName(),
			Comment:     in.GetSpec().GetComment(),
			Description: in.GetSpec().GetDescription(),
			Transports:  transports,
		},
	}
	objectMetaFromProto(&out.ObjectMeta, in.GetMetadata())

	return out
}

func serviceTransportToProto(in v1alpha1.ServiceTransport) *common.Transport {
	entries := make([]*common.Transport_Entry, 0, len(in.Entries))
	for _, e := range in.Entries {
		entries = append(entries, serviceTransportEntryToProto(e))
	}

	return &common.Transport{
		Protocol: protocolToProto(in.Protocol),
		Ipv:      ipAddrFamilyToProto(in.IPv),
		Entries:  entries,
	}
}

func serviceTransportFromProto(in *common.Transport) v1alpha1.ServiceTransport {
	if in == nil {
		return v1alpha1.ServiceTransport{}
	}
	entries := make([]v1alpha1.ServiceTransportEntry, 0, len(in.GetEntries()))
	for _, e := range in.GetEntries() {
		entries = append(entries, serviceTransportEntryFromProto(e))
	}

	return v1alpha1.ServiceTransport{
		Protocol: protocolFromProto(in.GetProtocol()),
		IPv:      ipAddrFamilyFromProto(in.GetIpv()),
		Entries:  entries,
	}
}

func serviceTransportEntryToProto(in v1alpha1.ServiceTransportEntry) *common.Transport_Entry {
	return &common.Transport_Entry{
		Description: in.Description,
		Comment:     in.Comment,
		Ports:       in.Ports,
		Types:       in.Types,
	}
}

func serviceTransportEntryFromProto(in *common.Transport_Entry) v1alpha1.ServiceTransportEntry {
	if in == nil {
		return v1alpha1.ServiceTransportEntry{}
	}

	return v1alpha1.ServiceTransportEntry{
		Description: in.GetDescription(),
		Comment:     in.GetComment(),
		Ports:       in.GetPorts(),
		Types:       in.GetTypes(),
	}
}
