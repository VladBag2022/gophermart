package server

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	Accrual string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func NewConfig() (*Config, error) {
	config := &Config{}
	if err := env.Parse(config); err != nil {
		return nil, err
	}
	return config, nil
}
