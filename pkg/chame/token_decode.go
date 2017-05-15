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
	"github.com/yosida95/chame/pkg/metadata"
	"golang.org/x/net/context"
)

func DecodeToken(ctx context.Context, store Store, tokenString string) (string, error) {
	now := metadata.Time(ctx)
	token, err := jwt.ParseWithClaims(tokenString, &Token{now: now}, func(token *jwt.Token) (interface{}, error) {
		claim := token.Claims.(*Token)
		kid, _ := token.Header["kid"].(string)
		key, err := store.GetVerifyingKey(claim.Issuer, kid)
		if err != nil {
			return nil, errors.Wrap(err, "chame: failed to retrieve keys to verify signed token")
		}
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			switch key := key.(type) {
			default:
				return nil, errors.Wrap(err, "chame: incompatible key algorithms")
			case []byte:
				return key, nil
			}
		case *jwt.SigningMethodRSA:
			switch key := key.(type) {
			default:
				return nil, errors.Wrap(err, "chame: incompatible key algorithms")
			case *rsa.PublicKey:
				return key, nil
			}
		case *jwt.SigningMethodECDSA:
			switch key := key.(type) {
			default:
				return nil, errors.Wrap(err, "chame: incompatible key algorithms")
			case *ecdsa.PublicKey:
				return key, nil
			}
		default:
			return nil, errors.Wrap(err, "chame: incompatible key algorithms")
		}
	})
	if err != nil {
		if _, ok := err.(interface {
			Cause() error
		}); ok {
			return "", err
		}
		return "", errors.Wrap(err, "chame: failed to decode signed token")
	}
	claims := token.Claims.(*Token)
	return claims.Subject, nil
}
