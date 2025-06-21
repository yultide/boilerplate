package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

var (
	version              = "dev"
	builtAt              = time.Now().Format(time.RFC3339)
	maxRequestBody int64 = 1024 * 1024 // 1 MB
	requestTimeout       = 3 * time.Second
)

type Config struct {
	Version        string        `json:"version"`
	BuiltAt        string        `json:"builtAt"`
	MaxRequestBody int64         `json:"maxRequestBody"`
	RequestTimeout time.Duration `json:"requestTimeout"`
	Host           string        `env:"HOST" envDefault:"localhost" json:"host"`
	Port           int32         `env:"PORT" envDefault:"4000" json:"port"`
}

func LoadConfig() *Config {
	cfg := &Config{
		Version:        version,
		BuiltAt:        builtAt,
		MaxRequestBody: maxRequestBody,
		RequestTimeout: requestTimeout,
	}
	env.Parse(cfg)
	return cfg
}
