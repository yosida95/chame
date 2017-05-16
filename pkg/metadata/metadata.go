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

package metadata

import (
	"net/http"
	"sync"

	"golang.org/x/net/context"
)

var (
	ctxsmu sync.Mutex
	ctxs   = make(map[*http.Request]context.Context)
)

func New(parent context.Context) context.Context {
	return withTime(parent)
}

func FromRequest(req *http.Request) context.Context {
	ctxsmu.Lock()
	if ctx := ctxs[req]; ctx != nil {
		ctxsmu.Unlock()
		return ctx
	}

	ctx := New(context.Background())
	setLocked(req, ctx)
	ctxsmu.Unlock()
	return ctx
}

func setLocked(req *http.Request, ctx context.Context) {
	ctxs[req] = ctx
}

func Set(req *http.Request, ctx context.Context) {
	ctxsmu.Lock()
	setLocked(req, ctx)
	ctxsmu.Unlock()
}

func Release(req *http.Request) {
	ctxsmu.Lock()
	delete(ctxs, req)
	ctxsmu.Unlock()
}
