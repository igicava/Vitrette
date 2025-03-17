package redis

import (
	"time"
)

const DURATION = time.Second * 5

type Config struct {
	Port     string `yaml:"PORT" env:"REDIS_PORT" env-default:"6379"`
	Password string `yaml:"PASSWORD" env:"REDIS_PASSWORD" env-default:"1234"`
	DB       int    `yaml:"DB" env:"REDIS_DB" env-default:"0"`
}
