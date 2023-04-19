// Copyright 2017 Kohei YOSHIDA <https://yosida95.com/>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package chame

import (
	"net/http"
	"strings"
	"time"
)

var DefaultHTTPClient = &http.Client{
	Transport: http.DefaultTransport,
	Timeout:   30 * time.Second,
}

const (
	headerKeyContentType = "Content-Type"
)

func canonicalizedMIMEHeaderKeys(v []string) []string {
	for i := range v {
		v[i] = http.CanonicalHeaderKey(v[i])
	}
	return v
}

func copyHeader(dest http.Header, src http.Header) {
	for key, srcv := range src {
		dest[key] = append(dest[key], srcv...)
	}
}

func copyHeadersOnlyIn(dest http.Header, src http.Header, allowlist []string) {
	for _, key := range allowlist {
		srcv := src[key]
		if len(srcv) > 0 {
			destv := dest[key]
			if newSize := len(destv) + len(srcv); newSize > cap(destv) {
				dest[key] = append(make([]string, 0, newSize), destv...)
			}
			dest[key] = append(dest[key], srcv...)
		}
	}
}

var passThroughReqHeaders = canonicalizedMIMEHeaderKeys([]string{
	"Accept",
	"Cache-Control",
	"If-Modified-Since",
	"If-None-Match",
})

// Deprecated. This method is mainly for internal use and is no longer used
// internally.
func CopyRequestHeaders(dest *http.Request, src *http.Request) {
	copyHeadersOnlyIn(dest.Header, src.Header, passThroughReqHeaders)
}

var passThroughRespHeaders = canonicalizedMIMEHeaderKeys([]string{
	"Cache-Control",
	"Content-Encoding",
	"Content-Length",
	"Content-Type",
	"Etag",
	"Expires",
	"Last-Modified",
	"Transfer-Encoding",
})

// Deprecated. This method is mainly for internal use and is no longer used
// internally.
func CopyResponseHeaders(w http.ResponseWriter, resp *http.Response) {
	copyHeadersOnlyIn(w.Header(), resp.Header, passThroughRespHeaders)
}

// Deprecated. This method is mainly for internal use and is no longer used
// internally.
func IsAcceptableContentType(ctype string) bool {
	// As https://tools.ietf.org/html/rfc2045#section-5.1 said,
	// it is case-insensitive.
	ctype = strings.ToLower(ctype)
	for i := range defaultContentType {
		if ctype == defaultContentType[i] {
			return true
		}
	}
	return false
}

var securityHeaders = map[string]string{
	"Content-Security-Policy": "default-src 'none'; img-src data:; style-src 'unsafe-inline'",
	"X-Content-Type-Options":  "nosniff",
	"X-Frame-Options":         "DENY",
	"X-XSS-Protection":        "1; mode=block",
}

func emitCommonHeaders(h http.Header) {
	h.Set("Server", "chame")
	for k := range securityHeaders {
		h.Set(k, securityHeaders[k])
	}
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func httpErrorIfMethodNotAllowed(w http.ResponseWriter, req *http.Request, allowed ...string) bool {
	foundOptions := false
	for i := range allowed {
		if allowed[i] == req.Method {
			return true
		}
		if allowed[i] == http.MethodOptions {
			foundOptions = true
		}
	}
	if !foundOptions {
		allowed = append(allowed, http.MethodOptions)
	}

	w.Header().Set("Allow", strings.Join(allowed, ","))
	if req.Method == http.MethodOptions {
		// NOTE(yosida95): foundOptions is always false here.
		w.WriteHeader(http.StatusNoContent)
	} else {
		httpError(w, http.StatusMethodNotAllowed)
	}
	return false
}
