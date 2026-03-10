package mock

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"sgroups.io/sgroups-k8s-api/internal/backend"

	commonpb "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
)

type MockBackend struct {
	mu            sync.RWMutex
	version       int64
	namespaces    map[string]*sgroupsv1.Namespace
	addressGroups map[string]*sgroupsv1.AddressGroup
	networks      map[string]*sgroupsv1.Network
	hosts         map[string]*sgroupsv1.Host
	hostBindings  map[string]*sgroupsv1.HostBinding
	nsHub         *namespaceWatchHub
	agHub         *addressGroupWatchHub
	nwHub         *networkWatchHub
	hostHub       *hostWatchHub
	hbHub         *hostBindingWatchHub
}

func New() *MockBackend {
	return &MockBackend{
		namespaces:    make(map[string]*sgroupsv1.Namespace),
		addressGroups: make(map[string]*sgroupsv1.AddressGroup),
		networks:      make(map[string]*sgroupsv1.Network),
		hosts:         make(map[string]*sgroupsv1.Host),
		hostBindings:  make(map[string]*sgroupsv1.HostBinding),
		nsHub:         newNamespaceWatchHub(),
		agHub:         newAddressGroupWatchHub(),
		nwHub:         newNetworkWatchHub(),
		hostHub:       newHostWatchHub(),
		hbHub:         newHostBindingWatchHub(),
	}
}

func (m *MockBackend) UpsertNamespaces(ctx context.Context, req *sgroupsv1.NamespaceReq_Upsert) (*sgroupsv1.NamespaceResp_Upsert, error) {
	if req == nil || len(req.Namespaces) == 0 {
		return nil, errors.New("namespaces are required")
	}
	m.mu.Lock()
	added := make([]*sgroupsv1.Namespace, 0, len(req.Namespaces))
	modified := make([]*sgroupsv1.Namespace, 0, len(req.Namespaces))
	resp := &sgroupsv1.NamespaceResp_Upsert{Namespaces: make([]*sgroupsv1.Namespace, 0, len(req.Namespaces))}
	for _, ns := range req.Namespaces {
		if ns == nil || ns.Metadata == nil {
			m.mu.Unlock()

			return nil, errors.New("namespace metadata is required")
		}
		stored, isNew := m.upsertNamespaceLocked(ns)
		resp.Namespaces = append(resp.Namespaces, cloneNamespace(stored))
		if isNew {
			added = append(added, cloneNamespace(stored))
		} else {
			modified = append(modified, cloneNamespace(stored))
		}
	}
	m.mu.Unlock()

	if len(added) > 0 {
		m.nsHub.publish(&sgroupsv1.NamespaceResp_Watch{Type: commonpb.WatchEventType_ADDED, Namespaces: added})
	}
	if len(modified) > 0 {
		m.nsHub.publish(&sgroupsv1.NamespaceResp_Watch{Type: commonpb.WatchEventType_MODIFIED, Namespaces: modified})
	}

	return resp, nil
}

func (m *MockBackend) DeleteNamespaces(ctx context.Context, req *sgroupsv1.NamespaceReq_Delete) error {
	if req == nil || len(req.Namespaces) == 0 {
		return errors.New("namespaces are required")
	}
	m.mu.Lock()
	deleted := make([]*sgroupsv1.Namespace, 0, len(req.Namespaces))
	for _, ns := range req.Namespaces {
		if ns == nil || ns.Metadata == nil {
			m.mu.Unlock()

			return errors.New("namespace delete metadata is required")
		}
		uid := ns.Metadata.Uid
		name := ns.Metadata.Name
		if uid == "" && name == "" {
			m.mu.Unlock()

			return errors.New("namespace delete requires uid or name")
		}
		if uid != "" {
			if stored, ok := m.namespaces[uid]; ok {
				delete(m.namespaces, uid)
				deleted = append(deleted, cloneNamespace(stored))
			}

			continue
		}
		for id, stored := range m.namespaces {
			if stored.Metadata != nil && stored.Metadata.Name == name {
				delete(m.namespaces, id)
				deleted = append(deleted, cloneNamespace(stored))

				break
			}
		}
	}
	m.mu.Unlock()

	if len(deleted) > 0 {
		m.nsHub.publish(&sgroupsv1.NamespaceResp_Watch{Type: commonpb.WatchEventType_DELETED, Namespaces: deleted})
	}

	return nil
}

func (m *MockBackend) ListNamespaces(ctx context.Context, req *sgroupsv1.NamespaceReq_List) (*sgroupsv1.NamespaceResp_List, error) {
	if req == nil || len(req.Selectors) == 0 {
		return nil, errors.New("selectors are required")
	}
	m.mu.RLock()
	items := make([]*sgroupsv1.Namespace, 0, len(m.namespaces))
	for _, ns := range m.namespaces {
		items = append(items, ns)
	}
	m.mu.RUnlock()

	filtered := filterNamespaces(req.Selectors, items)

	return &sgroupsv1.NamespaceResp_List{
		ResourceVersion: strconv.FormatInt(atomic.LoadInt64(&m.version), 10),
		Namespaces:      filtered,
	}, nil
}

func (m *MockBackend) WatchNamespaces(ctx context.Context, req *sgroupsv1.NamespaceReq_Watch) (backend.WatchStream[*sgroupsv1.NamespaceResp_Watch], error) {
	if req == nil || len(req.Selectors) == 0 {
		return backend.WatchStream[*sgroupsv1.NamespaceResp_Watch]{}, errors.New("selectors are required")
	}
	ch, cancel := m.nsHub.subscribe(req.Selectors, 16)

	m.mu.RLock()
	items := make([]*sgroupsv1.Namespace, 0, len(m.namespaces))
	for _, ns := range m.namespaces {
		items = append(items, ns)
	}
	m.mu.RUnlock()

	snapshot := filterNamespaces(req.Selectors, items)
	ch <- &sgroupsv1.NamespaceResp_Watch{Type: commonpb.WatchEventType_ADDED, Namespaces: snapshot}

	return backend.WatchStream[*sgroupsv1.NamespaceResp_Watch]{
		C:     ch,
		Close: cancel,
	}, nil
}

