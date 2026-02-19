package tenant

import (
	"context"
	"errors"

	common "github.com/PRO-Robotech/sgroups-proto/pkg/api/common"
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"

	"sgroups.io/sgroups-k8s-api/internal/registry/base"
	"sgroups.io/sgroups-k8s-api/internal/registry/convert"
	regerrors "sgroups.io/sgroups-k8s-api/internal/registry/errors"
	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
	"sgroups.io/sgroups-k8s-api/pkg/client"
)

type backend struct {
	client *client.Client
}

func (b *backend) NamespaceScoped() bool { return false }

func (b *backend) Resource() schema.GroupResource {
	return v1alpha1.Resource(v1alpha1.ResourceTenants)
}

func (b *backend) List(ctx context.Context, sel base.Selection) (*v1alpha1.TenantList, error) {
	selector := &sgroupsv1.NamespaceReq_Selector{
		LabelSelector: sel.Labels,
	}
	if sel.Name != "" {
		selector.FieldSelector = &sgroupsv1.NamespaceReq_Selector_FieldSelector{Name: sel.Name}
	}

	req := &sgroupsv1.NamespaceReq_List{Selectors: []*sgroupsv1.NamespaceReq_Selector{selector}}
	resp, err := b.client.Namespaces.List(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	out := &v1alpha1.TenantList{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindTenantList,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Items: make([]v1alpha1.Tenant, 0, len(resp.GetNamespaces())),
	}
	out.ResourceVersion = resp.GetResourceVersion()
	for _, ns := range resp.GetNamespaces() {
		if converted := convert.TenantFromProto(ns); converted != nil {
			out.Items = append(out.Items, *converted)
		}
	}

	return out, nil
}

func (b *backend) Upsert(ctx context.Context, obj *v1alpha1.Tenant) (*v1alpha1.Tenant, error) {
	req := &sgroupsv1.NamespaceReq_Upsert{
		Namespaces: []*sgroupsv1.Namespace{convert.TenantToProto(obj)},
	}
	resp, err := b.client.Namespaces.Upsert(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), obj.Name)
	}
	if len(resp.GetNamespaces()) == 0 {
		return nil, apierrors.NewInternalError(errors.New("empty upsert response"))
	}

	return convert.TenantFromProto(resp.GetNamespaces()[0]), nil
}

func (b *backend) Delete(ctx context.Context, name, _ string) error {
	req := &sgroupsv1.NamespaceReq_Delete{
		Namespaces: []*sgroupsv1.NamespaceReq_Delete_Namespace{
			{Metadata: &sgroupsv1.NamespaceReq_Delete_MetadataScope{Name: name}},
		},
	}
	_, err := b.client.Namespaces.Delete(ctx, req)
	if err != nil {
		return regerrors.FromGRPC(err, b.Resource(), name)
	}

	return nil
}

func (b *backend) Watch(ctx context.Context, sel base.Selection, resourceVersion string) (watch.Interface, error) {
	selector := &sgroupsv1.NamespaceReq_Selector{
		LabelSelector: sel.Labels,
	}
	if sel.Name != "" {
		selector.FieldSelector = &sgroupsv1.NamespaceReq_Selector_FieldSelector{Name: sel.Name}
	}

	req := &sgroupsv1.NamespaceReq_Watch{
		ResourceVersion: resourceVersion,
		Selectors:       []*sgroupsv1.NamespaceReq_Selector{selector},
	}
	ctx, cancel := context.WithCancel(ctx)
	stream, err := b.client.Namespaces.Watch(ctx, req)
	if err != nil {
		cancel()

		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	return base.NewStreamWatch(
		cancel,
		stream.Recv,
		func(evt *sgroupsv1.NamespaceResp_Watch) common.WatchEventType {
			return evt.GetType()
		},
		func(evt *sgroupsv1.NamespaceResp_Watch) []*sgroupsv1.Namespace {
			return evt.GetNamespaces()
		},
		func(ns *sgroupsv1.Namespace) runtime.Object {
			return convert.TenantFromProto(ns)
		},
	), nil
}
