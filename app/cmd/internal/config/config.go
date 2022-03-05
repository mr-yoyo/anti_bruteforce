package config

import (
	"errors"
	"log"

	"github.com/spf13/viper"
)

var cfg *viper.Viper

func init() {
	cfg = viper.New()
	cfg.AddConfigPath(".")
	cfg.SetConfigType("yaml")
	cfg.SetConfigName("config")
}

func Get() *viper.Viper {
	return cfg
}

func Read(cfgFile string) {
	cfg.SetConfigFile(cfgFile)

	if err := cfg.ReadInConfig(); err != nil {
		var viperConfigFileNotFoundErr viper.ConfigFileNotFoundError

		if errors.As(err, &viperConfigFileNotFoundErr) {
			log.Fatal("fatal error, config file not found")
		} else {
			log.Fatalf("fatal error config file: %s \n", err)
		}
	}
}
