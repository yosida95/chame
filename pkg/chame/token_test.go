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
	"errors"
	"testing"
	"time"

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

func TestValidateClaims(t *testing.T) {
	now := time.Date(2023, 4, 19, 8, 0, 0, 0, time.UTC)
	for _, c := range []struct {
		in   Token
		code uint32
	}{
		{
			in:   Token{},
			code: 0,
		},
		{
			in: Token{
				ExpiresAt: toNumericDate(now.Add(-59 * time.Second)),
			},
			code: 0,
		},
		{
			in: Token{
				ExpiresAt: toNumericDate(now.Add(-60 * time.Second)),
			},
			code: jwt.ValidationErrorExpired,
		},
		{
			in: Token{
				IssuedAt: toNumericDate(now.Add(60 * time.Second)),
			},
			code: 0,
		},
		{
			in: Token{
				IssuedAt: toNumericDate(now.Add(61 * time.Second)),
			},
			code: jwt.ValidationErrorIssuedAt,
		},
		{
			in: Token{
				NotBefore: toNumericDate(now.Add(60 * time.Second)),
			},
			code: 0,
		},
		{
			in: Token{
				NotBefore: toNumericDate(now.Add(61 * time.Second)),
			},
			code: jwt.ValidationErrorNotValidYet,
		},
		{
			in: Token{
				ExpiresAt: toNumericDate(now.Add(-60 * time.Second)),
				IssuedAt:  toNumericDate(now.Add(61 * time.Second)),
				NotBefore: toNumericDate(now.Add(61 * time.Second)),
			},
			code: jwt.ValidationErrorExpired | jwt.ValidationErrorIssuedAt | jwt.ValidationErrorNotValidYet,
		},
	} {
		err := validateClaims(&c.in, now)
		if err != nil {
			var jwtErr *jwt.ValidationError
			if !errors.As(err, &jwtErr) || jwtErr.Errors != c.code {
				t.Errorf("expect %d, got %#v", c.code, err)
			}
			continue
		}
		if c.code != 0 {
			t.Errorf("expected error not occured")
		}
	}
}
