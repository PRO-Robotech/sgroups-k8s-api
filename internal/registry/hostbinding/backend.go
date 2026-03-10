package hostbinding

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

func (b *backend) NamespaceScoped() bool {
	return true
}

func (b *backend) Resource() schema.GroupResource {
	return v1alpha1.Resource(v1alpha1.ResourceHostBindings)
}

func (b *backend) List(ctx context.Context, sel base.Selection) (*v1alpha1.HostBindingList, error) {
	selector := &sgroupsv1.HostBindingReq_Selectors{
		LabelSelector: sel.Labels,
	}
	if sel.Name != "" || sel.Namespace != "" {
		selector.FieldSelector = &sgroupsv1.HostBindingReq_Selectors_FieldSelector{
			Name:      sel.Name,
			Namespace: sel.Namespace,
		}
	}

	req := &sgroupsv1.HostBindingReq_List{Selectors: []*sgroupsv1.HostBindingReq_Selectors{selector}}
	resp, err := b.client.HostBindings.List(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	out := &v1alpha1.HostBindingList{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindHostBindingList,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Items: make([]v1alpha1.HostBinding, 0, len(resp.GetHostBindings())),
	}
	out.ResourceVersion = resp.GetResourceVersion()
	for _, hb := range resp.GetHostBindings() {
		if converted := convert.HostBindingFromProto(hb); converted != nil {
			out.Items = append(out.Items, *converted)
		}
	}

	return out, nil
}

func (b *backend) Upsert(ctx context.Context, obj *v1alpha1.HostBinding) (*v1alpha1.HostBinding, error) {
	req := &sgroupsv1.HostBindingReq_Upsert{
		HostBindings: []*sgroupsv1.HostBinding{convert.HostBindingToProto(obj)},
	}
	resp, err := b.client.HostBindings.Upsert(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), obj.Name)
	}
	if len(resp.GetHostBindings()) == 0 {
		return nil, apierrors.NewInternalError(errors.New("empty upsert response"))
	}

	return convert.HostBindingFromProto(resp.GetHostBindings()[0]), nil
}

func (b *backend) Delete(ctx context.Context, name, namespace string) error {
	req := &sgroupsv1.HostBindingReq_Delete{
		HostBindings: []*sgroupsv1.HostBindingReq_Delete_HostBinding{
			{
				Metadata: &common.MetadataScope{
					Name:      name,
					Namespace: namespace,
				},
			},
		},
	}
	_, err := b.client.HostBindings.Delete(ctx, req)
	if err != nil {
		return regerrors.FromGRPC(err, b.Resource(), name)
	}

	return nil
}

func (b *backend) Watch(ctx context.Context, sel base.Selection, resourceVersion string) (watch.Interface, error) {
	selector := &sgroupsv1.HostBindingReq_Selectors{
		LabelSelector: sel.Labels,
	}
	if sel.Name != "" || sel.Namespace != "" {
		selector.FieldSelector = &sgroupsv1.HostBindingReq_Selectors_FieldSelector{
			Name:      sel.Name,
			Namespace: sel.Namespace,
		}
	}

	req := &sgroupsv1.HostBindingReq_Watch{
		ResourceVersion: resourceVersion,
		Selectors:       []*sgroupsv1.HostBindingReq_Selectors{selector},
	}
	ctx, cancel := context.WithCancel(ctx)
	stream, err := b.client.HostBindings.Watch(ctx, req)
	if err != nil {
		cancel()

		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	return base.NewStreamWatch(
		cancel,
		stream.Recv,
		func(evt *sgroupsv1.HostBindingResp_Watch) common.WatchEventType {
			return evt.GetType()
		},
		func(evt *sgroupsv1.HostBindingResp_Watch) []*sgroupsv1.HostBinding {
			return evt.GetHostBindings()
		},
		func(hb *sgroupsv1.HostBinding) runtime.Object {
			return convert.HostBindingFromProto(hb)
		},
	), nil
}
