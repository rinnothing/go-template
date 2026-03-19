package config

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type (
	Config struct {
		Logger `koanf:"logger"`
		HTTP   `koanf:"http"`
		PG     `koanf:"pg"`
		Redis  `koanf:"redis"`
		Kafka  `koanf:"kafka"`
	}

	Logger struct {
		Level    string `koanf:"level"`
		Filepath string `koanf:"filepath"`
	}

	PG struct {
		URL      string
		Host     string `koanf:"host"`
		Port     string `koanf:"port"`
		DB       string `koanf:"db"`
		User     string `koanf:"user"`
		Password string `koanf:"password"`
		MaxConn  string `koanf:"max_conn"`
	}

	Redis struct {
		Addr           string
		Host           string        `koanf:"host"`
		Port           string        `koanf:"port"`
		DB             int           `koanf:"db"`
		MaxActiveConns int           `koanf:"max_active_conns"`
		ExpirationTime time.Duration `koanf:"expiration_time"`
	}

	Kafka struct {
		WorkersNum int `koanf:"workers_num"`
	}

	HTTP struct {
		Port string `koanf:"port"`
	}
)

var (
	k      = koanf.New(".")
	parser = yaml.Parser()
)

const ENV_PREFIX = "GO_TMP_"

func New(path string) (*Config, error) {
	if err := k.Load(file.Provider(path), parser); err != nil {
		return nil, err
	}

	k.Load(env.Provider(".", env.Opt{
		Prefix: ENV_PREFIX,
		TransformFunc: func(k, v string) (string, any) {
			k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, ENV_PREFIX)), "_", ".")
			return k, v
		},
	}), nil)

	cfg := Config{}
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	err := checkPortFormat(cfg.HTTP.Port, "http.port")
	if err != nil {
		return nil, err
	}

	err = checkNumFormat(cfg.PG.MaxConn, "pg.max_conn")
	if err != nil {
		return nil, err
	}

	cfg.PG.URL = fillPostgresURL(
		cfg.PG.User,
		cfg.PG.Password,
		cfg.PG.Host,
		cfg.PG.Port,
		cfg.PG.DB,
		cfg.PG.MaxConn,
	)

	cfg.Redis.Addr = net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port)

	return &cfg, nil
}

var ErrPortFmt = errors.New("config variable isn't in port format")
var ErrNumFmt = errors.New("config variable isn't a natural number")

func checkPortFormat(port string, varName string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 0 || n >= 1<<16 {
		return fmt.Errorf("%w: var=%s, port=%s", ErrPortFmt, varName, port)
	}

	return nil
}

func checkNumFormat(num string, varName string) error {
	val, err := strconv.Atoi(num)
	if err != nil || val < 1 {
		if err == nil {
			err = errors.New("less then 1")
		}
		return fmt.Errorf("%w: var=%s, port=%s, %w", ErrNumFmt, varName, num, err)
	}

	return nil
}

func fillPostgresURL(user, password, host, port, db, maxConn string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&pool_max_conns=%s",
		user,
		password,
		host,
		port,
		db,
		maxConn,
	)
}
