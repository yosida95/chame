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

type Store interface {
	// GetVerifyingKey retrieves a key would be used to verify signed URLs by
	// a combination of Issuer (the "iss" claim) and Key ID (the "kid" header
	// value).
	// If an appropriate key is found, GetVerifyingKey returns a non-nil key
	// (the its type must be []byte for HMAC, *rsa.Publickey for RSA, or
	// *ecdsa.PublicKey for ECDSA) as the first return value, and nil err as
	// the second return value.
	// Otherwise GetVerifyingKey returns non-nil err as the second.
	GetVerifyingKey(iss string, kid string) (key interface{}, err error)
}
