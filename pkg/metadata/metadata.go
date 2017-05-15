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
	mu    sync.RWMutex
	table = make(map[*http.Request]context.Context)
)

func Acquire(req *http.Request) context.Context {
	if ctx := FromRequest(req); ctx != nil {
		return ctx
	}
	mu.Lock()
	ctx := context.Background()
	ctx = withTime(ctx)
	setLocked(req, ctx)
	mu.Unlock()
	return ctx
}

func Release(req *http.Request) {
	mu.Lock()
	delete(table, req)
	mu.Unlock()
}

func FromRequest(req *http.Request) context.Context {
	mu.RLock()
	defer mu.RUnlock()

	return table[req]
}

func setLocked(req *http.Request, ctx context.Context) {
	table[req] = ctx
}

func Set(req *http.Request, ctx context.Context) {
	mu.Lock()
	setLocked(req, ctx)
	mu.Unlock()
}
