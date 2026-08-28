package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const payload = `{"message":"hello gzip world, hello gzip world"}`

// echoHandler отдаёт payload с кодом 200.
func writeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(payload))
	})
}

func TestMiddleware_CompressesResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	Middleware(writeHandler()).ServeHTTP(rec, req)

	if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", enc)
	}

	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	data, _ := io.ReadAll(gr)
	if string(data) != payload {
		t.Errorf("decompressed body = %q, want %q", data, payload)
	}
}

func TestMiddleware_NoAcceptEncoding(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	Middleware(writeHandler()).ServeHTTP(rec, req)

	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Errorf("Content-Encoding = %q, want empty", enc)
	}
	if rec.Body.String() != payload {
		t.Errorf("body = %q, want plain %q", rec.Body.String(), payload)
	}
}

func TestMiddleware_DecompressesRequest(t *testing.T) {
	var got string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got = string(b)
		w.WriteHeader(http.StatusOK)
	})

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	zw.Write([]byte(payload))
	zw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	Middleware(handler).ServeHTTP(rec, req)

	if got != payload {
		t.Errorf("handler read %q, want decompressed %q", got, payload)
	}
}

func TestMiddleware_InvalidGzipRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("this is not gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	Middleware(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 on invalid gzip body", rec.Code)
	}
}
