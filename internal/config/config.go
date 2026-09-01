package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseUri          string
	AccrualSystemAddress string
	JwtSecret            string
}

const (
	runAddressFlagName           = "a"
	databaseUriFlagName          = "d"
	AccrualSystemAddressFlagName = "r"
)

func NewConfig() *Config {
	c := &Config{}

	flag.StringVar(&c.RunAddress, runAddressFlagName, "localhost:8080", "адрес и порт запуска HTTP-сервера")
	flag.StringVar(&c.DatabaseUri, databaseUriFlagName, "", "строка подключения к PostgreSQL")
	flag.StringVar(&c.AccrualSystemAddress, AccrualSystemAddressFlagName, "", "адрес системы расчёта начислений")

	return c
}

func (c *Config) ApplyEnv() {
	passed := map[string]bool{}

	flag.Visit(func(flag *flag.Flag) {
		passed[flag.Name] = true
	})

	if !passed[runAddressFlagName] {
		if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
			c.RunAddress = envRunAddress
		}
	}

	if !passed[databaseUriFlagName] {
		if envDatabaseUri := os.Getenv("DATABASE_URI"); envDatabaseUri != "" {
			c.DatabaseUri = envDatabaseUri
		}
	}

	if !passed[AccrualSystemAddressFlagName] {
		if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
			c.AccrualSystemAddress = envAccrualSystemAddress
		}
	}

	if envJwtSecret := os.Getenv("JWT_SECRET"); envJwtSecret != "" {
		c.JwtSecret = envJwtSecret
	}

	if c.JwtSecret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatal(err)
		}
		c.JwtSecret = hex.EncodeToString(b)
	}
}
