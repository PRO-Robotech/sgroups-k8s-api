package network

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
	return v1alpha1.Resource(v1alpha1.ResourceNetworks)
}

func (b *backend) List(ctx context.Context, sel base.Selection) (*v1alpha1.NetworkList, error) {
	resSelector := &common.ResSelector{
		LabelSelector: sel.Labels,
	}
	if sel.Name != "" || sel.Namespace != "" {
		resSelector.FieldSelector = &common.FieldSelector{
			Name:      sel.Name,
			Namespace: sel.Namespace,
		}
	}

	req := &sgroupsv1.NetworkReq_List{Selectors: []*common.ResSelector{resSelector}}
	resp, err := b.client.Networks.List(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	out := &v1alpha1.NetworkList{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindNetworkList,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Items: make([]v1alpha1.Network, 0, len(resp.GetNetworks())),
	}
	out.ResourceVersion = resp.GetResourceVersion()
	for _, nw := range resp.GetNetworks() {
		if converted := convert.NetworkFromProtoExt(nw); converted != nil {
			out.Items = append(out.Items, *converted)
		}
	}

	return out, nil
}

func (b *backend) Upsert(ctx context.Context, obj *v1alpha1.Network) (*v1alpha1.Network, error) {
	req := &sgroupsv1.NetworkReq_Upsert{
		Networks: []*sgroupsv1.Network{convert.NetworkToProto(obj)},
	}
	resp, err := b.client.Networks.Upsert(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), obj.Name)
	}
	if len(resp.GetNetworks()) == 0 {
		return nil, apierrors.NewInternalError(errors.New("empty upsert response"))
	}

	return convert.NetworkFromProto(resp.GetNetworks()[0]), nil
}

func (b *backend) Delete(ctx context.Context, name, namespace string) error {
	req := &sgroupsv1.NetworkReq_Delete{
		Networks: []*sgroupsv1.NetworkReq_Delete_Network{
			{
				Metadata: &common.MetadataScope{
					Name:      name,
					Namespace: namespace,
				},
			},
		},
	}
	_, err := b.client.Networks.Delete(ctx, req)
	if err != nil {
		return regerrors.FromGRPC(err, b.Resource(), name)
	}

	return nil
}

func (b *backend) Watch(ctx context.Context, sel base.Selection, resourceVersion string) (watch.Interface, error) {
	resSelector := &common.ResSelector{
		LabelSelector: sel.Labels,
	}
	if sel.Name != "" || sel.Namespace != "" {
		resSelector.FieldSelector = &common.FieldSelector{
			Name:      sel.Name,
			Namespace: sel.Namespace,
		}
	}

	req := &sgroupsv1.NetworkReq_Watch{
		ResourceVersion: resourceVersion,
		Selectors:       []*common.ResSelector{resSelector},
	}
	ctx, cancel := context.WithCancel(ctx)
	stream, err := b.client.Networks.Watch(ctx, req)
	if err != nil {
		cancel()

		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	return base.NewStreamWatch(
		cancel,
		stream.Recv,
		func(evt *sgroupsv1.NetworkResp_Watch) common.WatchEventType {
			return evt.GetType()
		},
		func(evt *sgroupsv1.NetworkResp_Watch) []*sgroupsv1.NetworkResp_NetworkExt {
			return evt.GetNetworks()
		},
		func(nw *sgroupsv1.NetworkResp_NetworkExt) runtime.Object {
			return convert.NetworkFromProtoExt(nw)
		},
	), nil
}