func (m *MockBackend) UpsertAddressGroups(ctx context.Context, req *sgroupsv1.AddressGroupReq_Upsert) (*sgroupsv1.AddressGroupResp_Upsert, error) {
	if req == nil || len(req.AddressGroups) == 0 {
		return nil, errors.New("address groups are required")
	}
	m.mu.Lock()
	added := make([]*sgroupsv1.AddressGroup, 0, len(req.AddressGroups))
	modified := make([]*sgroupsv1.AddressGroup, 0, len(req.AddressGroups))
	resp := &sgroupsv1.AddressGroupResp_Upsert{AddressGroups: make([]*sgroupsv1.AddressGroup, 0, len(req.AddressGroups))}
	for _, ag := range req.AddressGroups {
		if ag == nil || ag.Metadata == nil {
			m.mu.Unlock()

			return nil, errors.New("address group metadata is required")
		}
		stored, isNew := m.upsertAddressGroupLocked(ag)
		resp.AddressGroups = append(resp.AddressGroups, cloneAddressGroup(stored))
		if isNew {
			added = append(added, cloneAddressGroup(stored))
		} else {
			modified = append(modified, cloneAddressGroup(stored))
		}
	}
	m.mu.Unlock()

	if len(added) > 0 {
		m.agHub.publish(&sgroupsv1.AddressGroupResp_Watch{Type: commonpb.WatchEventType_ADDED, AddressGroups: toAddressGroupExtList(added)})
	}
	if len(modified) > 0 {
		m.agHub.publish(&sgroupsv1.AddressGroupResp_Watch{Type: commonpb.WatchEventType_MODIFIED, AddressGroups: toAddressGroupExtList(modified)})
	}

	return resp, nil
}

func (m *MockBackend) DeleteAddressGroups(ctx context.Context, req *sgroupsv1.AddressGroupReq_Delete) error {
	if req == nil || len(req.AddressGroups) == 0 {
		return errors.New("address groups are required")
	}
	m.mu.Lock()
	deleted := make([]*sgroupsv1.AddressGroup, 0, len(req.AddressGroups))
	for _, ag := range req.AddressGroups {
		if ag == nil || ag.Metadata == nil {
			m.mu.Unlock()

			return errors.New("address group delete metadata is required")
		}
		uid := ag.Metadata.Uid
		name := ag.Metadata.Name
		ns := ag.Metadata.Namespace
		if uid == "" && (name == "" || ns == "") {
			m.mu.Unlock()

			return errors.New("address group delete requires uid or name+namespace")
		}
		if uid != "" {
			if stored, ok := m.addressGroups[uid]; ok {
				delete(m.addressGroups, uid)
				deleted = append(deleted, cloneAddressGroup(stored))
			}

			continue
		}
		for id, stored := range m.addressGroups {
			if stored.Metadata != nil && stored.Metadata.Name == name && stored.Metadata.Namespace == ns {
				delete(m.addressGroups, id)
				deleted = append(deleted, cloneAddressGroup(stored))

				break
			}
		}
	}
	m.mu.Unlock()

	if len(deleted) > 0 {
		m.agHub.publish(&sgroupsv1.AddressGroupResp_Watch{Type: commonpb.WatchEventType_DELETED, AddressGroups: toAddressGroupExtList(deleted)})
	}

	return nil
}

func (m *MockBackend) ListAddressGroups(ctx context.Context, req *sgroupsv1.AddressGroupReq_List) (*sgroupsv1.AddressGroupResp_List, error) {
	if req == nil || len(req.Selectors) == 0 {
		return nil, errors.New("selectors are required")
	}
	m.mu.RLock()
	items := make([]*sgroupsv1.AddressGroup, 0, len(m.addressGroups))
	for _, ag := range m.addressGroups {
		items = append(items, ag)
	}
	m.mu.RUnlock()

	filtered := filterAddressGroups(req.Selectors, items)

	return &sgroupsv1.AddressGroupResp_List{
		ResourceVersion: strconv.FormatInt(atomic.LoadInt64(&m.version), 10),
		AddressGroups:   toAddressGroupExtList(filtered),
	}, nil
}

func (m *MockBackend) WatchAddressGroups(ctx context.Context, req *sgroupsv1.AddressGroupReq_Watch) (backend.WatchStream[*sgroupsv1.AddressGroupResp_Watch], error) {
	if req == nil || len(req.Selectors) == 0 {
		return backend.WatchStream[*sgroupsv1.AddressGroupResp_Watch]{}, errors.New("selectors are required")
	}
	ch, cancel := m.agHub.subscribe(req.Selectors, 16)

	m.mu.RLock()
	items := make([]*sgroupsv1.AddressGroup, 0, len(m.addressGroups))
	for _, ag := range m.addressGroups {
		items = append(items, ag)
	}
	m.mu.RUnlock()

	snapshot := filterAddressGroups(req.Selectors, items)
	ch <- &sgroupsv1.AddressGroupResp_Watch{Type: commonpb.WatchEventType_ADDED, AddressGroups: toAddressGroupExtList(snapshot)}

	return backend.WatchStream[*sgroupsv1.AddressGroupResp_Watch]{
		C:     ch,
		Close: cancel,
	}, nil
}

func (m *MockBackend) UpsertNetworks(ctx context.Context, req *sgroupsv1.NetworkReq_Upsert) (*sgroupsv1.NetworkResp_Upsert, error) {
	if req == nil || len(req.Networks) == 0 {
		return nil, errors.New("networks are required")
	}
	m.mu.Lock()
	added := make([]*sgroupsv1.Network, 0, len(req.Networks))
	modified := make([]*sgroupsv1.Network, 0, len(req.Networks))
	resp := &sgroupsv1.NetworkResp_Upsert{Networks: make([]*sgroupsv1.Network, 0, len(req.Networks))}
	for _, nw := range req.Networks {
		if nw == nil || nw.Metadata == nil {
			m.mu.Unlock()

			return nil, errors.New("network metadata is required")
		}
		stored, isNew := m.upsertNetworkLocked(nw)
		resp.Networks = append(resp.Networks, cloneNetwork(stored))
		if isNew {
			added = append(added, cloneNetwork(stored))
		} else {
			modified = append(modified, cloneNetwork(stored))
		}
	}
	m.mu.Unlock()

	if len(added) > 0 {
		m.nwHub.publish(&sgroupsv1.NetworkResp_Watch{Type: commonpb.WatchEventType_ADDED, Networks: toNetworkExtList(added)})
	}
	if len(modified) > 0 {
		m.nwHub.publish(&sgroupsv1.NetworkResp_Watch{Type: commonpb.WatchEventType_MODIFIED, Networks: toNetworkExtList(modified)})
	}

	return resp, nil
}

