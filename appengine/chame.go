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

package devserver

import (
	"net/http"
	"os"

	chame_appengine "github.com/yosida95/chame/pkg/appengine"
	"github.com/yosida95/chame/pkg/chame"
	"github.com/yosida95/chame/pkg/memstore"
	"github.com/yosida95/chame/pkg/metadata"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	fixedIssuer := os.Getenv("CHAME_ISSUER")
	fixedSecret := os.Getenv("CHAME_SECRET")

	cfg := &chame.Config{
		Store:         memstore.Fixed(fixedIssuer, []byte(fixedSecret)),
		NewHTTPClient: urlfetch.Client,
	}
	cfg.UseInterceptor(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := appengine.NewContext(req)
			ctx = chame.NewContextWithLogger(ctx, chame_appengine.NewLogger(ctx))

			metadata.Set(req, ctx)
			defer metadata.Release(req)
			next.ServeHTTP(w, req)
		})
	})
	chame := chame.New(cfg)
	http.Handle("/", chame)
}
