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
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/yosida95/chame/pkg/metadata"
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
		return fmt.Errorf("chame: expired token")
	}

	if exp := tok.Expiry.Time(); !exp.IsZero() && !exp.Add(leeway).After(tok.now) {
		return fmt.Errorf("chame: expired token")
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

var parser = &jwt.Parser{
	ValidMethods: []string{
		jwt.SigningMethodHS256.Name,
		jwt.SigningMethodHS384.Name,
		jwt.SigningMethodHS512.Name,
		jwt.SigningMethodRS256.Name,
		jwt.SigningMethodES384.Name,
		jwt.SigningMethodRS512.Name,
		jwt.SigningMethodES256.Name,
		jwt.SigningMethodES384.Name,
		jwt.SigningMethodES512.Name,
	},
}

func DecodeToken(ctx context.Context, store Store, tokenString string) (string, error) {
	now := metadata.Time(ctx)
	token, err := parser.ParseWithClaims(tokenString, &Token{now: now}, func(token *jwt.Token) (interface{}, error) {
		claim := token.Claims.(*Token)
		kid, _ := token.Header["kid"].(string)
		return store.GetVerifyingKey(claim.Issuer, kid)
	})
	if err != nil {
		if _, ok := err.(interface {
			Cause() error
		}); ok {
			return "", err
		}
		return "", fmt.Errorf("chame: failed to decode signed token: %w", err)
	}
	claims := token.Claims.(*Token)
	return claims.Subject, nil
}
