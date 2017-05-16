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

package chame

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Token struct {
	now time.Time

	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	NotBefore JWTEpoch `json:"nbf,omitempty"`
	Expiry    JWTEpoch `json:"exp,omitempty"`
}

func (tok *Token) Valid() error {
	const leeway = 1 * time.Minute

	if nbf := tok.NotBefore.Time(); !nbf.IsZero() && nbf.Add(-leeway).After(tok.now) {
		return errors.New("chame: expired token")
	}

	if exp := tok.Expiry.Time(); !exp.IsZero() && !exp.Add(leeway).After(tok.now) {
		return errors.New("chame: expired token")
	}

	return nil
}

type JWTEpoch int64

func (epoch JWTEpoch) Time() time.Time {
	if epoch == 0 {
		return time.Time{}
	}
	return time.Unix(int64(epoch), 0)
}

var _ json.Marshaler = (*JWTEpoch)(nil)

func (epoch JWTEpoch) MarshalJSON() ([]byte, error) {
	v := strconv.FormatInt(int64(epoch), 10)
	return []byte(v), nil
}

var _ json.Unmarshaler = (*JWTEpoch)(nil)

func (epoch *JWTEpoch) UnmarshalJSON(data []byte) error {
	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*epoch = JWTEpoch(i)
	return nil
}
