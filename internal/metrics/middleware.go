package metrics

import (
	"net/http"
	"strconv"
	"time"
)

func MetricsMiddleware(next http.Handler, handlerName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		RequestCounter.WithLabelValues(handlerName, r.Method, strconv.Itoa(rw.statusCode)).Inc()
		RequestDuration.WithLabelValues(handlerName, r.Method).Observe(duration)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
