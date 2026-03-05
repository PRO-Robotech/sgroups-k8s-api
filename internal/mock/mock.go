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
	nsHub         *namespaceWatchHub
	agHub         *addressGroupWatchHub
}

func New() *MockBackend {
	return &MockBackend{
		namespaces:    make(map[string]*sgroupsv1.Namespace),
		addressGroups: make(map[string]*sgroupsv1.AddressGroup),
		nsHub:         newNamespaceWatchHub(),
		agHub:         newAddressGroupWatchHub(),
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
		m.agHub.publish(&sgroupsv1.AddressGroupResp_Watch{Type: commonpb.WatchEventType_ADDED, AddressGroups: added})
	}
	if len(modified) > 0 {
		m.agHub.publish(&sgroupsv1.AddressGroupResp_Watch{Type: commonpb.WatchEventType_MODIFIED, AddressGroups: modified})
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
		m.agHub.publish(&sgroupsv1.AddressGroupResp_Watch{Type: commonpb.WatchEventType_DELETED, AddressGroups: deleted})
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
	ext := make([]*sgroupsv1.AddressGroupResp_List_AddressGroupExt, 0, len(filtered))
	for _, ag := range filtered {
		ext = append(ext, &sgroupsv1.AddressGroupResp_List_AddressGroupExt{
			Metadata: ag.GetMetadata(),
			Spec:     ag.GetSpec(),
		})
	}

	return &sgroupsv1.AddressGroupResp_List{
		ResourceVersion: strconv.FormatInt(atomic.LoadInt64(&m.version), 10),
		AddressGroups:   ext,
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
	ch <- &sgroupsv1.AddressGroupResp_Watch{Type: commonpb.WatchEventType_ADDED, AddressGroups: snapshot}

	return backend.WatchStream[*sgroupsv1.AddressGroupResp_Watch]{
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
		filtered := filterAddressGroups(sub.selectors, event.AddressGroups)
		if len(filtered) == 0 {
			continue
		}
		resp := &sgroupsv1.AddressGroupResp_Watch{Type: event.Type, AddressGroups: filtered}
		select {
		case sub.ch <- resp:
		default:
		}
	}
}
