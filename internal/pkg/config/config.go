package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
)

type Config struct {
	Database DatabaseConfig
	GRPC     GRPCConfig
	Grinex   GrinexConfig
	OTel     OTelConfig
	Logger   LoggerConfig
	Metrics  MetricsConfig
	Debug    DebugConfig
}

type DebugConfig struct {
	Port int
}

type MetricsConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type GRPCConfig struct {
	Port int
}

type GrinexConfig struct {
	BaseURL    string
	Timeout    time.Duration
	DepthLimit int
}

type OTelConfig struct {
	Endpoint string
	Insecure bool
}

type LoggerConfig struct {
	Level string
	Dev   bool
}

func Load() (*Config, error) {
	k := koanf.New(".")

	// Define CLI flags
	f := flag.NewFlagSet("app", flag.ContinueOnError)
	f.String("db-host", "localhost", "Database host")
	f.Int("db-port", 5432, "Database port")
	f.String("db-user", "postgres", "Database user")
	f.String("db-password", "", "Database password")
	f.String("db-name", "usdt_parser", "Database name")
	f.String("db-sslmode", "disable", "Database SSL mode")
	f.Int("grpc-port", 50051, "gRPC server port")
	f.String("grinex-base-url", "https://grinex.io", "Grinex API base URL")
	f.Duration("grinex-timeout", 10*time.Second, "Grinex API timeout")
	f.Int("grinex-depth-limit", 20, "Grinex depth API entry limit (0 = no limit)")
	f.Int("db-max-open-conns", 25, "Database max open connections")
	f.Int("db-max-idle-conns", 5, "Database max idle connections")
	f.Duration("db-conn-max-lifetime", 5*time.Minute, "Database connection max lifetime")
	f.String("otel-endpoint", "localhost:4317", "OTel collector endpoint")
	f.Bool("otel-insecure", true, "OTel insecure connection")
	f.String("log-level", "info", "Log level")
	f.Bool("log-dev", false, "Development logging mode")
	f.Int("metrics-port", 9090, "Prometheus metrics HTTP port")
	f.Int("debug-port", 6060, "pprof debug HTTP port (0 to disable)")
	_ = f.Parse([]string{})

	// Load env vars (prefix APP_, delimiter _)
	if err := k.Load(env.Provider("APP_", ".", func(s string) string {
		return strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, "APP_")), "_", ".",
		)
	}), nil); err != nil {
		return nil, err
	}

	// Load CLI flags (overrides env)
	if err := k.Load(posflag.ProviderWithFlag(f, ".", k, func(fl *flag.Flag) (string, interface{}) {
		key := strings.ReplaceAll(fl.Name, "-", ".")
		return key, posflag.FlagVal(f, fl)
	}), nil); err != nil {
		return nil, err
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:            k.String("db.host"),
			Port:            k.Int("db.port"),
			User:            k.String("db.user"),
			Password:        k.String("db.password"),
			Name:            k.String("db.name"),
			SSLMode:         k.String("db.sslmode"),
			MaxOpenConns:    k.Int("db.max.open.conns"),
			MaxIdleConns:    k.Int("db.max.idle.conns"),
			ConnMaxLifetime: k.Duration("db.conn.max.lifetime"),
		},
		GRPC: GRPCConfig{
			Port: k.Int("grpc.port"),
		},
		Grinex: GrinexConfig{
			BaseURL:    k.String("grinex.base.url"),
			Timeout:    k.Duration("grinex.timeout"),
			DepthLimit: k.Int("grinex.depth.limit"),
		},
		OTel: OTelConfig{
			Endpoint: k.String("otel.endpoint"),
			Insecure: k.Bool("otel.insecure"),
		},
		Logger: LoggerConfig{
			Level: k.String("log.level"),
			Dev:   k.Bool("log.dev"),
		},
		Metrics: MetricsConfig{
			Port: k.Int("metrics.port"),
		},
		Debug: DebugConfig{
			Port: k.Int("debug.port"),
		},
	}

	// Defaults for zero values
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.GRPC.Port == 0 {
		cfg.GRPC.Port = 50051
	}
	if cfg.Grinex.BaseURL == "" {
		cfg.Grinex.BaseURL = "https://grinex.io"
	}
	if cfg.Grinex.Timeout == 0 {
		cfg.Grinex.Timeout = 10 * time.Second
	}
	if cfg.Grinex.DepthLimit == 0 {
		cfg.Grinex.DepthLimit = 20
	}
	if cfg.OTel.Endpoint == "" {
		cfg.OTel.Endpoint = "localhost:4317"
	}
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 5
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 5 * time.Minute
	}
	if cfg.Metrics.Port == 0 {
		cfg.Metrics.Port = 9090
	}
	if cfg.Debug.Port == 0 {
		cfg.Debug.Port = 6060
	}

	return cfg, nil
}
