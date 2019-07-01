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
	"log"

	"github.com/spf13/cobra"
	"github.com/yosida95/chame/pkg/chame"
	"github.com/yosida95/chame/pkg/memstore"
)

func newEncodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "encode",
		Run: runEncode,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&flgFixedIssuer, "issuer", "https://chame.yosida95.com", "URL to identify token issuer")
	flags.StringVar(&flgFixedSecret, "secret", "dummysecret", "HMAC shared secret to sign/verify tokens")
	flags.StringVar(&flgUrlToEncode, "url", "https://example.com/", "URL to encode")
	return cmd
}

func runEncode(cmd *cobra.Command, args []string) {
	store := memstore.Fixed(flgFixedIssuer, []byte(flgFixedSecret))
	token := &chame.Token{
		Issuer:  flgFixedIssuer,
		Subject: flgUrlToEncode,
	}
	signed, err := chame.EncodeToken(context.Background(), store, token, "")
	if err != nil {
		log.Printf("failed to encode URL: %v", err)
	}

	fmt.Printf("%s\n", signed)
}
