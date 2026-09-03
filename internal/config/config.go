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
	DatabaseURI          string
	AccrualSystemAddress string
	JwtSecret            string
}

const (
	runAddressFlagName           = "a"
	databaseURIFlagName          = "d"
	AccrualSystemAddressFlagName = "r"
)

func NewConfig() *Config {
	c := &Config{}

	flag.StringVar(&c.RunAddress, runAddressFlagName, "localhost:8080", "адрес и порт запуска HTTP-сервера")
	flag.StringVar(&c.DatabaseURI, databaseURIFlagName, "", "строка подключения к PostgreSQL")
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

	if !passed[databaseURIFlagName] {
		if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
			c.DatabaseURI = envDatabaseURI
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
