package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// Scheme contains all registered types for this API group.
var Scheme = runtime.NewScheme()

// Codecs provides access to encoding and decoding for the scheme.
var Codecs = serializer.NewCodecFactory(Scheme)

// ParameterCodec handles versioning of query parameters.
var ParameterCodec = runtime.NewParameterCodec(Scheme)

func init() {
	_ = AddToScheme(Scheme)
}
