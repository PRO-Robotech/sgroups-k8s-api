package addressgroup

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"

	appbackend "sgroups.io/sgroups-k8s-api/internal/backend"
	"sgroups.io/sgroups-k8s-api/internal/mock"
	registryoptions "sgroups.io/sgroups-k8s-api/internal/registry/options"
	"sgroups.io/sgroups-k8s-api/internal/registry/testutil"
	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

func TestAddressGroupStorageCreateList(t *testing.T) {
	mb := mock.New()
	b := appbackend.Backend{Namespaces: mb, AddressGroups: mb}
	cli, cleanup := testutil.NewBufconnClient(t, b)
	defer cleanup()

	store := NewStorage(cli, registryoptions.StorageOptions{})

	ctx := apirequest.WithNamespace(context.Background(), "default")
	obj := &v1alpha1.AddressGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "ag-1"},
		Spec: v1alpha1.AddressGroupSpec{
			DisplayName:   "AG",
			DefaultAction: v1alpha1.ActionAllow,
		},
	}

	created, err := store.Create(ctx, obj, nil, &metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	ag, ok := created.(*v1alpha1.AddressGroup)
	if !ok {
		t.Fatalf("unexpected type: %T", created)
	}
	if ag.Name != "ag-1" {
		t.Fatalf("unexpected name: %q", ag.Name)
	}
	if ag.Namespace != "default" {
		t.Fatalf("unexpected namespace: %q", ag.Namespace)
	}

	listObj, err := store.List(ctx, &metainternalversion.ListOptions{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	list, ok := listObj.(*v1alpha1.AddressGroupList)
	if !ok {
		t.Fatalf("unexpected list type: %T", listObj)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list.Items))
	}
}

func TestAddressGroupStorageWatch(t *testing.T) {
	mb := mock.New()
	b := appbackend.Backend{Namespaces: mb, AddressGroups: mb}
	cli, cleanup := testutil.NewBufconnClient(t, b)
	defer cleanup()

	store := NewStorage(cli, registryoptions.StorageOptions{})
	ctx := apirequest.WithNamespace(context.Background(), "default")

	// Create initial address group.
	_, err := store.Create(ctx, &v1alpha1.AddressGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "ag-1"},
		Spec: v1alpha1.AddressGroupSpec{
			DisplayName:   "AG 1",
			DefaultAction: v1alpha1.ActionAllow,
		},
	}, nil, &metav1.CreateOptions{})
	require.NoError(t, err)

	// List to get ResourceVersion for Watch resumption.
	listObj, err := store.List(ctx, &metainternalversion.ListOptions{})
	require.NoError(t, err)
	list := listObj.(*v1alpha1.AddressGroupList)
	rv := list.ResourceVersion
	require.NotEmpty(t, rv)

	// Start watch with ResourceVersion from List.
	watcher, err := store.Watch(ctx, &metainternalversion.ListOptions{ResourceVersion: rv})
	require.NoError(t, err)
	defer watcher.Stop()

	// Read initial snapshot (ADDED event for existing ag-1).
	var evt watch.Event
	require.Eventually(t, func() bool {
		select {
		case e, ok := <-watcher.ResultChan():
			if !ok {
				return false
			}
			evt = e

			return true
		default:
			return false
		}
	}, 5*time.Second, 50*time.Millisecond, "expected initial ADDED event")
	require.Equal(t, watch.Added, evt.Type)
	ag := evt.Object.(*v1alpha1.AddressGroup)
	require.Equal(t, "ag-1", ag.Name)

	// Create a second address group to trigger a watch event.
	_, err = store.Create(ctx, &v1alpha1.AddressGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "ag-2"},
		Spec: v1alpha1.AddressGroupSpec{
			DisplayName:   "AG 2",
			DefaultAction: v1alpha1.ActionDeny,
		},
	}, nil, &metav1.CreateOptions{})
	require.NoError(t, err)

	// Read the ADDED event for ag-2.
	require.Eventually(t, func() bool {
		select {
		case e, ok := <-watcher.ResultChan():
			if !ok {
				return false
			}
			evt = e

			return true
		default:
			return false
		}
	}, 5*time.Second, 50*time.Millisecond, "expected ADDED event for ag-2")
	require.Equal(t, watch.Added, evt.Type)
	ag = evt.Object.(*v1alpha1.AddressGroup)
	require.Equal(t, "ag-2", ag.Name)

	// Stop the watcher and verify channel is closed.
	watcher.Stop()
	require.Eventually(t, func() bool {
		_, ok := <-watcher.ResultChan()

		return !ok
	}, 5*time.Second, 50*time.Millisecond, "expected ResultChan to be closed after Stop")
}