func (m *MockBackend) DeleteNetworks(ctx context.Context, req *sgroupsv1.NetworkReq_Delete) error {
	if req == nil || len(req.Networks) == 0 {
		return errors.New("networks are required")
	}
	m.mu.Lock()
	deleted := make([]*sgroupsv1.Network, 0, len(req.Networks))
	for _, nw := range req.Networks {
		if nw == nil || nw.Metadata == nil {
			m.mu.Unlock()

			return errors.New("network delete metadata is required")
		}
		uid := nw.Metadata.Uid
		name := nw.Metadata.Name
		ns := nw.Metadata.Namespace
		if uid == "" && (name == "" || ns == "") {
			m.mu.Unlock()

			return errors.New("network delete requires uid or name+namespace")
		}
		if uid != "" {
			if stored, ok := m.networks[uid]; ok {
				delete(m.networks, uid)
				deleted = append(deleted, cloneNetwork(stored))
			}

			continue
		}
		for id, stored := range m.networks {
			if stored.Metadata != nil && stored.Metadata.Name == name && stored.Metadata.Namespace == ns {
				delete(m.networks, id)
				deleted = append(deleted, cloneNetwork(stored))

				break
			}
		}
	}
	m.mu.Unlock()

	if len(deleted) > 0 {
		m.nwHub.publish(&sgroupsv1.NetworkResp_Watch{Type: commonpb.WatchEventType_DELETED, Networks: toNetworkExtList(deleted)})
	}

	return nil
}

func (m *MockBackend) ListNetworks(ctx context.Context, req *sgroupsv1.NetworkReq_List) (*sgroupsv1.NetworkResp_List, error) {
	if req == nil || len(req.Selectors) == 0 {
		return nil, errors.New("selectors are required")
	}
	m.mu.RLock()
	items := make([]*sgroupsv1.Network, 0, len(m.networks))
	for _, nw := range m.networks {
		items = append(items, nw)
	}
	m.mu.RUnlock()

	filtered := filterNetworks(req.Selectors, items)

	return &sgroupsv1.NetworkResp_List{
		ResourceVersion: strconv.FormatInt(atomic.LoadInt64(&m.version), 10),
		Networks:        toNetworkExtList(filtered),
	}, nil
}

func (m *MockBackend) WatchNetworks(ctx context.Context, req *sgroupsv1.NetworkReq_Watch) (backend.WatchStream[*sgroupsv1.NetworkResp_Watch], error) {
	if req == nil || len(req.Selectors) == 0 {
		return backend.WatchStream[*sgroupsv1.NetworkResp_Watch]{}, errors.New("selectors are required")
	}
	ch, cancel := m.nwHub.subscribe(req.Selectors, 16)

	m.mu.RLock()
	items := make([]*sgroupsv1.Network, 0, len(m.networks))
	for _, nw := range m.networks {
		items = append(items, nw)
	}
	m.mu.RUnlock()

	snapshot := filterNetworks(req.Selectors, items)
	ch <- &sgroupsv1.NetworkResp_Watch{Type: commonpb.WatchEventType_ADDED, Networks: toNetworkExtList(snapshot)}

	return backend.WatchStream[*sgroupsv1.NetworkResp_Watch]{
		C:     ch,
		Close: cancel,
	}, nil
}

func (m *MockBackend) UpsertHosts(ctx context.Context, req *sgroupsv1.HostReq_Upsert) (*sgroupsv1.HostResp_Upsert, error) {
	if req == nil || len(req.Hosts) == 0 {
		return nil, errors.New("hosts are required")
	}
	m.mu.Lock()
	added := make([]*sgroupsv1.Host, 0, len(req.Hosts))
	modified := make([]*sgroupsv1.Host, 0, len(req.Hosts))
	resp := &sgroupsv1.HostResp_Upsert{Hosts: make([]*sgroupsv1.Host, 0, len(req.Hosts))}
	for _, h := range req.Hosts {
		if h == nil || h.Metadata == nil {
			m.mu.Unlock()

			return nil, errors.New("host metadata is required")
		}
		stored, isNew := m.upsertHostLocked(h)
		resp.Hosts = append(resp.Hosts, cloneHost(stored))
		if isNew {
			added = append(added, cloneHost(stored))
		} else {
			modified = append(modified, cloneHost(stored))
		}
	}
	m.mu.Unlock()

	if len(added) > 0 {
		m.hostHub.publish(&sgroupsv1.HostResp_Watch{Type: commonpb.WatchEventType_ADDED, Hosts: toHostExtList(added)})
	}
	if len(modified) > 0 {
		m.hostHub.publish(&sgroupsv1.HostResp_Watch{Type: commonpb.WatchEventType_MODIFIED, Hosts: toHostExtList(modified)})
	}

	return resp, nil
}

func (m *MockBackend) DeleteHosts(ctx context.Context, req *sgroupsv1.HostReq_Delete) error {
	if req == nil || len(req.Hosts) == 0 {
		return errors.New("hosts are required")
	}
	m.mu.Lock()
	deleted := make([]*sgroupsv1.Host, 0, len(req.Hosts))
	for _, h := range req.Hosts {
		if h == nil || h.Metadata == nil {
			m.mu.Unlock()

			return errors.New("host delete metadata is required")
		}
		uid := h.Metadata.Uid
		name := h.Metadata.Name
		ns := h.Metadata.Namespace
		if uid == "" && (name == "" || ns == "") {
			m.mu.Unlock()

			return errors.New("host delete requires uid or name+namespace")
		}
		if uid != "" {
			if stored, ok := m.hosts[uid]; ok {
				delete(m.hosts, uid)
				deleted = append(deleted, cloneHost(stored))
			}

			continue
		}
		for id, stored := range m.hosts {
			if stored.Metadata != nil && stored.Metadata.Name == name && stored.Metadata.Namespace == ns {
				delete(m.hosts, id)
				deleted = append(deleted, cloneHost(stored))

				break
			}
		}
	}
	m.mu.Unlock()

	if len(deleted) > 0 {
		m.hostHub.publish(&sgroupsv1.HostResp_Watch{Type: commonpb.WatchEventType_DELETED, Hosts: toHostExtList(deleted)})
	}

	return nil
}

