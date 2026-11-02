package middlewares

import (
	"expvar"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	logger         = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	requestsStatus = expvar.NewMap("http_requests_by_status")
	errors5xx      = expvar.NewInt("http_errors_5xx")
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		requestsStatus.Add(strconv.Itoa(rec.status), 1)
		if rec.status >= 500 {
			errors5xx.Add(1)
		}
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status, "duration_ms", time.Since(start).Milliseconds())
	})
}
