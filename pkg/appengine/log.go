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

package appengine

import (
	"github.com/yosida95/chame/pkg/chame"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type logger struct {
	ctx context.Context
}

func NewLogger(ctx context.Context) chame.Logger {
	return logger{ctx: ctx}
}

func (l logger) Debugf(format string, v ...interface{}) {
	log.Debugf(l.ctx, format, v...)
}

func (l logger) Infof(format string, v ...interface{}) {
	log.Infof(l.ctx, format, v...)
}

func (l logger) Warningf(format string, v ...interface{}) {
	log.Warningf(l.ctx, format, v...)
}

func (l logger) Errorf(format string, v ...interface{}) {
	log.Errorf(l.ctx, format, v...)
}
