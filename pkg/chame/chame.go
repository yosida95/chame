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
	"log"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	//lint:ignore SA1019 backward compatibility
	"github.com/yosida95/chame/pkg/metadata"
)

type Chame struct {
	Proxy Proxy
	Store Store

	// If ContentType is not nil, it overrides the default list of Content-Type
	// to be proxied.
	ContentType []string
	// ExtraContentType is a list of Content-Type that to be proxied in addition
	// to the ContentType. In contrast to ContentType, ExtraContentType does not
	// override the default list.
	ExtraContentType []string

	ctypes map[string]struct{}
	once   sync.Once
}

// Deprecated: Instantiate Chame directly.
func New(cfg *Config) http.Handler {
	chame := &Chame{
		Proxy:       cfg.Proxy,
		Store:       cfg.Store,
		ContentType: cfg.ProxyContentType,
	}
	if chame.Proxy == nil {
		chame.Proxy = &HTTPProxy{
			httpCFactory: cfg.NewHTTPClient,
		}
	}
	return cfg.applyInterceptors(chame)
}

func (chame *Chame) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = path.Clean(p)
	if n := len(p); len(proxyPrefix)-n == 1 && proxyPrefix[:n] == p {
		p = proxyPrefix
	}
	if p != req.URL.Path {
		req.URL.Path = p
		http.Redirect(w, req, req.URL.String(), http.StatusPermanentRedirect)
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
	hdr := w.Header()
	emitCommonHeaders(hdr)
	if req.URL.Path != "/" {
		httpError(w, http.StatusNotFound)
		return
	}
	if !httpErrorIfMethodNotAllowed(w, req, http.MethodGet) {
		return
	}
	hdr.Set(headerKeyContentType, "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello, this is chame!\nVisit https://github.com/yosida95/chame.\n")
}

const proxyPrefix = "/i/"

func (chame *Chame) ServeProxy(w http.ResponseWriter, userReq *http.Request) {
	emitCommonHeaders(w.Header())
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
	decoded, err := DecodeToken(ctx, chame.Store, signedURL)
	if err != nil {
		log.Printf("chame: DecodeToken error: %v", err)
		httpError(w, http.StatusBadRequest)
		return
	}
	reqUrl, err := url.Parse(decoded)
	if err != nil {
		log.Printf("chame: malformed URL: %v", err)
		httpError(w, http.StatusBadRequest)
		return
	}

	filtered := make(http.Header)
	copyHeadersOnlyIn(filtered, userReq.Header, passThroughReqHeaders)

	w = chame.newResponseWriter(w)
	chame.Proxy.Do(w, &ProxyRequest{
		Context: ctx,
		URL:     reqUrl,
		Header:  filtered,
	})
}

// checkContentType checks if the given ctype is allowed to be proxied. ctype
// must be in lowercase and should not contain any parameters.
func (chame *Chame) checkContentType(ctype string) bool {
	chame.once.Do(func() {
		chame.ctypes = map[string]struct{}{}
		in := chame.ContentType
		if in == nil {
			in = defaultProxyContentType
		}
		for _, in := range in {
			chame.ctypes[strings.ToLower(in)] = struct{}{}
		}
		for _, in := range chame.ExtraContentType {
			chame.ctypes[strings.ToLower(in)] = struct{}{}
		}
	})
	_, found := chame.ctypes[ctype]
	return found
}

type responseWriter struct {
	http.ResponseWriter
	headers http.Header
	checkCT func(string) bool

	once    sync.Once
	discard bool
}

var _ http.ResponseWriter = (*responseWriter)(nil)

func (chame *Chame) newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,

		headers: make(http.Header),
		checkCT: chame.checkContentType,
	}
}

func (w *responseWriter) Header() http.Header { return w.headers }

type RawHeader interface {
	RawHeader() http.Header
}

func (w *responseWriter) RawHeader() http.Header {
	return w.ResponseWriter.Header()
}

func (w *responseWriter) WriteHeader(code int) {
	const cl = "Content-Length"
	w.once.Do(func() {
		// NOTE(yosida95): override fields set by RawHeader().
		dest := w.ResponseWriter.Header()
		emitCommonHeaders(dest)
		copyHeadersOnlyIn(dest, w.headers, passThroughRespHeaders)

		ctype := dest.Get(headerKeyContentType)
		parsed, _, err := mime.ParseMediaType(ctype)
		if err != nil || !w.checkCT(parsed) {
			switch {
			case parsed == "text/plain" && code >= 400:
				// special handling for error responses
			case ctype == "" && code == http.StatusNotModified:
				w.discard = true
				dest.Del(cl)
			default:
				log.Printf("chame: unacceptable Content-Type: %q", ctype)
				w.discard = true
				dest.Del(cl)
				httpError(w.ResponseWriter, http.StatusBadGateway)
				return
			}
		}
		w.ResponseWriter.WriteHeader(code)
	})
}

func (w *responseWriter) Write(p []byte) (int, error) {
	w.WriteHeader(http.StatusOK)
	if w.discard {
		return len(p), nil
	}
	return w.ResponseWriter.Write(p)
}
