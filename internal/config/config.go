package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type ConfigENV struct {
	ServerAddress  string `env:"RUN_ADDRESS"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DBDsn          string `env:"DATABASE_URI"`
	SecretKey      string `env:"SECRET_KEY"`
}

var Options struct {
	FlagServerAddress  string
	FlagAccrualAddress string
	FlagDBDsn          string
	FlagSecretKey      string
}

func ParseFlags() error {
	if Options.FlagServerAddress != "" {
		return nil
	}

	flag.StringVar(&Options.FlagServerAddress, "a", "http://localhost:8080", "address to run server")
	flag.StringVar(&Options.FlagDBDsn, "d", "", "Database DSN")
	flag.StringVar(&Options.FlagAccrualAddress, "r", "http://localhost", "Accrual system address")
	flag.StringVar(&Options.FlagSecretKey, "sk", "verycomplexsecretkey", "Secret key")
	flag.Parse()

	var cfg ConfigENV

	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("Ошибка при парсинге переменных окружения %s", err.Error())
		return err
	}

	if cfg.ServerAddress != "" {
		Options.FlagServerAddress = cfg.ServerAddress
	}

	if cfg.AccrualAddress != "" {
		Options.FlagAccrualAddress = cfg.AccrualAddress
	}

	if cfg.DBDsn != "" {
		Options.FlagDBDsn = cfg.DBDsn
	}

	if cfg.SecretKey != "" {
		Options.FlagSecretKey = cfg.SecretKey
	}

	return nil
}
