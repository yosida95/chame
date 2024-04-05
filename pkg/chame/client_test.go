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
)

func TestClientBaseURL(t *testing.T) {
	const baseURL = "https://chame.yosida95.com"
	client, err := NewClient(baseURL, "https://chame.yosida95.com", store)
	if err != nil {
		t.Errorf("failed to get new Client: %v", err)
		return
	}

	if have := client.BaseURL(); have != baseURL {
		t.Errorf("expect %q, got %q", baseURL, have)
	}
}

func TestClientSign(t *testing.T) {
	client, err := NewClient("https://chame.yosida95.com", "https://chame.yosida95.com", store)
	if err != nil {
		t.Errorf("failed to get new Client: %v", err)
		return
	}

	for i, c := range testEncodeTokenCases {
		expected := "https://chame.yosida95.com/i/" + c.Expected

		signed, err := client.Sign(context.Background(), c.URL, SignOption{
			JwtKid:    c.Kid,
			NotBefore: fromNumericDate(c.NotBefore),
			NotAfter:  fromNumericDate(c.NotAfter),
		})
		if err != nil {
			t.Errorf("%d: unexpected error: %v", i, err)
			continue
		}

		if signed != expected {
			t.Errorf("%d: expected %q, have: %q", i, expected, signed)
			continue
		}
	}
}