func (m *MockBackend) ListHosts(ctx context.Context, req *sgroupsv1.HostReq_List) (*sgroupsv1.HostResp_List, error) {
	if req == nil || len(req.Selectors) == 0 {
		return nil, errors.New("selectors are required")
	}
	m.mu.RLock()
	items := make([]*sgroupsv1.Host, 0, len(m.hosts))
	for _, h := range m.hosts {
		items = append(items, h)
	}
	m.mu.RUnlock()

	filtered := filterHosts(req.Selectors, items)

	return &sgroupsv1.HostResp_List{
		ResourceVersion: strconv.FormatInt(atomic.LoadInt64(&m.version), 10),
		Hosts:           toHostExtList(filtered),
	}, nil
}

func (m *MockBackend) WatchHosts(ctx context.Context, req *sgroupsv1.HostReq_Watch) (backend.WatchStream[*sgroupsv1.HostResp_Watch], error) {
	if req == nil || len(req.Selectors) == 0 {
		return backend.WatchStream[*sgroupsv1.HostResp_Watch]{}, errors.New("selectors are required")
	}
	ch, cancel := m.hostHub.subscribe(req.Selectors, 16)

	m.mu.RLock()
	items := make([]*sgroupsv1.Host, 0, len(m.hosts))
	for _, h := range m.hosts {
		items = append(items, h)
	}
	m.mu.RUnlock()

	snapshot := filterHosts(req.Selectors, items)
	ch <- &sgroupsv1.HostResp_Watch{Type: commonpb.WatchEventType_ADDED, Hosts: toHostExtList(snapshot)}

	return backend.WatchStream[*sgroupsv1.HostResp_Watch]{
		C:     ch,
		Close: cancel,
	}, nil
}

func (m *MockBackend) UpsertHostBindings(ctx context.Context, req *sgroupsv1.HostBindingReq_Upsert) (*sgroupsv1.HostBindingResp_Upsert, error) {
	if req == nil || len(req.HostBindings) == 0 {
		return nil, errors.New("host bindings are required")
	}
	m.mu.Lock()
	added := make([]*sgroupsv1.HostBinding, 0, len(req.HostBindings))
	modified := make([]*sgroupsv1.HostBinding, 0, len(req.HostBindings))
	resp := &sgroupsv1.HostBindingResp_Upsert{HostBindings: make([]*sgroupsv1.HostBinding, 0, len(req.HostBindings))}
	for _, hb := range req.HostBindings {
		if hb == nil || hb.Metadata == nil {
			m.mu.Unlock()

			return nil, errors.New("host binding metadata is required")
		}
		stored, isNew := m.upsertHostBindingLocked(hb)
		resp.HostBindings = append(resp.HostBindings, cloneHostBinding(stored))
		if isNew {
			added = append(added, cloneHostBinding(stored))
		} else {
			modified = append(modified, cloneHostBinding(stored))
		}
	}
	m.mu.Unlock()

	if len(added) > 0 {
		m.hbHub.publish(&sgroupsv1.HostBindingResp_Watch{Type: commonpb.WatchEventType_ADDED, HostBindings: added})
	}
	if len(modified) > 0 {
		m.hbHub.publish(&sgroupsv1.HostBindingResp_Watch{Type: commonpb.WatchEventType_MODIFIED, HostBindings: modified})
	}

	return resp, nil
}

func (m *MockBackend) DeleteHostBindings(ctx context.Context, req *sgroupsv1.HostBindingReq_Delete) error {
	if req == nil || len(req.HostBindings) == 0 {
		return errors.New("host bindings are required")
	}
	m.mu.Lock()
	deleted := make([]*sgroupsv1.HostBinding, 0, len(req.HostBindings))
	for _, hb := range req.HostBindings {
		if hb == nil || hb.Metadata == nil {
			m.mu.Unlock()

			return errors.New("host binding delete metadata is required")
		}
		uid := hb.Metadata.Uid
		name := hb.Metadata.Name
		ns := hb.Metadata.Namespace
		if uid == "" && (name == "" || ns == "") {
			m.mu.Unlock()

			return errors.New("host binding delete requires uid or name+namespace")
		}
		if uid != "" {
			if stored, ok := m.hostBindings[uid]; ok {
				delete(m.hostBindings, uid)
				deleted = append(deleted, cloneHostBinding(stored))
			}

			continue
		}
		for id, stored := range m.hostBindings {
			if stored.Metadata != nil && stored.Metadata.Name == name && stored.Metadata.Namespace == ns {
				delete(m.hostBindings, id)
				deleted = append(deleted, cloneHostBinding(stored))

				break
			}
		}
	}
	m.mu.Unlock()

	if len(deleted) > 0 {
		m.hbHub.publish(&sgroupsv1.HostBindingResp_Watch{Type: commonpb.WatchEventType_DELETED, HostBindings: deleted})
	}

	return nil
}

func (m *MockBackend) ListHostBindings(ctx context.Context, req *sgroupsv1.HostBindingReq_List) (*sgroupsv1.HostBindingResp_List, error) {
	if req == nil || len(req.Selectors) == 0 {
		return nil, errors.New("selectors are required")
	}
	m.mu.RLock()
	items := make([]*sgroupsv1.HostBinding, 0, len(m.hostBindings))
	for _, hb := range m.hostBindings {
		items = append(items, hb)
	}
	m.mu.RUnlock()

	filtered := filterHostBindings(req.Selectors, items)

	return &sgroupsv1.HostBindingResp_List{
		ResourceVersion: strconv.FormatInt(atomic.LoadInt64(&m.version), 10),
		HostBindings:    filtered,
	}, nil
}

func (m *MockBackend) WatchHostBindings(ctx context.Context, req *sgroupsv1.HostBindingReq_Watch) (backend.WatchStream[*sgroupsv1.HostBindingResp_Watch], error) {
	if req == nil || len(req.Selectors) == 0 {
		return backend.WatchStream[*sgroupsv1.HostBindingResp_Watch]{}, errors.New("selectors are required")
	}
	ch, cancel := m.hbHub.subscribe(req.Selectors, 16)

	m.mu.RLock()
	items := make([]*sgroupsv1.HostBinding, 0, len(m.hostBindings))
	for _, hb := range m.hostBindings {
		items = append(items, hb)
	}
	m.mu.RUnlock()

	snapshot := filterHostBindings(req.Selectors, items)
	ch <- &sgroupsv1.HostBindingResp_Watch{Type: commonpb.WatchEventType_ADDED, HostBindings: snapshot}

	return backend.WatchStream[*sgroupsv1.HostBindingResp_Watch]{
		C:     ch,
		Close: cancel,
	}, nil
}

