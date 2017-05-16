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
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/yosida95/chame/pkg/chame"
	"github.com/yosida95/chame/pkg/memstore"
	"golang.org/x/net/context"
)

func newDecodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "decode",
		Run: runDecode,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&flgFixedIssuer, "issuer", "https://chame.yosida95.com", "URL to identify token issuer")
	flags.StringVar(&flgFixedSecret, "secret", "dummysecret", "HMAC shared secret to sign/verify tokens")
	flags.StringVar(&flgTokenToDecode, "token", "", "signed token to decode")
	return cmd
}

func runDecode(cmd *cobra.Command, args []string) {
	store := memstore.Fixed(flgFixedIssuer, []byte(flgFixedSecret))
	decoded, err := chame.DecodeToken(context.Background(), store, flgTokenToDecode)
	if err != nil {
		log.Printf("failed to decode token: %v", err)
	}

	fmt.Printf("%s\n", decoded)
}