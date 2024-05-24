package test

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const sampling = 1000

// NewLogger creates a new logger.
//
// If debug, development logger is created.
func NewLogger(debug bool) *zap.Logger {
	l := zap.L()

	if debug {
		cfg := zap.NewDevelopmentConfig()
		cfg.Sampling = &zap.SamplingConfig{
			Initial:    sampling,
			Thereafter: sampling,
		}

		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		log, err := cfg.Build()
		if err != nil {
			panic("could not prepare logger: " + err.Error())
		}

		return log
	}

	return l
}
