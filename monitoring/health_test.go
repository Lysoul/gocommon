package monitoring_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lysoul/gocommon/monitoring"
)

func TestHealthCheck(t *testing.T) {
	monitoring.AddCheck("test", false, 0, func(context.Context) error {
		return nil
	})
	h := monitoring.HealthHandler(monitoring.HealthConfig{
		PathPrefix:  "/api",
		ServiceName: "service",
		Version:     "1.0.0",
	})

	w := performRequest(h, "GET", "/api/health")
	if w.Code != 200 {
		t.Errorf("expected status code to be 200, got %d", w.Code)
	}
	log.Println(w.Body.String())
}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
