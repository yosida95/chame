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

package memstore

import (
	"fmt"
	"sync"

	"github.com/yosida95/chame/pkg/chame"
)

type entryKey struct {
	iss string
	kid string
}

type MemStore struct {
	mu     sync.RWMutex
	values map[entryKey]interface{}
}

func New() chame.Store {
	return &MemStore{}
}

func Fixed(iss string, secret interface{}) chame.Store {
	ms := New().(*MemStore)
	ms.Set(iss, "", secret)
	return ms
}

func (ms *MemStore) Get(iss string, kid string) interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.values[entryKey{iss, kid}]
}

func (ms *MemStore) Set(iss string, kid string, value interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.values == nil {
		ms.values = make(map[entryKey]interface{})
	}
	ms.values[entryKey{iss, kid}] = value
}

func (ms *MemStore) GetVerifyingKey(iss string, kid string) (interface{}, error) {
	key := ms.Get(iss, kid)
	if key == nil {
		return nil, fmt.Errorf("chame: key not found")
	}
	return key, nil
}

func (ms *MemStore) GetSigningKey(iss string, kid string) (interface{}, error) {
	key := ms.Get(iss, kid)
	if key == nil {
		return nil, fmt.Errorf("chame: key not found")
	}
	return key, nil
}
