package addressgroup

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

func (b *backend) NamespaceScoped() bool { return true }

func (b *backend) Resource() schema.GroupResource {
	return v1alpha1.Resource(v1alpha1.ResourceAddressGroups)
}

func (b *backend) List(ctx context.Context, sel base.Selection) (*v1alpha1.AddressGroupList, error) {
	resSelector := &common.ResSelector{
		LabelSelector: sel.Labels,
	}
	if sel.Name != "" || sel.Namespace != "" {
		resSelector.FieldSelector = &common.FieldSelector{
			Name:      sel.Name,
			Namespace: sel.Namespace,
		}
	}

	req := &sgroupsv1.AddressGroupReq_List{Selectors: []*common.ResSelector{resSelector}}
	resp, err := b.client.AddressGroups.List(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	out := &v1alpha1.AddressGroupList{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindAddressGroupList,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Items: make([]v1alpha1.AddressGroup, 0, len(resp.GetAddressGroups())),
	}
	out.ResourceVersion = resp.GetResourceVersion()
	for _, ag := range resp.GetAddressGroups() {
		if converted := convert.AddressGroupFromProtoExt(ag); converted != nil {
			out.Items = append(out.Items, *converted)
		}
	}

	return out, nil
}

func (b *backend) Upsert(ctx context.Context, obj *v1alpha1.AddressGroup) (*v1alpha1.AddressGroup, error) {
	req := &sgroupsv1.AddressGroupReq_Upsert{
		AddressGroups: []*sgroupsv1.AddressGroup{convert.AddressGroupToProto(obj)},
	}
	resp, err := b.client.AddressGroups.Upsert(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), obj.Name)
	}
	if len(resp.GetAddressGroups()) == 0 {
		return nil, apierrors.NewInternalError(errors.New("empty upsert response"))
	}

	return convert.AddressGroupFromProto(resp.GetAddressGroups()[0]), nil
}

func (b *backend) Delete(ctx context.Context, name, namespace string) error {
	req := &sgroupsv1.AddressGroupReq_Delete{
		AddressGroups: []*sgroupsv1.AddressGroupReq_Delete_AddressGroup{
			{
				Metadata: &common.MetadataScope{
					Name:      name,
					Namespace: namespace,
				},
			},
		},
	}
	_, err := b.client.AddressGroups.Delete(ctx, req)
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

	req := &sgroupsv1.AddressGroupReq_Watch{
		ResourceVersion: resourceVersion,
		Selectors:       []*common.ResSelector{resSelector},
	}
	ctx, cancel := context.WithCancel(ctx)
	stream, err := b.client.AddressGroups.Watch(ctx, req)
	if err != nil {
		cancel()

		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	return base.NewStreamWatch(
		cancel,
		stream.Recv,
		func(evt *sgroupsv1.AddressGroupResp_Watch) common.WatchEventType {
			return evt.GetType()
		},
		func(evt *sgroupsv1.AddressGroupResp_Watch) []*sgroupsv1.AddressGroup {
			return evt.GetAddressGroups()
		},
		func(ag *sgroupsv1.AddressGroup) runtime.Object {
			return convert.AddressGroupFromProto(ag)
		},
	), nil
}
