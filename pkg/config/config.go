package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type DatabaseSettings struct {
	URL string `toml:"url" env:"DATABASE_URL" env-required:"true" env-description:"Postgres database URL"`
}

type AppSettings struct {
	Environment      string           `toml:"environment" env:"ENVIRONMENT" env-default:"development" env-description:"Application environment - production or development"`
	DatabaseSettings DatabaseSettings `toml:"database"`
}

func Init(path string) (*AppSettings, error) {
	var cfg AppSettings
	if path == "" {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, err
		}
		return &cfg, nil

	} else {
		if err := cleanenv.ReadConfig(path, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil

	}
}
