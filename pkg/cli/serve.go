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
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/yosida95/chame/pkg/chame"
	"github.com/yosida95/chame/pkg/metadata"
	"golang.org/x/sync/errgroup"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "serve",
		Run: runServe,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&cmdflg.Serve.Address, "listen", "0.0.0.0:8080", "address and port chame will accept requests")
	flags.StringVar(&cmdflg.Issuer, "issuer", "https://chame.yosida95.com", "URL to identify token issuer")
	flags.StringVar(&cmdflg.Secret, "secret", "dummysecret", "HMAC shared secret to sign/verify tokens")
	return cmd
}

func runServe(cmd *cobra.Command, args []string) {
	cfg := &chame.Config{
		Store: FixedStoreFromConfig(cmdflg),
	}
	cfg.UseInterceptor(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := metadata.New(req.Context())
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	chame := chame.New(cfg)

	srv := &http.Server{
		Addr:    cmdflg.Serve.Address,
		Handler: chame,
	}

	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		glog.Infof("chame: Listen on %q", srv.Addr)

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	group.Go(func() error {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			glog.Warningf("chame: Error on (*http.Server).Shutdown: %v", err)
		}
		return nil
	})
	group.Go(func() error {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigch)

		select {
		case sig := <-sigch:
			glog.Infof("chame: Received %s signal", sig)
			cancel()
			return nil
		case <-ctx.Done():
			return nil
		}
	})

	if err := group.Wait(); err != nil {
		glog.Exitln(err)
	}
}