func (m *MockBackend) upsertNamespaceLocked(ns *sgroupsv1.Namespace) (*sgroupsv1.Namespace, bool) {
	clone := cloneNamespace(ns)
	uid := clone.Metadata.Uid
	if uid == "" {
		uid = uuid.NewString()
		clone.Metadata.Uid = uid
	}
	existing, ok := m.namespaces[uid]
	if ok && existing != nil && existing.Metadata != nil {
		clone.Metadata.CreationTimestamp = existing.Metadata.CreationTimestamp
	} else if clone.Metadata.CreationTimestamp == nil {
		clone.Metadata.CreationTimestamp = timestamppb.Now()
	}
	clone.Metadata.ResourceVersion = strconv.FormatInt(atomic.AddInt64(&m.version, 1), 10)
	m.namespaces[uid] = clone

	return clone, !ok
}

func (m *MockBackend) upsertAddressGroupLocked(ag *sgroupsv1.AddressGroup) (*sgroupsv1.AddressGroup, bool) {
	clone := cloneAddressGroup(ag)
	uid := clone.Metadata.Uid
	if uid == "" {
		uid = uuid.NewString()
		clone.Metadata.Uid = uid
	}
	existing, ok := m.addressGroups[uid]
	if ok && existing != nil && existing.Metadata != nil {
		clone.Metadata.CreationTimestamp = existing.Metadata.CreationTimestamp
	} else if clone.Metadata.CreationTimestamp == nil {
		clone.Metadata.CreationTimestamp = timestamppb.Now()
	}
	clone.Metadata.ResourceVersion = strconv.FormatInt(atomic.AddInt64(&m.version, 1), 10)
	m.addressGroups[uid] = clone

	return clone, !ok
}

func (m *MockBackend) upsertHostLocked(h *sgroupsv1.Host) (*sgroupsv1.Host, bool) {
	clone := cloneHost(h)
	uid := clone.Metadata.Uid
	if uid == "" {
		uid = uuid.NewString()
		clone.Metadata.Uid = uid
	}
	existing, ok := m.hosts[uid]
	if ok && existing != nil && existing.Metadata != nil {
		clone.Metadata.CreationTimestamp = existing.Metadata.CreationTimestamp
	} else if clone.Metadata.CreationTimestamp == nil {
		clone.Metadata.CreationTimestamp = timestamppb.Now()
	}
	clone.Metadata.ResourceVersion = strconv.FormatInt(atomic.AddInt64(&m.version, 1), 10)
	m.hosts[uid] = clone

	return clone, !ok
}

func (m *MockBackend) upsertHostBindingLocked(hb *sgroupsv1.HostBinding) (*sgroupsv1.HostBinding, bool) {
	clone := cloneHostBinding(hb)
	uid := clone.Metadata.Uid
	if uid == "" {
		uid = uuid.NewString()
		clone.Metadata.Uid = uid
	}
	existing, ok := m.hostBindings[uid]
	if ok && existing != nil && existing.Metadata != nil {
		clone.Metadata.CreationTimestamp = existing.Metadata.CreationTimestamp
	} else if clone.Metadata.CreationTimestamp == nil {
		clone.Metadata.CreationTimestamp = timestamppb.Now()
	}
	clone.Metadata.ResourceVersion = strconv.FormatInt(atomic.AddInt64(&m.version, 1), 10)
	m.hostBindings[uid] = clone

	return clone, !ok
}

func (m *MockBackend) upsertNetworkLocked(nw *sgroupsv1.Network) (*sgroupsv1.Network, bool) {
	clone := cloneNetwork(nw)
	uid := clone.Metadata.Uid
	if uid == "" {
		uid = uuid.NewString()
		clone.Metadata.Uid = uid
	}
	existing, ok := m.networks[uid]
	if ok && existing != nil && existing.Metadata != nil {
		clone.Metadata.CreationTimestamp = existing.Metadata.CreationTimestamp
	} else if clone.Metadata.CreationTimestamp == nil {
		clone.Metadata.CreationTimestamp = timestamppb.Now()
	}
	clone.Metadata.ResourceVersion = strconv.FormatInt(atomic.AddInt64(&m.version, 1), 10)
	m.networks[uid] = clone

	return clone, !ok
}

func cloneNamespace(ns *sgroupsv1.Namespace) *sgroupsv1.Namespace {
	if ns == nil {
		return nil
	}

	return proto.Clone(ns).(*sgroupsv1.Namespace) //nolint:forcetypeassert,errcheck // proto.Clone preserves concrete type
}

func cloneAddressGroup(ag *sgroupsv1.AddressGroup) *sgroupsv1.AddressGroup {
	if ag == nil {
		return nil
	}

	return proto.Clone(ag).(*sgroupsv1.AddressGroup) //nolint:forcetypeassert,errcheck // proto.Clone preserves concrete type
}

func toAddressGroupExtList(items []*sgroupsv1.AddressGroup) []*sgroupsv1.AddressGroupResp_AddressGroupExt {
	out := make([]*sgroupsv1.AddressGroupResp_AddressGroupExt, 0, len(items))
	for _, ag := range items {
		if ag == nil {
			continue
		}
		out = append(out, &sgroupsv1.AddressGroupResp_AddressGroupExt{
			Metadata: ag.GetMetadata(),
			Spec:     ag.GetSpec(),
		})
	}

	return out
}

func fromAddressGroupExtList(items []*sgroupsv1.AddressGroupResp_AddressGroupExt) []*sgroupsv1.AddressGroup {
	out := make([]*sgroupsv1.AddressGroup, 0, len(items))
	for _, ag := range items {
		if ag == nil {
			continue
		}
		out = append(out, &sgroupsv1.AddressGroup{
			Metadata: ag.GetMetadata(),
			Spec:     ag.GetSpec(),
		})
	}

	return out
}

func cloneNetwork(nw *sgroupsv1.Network) *sgroupsv1.Network {
	if nw == nil {
		return nil
	}

	return proto.Clone(nw).(*sgroupsv1.Network) //nolint:forcetypeassert,errcheck // proto.Clone preserves concrete type
}

