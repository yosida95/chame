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
	"net/textproto"
	"strings"
)

const (
	headerKeyAllow       = "Allow"
	headerKeyContentType = "Content-Type"
	headerKeyServer      = "Server"
)

func canonicalizedMIMEHeaderKeys(v []string) []string {
	for i := range v {
		v[i] = textproto.CanonicalMIMEHeaderKey(v[i])
	}
	return v
}

func copyHeadersOnlyIn(dest http.Header, src http.Header, whitelist []string) {
	for k := range src {
		found := false
		for i := range whitelist {
			if textproto.CanonicalMIMEHeaderKey(k) == whitelist[i] {
				found = true
				break
			}
		}
		if found {
			vv := src[k]
			for i := range vv {
				dest.Add(k, vv[i])
			}
		}
	}
}

var passThroughReqHeaders = canonicalizedMIMEHeaderKeys([]string{
	"Accept",
	"Cache-Control",
	"If-Modified-Since",
	"If-None-Match",
})

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

func CopyResponseHeaders(w http.ResponseWriter, resp *http.Response) {
	copyHeadersOnlyIn(w.Header(), resp.Header, passThroughRespHeaders)
}

// Values below ware taken from
// https://github.com/atmos/camo/blob/bd731cff64fd61a7ee4ea7dd6e96b8e0b69c3da0/mime-types.json
var acceptedContentTypes = []string{
	"image/bmp",
	"image/cgm",
	"image/g3fax",
	"image/gif",
	"image/ief",
	"image/jp2",
	"image/jpeg",
	"image/jpg",
	"image/pict",
	"image/png",
	"image/prs.btif",
	"image/svg+xml",
	"image/tiff",
	"image/vnd.adobe.photoshop",
	"image/vnd.djvu",
	"image/vnd.dwg",
	"image/vnd.dxf",
	"image/vnd.fastbidsheet",
	"image/vnd.fpx",
	"image/vnd.fst",
	"image/vnd.fujixerox.edmics-mmr",
	"image/vnd.fujixerox.edmics-rlc",
	"image/vnd.microsoft.icon",
	"image/vnd.ms-modi",
	"image/vnd.net-fpx",
	"image/vnd.wap.wbmp",
	"image/vnd.xiff",
	"image/webp",
	"image/x-cmu-raster",
	"image/x-cmx",
	"image/x-icon",
	"image/x-macpaint",
	"image/x-pcx",
	"image/x-pict",
	"image/x-portable-anymap",
	"image/x-portable-bitmap",
	"image/x-portable-graymap",
	"image/x-portable-pixmap",
	"image/x-quicktime",
	"image/x-rgb",
	"image/x-xbitmap",
	"image/x-xpixmap",
	"image/x-xwindowdump",
}

func IsAcceptedContentType(ctype string) bool {
	// As https://tools.ietf.org/html/rfc2045#section-5.1 said,
	// it is case-insensitive.
	ctype = strings.ToLower(ctype)
	for i := range acceptedContentTypes {
		if ctype == acceptedContentTypes[i] {
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

func emitCommonHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set(headerKeyServer, "chame")
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

	w.Header().Set(headerKeyAllow, strings.Join(allowed, ","))
	if req.Method == http.MethodOptions {
		// NOTE(yosida95): foundOptions is always false here.
		w.WriteHeader(http.StatusNoContent)
	} else {
		httpError(w, http.StatusMethodNotAllowed)
	}
	return false
}
