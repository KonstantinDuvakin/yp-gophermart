// Package gzip содержит HTTP-middleware для gzip-сжатия ответов и распаковки
// сжатых тел запросов.
package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// CompressWriter оборачивает http.ResponseWriter и сжимает тело ответа gzip.
type CompressWriter struct {
	w  http.ResponseWriter
	cw *gzip.Writer
}

// NewCompressWriter создаёт CompressWriter поверх w.
func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	cw := gzip.NewWriter(w)
	return &CompressWriter{w, cw}
}

// Header возвращает заголовки исходного ResponseWriter.
func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

// Write пишет данные в gzip-поток.
func (c *CompressWriter) Write(data []byte) (int, error) {
	return c.cw.Write(data)
}

// WriteHeader выставляет Content-Encoding: gzip и код ответа.
func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode != http.StatusNoContent {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip-поток, дописывая его хвост.
func (c *CompressWriter) Close() error {
	return c.cw.Close()
}

// CompressReader оборачивает тело запроса и распаковывает gzip на лету.
type CompressReader struct {
	r  io.ReadCloser
	cr *gzip.Reader
}

// NewCompressReader создаёт CompressReader поверх сжатого тела r.
func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	cr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:  r,
		cr: cr,
	}, nil
}

// Read читает распакованные данные.
func (c *CompressReader) Read(p []byte) (n int, err error) {
	return c.cr.Read(p)
}

// Close закрывает исходное тело и gzip-ридер.
func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.cr.Close()
}

// Middleware сжимает ответ gzip, если клиент прислал Accept-Encoding: gzip, и
// распаковывает тело запроса, если оно пришло с Content-Encoding: gzip.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		ow := rw

		acceptEnc := r.Header.Get("Accept-Encoding")
		isAcceptGzip := strings.Contains(acceptEnc, "gzip")
		if isAcceptGzip {
			cw := NewCompressWriter(rw)

			ow = cw

			defer cw.Close()
		}

		contentEnc := r.Header.Get("Content-Encoding")
		isContentGzip := strings.Contains(contentEnc, "gzip")
		if isContentGzip {
			cr, err := NewCompressReader(r.Body)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr

			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}
