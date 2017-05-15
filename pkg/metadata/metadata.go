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

func New(req *http.Request) context.Context {
	mu.Lock()
	defer mu.Unlock()

	if ctx := table[req]; ctx != nil {
		return ctx
	}
	ctx := context.Background()
	ctx = withTime(ctx)
	return ctx
}

func FromRequest(req *http.Request) context.Context {
	mu.RLock()
	ctx := table[req]
	mu.RUnlock()
	if ctx == nil {
		return New(req)
	}
	return ctx
}
