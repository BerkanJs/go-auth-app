package shared

import (
	"log"
	"net/http"
	"time"
)

// loggingResponseWriter, status code ve yazılan byte sayısını takip eder.
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	// Henüz status set edilmediyse, varsayılan 200 kabul edilir.
	if lrw.status == 0 {
		lrw.status = http.StatusOK
	}
	n, err := lrw.ResponseWriter.Write(b)
	lrw.bytes += n
	return n, err
}

// LoggingMiddleware her isteği method, path, status ve süre bilgisiyle loglar.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		log.Printf(
			"http request: method=%s path=%s status=%d duration=%s bytes=%d",
			r.Method,
			r.URL.Path,
			lrw.status,
			duration,
			lrw.bytes,
		)
	})
}

