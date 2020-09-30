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
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/yosida95/chame/pkg/chame"
)

func newEncodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "encode",
		Run: runEncode,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&cmdflg.Issuer, "issuer", "https://chame.yosida95.com", "URL to identify token issuer")
	flags.StringVar(&cmdflg.Secret, "secret", "dummysecret", "HMAC shared secret to sign/verify tokens")
	flags.StringVar(&cmdflg.Encode.URL, "url", "https://example.com/", "URL to encode")
	return cmd
}

func runEncode(cmd *cobra.Command, args []string) {
	store := FixedStoreFromConfig(cmdflg)
	token := &chame.Token{
		Issuer:  cmdflg.Issuer,
		Subject: cmdflg.Encode.URL,
	}
	signed, err := chame.EncodeToken(context.Background(), store, token, "")
	if err != nil {
		glog.Exitf("failed to encode URL: %v", err)
		return
	}

	fmt.Fprintf(os.Stdout, "%s\n", signed)
}
