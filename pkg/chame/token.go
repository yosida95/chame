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
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Token = jwt.RegisteredClaims

func JWTEpoch(unix int64) *jwt.NumericDate {
	if unix == 0 {
		return nil
	}
	return jwt.NewNumericDate(time.Unix(unix, 0))
}

func fromNumericDate(t *jwt.NumericDate) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.Time
}

func toNumericDate(t time.Time) *jwt.NumericDate {
	if t.IsZero() {
		return nil
	}
	return jwt.NewNumericDate(t)
}

func EncodeToken(ctx context.Context, store Store, token *Token, kid string) (string, error) {
	key, err := store.GetSigningKey(token.Issuer, kid)
	if err != nil {
		return "", fmt.Errorf("chame: failed to retrieve a signing key: %w", err)
	}

	var mech jwt.SigningMethod
	switch key := key.(type) {
	default:
		return "", fmt.Errorf("chame: unsupported key algorithm")
	case []byte:
		mech = jwt.SigningMethodHS256
	case *rsa.PrivateKey:
		mech = jwt.SigningMethodRS256
	case *ecdsa.PrivateKey:
		switch key.Curve {
		case elliptic.P256():
			mech = jwt.SigningMethodES256
		case elliptic.P384():
			mech = jwt.SigningMethodES384
		case elliptic.P521():
			mech = jwt.SigningMethodES512
		default:
			return "", fmt.Errorf("chame: unsupported elliptic curve")
		}
	}

	jwtobj := jwt.NewWithClaims(mech, token)
	if kid != "" {
		jwtobj.Header["kid"] = kid
	}
	signed, err := jwtobj.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("chame: failed to sign a token: %w", err)
	}
	return signed, nil
}

var parserPool = sync.Pool{
	New: func() interface{} {
		return jwt.NewParser(
			jwt.WithValidMethods([]string{
				jwt.SigningMethodHS256.Name,
				jwt.SigningMethodHS384.Name,
				jwt.SigningMethodHS512.Name,
				jwt.SigningMethodRS256.Name,
				jwt.SigningMethodES384.Name,
				jwt.SigningMethodRS512.Name,
				jwt.SigningMethodES256.Name,
				jwt.SigningMethodES384.Name,
				jwt.SigningMethodES512.Name,
			}),
			jwt.WithoutClaimsValidation())
	},
}

func DecodeToken(ctx context.Context, store Store, tokenString string) (string, error) {
	const (
		leeway = 1 * time.Minute
	)
	parser := parserPool.Get().(*jwt.Parser)
	defer parserPool.Put(parser)
	claims := Token{}
	_, err := parser.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		kid, _ := token.Header["kid"].(string)
		return store.GetVerifyingKey(claims.Issuer, kid)
	})
	if err != nil {
		if _, ok := err.(interface {
			Cause() error
		}); ok {
			return "", err
		}
		return "", fmt.Errorf("chame: failed to decode signed token: %w", err)
	}

	now := time.Now()
	nowMin, nowMax := now.Add(-leeway), now.Add(leeway)
	if !claims.VerifyNotBefore(nowMax, false) ||
		!claims.VerifyIssuedAt(nowMax, false) ||
		!claims.VerifyExpiresAt(nowMin, false) {
		return "", fmt.Errorf("chame: failed to decode signed token: %w", err)
	}

	return claims.Subject, nil
}
