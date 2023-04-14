package chame

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestResponseWriter(t *testing.T) {
	chame := &Chame{
		ctypes: []string{"image/jpeg"},
	}
	for _, c := range []struct {
		codeIn, codeOut int
		typeIn, typeOut string
		textIn, textOut string
	}{
		{
			codeIn:  200,
			codeOut: 200,
			typeIn:  "image/jpeg",
			typeOut: "image/jpeg",
			textIn:  "response",
			textOut: "response",
		},
		{
			codeIn:  200,
			codeOut: http.StatusBadGateway,
			typeIn:  "text/html",
			typeOut: "text/plain; charset=utf-8",
			textIn:  "response",
			textOut: "Bad Gateway\n",
		},
		{
			codeIn:  400,
			codeOut: 400,
			typeIn:  "text/plain; charset=US-ASCII",
			typeOut: "text/plain; charset=US-ASCII",
			textIn:  "error message",
			textOut: "error message",
		},
		{
			codeIn:  400,
			codeOut: http.StatusBadGateway,
			typeIn:  "text/html",
			typeOut: "text/plain; charset=utf-8",
			textIn:  "HTML error",
			textOut: "Bad Gateway\n",
		},
		{
			codeIn:  http.StatusNotModified,
			codeOut: http.StatusNotModified,
			typeIn:  "",
			typeOut: "",
			textIn:  "",
			textOut: "",
		},
		{
			codeIn:  http.StatusNotModified,
			codeOut: http.StatusNotModified,
			typeIn:  "image/jpeg",
			typeOut: "image/jpeg",
			textIn:  "",
			textOut: "",
		},
		{
			codeIn:  http.StatusNotModified,
			codeOut: http.StatusBadGateway,
			typeIn:  "text/html",
			typeOut: "text/plain; charset=utf-8",
			textIn:  "",
			textOut: "Bad Gateway\n",
		},
	} {
		t.Logf("%d | %q", c.codeIn, c.textIn)
		out := httptest.NewRecorder()
		w := chame.newResponseWriter(out)
		w.Header().Set("Content-Type", c.typeIn)
		w.Header().Set("Content-Length", strconv.Itoa(len(c.textIn)))
		w.WriteHeader(c.codeIn)
		fmt.Fprintf(w, c.textIn)

		if c.codeOut != out.Code {
			t.Errorf("expect %d, got %d", c.codeOut, out.Code)
		}
		if have := out.Header().Get("Content-Type"); c.typeOut != have {
			t.Errorf("expect %q, got %q", c.typeOut, have)
		}
		switch l := out.Header().Get("Content-Length"); {
		case out.Code == http.StatusNotModified:
			if l != "" && l != "0" {
				t.Errorf("Content-Length must removed")
			}
		case l == "" && c.textIn == c.textOut:
			t.Errorf("Content-Length must retained")
		case l != "" && c.textIn != c.textOut:
			t.Errorf("Content-Length must removed")
		}
		if have := out.Body.String(); c.textOut != have {
			t.Errorf("expect %q, got %q", c.textOut, have)
		}
	}
}

func TestRawHeader(t *testing.T) {
	chame := &Chame{
		ctypes: []string{"image/jpeg"},
	}
	for _, c := range []struct {
		std    http.Header
		raw    http.Header
		expect http.Header
	}{
		{
			std: http.Header{
				"Content-Type":        []string{"image/jpeg"},
				"Content-Disposition": []string{"inline"},
				"X-Frame-Options":     []string{"SAMEORIGIN"},
			},
			raw: http.Header{},
			expect: http.Header{
				"Content-Type": []string{"image/jpeg"},
			},
		},
		{
			std: http.Header{
				"Content-Type": []string{"image/jpeg"},
			},
			raw: http.Header{
				"Content-Disposition": []string{"inline"},
				"X-Frame-Options":     []string{"SAMEORIGIN"},
			},
			expect: http.Header{
				"Content-Type":        []string{"image/jpeg"},
				"Content-Disposition": []string{"inline"},
			},
		},
	} {
		out := httptest.NewRecorder()
		var w http.ResponseWriter = chame.newResponseWriter(out)
		std := w.Header()
		copyHeader(std, c.std)
		raw := w.(RawHeader).RawHeader()
		copyHeader(raw, c.raw)
		w.WriteHeader(http.StatusOK)

		emitCommonHeaders(c.expect)
		if have := out.Header(); !cmp.Equal(c.expect, have) {
			t.Error(cmp.Diff(c.expect, have))
		}
	}
}
