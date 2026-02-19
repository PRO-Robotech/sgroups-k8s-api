package filters

import (
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/endpoints/handlers/negotiation"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"

	"sgroups.io/sgroups-k8s-api/pkg/apis/sgroups/v1alpha1"
)

// RejectApplyPatch blocks server-side apply patch types.
func RejectApplyPatch(serializer runtime.NegotiatedSerializer) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPatch {
				contentType := r.Header.Get("Content-Type")
				if idx := strings.Index(contentType, ";"); idx > 0 {
					contentType = contentType[:idx]
				}
				if contentType == string(types.ApplyYAMLPatchType) || contentType == string(types.ApplyCBORPatchType) {
					err := negotiation.NewUnsupportedMediaTypeError([]string{
						string(types.JSONPatchType),
						string(types.MergePatchType),
						string(types.StrategicMergePatchType),
					})
					responsewriters.ErrorNegotiated(err, serializer, v1alpha1.SchemeGroupVersion, w, r)

					return
				}
			}
			handler.ServeHTTP(w, r)
		})
	}
}
