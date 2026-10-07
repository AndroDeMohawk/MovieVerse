package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string            `yaml:"env" env-default:"local"`
	GRPC        GRPCConfig        `yaml:"grpc"`
	Postgres    PostgresConfig    `yaml:"postgres"`
	Redis       RedisConfig       `yaml:"redis"`
	Kafka       KafkaConfig       `yaml:"kafka"`
	AuthService AuthServiceConfig `yaml:"auth_service"`
	HTTP        HTTPConfig        `yaml:"http"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-default:"44044"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type HTTPConfig struct {
	Port int `yaml:"port" env-default:"8085"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"1291"`
	User     string `yaml:"user" env-default:"postgres"`
	Password string `yaml:"password" env-required:"true"`
	DBName   string `yaml:"dbname" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env-default:"disable"`
}

type RedisConfig struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5555"`
	Password string `yaml:"password" env-default:""`
}

type KafkaConfig struct {
	Brokers       []string `yaml:"brokers" env-required:"true"`
	CommentsTopic string   `yaml:"comments_topic" env-default:"movie-comments"`
}

type AuthServiceConfig struct {
	Address   string `yaml:"address" env-required:"true"`
	AppID     int32  `yaml:"app_id" env-required:"true"`
	AppSecret string `yaml:"app_secret" env-required:"true"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
