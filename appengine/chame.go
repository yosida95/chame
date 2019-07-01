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
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/yosida95/chame/pkg/chame"
	"github.com/yosida95/chame/pkg/cli"
	"github.com/yosida95/chame/pkg/metadata"
)

var cmdflg = cli.Config{
	Issuer: os.Getenv("CHAME_ISSUER"),
	Secret: os.Getenv("CHAME_SECRET"),
	Serve: struct {
		Address string
	}{
		Address: fmt.Sprintf(":%s", os.Getenv("PORT")),
	},
}

func main() {
	flag.CommandLine.Parse([]string{"-logtostderr"})
	defer glog.Flush()

	cfg := &chame.Config{
		Store: cli.FixedStoreFromConfig(cmdflg),
	}
	cfg.UseInterceptor(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := metadata.New(req.Context())

			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	chame := chame.New(cfg)
	if err := http.ListenAndServe(cmdflg.Serve.Address, chame); err != nil {
		glog.Fatal(err)
	}
}