func toNetworkExtList(items []*sgroupsv1.Network) []*sgroupsv1.NetworkResp_NetworkExt {
	out := make([]*sgroupsv1.NetworkResp_NetworkExt, 0, len(items))
	for _, nw := range items {
		if nw == nil {
			continue
		}
		out = append(out, &sgroupsv1.NetworkResp_NetworkExt{
			Metadata: nw.GetMetadata(),
			Spec:     nw.GetSpec(),
		})
	}

	return out
}

func cloneHost(h *sgroupsv1.Host) *sgroupsv1.Host {
	if h == nil {
		return nil
	}

	return proto.Clone(h).(*sgroupsv1.Host) //nolint:forcetypeassert,errcheck // proto.Clone preserves concrete type
}

func cloneHostBinding(hb *sgroupsv1.HostBinding) *sgroupsv1.HostBinding {
	if hb == nil {
		return nil
	}

	return proto.Clone(hb).(*sgroupsv1.HostBinding) //nolint:forcetypeassert,errcheck // proto.Clone preserves concrete type
}

func toHostExtList(items []*sgroupsv1.Host) []*sgroupsv1.HostResp_HostExt {
	out := make([]*sgroupsv1.HostResp_HostExt, 0, len(items))
	for _, h := range items {
		if h == nil {
			continue
		}
		out = append(out, &sgroupsv1.HostResp_HostExt{
			Metadata: h.GetMetadata(),
			Spec:     h.GetSpec(),
		})
	}

	return out
}

func fromHostExtList(items []*sgroupsv1.HostResp_HostExt) []*sgroupsv1.Host {
	out := make([]*sgroupsv1.Host, 0, len(items))
	for _, h := range items {
		if h == nil {
			continue
		}
		out = append(out, &sgroupsv1.Host{
			Metadata: h.GetMetadata(),
			Spec:     h.GetSpec(),
		})
	}

	return out
}

func filterHosts(selectors []*commonpb.ResSelector, items []*sgroupsv1.Host) []*sgroupsv1.Host {
	if len(selectors) == 0 {
		return nil
	}
	result := make([]*sgroupsv1.Host, 0)
	for _, h := range items {
		if h == nil || h.Metadata == nil {
			continue
		}
		if matchHostSelectors(h, selectors) {
			result = append(result, cloneHost(h))
		}
	}

	return result
}

func matchHostSelectors(h *sgroupsv1.Host, selectors []*commonpb.ResSelector) bool {
	for _, sel := range selectors {
		if sel == nil {
			continue
		}
		if matchHostSelector(h, sel) {
			return true
		}
	}

	return false
}

func matchHostSelector(h *sgroupsv1.Host, sel *commonpb.ResSelector) bool {
	if sel == nil {
		return false
	}
	if fs := sel.FieldSelector; fs != nil {
		if fs.Name != "" && h.Metadata.Name != fs.Name {
			return false
		}
		if fs.Namespace != "" && h.Metadata.Namespace != fs.Namespace {
			return false
		}
		if len(fs.Refs) > 0 {
			matched := false
			for _, ref := range fs.Refs {
				if ref == nil {
					continue
				}
				if ref.Name != "" && h.Metadata.Name != ref.Name {
					continue
				}
				if ref.Namespace != "" && h.Metadata.Namespace != ref.Namespace {
					continue
				}
				matched = true

				break
			}
			if !matched {
				return false
			}
		}
	}
	if len(sel.LabelSelector) > 0 {
		if !matchLabels(h.Metadata.Labels, sel.LabelSelector) {
			return false
		}
	}

	return true
}

func filterHostBindings(selectors []*sgroupsv1.HostBindingReq_Selectors, items []*sgroupsv1.HostBinding) []*sgroupsv1.HostBinding {
	if len(selectors) == 0 {
		return nil
	}
	result := make([]*sgroupsv1.HostBinding, 0)
	for _, hb := range items {
		if hb == nil || hb.Metadata == nil {
			continue
		}
		if matchHostBindingSelectors(hb, selectors) {
			result = append(result, cloneHostBinding(hb))
		}
	}

	return result
}

func matchHostBindingSelectors(hb *sgroupsv1.HostBinding, selectors []*sgroupsv1.HostBindingReq_Selectors) bool {
	for _, sel := range selectors {
		if sel == nil {
			continue
		}
		if matchHostBindingSelector(hb, sel) {
			return true
		}
	}

	return false
}

func matchHostBindingSelector(hb *sgroupsv1.HostBinding, sel *sgroupsv1.HostBindingReq_Selectors) bool {
	if sel == nil {
		return false
	}
	if fs := sel.FieldSelector; fs != nil {
		if fs.Name != "" && hb.Metadata.Name != fs.Name {
			return false
		}
		if fs.Namespace != "" && hb.Metadata.Namespace != fs.Namespace {
			return false
		}
		if fs.AddressGroup != nil {
			ag := hb.GetSpec().GetAddressGroup()
			if ag == nil {
				return false
			}
			if fs.AddressGroup.Name != "" && ag.Name != fs.AddressGroup.Name {
				return false
			}
			if fs.AddressGroup.Namespace != "" && ag.Namespace != fs.AddressGroup.Namespace {
				return false
			}
		}
		if fs.Host != nil {
			h := hb.GetSpec().GetHost()
			if h == nil {
				return false
			}
			if fs.Host.Name != "" && h.Name != fs.Host.Name {
				return false
			}
			if fs.Host.Namespace != "" && h.Namespace != fs.Host.Namespace {
				return false
			}
		}
	}
	if len(sel.LabelSelector) > 0 {
		if !matchLabels(hb.Metadata.Labels, sel.LabelSelector) {
			return false
		}
	}

	return true
}

func filterNamespaces(selectors []*sgroupsv1.NamespaceReq_Selector, items []*sgroupsv1.Namespace) []*sgroupsv1.Namespace {
	if len(selectors) == 0 {
		return nil
	}
	result := make([]*sgroupsv1.Namespace, 0)
	for _, ns := range items {
		if ns == nil || ns.Metadata == nil {
			continue
		}
		if matchNamespaceSelectors(ns, selectors) {
			result = append(result, cloneNamespace(ns))
		}
	}

	return result
}

func matchNamespaceSelectors(ns *sgroupsv1.Namespace, selectors []*sgroupsv1.NamespaceReq_Selector) bool {
	for _, sel := range selectors {
		if sel == nil {
			continue
		}
		if matchNamespaceSelector(ns, sel) {
			return true
		}
	}

	return false
}

