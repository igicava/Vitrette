package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"lyceum/pkg/postgres"
	"lyceum/pkg/redis"
)

type Config struct {
	Postgres postgres.Config `yaml:"POSTGRES" env:"POSTGRES"`
	Redis    redis.Config    `yaml:"REDIS" env:"REDIS"`

	GRPCPort    int    `yaml:"GRPC_PORT" env:"GRPC_PORT" env-default:"50051"`
	GATEWAYPort string `yaml:"GRPC_GATEWAY_PORT" env:"GRPC_GATEWAY_PORT" env-default:"8081"`
}

func New() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
