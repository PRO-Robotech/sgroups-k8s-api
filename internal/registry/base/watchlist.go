package base

import (
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// WatchListWatch wraps a live-events watch.Interface and prepends synthetic
// ADDED events (from a List snapshot) followed by a BOOKMARK event,
// implementing the K8s WatchList protocol (sendInitialEvents=true).
type WatchListWatch struct {
	inner    watch.Interface
	resultCh chan watch.Event
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewWatchListWatch creates a watch.Interface that first sends ADDED events for
// each item in initialItems, then a BOOKMARK with the initial-events-end
// annotation, then forwards all events from inner.
func NewWatchListWatch(
	inner watch.Interface,
	initialItems []runtime.Object,
	newObj func() runtime.Object,
	listRV string,
) watch.Interface {
	w := &WatchListWatch{
		inner:    inner,
		resultCh: make(chan watch.Event, len(initialItems)+16),
		stopCh:   make(chan struct{}),
	}
	go w.run(initialItems, newObj, listRV)
	return w
}

// Stop terminates the watch and the inner watch.
func (w *WatchListWatch) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopCh)
		w.inner.Stop()
	})
}

// ResultChan returns the channel of watch events.
func (w *WatchListWatch) ResultChan() <-chan watch.Event {
	return w.resultCh
}

func (w *WatchListWatch) run(
	initialItems []runtime.Object,
	newObj func() runtime.Object,
	listRV string,
) {
	defer close(w.resultCh)

	// Phase 1: Send initial items as ADDED events.
	for _, item := range initialItems {
		select {
		case w.resultCh <- watch.Event{Type: watch.Added, Object: item}:
		case <-w.stopCh:
			return
		}
	}

	// Phase 2: Send BOOKMARK with initial-events-end annotation.
	bookmark := newObj()
	if a, err := meta.Accessor(bookmark); err == nil {
		a.SetResourceVersion(listRV)
		a.SetAnnotations(map[string]string{
			metav1.InitialEventsAnnotationKey: "true",
		})
	}
	select {
	case w.resultCh <- watch.Event{Type: watch.Bookmark, Object: bookmark}:
	case <-w.stopCh:
		return
	}

	// Phase 3: Forward live events from inner watch.
	innerCh := w.inner.ResultChan()
	for {
		select {
		case <-w.stopCh:
			return
		case evt, ok := <-innerCh:
			if !ok {
				return
			}
			select {
			case w.resultCh <- evt:
			case <-w.stopCh:
				return
			}
		}
	}
}
