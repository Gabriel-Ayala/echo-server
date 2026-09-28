package echo_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"web-server-eccho/pkg/echo"
)

func TestEcho(t *testing.T) {
	for _, tc := range []struct{ name, method, path, body string }{
		{"root", "GET", "/", ""},
		{"nested", "POST", "/a/b?tag=one&tag=two&empty=", `{"hello":"world"}`},
		{"plain text", "PUT", "/text", "hello\nworld"},
		{"no redirect", "PATCH", "/a//b/../c?tag=one&tag=two&empty=", "raw"},
		{"delete", "DELETE", "/remove", "payload"},
		{"options", "OPTIONS", "/anything", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "http://example.com"+tc.path, strings.NewReader(tc.body))
			req.Header.Add("X-Test", "one")
			req.Header.Add("X-Test", "two")
			rec := httptest.NewRecorder()
			echo.NewHandler().ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("content type = %q", got)
			}
			var got echo.EchoResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			headers := req.Header.Clone()
			headers.Set("Host", req.Host)
			want := echo.EchoResponse{Headers: headers, Params: req.URL.Query(), Body: tc.body, Path: req.URL.Path}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %#v, want %#v", got, want)
			}
		})
	}
}

type brokenBody struct{ closed bool }

func (*brokenBody) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (b *brokenBody) Close() error           { b.closed = true; return nil }

func TestBodyReadFailure(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	body := &brokenBody{}
	req.Body = body
	rec := httptest.NewRecorder()
	echo.NewHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	if !body.closed {
		t.Fatal("request body was not closed")
	}
}
