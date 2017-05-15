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
	"fmt"
	"net"
	"net/http"
	"time"
)

var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
	DualStack: true,
}

var DefaultHTTPClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: func(network string, addr string) (net.Conn, error) {
			switch addr {
			case "tcp", "tcp4", "tcp6":
				raddr, err := net.ResolveTCPAddr(network, addr)
				if err != nil {
					return nil, err
				}
				ip := raddr.IP
				if ip4 := ip.To4(); ip4 != nil {
					ip = ip4
				}
				addr = fmt.Sprintf("%s:%d", ip.String(), raddr.Port)
			}
			return dialer.Dial(network, addr)
		},
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}
