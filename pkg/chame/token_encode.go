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
	"crypto/ecdsa"
	"crypto/rsa"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func EncodeToken(ctx context.Context, store Store, token *Token, kid string) (string, error) {
	key, err := store.GetVerifyingKey(token.Issuer, kid)
	if err != nil {
		return "", errors.Wrap(err, "chame: failed to retrieve a signing key")
	}

	var mech jwt.SigningMethod
	switch key.(type) {
	default:
		return "", errors.New("chame: unsupported key algorithm")
	case []byte:
		mech = jwt.SigningMethodHS256
	case *rsa.PrivateKey:
		mech = jwt.SigningMethodRS256
	case *ecdsa.PrivateKey:
		mech = jwt.SigningMethodES256
	}

	jwtobj := jwt.NewWithClaims(mech, token)
	if kid != "" {
		jwtobj.Header["kid"] = kid
	}
	signed, err := jwtobj.SignedString(key)
	if err != nil {
		return "", errors.Wrap(err, "chame: failed to sign a token")
	}
	return signed, nil
}
