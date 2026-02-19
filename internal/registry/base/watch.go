package base

import (
	"context"

	common "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// ToWatchType converts proto watch event type to Kubernetes watch event type.
func ToWatchType(t common.WatchEventType) watch.EventType {
	switch t {
	case common.WatchEventType_ADDED:
		return watch.Added
	case common.WatchEventType_MODIFIED:
		return watch.Modified
	case common.WatchEventType_DELETED:
		return watch.Deleted
	default:
		return watch.Error
	}
}

// StreamWatch adapts a gRPC stream into a Kubernetes watch.Interface.
type StreamWatch[E any, T any] struct {
	resultCh chan watch.Event
	stopCh   chan struct{}
	cancel   context.CancelFunc
	recv     func() (*E, error)
	getType  func(*E) common.WatchEventType
	list     func(*E) []T
	convert  func(T) runtime.Object
}

// NewStreamWatch creates a watch.Interface from a gRPC stream.
func NewStreamWatch[E any, T any](
	cancel context.CancelFunc,
	recv func() (*E, error),
	getType func(*E) common.WatchEventType,
	list func(*E) []T,
	convert func(T) runtime.Object,
) watch.Interface {
	w := &StreamWatch[E, T]{
		resultCh: make(chan watch.Event, 16),
		stopCh:   make(chan struct{}),
		cancel:   cancel,
		recv:     recv,
		getType:  getType,
		list:     list,
		convert:  convert,
	}
	go w.run()

	return w
}

// Stop terminates the watch stream.
func (w *StreamWatch[E, T]) Stop() {
	select {
	case <-w.stopCh:
		return
	default:
		close(w.stopCh)
		if w.cancel != nil {
			w.cancel()
		}
	}
}

// ResultChan returns the watch event channel.
func (w *StreamWatch[E, T]) ResultChan() <-chan watch.Event {
	return w.resultCh
}

func (w *StreamWatch[E, T]) run() {
	defer close(w.resultCh)
	for {
		select {
		case <-w.stopCh:
			return
		default:
		}
		evt, err := w.recv()
		if err != nil {
			return
		}
		watchType := ToWatchType(w.getType(evt))
		for _, item := range w.list(evt) {
			obj := w.convert(item)
			if obj == nil {
				continue
			}
			w.resultCh <- watch.Event{Type: watchType, Object: obj}
		}
	}
}
