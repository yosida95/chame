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
	"crypto/rsa"
	"encoding/json"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
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

func EncodeToken(ctx context.Context, store Store, token *Token, kid string) (string, error) {
	key, err := store.GetSigningKey(token.Issuer, kid)
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
