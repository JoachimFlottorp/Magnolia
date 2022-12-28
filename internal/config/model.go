package config

import (
	"flag"
	"os"

	"github.com/pelletier/go-toml/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	cfg   = flag.String("config", "config.toml", "Path to the config file")
	debug = flag.Bool("debug", false, "Enable debug logging")
)

type Config struct {
	Redis struct {
		Username string `toml:"username"`
		Password string `toml:"password"`
		Database int    `toml:"database"`
		Address  string `toml:"address"`
	} `toml:"redis"`
	Mongo struct {
		Username string `toml:"username"`
		Password string `toml:"password"`
		Address  string `toml:"address"`
		SRV      bool   `toml:"srv"`
		DB       string `toml:"db"`
	}
	RabbitMQ struct {
		URI string `toml:"uri"`
	} `toml:"rmq"`
	Markov struct {
		HealthAddress string `toml:"health_address"`
		HealthBind    int    `toml:"health_bind"`
	} `toml:"markov"`
	Http struct {
		Port      int    `toml:"port"`
		PublicURL string `toml:"public_url"`
	} `toml:"http"`
	Twitch struct {
		Bot struct {
			Username string   `toml:"username"`
			Password string   `toml:"password"`
			Admins   []string `toml:"admins"`
			Prefix   string   `toml:"prefix"`
		} `toml:"bot"`
	} `toml:"twitch"`
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
func CreateConfig() (*Config, error) {
	flag.Parse()

	bytes, err := os.ReadFile(*cfg)
	if err != nil {
		return nil, err
	}
	var config Config

	err = toml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, ReplaceZapGlobal(*debug)
}
