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
	"time"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/yosida95/chame/pkg/chame"
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

func runServe(*cobra.Command, []string) {
	srv := &http.Server{
		Addr: cmdflg.Serve.Address,
		Handler: &chame.Chame{
			Proxy: &chame.HTTPProxy{},
			Store: FixedStoreFromConfig(cmdflg),
		},
	}

	errch := make(chan error, 1)
	go func(errch chan<- error) {
		glog.Infof("chame: Listen on %q", srv.Addr)

		errch <- srv.ListenAndServe()
	}(errch)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigch)

	select {
	case sig := <-sigch:
		glog.Infof("chame: Received %s signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			glog.Warningf("chame: Error on (*http.Server).Shutdown: %v", err)
			srv.Close()
		}
		<-errch
	case err := <-errch:
		glog.Exitln(err)
	}
}
