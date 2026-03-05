package filters

import (
	"net/http"
	"net/url"
	"strings"

	"k8s.io/klog/v2"
)

// extraHeaderPrefix is the prefix kube-apiserver uses for forwarding
// user.Info.Extra via request-header authentication.
const extraHeaderPrefix = "X-Remote-Extra-"

// StripExtraAuthHeaders removes X-Remote-Extra-* headers whose decoded
// key contains characters not permitted by the request-header authenticator
// (allowed set: [0-9a-z-_.]).
func StripExtraAuthHeaders(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key := range r.Header {
			if !strings.HasPrefix(key, extraHeaderPrefix) {
				continue
			}
			encodedKey := key[len(extraHeaderPrefix):]
			decodedKey, err := url.QueryUnescape(encodedKey)
			if err != nil {
				klog.V(4).Infof("[strip-headers] removing header with un-decodable key: %q", key)
				r.Header.Del(key)

				continue
			}
			decodedKey = strings.ToLower(decodedKey)
			if !isValidExtraKey(decodedKey) {
				klog.V(4).Infof("[strip-headers] removing X-Remote-Extra header with invalid key chars: %q (decoded: %q)", key, decodedKey)
				r.Header.Del(key)
			}
		}
		handler.ServeHTTP(w, r)
	})
}

// isValidExtraKey checks whether key contains only characters in [0-9a-z-_.].
func isValidExtraKey(key string) bool {
	for _, c := range key {
		if (c < '0' || c > '9') && (c < 'a' || c > 'z') && c != '-' && c != '_' && c != '.' {
			return false
		}
	}

	return true
}
