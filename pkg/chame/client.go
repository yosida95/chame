// Copyright 2020 Kohei YOSHIDA <https://yosida95.com/>.
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
	"fmt"
	"net/url"
	"path"
	"time"
)

type Client struct {
	baseUrl string
	issuer  string
	store   Store
}

func NewClient(baseUrl string, issuer string, store Store) (*Client, error) {
	parsed, err := url.ParseRequestURI(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("unrecognized base URL: %w", err)
	}
	if parsed.ForceQuery ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return nil, fmt.Errorf("Base URL must not contain query nor fragment")
	}
	// path.Join trims the ending slash.
	parsed.Path = path.Join(parsed.Path, proxyPrefix) + "/"

	return &Client{
		baseUrl: parsed.String(),
		issuer:  issuer,
		store:   store,
	}, nil
}

func (cli *Client) Sign(ctx context.Context, url string, opts SignOption) (string, error) {
	var exp JWTEpoch
	if !opts.Expiry.IsZero() {
		exp = JWTEpoch(opts.Expiry.Unix())
	}

	var nbf JWTEpoch
	if !opts.Expiry.IsZero() {
		nbf = JWTEpoch(opts.Expiry.Unix())
	}

	signed, err := EncodeToken(ctx, cli.store, &Token{
		Issuer:    cli.issuer,
		Subject:   url,
		NotBefore: nbf,
		Expiry:    exp,
	}, opts.JwtKid)
	if err != nil {
		return "", err
	}

	return cli.baseUrl + signed, nil
}

type SignOption struct {
	JwtKid    string
	NotBefore time.Time
	Expiry    time.Time
}
