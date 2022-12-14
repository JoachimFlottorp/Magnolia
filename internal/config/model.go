package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Redis struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Database int    `json:"database"`
		Address  string `json:"address"`
	} `json:"redis"`
	Mongo struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Address  string `json:"address"`
		SRV      bool   `json:"srv"`
		DB       string `json:"db"`
	}
	RabbitMQ struct {
		URI string `json:"uri"`
	} `json:"rmq"`
	Markov struct {
		HealthAddress string `json:"health_address"`
		HealthBind    int    `json:"health_bind"`
	} `json:"markov"`
	Http struct {
		Port      int    `json:"port"`
		PublicURL string `json:"public_url"`
	} `json:"http"`
	Twitch struct {
		Bot struct {
			Username string   `json:"username"`
			Password string   `json:"password"`
			Admins   []string `json:"admins"`
			Prefix   string   `json:"prefix"`
		} `json:"bot"`
	} `json:"twitch"`
}

func ReplaceZapGlobal(isDebug bool) error {
	config := &zap.Config{
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if isDebug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	if a, err := config.Build(); err != nil {
		return err
	} else {
		zap.ReplaceGlobals(a)
	}

	return nil
}