func matchNamespaceSelector(ns *sgroupsv1.Namespace, sel *sgroupsv1.NamespaceReq_Selector) bool {
	if sel == nil {
		return false
	}
	if fs := sel.FieldSelector; fs != nil && fs.Name != "" {
		if ns.Metadata == nil || ns.Metadata.Name != fs.Name {
			return false
		}
	}
	if len(sel.LabelSelector) > 0 {
		if ns.Metadata == nil || !matchLabels(ns.Metadata.Labels, sel.LabelSelector) {
			return false
		}
	}

	return true
}

func filterAddressGroups(selectors []*commonpb.ResSelector, items []*sgroupsv1.AddressGroup) []*sgroupsv1.AddressGroup {
	if len(selectors) == 0 {
		return nil
	}
	result := make([]*sgroupsv1.AddressGroup, 0)
	for _, ag := range items {
		if ag == nil || ag.Metadata == nil {
			continue
		}
		if matchAddressGroupSelectors(ag, selectors) {
			result = append(result, cloneAddressGroup(ag))
		}
	}

	return result
}

func matchAddressGroupSelectors(ag *sgroupsv1.AddressGroup, selectors []*commonpb.ResSelector) bool {
	for _, sel := range selectors {
		if sel == nil {
			continue
		}
		if matchAddressGroupSelector(ag, sel) {
			return true
		}
	}

	return false
}

func matchAddressGroupSelector(ag *sgroupsv1.AddressGroup, sel *commonpb.ResSelector) bool {
	if sel == nil {
		return false
	}
	if fs := sel.FieldSelector; fs != nil {
		if fs.Name != "" && ag.Metadata.Name != fs.Name {
			return false
		}
		if fs.Namespace != "" && ag.Metadata.Namespace != fs.Namespace {
			return false
		}
		if len(fs.Refs) > 0 {
			matched := false
			for _, ref := range fs.Refs {
				if ref == nil {
					continue
				}
				if ref.Name != "" && ag.Metadata.Name != ref.Name {
					continue
				}
				if ref.Namespace != "" && ag.Metadata.Namespace != ref.Namespace {
					continue
				}
				matched = true

				break
			}
			if !matched {
				return false
			}
		}
	}
	if len(sel.LabelSelector) > 0 {
		if !matchLabels(ag.Metadata.Labels, sel.LabelSelector) {
			return false
		}
	}

	return true
}

func filterNetworks(selectors []*commonpb.ResSelector, items []*sgroupsv1.Network) []*sgroupsv1.Network {
	if len(selectors) == 0 {
		return nil
	}
	result := make([]*sgroupsv1.Network, 0)
	for _, nw := range items {
		if nw == nil || nw.Metadata == nil {
			continue
		}
		if matchNetworkSelectors(nw, selectors) {
			result = append(result, cloneNetwork(nw))
		}
	}

	return result
}

func matchNetworkSelectors(nw *sgroupsv1.Network, selectors []*commonpb.ResSelector) bool {
	for _, sel := range selectors {
		if sel == nil {
			continue
		}
		if matchNetworkSelector(nw, sel) {
			return true
		}
	}

	return false
}

func matchNetworkSelector(nw *sgroupsv1.Network, sel *commonpb.ResSelector) bool {
	if sel == nil {
		return false
	}
	if fs := sel.FieldSelector; fs != nil {
		if fs.Name != "" && nw.Metadata.Name != fs.Name {
			return false
		}
		if fs.Namespace != "" && nw.Metadata.Namespace != fs.Namespace {
			return false
		}
		if len(fs.Refs) > 0 {
			matched := false
			for _, ref := range fs.Refs {
				if ref == nil {
					continue
				}
				if ref.Name != "" && nw.Metadata.Name != ref.Name {
					continue
				}
				if ref.Namespace != "" && nw.Metadata.Namespace != ref.Namespace {
					continue
				}
				matched = true

				break
			}
			if !matched {
				return false
			}
		}
	}
	if len(sel.LabelSelector) > 0 {
		if !matchLabels(nw.Metadata.Labels, sel.LabelSelector) {
			return false
		}
	}

	return true
}

func matchLabels(labels map[string]string, selector map[string]string) bool {
	for k, v := range selector {
		if labels == nil || labels[k] != v {
			return false
		}
	}

	return true
}

type namespaceWatchHub struct {
	mu     sync.Mutex
	nextID int64
	subs   map[int64]*namespaceWatchSub
}

type namespaceWatchSub struct {
	selectors []*sgroupsv1.NamespaceReq_Selector
	ch        chan *sgroupsv1.NamespaceResp_Watch
}

func newNamespaceWatchHub() *namespaceWatchHub {
	return &namespaceWatchHub{subs: make(map[int64]*namespaceWatchSub)}
}

func (h *namespaceWatchHub) subscribe(selectors []*sgroupsv1.NamespaceReq_Selector, buffer int) (chan *sgroupsv1.NamespaceResp_Watch, func()) {
	if buffer <= 0 {
		buffer = 1
	}
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	ch := make(chan *sgroupsv1.NamespaceResp_Watch, buffer)
	h.subs[id] = &namespaceWatchSub{selectors: selectors, ch: ch}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.subs, id)
		h.mu.Unlock()
	}
}

func (h *namespaceWatchHub) publish(event *sgroupsv1.NamespaceResp_Watch) {
	if event == nil {
		return
	}
	h.mu.Lock()
	subs := make([]*namespaceWatchSub, 0, len(h.subs))
	for _, sub := range h.subs {
		subs = append(subs, sub)
	}
	h.mu.Unlock()

	for _, sub := range subs {
		filtered := filterNamespaces(sub.selectors, event.Namespaces)
		if len(filtered) == 0 {
			continue
		}
		resp := &sgroupsv1.NamespaceResp_Watch{Type: event.Type, Namespaces: filtered}
		select {
		case sub.ch <- resp:
		default:
		}
	}
}

type addressGroupWatchHub struct {
	mu     sync.Mutex
	nextID int64
	subs   map[int64]*addressGroupWatchSub
}

type addressGroupWatchSub struct {
	selectors []*commonpb.ResSelector
	ch        chan *sgroupsv1.AddressGroupResp_Watch
}

func newAddressGroupWatchHub() *addressGroupWatchHub {
	return &addressGroupWatchHub{subs: make(map[int64]*addressGroupWatchSub)}
}

