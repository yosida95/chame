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
	"io"
	"mime"
	"net/http"

	"github.com/yosida95/chame/pkg/metadata"
	"golang.org/x/net/context"
)

type Chame struct {
	store        Store
	httpCFactory func(context.Context) *http.Client
}

func New(cfg *Config) http.Handler {
	chame := &Chame{
		store:        cfg.Store,
		httpCFactory: cfg.NewHTTPClient,
	}
	if chame.httpCFactory == nil {
		chame.httpCFactory = func(context.Context) *http.Client {
			return DefaultHTTPClient
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/", cfg.applyInterceptors(http.HandlerFunc(chame.ServeHome)))
	mux.Handle(proxyPrefix, cfg.applyInterceptors(http.HandlerFunc(chame.ServeProxy)))
	return mux
}

func (chame *Chame) ServeHome(w http.ResponseWriter, req *http.Request) {
	if !httpRespondIfMethodNotAllowed(w, req, http.MethodGet) {
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello, this is chame!\nSee https://github.com/yosida95/chame.")
}

const proxyPrefix = "/i/"

func (chame *Chame) ServeProxy(w http.ResponseWriter, userReq *http.Request) {
	if !httpRespondIfMethodNotAllowed(w, userReq, http.MethodGet) {
		return
	}
	ctx := metadata.FromRequest(userReq)
	log := LoggerFromContext(ctx)

	signedURL := userReq.URL.Path[len(proxyPrefix):]
	url, err := DecodeToken(ctx, chame.store, signedURL)
	if err != nil {
		log.Errorf("chame: failed to decode signed token %q: %v", signedURL, err)
		httpError(w, http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Errorf("chame: failed to constract a HTTP request to fetch origin: %v", err)
		httpError(w, http.StatusBadRequest)
		return
	}
	CopyRequestHeaders(req, userReq)

	httpC := chame.httpCFactory(ctx)
	resp, err := httpC.Do(req)
	if err != nil {
		log.Errorf("chame: failed to fetch the original: %v", err)
		httpError(w, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	switch code := resp.StatusCode; code {
	case http.StatusOK:
		ctype, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil || !IsAcceptedContentType(ctype) {
			log.Infof("chame: unacceptable Content-Type")
			httpError(w, http.StatusBadRequest)
			return
		}

		CopyResponseHeaders(w, resp)
		EmitSecurityHeaders(w)
		w.WriteHeader(code)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Errorf("chame: failed to forward origin response to the client: %v", err)
			return
		}
	case http.StatusNotModified:
		CopyResponseHeaders(w, resp)
		EmitSecurityHeaders(w)
		w.WriteHeader(code)
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, 308: // http.StatusPermanentRedirect
		// max redirects exceeded
		httpError(w, http.StatusBadGateway)
	default:
		httpError(w, code)
	}
}
