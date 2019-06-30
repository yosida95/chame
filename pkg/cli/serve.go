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

package cli

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/yosida95/chame/pkg/chame"
	"github.com/yosida95/chame/pkg/memstore"
	"github.com/yosida95/chame/pkg/metadata"
	"github.com/yosida95/chame/pkg/stdlogger"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "serve",
		Run: runServe,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&flgListenAddr, "listen", "0.0.0.0:8080", "address and port chame will accept requests")
	flags.StringVar(&flgFixedIssuer, "issuer", "https://chame.yosida95.com", "URL to identify token issuer")
	flags.StringVar(&flgFixedSecret, "secret", "dummysecret", "HMAC shared secret to sign/verify tokens")
	return cmd
}

func runServe(cmd *cobra.Command, args []string) {
	cfg := &chame.Config{
		Store: memstore.Fixed(flgFixedIssuer, []byte(flgFixedSecret)),
	}
	cfg.UseInterceptor(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := metadata.New(req.Context())
			ctx = chame.NewContextWithLogger(ctx, stdlogger.Logger)

			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	chame := chame.New(cfg)

	log.Printf("chame: listen chame on %q", flgListenAddr)
	if err := http.ListenAndServe(flgListenAddr, chame); err != nil {
		log.Printf("chame: failed to accept requests: %v", err)
	}
}
