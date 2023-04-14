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
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/golang/glog"
	//lint:ignore SA1019 backward compatibility
	"github.com/yosida95/chame/pkg/metadata"
)

type Chame struct {
	store  Store
	proxy  Proxy
	ctypes []string
}

func New(cfg *Config) http.Handler {
	chame := &Chame{
		store: cfg.Store,
		proxy: cfg.Proxy,
	}
	if chame.proxy == nil {
		chame.proxy = &HTTPProxy{
			httpCFactory: cfg.NewHTTPClient,
		}
	}

	ctypes := cfg.ProxyContentType
	if ctypes == nil {
		ctypes = defaultProxyContentType
	}
	chame.ctypes = make([]string, len(ctypes))
	for i, ctype := range ctypes {
		chame.ctypes[i] = strings.ToLower(ctype)
	}

	return chame
}

func (chame *Chame) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = path.Clean(p)
	if p != req.URL.Path {
		http.Redirect(w, req, p, http.StatusPermanentRedirect)
		return
	}
	switch {
	case p == "/":
		chame.ServeHome(w, req)
	case strings.HasPrefix(p, proxyPrefix):
		chame.ServeProxy(w, req)
	default:
		http.NotFound(w, req)
	}
}

func (chame *Chame) ServeHome(w http.ResponseWriter, req *http.Request) {
	emitCommonHeaders(w)
	if req.URL.Path != "/" {
		httpError(w, http.StatusNotFound)
		return
	}
	if !httpErrorIfMethodNotAllowed(w, req, http.MethodGet) {
		return
	}
	w.Header().Set(headerKeyContentType, "text/plain;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello, this is chame!\nVisit https://github.com/yosida95/chame.\n")
}

const proxyPrefix = "/i/"

func (chame *Chame) ServeProxy(w http.ResponseWriter, userReq *http.Request) {
	emitCommonHeaders(w)
	if !strings.HasPrefix(userReq.URL.Path, proxyPrefix) {
		http.NotFound(w, userReq)
		return
	}
	if !httpErrorIfMethodNotAllowed(w, userReq, http.MethodGet) {
		return
	}
	ctx := userReq.Context()
	//lint:ignore SA1019 backward compatibility
	if time := metadata.Time(ctx); time.IsZero() {
		ctx = metadata.New(ctx) //lint:ignore SA1019 backward compatibility
	}
	signedURL := userReq.URL.Path[len(proxyPrefix):]
	decoded, err := DecodeToken(ctx, chame.store, signedURL)
	if err != nil {
		glog.Errorf("chame: failed to decode signed token %q: %v", signedURL, err)
		httpError(w, http.StatusBadRequest)
		return
	}
	reqUrl, err := url.Parse(decoded)
	if err != nil {
		glog.Errorf("chame: malformed URL: %v", err)
		httpError(w, http.StatusBadRequest)
		return
	}

	filtered := make(http.Header)
	copyHeadersOnlyIn(filtered, userReq.Header, passThroughReqHeaders)

	w = chame.newResponseWriter(w)
	chame.proxy.Do(w, &ProxyRequest{
		Context: ctx,
		URL:     reqUrl,
		Header:  filtered,
	})
}

func (chame *Chame) isAcceptableContentType(ctype string) bool {
	// As https://tools.ietf.org/html/rfc2045#section-5.1 said,
	// it is case-insensitive.
	ctype = strings.ToLower(ctype)
	for _, allowed := range chame.ctypes {
		if ctype == allowed {
			return true
		}
	}
	return false
}

type responsewriter struct {
	http.ResponseWriter
	once    sync.Once
	headers http.Header
	discard bool

	isAcceptableContentType func(string) bool
}

var _ http.ResponseWriter = (*responsewriter)(nil)

func (chame *Chame) newResponseWriter(w http.ResponseWriter) *responsewriter {
	return &responsewriter{
		ResponseWriter: w,
		headers:        make(http.Header),

		isAcceptableContentType: chame.isAcceptableContentType,
	}
}

func (w *responsewriter) Header() http.Header { return w.headers }

func (w *responsewriter) WriteHeader(code int) {
	w.once.Do(func() {
		dest := w.ResponseWriter.Header()
		src := w.headers
		copyHeadersOnlyIn(dest, src, passThroughRespHeaders)

		if ctype := dest.Get(headerKeyContentType); ctype != "" {
			ctype, _, err := mime.ParseMediaType(ctype)
			if err != nil || !w.isAcceptableContentType(ctype) {
				// special handling for error responses
				if !(code >= http.StatusBadRequest && ctype == "text/plain") {
					glog.Infof("chame: unacceptable Content-Type")
					dest.Del(headerKeyContentLength)
					httpError(w.ResponseWriter, http.StatusBadRequest)
					w.discard = true
					return
				}
			}
		} else {
			w.discard = true
			if code != http.StatusNotModified {
				glog.Infof("chame: Content-Type not present")
				dest.Del(headerKeyContentLength)
				httpError(w.ResponseWriter, http.StatusBadRequest)
				return
			}
		}

		w.ResponseWriter.WriteHeader(code)
	})
}

func (w *responsewriter) Write(p []byte) (int, error) {
	w.WriteHeader(http.StatusOK)
	if w.discard {
		return len(p), nil
	}
	return w.ResponseWriter.Write(p)
}