func (h *addressGroupWatchHub) subscribe(selectors []*commonpb.ResSelector, buffer int) (chan *sgroupsv1.AddressGroupResp_Watch, func()) {
	if buffer <= 0 {
		buffer = 1
	}
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	ch := make(chan *sgroupsv1.AddressGroupResp_Watch, buffer)
	h.subs[id] = &addressGroupWatchSub{selectors: selectors, ch: ch}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.subs, id)
		h.mu.Unlock()
	}
}

func (h *addressGroupWatchHub) publish(event *sgroupsv1.AddressGroupResp_Watch) {
	if event == nil {
		return
	}
	h.mu.Lock()
	subs := make([]*addressGroupWatchSub, 0, len(h.subs))
	for _, sub := range h.subs {
		subs = append(subs, sub)
	}
	h.mu.Unlock()

	for _, sub := range subs {
		filtered := filterAddressGroups(sub.selectors, fromAddressGroupExtList(event.AddressGroups))
		if len(filtered) == 0 {
			continue
		}
		resp := &sgroupsv1.AddressGroupResp_Watch{Type: event.Type, AddressGroups: toAddressGroupExtList(filtered)}
		select {
		case sub.ch <- resp:
		default:
		}
	}
}

type networkWatchHub struct {
	mu     sync.Mutex
	nextID int64
	subs   map[int64]*networkWatchSub
}

type networkWatchSub struct {
	selectors []*commonpb.ResSelector
	ch        chan *sgroupsv1.NetworkResp_Watch
}

func newNetworkWatchHub() *networkWatchHub {
	return &networkWatchHub{subs: make(map[int64]*networkWatchSub)}
}

func (h *networkWatchHub) subscribe(selectors []*commonpb.ResSelector, buffer int) (chan *sgroupsv1.NetworkResp_Watch, func()) {
	if buffer <= 0 {
		buffer = 1
	}
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	ch := make(chan *sgroupsv1.NetworkResp_Watch, buffer)
	h.subs[id] = &networkWatchSub{selectors: selectors, ch: ch}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.subs, id)
		h.mu.Unlock()
	}
}

func (h *networkWatchHub) publish(event *sgroupsv1.NetworkResp_Watch) {
	if event == nil {
		return
	}
	h.mu.Lock()
	subs := make([]*networkWatchSub, 0, len(h.subs))
	for _, sub := range h.subs {
		subs = append(subs, sub)
	}
	h.mu.Unlock()

	for _, sub := range subs {
		filtered := filterNetworks(sub.selectors, fromNetworkExtList(event.Networks))
		if len(filtered) == 0 {
			continue
		}
		resp := &sgroupsv1.NetworkResp_Watch{Type: event.Type, Networks: toNetworkExtList(filtered)}
		select {
		case sub.ch <- resp:
		default:
		}
	}
}

func fromNetworkExtList(items []*sgroupsv1.NetworkResp_NetworkExt) []*sgroupsv1.Network {
	out := make([]*sgroupsv1.Network, 0, len(items))
	for _, nw := range items {
		if nw == nil {
			continue
		}
		out = append(out, &sgroupsv1.Network{
			Metadata: nw.GetMetadata(),
			Spec:     nw.GetSpec(),
		})
	}

	return out
}

type hostWatchHub struct {
	mu     sync.Mutex
	nextID int64
	subs   map[int64]*hostWatchSub
}

type hostWatchSub struct {
	selectors []*commonpb.ResSelector
	ch        chan *sgroupsv1.HostResp_Watch
}

func newHostWatchHub() *hostWatchHub {
	return &hostWatchHub{subs: make(map[int64]*hostWatchSub)}
}

func (h *hostWatchHub) subscribe(selectors []*commonpb.ResSelector, buffer int) (chan *sgroupsv1.HostResp_Watch, func()) {
	if buffer <= 0 {
		buffer = 1
	}
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	ch := make(chan *sgroupsv1.HostResp_Watch, buffer)
	h.subs[id] = &hostWatchSub{selectors: selectors, ch: ch}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.subs, id)
		h.mu.Unlock()
	}
}

func (h *hostWatchHub) publish(event *sgroupsv1.HostResp_Watch) {
	if event == nil {
		return
	}
	h.mu.Lock()
	subs := make([]*hostWatchSub, 0, len(h.subs))
	for _, sub := range h.subs {
		subs = append(subs, sub)
	}
	h.mu.Unlock()

	for _, sub := range subs {
		filtered := filterHosts(sub.selectors, fromHostExtList(event.Hosts))
		if len(filtered) == 0 {
			continue
		}
		resp := &sgroupsv1.HostResp_Watch{Type: event.Type, Hosts: toHostExtList(filtered)}
		select {
		case sub.ch <- resp:
		default:
		}
	}
}

type hostBindingWatchHub struct {
	mu     sync.Mutex
	nextID int64
	subs   map[int64]*hostBindingWatchSub
}

type hostBindingWatchSub struct {
	selectors []*sgroupsv1.HostBindingReq_Selectors
	ch        chan *sgroupsv1.HostBindingResp_Watch
}

func newHostBindingWatchHub() *hostBindingWatchHub {
	return &hostBindingWatchHub{subs: make(map[int64]*hostBindingWatchSub)}
}

func (h *hostBindingWatchHub) subscribe(selectors []*sgroupsv1.HostBindingReq_Selectors, buffer int) (chan *sgroupsv1.HostBindingResp_Watch, func()) {
	if buffer <= 0 {
		buffer = 1
	}
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	ch := make(chan *sgroupsv1.HostBindingResp_Watch, buffer)
	h.subs[id] = &hostBindingWatchSub{selectors: selectors, ch: ch}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.subs, id)
		h.mu.Unlock()
	}
}

func (h *hostBindingWatchHub) publish(event *sgroupsv1.HostBindingResp_Watch) {
	if event == nil {
		return
	}
	h.mu.Lock()
	subs := make([]*hostBindingWatchSub, 0, len(h.subs))
	for _, sub := range h.subs {
		subs = append(subs, sub)
	}
	h.mu.Unlock()

	for _, sub := range subs {
		filtered := filterHostBindings(sub.selectors, event.HostBindings)
		if len(filtered) == 0 {
			continue
		}
		resp := &sgroupsv1.HostBindingResp_Watch{Type: event.Type, HostBindings: filtered}
		select {
		case sub.ch <- resp:
		default:
		}
	}
}
