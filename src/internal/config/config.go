package config

import (
	"flag"
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type api struct {
	Host string `yaml:"host" env:"API_HOST" env-default:"127.0.0.1"`
	Port string `yaml:"port" env:"API_PORT" env-default:"8000"`
}

type websocket struct {
	Port string `yaml:"port" env:"WS_PORT" env-default:"8001"`
}

type eventbus struct {
	Timeout time.Duration `yaml:"timeout" env:"EVENTBUS_TIMEOUT" env-default:"2s"`
}

type database struct {
	URL      string `yaml:"url" env:"DB_URL" env-required:"true"`
	LogLevel string `yaml:"log_level" env:"DB_LOG_LEVEL"`
}

type secure struct {
	Hash struct {
		Salt string `yaml:"salt" env:"HASH_SALT" env-required:"true"`
	} `yaml:"hash"`

	CORS struct {
		Origins        []string `yaml:"origins"`
		AllowedMethods []string `yaml:"allowed_methods"`
		AllowedHeaders []string `yaml:"allowed_headers"`
	} `yaml:"cors"`
}

type media struct {
	Url          string   `yaml:"url" env:"MEDIA_URL" env-default:"/media/"`
	Path         string   `yaml:"path" env:"MEDIA_PATH" env-default:"assset/media"`
	AllowedMedia []string `yaml:"allowed_media" env-required:"true"`
	MaxSize      string   `yaml:"max_size" env:"MEDIA_MAX_SIZE" env-default:"500M"`
}

type static struct {
	URL  string `yaml:"url" env:"STATIC_URL" env-default:"/static/"`
	Path string `yaml:"path" env:"STATIC_PATH" env-default:"assets/static"`
}

type logger struct {
	Level  string `yaml:"level" env:"LOGGER_LEVEL" env-default:"info"`
	Folder string `yaml:"folder" env:"LOGGER_FOLDER" env-default:"assets/log"`
}

type liveKit struct {
	ApiKey    string `yaml:"api_key" env:"LIVEKIT_API_KEY"`
	ApiSecret string `yaml:"api_secret" env:"LIVEKIT_API_SECRET"`
}

type Config struct {
	Api       api       `yaml:"api"`
	Websocket websocket `yaml:"websocket"`
	EventBus  eventbus  `yaml:"eventbus"`
	Database  database  `yaml:"database"`
	Secure    secure    `yaml:"secure"`
	Media     media     `yaml:"media"`
	Static    static    `yaml:"static"`
	Logger    logger    `yaml:"logger"`
	LiveKit   liveKit   `yaml:"livekit"`
}

var cfg *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		cfg = &Config{}

		path := getConfigPath()

		if err := cleanenv.ReadConfig(path, cfg); err != nil {
			panic(err)
		}
	})

	return cfg
}

func getConfigPath() string {
	var path string
	flag.StringVar(&path, "config", "../config/config.yaml", "set config file")

	envPath := os.Getenv("CONFIG_PATH")

	if len(envPath) > 0 {
		path = envPath
	}

	return path
}
