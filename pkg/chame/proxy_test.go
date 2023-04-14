package chame

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHTTPProxy(t *testing.T) {
	const (
		textPlain = "text/plain"
		content   = "OK\n"
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/OK", func(w http.ResponseWriter, _ *http.Request) {
		h := w.Header()
		h.Set("Content-Type", textPlain)
		fmt.Fprint(w, content)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	srvUrl, _ := url.Parse(srv.URL)

	proxy := &HTTPProxy{HTTPClient: srv.Client()}
	w := httptest.NewRecorder()
	proxy.Do(w, &ProxyRequest{
		Context: context.Background(),
		URL:     srvUrl.JoinPath("/OK"),
		Header:  http.Header{},
	})

	if w.Code != http.StatusOK {
		t.Errorf("expect %d, got %d", http.StatusOK, w.Code)
	}
	if have := w.Header().Get("Content-Type"); textPlain != have {
		t.Errorf("expect %q, got %q", textPlain, have)
	}
	if have := w.Body.String(); content != have {
		t.Errorf("expect %q, got %q", content, have)
	}
}
