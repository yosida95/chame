// Copyright 2020 Kohei YOSHIDA <https://yosida95.com/>.
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

package chame

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v4"
)

const (
	defaultIss = "https://chame.yosida95.com"
	defaultKid = "default"
)

var testEncodeTokenCases = []struct {
	Kid       string
	URL       string
	NotBefore *jwt.NumericDate
	NotAfter  *jwt.NumericDate

	Expected string
}{
	{
		Kid:      defaultKid,
		URL:      "https://example.com/foo.png",
		Expected: "eyJhbGciOiJIUzI1NiIsImtpZCI6ImRlZmF1bHQiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2NoYW1lLnlvc2lkYTk1LmNvbSIsInN1YiI6Imh0dHBzOi8vZXhhbXBsZS5jb20vZm9vLnBuZyJ9.xsOjAkiCiQrgoQTAAESJbE4GDbl8j5cohYnIZJ7GgWA",
	},
}

func TestEncodeToken(t *testing.T) {
	for i, c := range testEncodeTokenCases {
		encoded, err := EncodeToken(context.Background(), store, &Token{
			Issuer:    defaultIss,
			Subject:   c.URL,
			NotBefore: c.NotBefore,
			ExpiresAt: c.NotAfter,
		}, c.Kid)
		if err != nil {
			t.Errorf("%d: unexpected error: %v", i, err)
			continue
		}
		if encoded != c.Expected {
			t.Errorf("%d: expected %q, have %q", i, c.Expected, encoded)
			continue
		}
	}
}
