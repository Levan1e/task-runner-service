package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type RedisConfig struct {
	URL      string `yaml:"url"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type BrokerConfig struct {
	Broker        string `yaml:"broker"`
	DefaultQueue  string `yaml:"default_queue"`
	ResultBackend string `yaml:"result_backend"`
}

type Config struct {
	Server *ServerConfig `yaml:"server"`
	Redis  *RedisConfig  `yaml:"redis"`
	Broker *BrokerConfig `yaml:"broker"`
}

func ParseConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	target := &Config{}
	if err := yaml.NewDecoder(f).Decode(target); err != nil {
		return nil, err
	}

	return target, nil
}
