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

// Deprecated: DO NOT USE
package metadata

import (
	"context"
	"time"
)

type timeKey struct{}

// Deprecated: DO NOT USE.
func WithTime(ctx context.Context) context.Context {
	return context.WithValue(ctx, timeKey{}, time.Now().UTC())
}

// Deprecated: DO NOT USE.
func Time(ctx context.Context) time.Time {
	if t, ok := ctx.Value(timeKey{}).(time.Time); ok {
		return t
	}
	return time.Time{}
}
