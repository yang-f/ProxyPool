package util

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Config struct defines the config structure
type Config struct {
	Mongo MongoConfig `json:"mongo"`
	Redis RedisConfig `json:"redis"`
	Host  string      `json:"host"`
}

// MongoConfig has config values for Mongo
type MongoConfig struct {
	Addr  string `json:"addr"`
	DB    string `json:"db"`
	Table string `json:"table"`
	Event string `json:"event"`
}

// RedisConfig has config values for Redis
type RedisConfig struct {
	Addr string `json:"addr"`
	Key  string `json:"key"`
}

// NewConfig parses config file and return Config struct
func NewConfig() *Config {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatalln("Read config file error.")
	}
	config := &Config{}
	err = json.Unmarshal(file, config)
	if err != nil {
		panic(err)
	}
	return config
}
