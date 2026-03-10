package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Resource kind constants.
const (
	KindTenant     = "Tenant"
	KindTenantList = "TenantList"

	KindAddressGroup     = "AddressGroup"
	KindAddressGroupList = "AddressGroupList"

	KindNetwork     = "Network"
	KindNetworkList = "NetworkList"

	KindHost     = "Host"
	KindHostList = "HostList"

	KindHostBinding     = "HostBinding"
	KindHostBindingList = "HostBindingList"
)

// Resource plural name constants.
const (
	ResourceTenants       = "tenants"
	ResourceAddressGroups = "addressgroups"
	ResourceNetworks      = "networks"
	ResourceHosts         = "hosts"
	ResourceHostBindings  = "hostbindings"
)

// Action represents the default action for an AddressGroup.
type Action string

const (
	ActionUnknown Action = "UNKNOWN"
	ActionAllow   Action = "ALLOW"
	ActionDeny    Action = "DENY"
)

// TenantSpec defines the desired state of a Tenant.
type TenantSpec struct {
	DisplayName string `json:"displayName,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Description string `json:"description,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Tenant represents a tenant (namespace) resource.
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TenantSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// TenantList is a list of Tenant resources.
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

// AddressGroupSpec defines the desired state of an AddressGroup.
type AddressGroupSpec struct {
	DisplayName   string `json:"displayName,omitempty"`
	Comment       string `json:"comment,omitempty"`
	Description   string `json:"description,omitempty"`
	DefaultAction Action `json:"defaultAction,omitempty"`
	Logs          bool   `json:"logs,omitempty"`
	Trace         bool   `json:"trace,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AddressGroup represents an address group resource.
type AddressGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AddressGroupSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AddressGroupList is a list of AddressGroup resources.
type AddressGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AddressGroup `json:"items"`
}

// NetworkSpec defines the desired state of a Network.
type NetworkSpec struct {
	DisplayName string `json:"displayName,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Description string `json:"description,omitempty"`
	CIDR        string `json:"CIDR,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Network represents a network resource.
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NetworkSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkList is a list of Network resources.
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}

// ResourceIdentifier identifies a resource by name and namespace.
type ResourceIdentifier struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// HostSpec defines the desired state of a Host.
type HostSpec struct {
	DisplayName string `json:"displayName,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Description string `json:"description,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Host represents a host resource.
type Host struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              HostSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// HostList is a list of Host resources.
type HostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Host `json:"items"`
}

// HostBindingSpec defines the desired state of a HostBinding.
type HostBindingSpec struct {
	DisplayName  string             `json:"displayName,omitempty"`
	Comment      string             `json:"comment,omitempty"`
	Description  string             `json:"description,omitempty"`
	AddressGroup ResourceIdentifier `json:"addressGroup,omitempty"`
	Host         ResourceIdentifier `json:"host,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// HostBinding represents a host binding resource.
type HostBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              HostBindingSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// HostBindingList is a list of HostBinding resources.
type HostBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostBinding `json:"items"`
}

// ---------- OpenAPIModelName ----------
// The Kubernetes DefinitionNamer converts Go import paths (slashes) to
// dot-separated names. Types must implement OpenAPIModelName to match,
// otherwise $ref pointers contain JSON-Pointer-escaped slashes (~1)
// that the field manager cannot resolve.

// OpenAPIPrefix is the dot-separated prefix for OpenAPI model names.
// Exported for use by hack/openapi-spec.
const OpenAPIPrefix = "sgroups.io.sgroups-k8s-api.pkg.apis.sgroups.v1alpha1."

func (Tenant) OpenAPIModelName() string             { return OpenAPIPrefix + "Tenant" }
func (TenantList) OpenAPIModelName() string         { return OpenAPIPrefix + "TenantList" }
func (TenantSpec) OpenAPIModelName() string         { return OpenAPIPrefix + "TenantSpec" }
func (AddressGroup) OpenAPIModelName() string       { return OpenAPIPrefix + "AddressGroup" }
func (AddressGroupList) OpenAPIModelName() string   { return OpenAPIPrefix + "AddressGroupList" }
func (AddressGroupSpec) OpenAPIModelName() string   { return OpenAPIPrefix + "AddressGroupSpec" }
func (Network) OpenAPIModelName() string            { return OpenAPIPrefix + "Network" }
func (NetworkList) OpenAPIModelName() string        { return OpenAPIPrefix + "NetworkList" }
func (NetworkSpec) OpenAPIModelName() string        { return OpenAPIPrefix + "NetworkSpec" }
func (Host) OpenAPIModelName() string               { return OpenAPIPrefix + "Host" }
func (HostList) OpenAPIModelName() string           { return OpenAPIPrefix + "HostList" }
func (HostSpec) OpenAPIModelName() string           { return OpenAPIPrefix + "HostSpec" }
func (HostBinding) OpenAPIModelName() string        { return OpenAPIPrefix + "HostBinding" }
func (HostBindingList) OpenAPIModelName() string    { return OpenAPIPrefix + "HostBindingList" }
func (HostBindingSpec) OpenAPIModelName() string    { return OpenAPIPrefix + "HostBindingSpec" }
func (ResourceIdentifier) OpenAPIModelName() string { return OpenAPIPrefix + "ResourceIdentifier" }
