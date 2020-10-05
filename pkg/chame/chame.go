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
	"sync"

	"github.com/golang/glog"
)

type Chame struct {
	store Store
	proxy Proxy
}

func New(cfg *Config) http.Handler {
	proxy := &HTTPProxy{
		httpCFactory: cfg.NewHTTPClient,
	}
	chame := &Chame{
		store: cfg.Store,
		proxy: proxy,
	}

	mux := http.NewServeMux()
	mux.Handle("/", cfg.applyInterceptors(http.HandlerFunc(chame.ServeHome)))
	mux.Handle(proxyPrefix, cfg.applyInterceptors(http.HandlerFunc(chame.ServeProxy)))
	return mux
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
	fmt.Fprintf(w, "Hello, this is chame!\nVisit https://github.com/yosida95/chame.")
}

const proxyPrefix = "/i/"

func (chame *Chame) ServeProxy(w http.ResponseWriter, userReq *http.Request) {
	emitCommonHeaders(w)
	if !httpErrorIfMethodNotAllowed(w, userReq, http.MethodGet) {
		return
	}
	ctx, cancel := newProxyContext(userReq.Context())
	defer cancel()

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

	w = newResponseWriter(w)
	chame.proxy.Do(w, &ProxyRequest{
		Context: ctx,
		URL:     reqUrl,
		Header:  filtered,
	})
}

type responsewriter struct {
	http.ResponseWriter
	once    sync.Once
	headers http.Header
	discard bool
}

var _ http.ResponseWriter = (*responsewriter)(nil)

func newResponseWriter(w http.ResponseWriter) *responsewriter {
	return &responsewriter{
		ResponseWriter: w,
		headers:        make(http.Header),
	}
}

func (w *responsewriter) Header() http.Header { return w.headers }

func (w *responsewriter) WriteHeader(code int) {
	w.once.Do(func() {
		dest := w.ResponseWriter.Header()
		src := w.headers
		copyHeadersOnlyIn(dest, src, passThroughRespHeaders)

		ctype, _, err := mime.ParseMediaType(dest.Get(headerKeyContentType))
		if err != nil || !IsAcceptableContentType(ctype) {
			glog.Infof("chame: unacceptable Content-Type")
			dest.Del(headerKeyContentType)
			dest.Del(headerKeyContentLength)
			httpError(w.ResponseWriter, http.StatusBadRequest)
			w.discard = true
			return
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
