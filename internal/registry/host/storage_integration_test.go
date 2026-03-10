package host

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

func TestHostStorageCreateList(t *testing.T) {
	mb := mock.New()
	b := appbackend.Backend{Namespaces: mb, AddressGroups: mb, Networks: mb, Hosts: mb, HostBindings: mb}
	cli, cleanup := testutil.NewBufconnClient(t, b)
	defer cleanup()

	store := NewStorage(cli, registryoptions.StorageOptions{})

	ctx := apirequest.WithNamespace(context.Background(), "default")
	obj := &v1alpha1.Host{
		ObjectMeta: metav1.ObjectMeta{Name: "host-1"},
		Spec: v1alpha1.HostSpec{
			DisplayName: "Host 1",
		},
	}

	created, err := store.Create(ctx, obj, nil, &metav1.CreateOptions{})
	require.NoError(t, err)

	h, ok := created.(*v1alpha1.Host)
	require.True(t, ok)
	require.Equal(t, "host-1", h.Name)
	require.Equal(t, "default", h.Namespace)
	require.Equal(t, "Host 1", h.Spec.DisplayName)

	listObj, err := store.List(ctx, &metainternalversion.ListOptions{})
	require.NoError(t, err)
	list, ok := listObj.(*v1alpha1.HostList)
	require.True(t, ok)
	require.Len(t, list.Items, 1)
}

func TestHostStorageWatch(t *testing.T) {
	mb := mock.New()
	b := appbackend.Backend{Namespaces: mb, AddressGroups: mb, Networks: mb, Hosts: mb, HostBindings: mb}
	cli, cleanup := testutil.NewBufconnClient(t, b)
	defer cleanup()

	store := NewStorage(cli, registryoptions.StorageOptions{})
	ctx := apirequest.WithNamespace(context.Background(), "default")

	_, err := store.Create(ctx, &v1alpha1.Host{
		ObjectMeta: metav1.ObjectMeta{Name: "host-1"},
		Spec: v1alpha1.HostSpec{
			DisplayName: "Host 1",
		},
	}, nil, &metav1.CreateOptions{})
	require.NoError(t, err)

	listObj, err := store.List(ctx, &metainternalversion.ListOptions{})
	require.NoError(t, err)
	list := listObj.(*v1alpha1.HostList)
	rv := list.ResourceVersion
	require.NotEmpty(t, rv)

	watcher, err := store.Watch(ctx, &metainternalversion.ListOptions{ResourceVersion: rv})
	require.NoError(t, err)
	defer watcher.Stop()

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
	h := evt.Object.(*v1alpha1.Host)
	require.Equal(t, "host-1", h.Name)

	_, err = store.Create(ctx, &v1alpha1.Host{
		ObjectMeta: metav1.ObjectMeta{Name: "host-2"},
		Spec: v1alpha1.HostSpec{
			DisplayName: "Host 2",
		},
	}, nil, &metav1.CreateOptions{})
	require.NoError(t, err)

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
	}, 5*time.Second, 50*time.Millisecond, "expected ADDED event for host-2")
	require.Equal(t, watch.Added, evt.Type)
	h = evt.Object.(*v1alpha1.Host)
	require.Equal(t, "host-2", h.Name)

	watcher.Stop()
	require.Eventually(t, func() bool {
		_, ok := <-watcher.ResultChan()

		return !ok
	}, 5*time.Second, 50*time.Millisecond, "expected ResultChan to be closed after Stop")
}
