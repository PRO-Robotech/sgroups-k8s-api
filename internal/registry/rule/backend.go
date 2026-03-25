package rule

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
	return v1alpha1.Resource(v1alpha1.ResourceRules)
}

func (b *backend) List(ctx context.Context, sel base.Selection) (*v1alpha1.RuleList, error) {
	selector := &sgroupsv1.RuleReq_Selectors{
		LabelSelector: sel.Labels,
	}
	if sel.Name != "" || sel.Namespace != "" {
		selector.FieldSelector = &sgroupsv1.RuleReq_Selectors_FieldSelector{
			Name:      sel.Name,
			Namespace: sel.Namespace,
		}
	}

	req := &sgroupsv1.RuleReq_List{Selectors: []*sgroupsv1.RuleReq_Selectors{selector}}
	resp, err := b.client.Rules.List(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	out := &v1alpha1.RuleList{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.KindRuleList,
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		Items: make([]v1alpha1.Rule, 0, len(resp.GetRules())),
	}
	out.ResourceVersion = resp.GetResourceVersion()
	for _, r := range resp.GetRules() {
		if converted := convert.RuleFromProto(r); converted != nil {
			out.Items = append(out.Items, *converted)
		}
	}

	return out, nil
}

func (b *backend) Upsert(ctx context.Context, obj *v1alpha1.Rule) (*v1alpha1.Rule, error) {
	req := &sgroupsv1.RuleReq_Upsert{
		Rules: []*sgroupsv1.Rule{convert.RuleToProto(obj)},
	}
	resp, err := b.client.Rules.Upsert(ctx, req)
	if err != nil {
		return nil, regerrors.FromGRPC(err, b.Resource(), obj.Name)
	}
	if len(resp.GetRules()) == 0 {
		return nil, apierrors.NewInternalError(errors.New("empty upsert response"))
	}

	return convert.RuleFromProto(resp.GetRules()[0]), nil
}

func (b *backend) Delete(ctx context.Context, name, namespace string) error {
	req := &sgroupsv1.RuleReq_Delete{
		Rules: []*sgroupsv1.RuleReq_Delete_Rule{
			{
				Metadata: &common.MetadataScope{
					Name:      name,
					Namespace: namespace,
				},
			},
		},
	}
	_, err := b.client.Rules.Delete(ctx, req)
	if err != nil {
		return regerrors.FromGRPC(err, b.Resource(), name)
	}

	return nil
}

func (b *backend) Watch(ctx context.Context, sel base.Selection, resourceVersion string) (watch.Interface, error) {
	selector := &sgroupsv1.RuleReq_Selectors{
		LabelSelector: sel.Labels,
	}
	if sel.Name != "" || sel.Namespace != "" {
		selector.FieldSelector = &sgroupsv1.RuleReq_Selectors_FieldSelector{
			Name:      sel.Name,
			Namespace: sel.Namespace,
		}
	}

	req := &sgroupsv1.RuleReq_Watch{
		ResourceVersion: resourceVersion,
		Selectors:       []*sgroupsv1.RuleReq_Selectors{selector},
	}
	ctx, cancel := context.WithCancel(ctx)
	stream, err := b.client.Rules.Watch(ctx, req)
	if err != nil {
		cancel()

		return nil, regerrors.FromGRPC(err, b.Resource(), sel.Name)
	}

	return base.NewStreamWatch(
		cancel,
		stream.Recv,
		func(evt *sgroupsv1.RuleResp_Watch) common.WatchEventType {
			return evt.GetType()
		},
		func(evt *sgroupsv1.RuleResp_Watch) []*sgroupsv1.Rule {
			return evt.GetRules()
		},
		func(r *sgroupsv1.Rule) runtime.Object {
			return convert.RuleFromProto(r)
		},
	), nil
}
