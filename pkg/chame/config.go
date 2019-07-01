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
	"context"
	"net/http"
)

type Config struct {
	Store         Store
	NewHTTPClient func(context.Context) *http.Client

	iceptors []func(http.Handler) http.Handler
}

func (cfg *Config) UseInterceptor(iceptor func(http.Handler) http.Handler) {
	cfg.iceptors = append(cfg.iceptors, iceptor)
}

func (cfg *Config) applyInterceptors(handler http.Handler) http.Handler {
	for i := len(cfg.iceptors) - 1; i >= 0; i-- {
		handler = cfg.iceptors[i](handler)
	}
	return handler
}
