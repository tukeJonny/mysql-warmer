package mysql

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	MySQL MySQLConfig
}

type MySQLConfig struct {
	Username string
	Password string
	Hostname string
	Port     int
	DbName   string
	UnixSock string
}

func GetMySQLConfig() (config Config) {
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		log.Fatalf("Failed to decode config.toml")
	}

	return
}
