package monitoring

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	LogMode               string `envconfig:"LOG_MODE" default:"production"`
	LogLevel              string `envconfig:"LOG_LEVEL" default:"INFO"`
	LogEncoding           string `envconfig:"LOG_ENCODING" default:"json"`
	LogSamplingInitial    int    `envconfig:"LOG_SAMPLING_INITIAL" default:"100"`
	LogSamplingThereafter int    `envconfig:"LOG_SAMPLING_THEREAFTER" default:"100"`
}

//nolint:gochecknoglobals // this is a singleton
var logger = sync.OnceValue(func() *otelzap.Logger {
	config := LogConfig{}
	envconfig.MustProcess("", &config)
	level, err := zap.ParseAtomicLevel(config.LogLevel)
	if err != nil {
		panic(err)
	}

	var encoderConfig zapcore.EncoderConfig
	if config.LogMode == "dev" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
	}

	zapLogger, err := zap.Config{
		Level:       level,
		Development: config.LogMode == "dev",
		Sampling: &zap.SamplingConfig{
			Initial:    config.LogSamplingInitial,
			Thereafter: config.LogSamplingThereafter,
		},
		Encoding:         config.LogEncoding,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	if err != nil {
		panic(err)
	}
	return otelzap.New(zapLogger.With(zap.Namespace("context")))
})

func Logger() *otelzap.Logger {
	return logger()
}
