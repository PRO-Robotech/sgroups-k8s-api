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
)

// Resource plural name constants.
const (
	ResourceTenants       = "tenants"
	ResourceAddressGroups = "addressgroups"
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

// ---------- OpenAPIModelName ----------
// The Kubernetes DefinitionNamer converts Go import paths (slashes) to
// dot-separated names. Types must implement OpenAPIModelName to match,
// otherwise $ref pointers contain JSON-Pointer-escaped slashes (~1)
// that the field manager cannot resolve.

// OpenAPIPrefix is the dot-separated prefix for OpenAPI model names.
// Exported for use by hack/openapi-spec.
const OpenAPIPrefix = "sgroups.io.sgroups-k8s-api.pkg.apis.sgroups.v1alpha1."

func (Tenant) OpenAPIModelName() string           { return OpenAPIPrefix + "Tenant" }
func (TenantList) OpenAPIModelName() string       { return OpenAPIPrefix + "TenantList" }
func (TenantSpec) OpenAPIModelName() string       { return OpenAPIPrefix + "TenantSpec" }
func (AddressGroup) OpenAPIModelName() string     { return OpenAPIPrefix + "AddressGroup" }
func (AddressGroupList) OpenAPIModelName() string { return OpenAPIPrefix + "AddressGroupList" }
func (AddressGroupSpec) OpenAPIModelName() string { return OpenAPIPrefix + "AddressGroupSpec" }
