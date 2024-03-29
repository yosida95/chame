// Copyright 2020 Kohei YOSHIDA <https://yosida95.com/>.
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
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Proxy interface {
	Do(http.ResponseWriter, *ProxyRequest)
}

type ProxyRequest struct {
	Context context.Context
	URL     *url.URL
	Header  http.Header
}

type HTTPProxy struct {
	HTTPClient *http.Client
	// Deprecated.
	httpCFactory func(context.Context) *http.Client
}

var _ Proxy = (*HTTPProxy)(nil)

func (f *HTTPProxy) Do(w http.ResponseWriter, userReq *ProxyRequest) {
	req, err := http.NewRequestWithContext(
		userReq.Context, http.MethodGet, userReq.URL.String(), nil)
	if err != nil {
		log.Printf("chame: failed to constract a HTTP request to fetch origin: %v", err)
		httpError(w, http.StatusBadRequest)
		return
	}
	req.Header = userReq.Header

	httpC := f.HTTPClient
	if httpC == nil {
		if f.httpCFactory != nil {
			httpC = f.httpCFactory(userReq.Context)
		} else {
			httpC = DefaultHTTPClient
		}
	}
	resp, err := httpC.Do(req)
	if err != nil {
		log.Printf("chame: failed to fetch the original: %v", err)
		httpError(w, http.StatusInternalServerError)
		return
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	switch code := resp.StatusCode; code {
	case http.StatusOK:
		copyHeader(w.Header(), resp.Header)
		w.WriteHeader(code)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("chame: failed to forward origin response to the client: %v", err)
			return
		}
	case http.StatusNotModified:
		copyHeader(w.Header(), resp.Header)
		w.WriteHeader(code)
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		log.Printf("chame: too many redirects: %q", req.URL)
		httpError(w, http.StatusBadGateway)
	default:
		log.Printf("chame: unexpected HTTP status(%d): %q", code, req.URL)
		httpError(w, code)
	}
}
