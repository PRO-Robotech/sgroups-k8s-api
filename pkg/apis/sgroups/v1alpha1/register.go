package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupName is the API group for SGroups.
const GroupName = "sgroups.io"

// SchemeGroupVersion is the group/version used to register these objects.
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

// SchemeBuilder registers our types.
var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// AddToScheme adds all types to the scheme.
var AddToScheme = SchemeBuilder.AddToScheme

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion, KnownTypes()...)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	// Register core meta types (ListOptions, Status, etc.) for the
	// unversioned "v1" group — required by InstallAPIGroup.
	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})

	return nil
}

// Resource returns a GroupResource for the given resource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Kind returns a GroupKind for the given kind.
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}
