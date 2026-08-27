package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseUri          string
	AccrualSystemAddress string
	JwtSecret            string
}

func NewConfig() *Config {
	c := &Config{}

	flag.StringVar(&c.RunAddress, "a", "localhost:8080", "адрес и порт запуска HTTP-сервера")
	flag.StringVar(&c.DatabaseUri, "d", "", "строка подключения к PostgreSQL")
	flag.StringVar(&c.AccrualSystemAddress, "r", "", "адрес системы расчёта начислений")
	flag.StringVar(&c.JwtSecret, "s", "", "секрет для JWT токена")

	return c
}

func (c *Config) ApplyEnv() {
	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		c.RunAddress = envRunAddress
	}

	if envDatabaseUri := os.Getenv("DATABASE_URI"); envDatabaseUri != "" {
		c.DatabaseUri = envDatabaseUri
	}

	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		c.AccrualSystemAddress = envAccrualSystemAddress
	}

	if envJwtSecret := os.Getenv("JWT_SECRET"); envJwtSecret != "" {
		c.JwtSecret = envJwtSecret
	}

	if c.JwtSecret == "" {
		c.JwtSecret = "change_me"
	}
}
