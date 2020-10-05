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
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/yosida95/chame/pkg/chame"
	"github.com/yosida95/chame/pkg/cli"
	"github.com/yosida95/chame/pkg/metadata"
	"golang.org/x/sync/errgroup"
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
	srv := &http.Server{
		Addr:    cmdflg.Serve.Address,
		Handler: chame,
	}

	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		glog.Infof("Listen on %q", srv.Addr)
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
