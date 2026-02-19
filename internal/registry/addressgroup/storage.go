package addressgroup

import (
	"sgroups.io/sgroups-k8s-api/internal/registry/generic"
	"sgroups.io/sgroups-k8s-api/internal/registry/options"
	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
	"sgroups.io/sgroups-k8s-api/pkg/client"
)

// NewStorage creates an AddressGroup storage backed by gRPC.
func NewStorage(c *client.Client, opts options.StorageOptions) *generic.Storage[*v1alpha1.AddressGroup, *v1alpha1.AddressGroupList] {
	return generic.NewStorage(
		&backend{client: c},
		opts,
		&addressGroupStrategy{},
		func() *v1alpha1.AddressGroup { return &v1alpha1.AddressGroup{} },
		func() *v1alpha1.AddressGroupList { return &v1alpha1.AddressGroupList{} },
		"addressgroup",
	)
}
