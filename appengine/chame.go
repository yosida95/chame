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

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/yosida95/chame/pkg/chame"
	"github.com/yosida95/chame/pkg/memstore"
	"github.com/yosida95/chame/pkg/metadata"
	"github.com/yosida95/chame/pkg/stdlogger"
)

func main() {
	flag.Parse()

	port := os.Getenv("PORT")
	fixedIssuer := os.Getenv("CHAME_ISSUER")
	fixedSecret := os.Getenv("CHAME_SECRET")

	cfg := &chame.Config{
		Store: memstore.Fixed(fixedIssuer, []byte(fixedSecret)),
	}
	cfg.UseInterceptor(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := metadata.New(req.Context())
			ctx = chame.NewContextWithLogger(ctx, stdlogger.Logger)

			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	chame := chame.New(cfg)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), chame); err != nil {
		log.Fatal(err)
	}
}
