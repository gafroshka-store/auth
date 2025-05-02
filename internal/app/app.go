package app

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ConfigRedis ConfigRedis `yaml:"redis"`
	TokenSecret string      `yaml:"token_secret"`
	ServerPort  uint        `yaml:"server_port"`
}

type ConfigRedis struct {
	Host     string `yaml:"host"`
	Port     uint   `yaml:"port"`
	DB       int    `yaml:"db"`
	Password string `yaml:"password"`
}

func NewConfig(configPath string) (*Config, error) {
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err = yaml.Unmarshal(configFile, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
