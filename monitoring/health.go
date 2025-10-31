package monitoring

import (
	"net/http"
	"time"

	"github.com/hellofresh/health-go/v5"
)

//nolint:gochecknoglobals // this is a singleton
var checks = []health.Config{}

type HealthConfig struct {
	PathPrefix  string
	ServiceName string `envconfig:"SERVICE_NAME" default:"service"`
	Version     string
}

func AddCheck(name string, skipOnErr bool, timeout time.Duration, check health.CheckFunc) {
	checks = append(checks, health.Config{
		Name:      name,
		SkipOnErr: skipOnErr,
		Check:     check,
		Timeout:   timeout,
	})
}

func HealthHandler(config HealthConfig) http.Handler {
	h, _ := health.New(
		health.WithComponent(health.Component{
			Name:    config.ServiceName,
			Version: config.Version,
		}),
	)
	for _, c := range checks {
		h.Register(c)
	}
	return h.Handler()
}
