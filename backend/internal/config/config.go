package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppEnv        string `env:"APP_ENV" envDefault:"development"`
	ServerPort    string `env:"SERVER_PORT" envDefault:"8080"`
	DBHost        string `env:"DB_HOST" envDefault:"localhost"`
	DBPort        string `env:"DB_PORT" envDefault:"5432"`
	DBUser        string `env:"DB_USER" envDefault:"ldstoreinventory_user"`
	DBPassword    string `env:"DB_PASSWORD" envDefault:"ldstoreinventory_pwd"`
	DBName        string `env:"DB_NAME" envDefault:"ldstoreinventory_db"`
	JWTSecret     string `env:"JWT_SECRET" envDefault:"change_me_to_a_long_random_string"`
	TokenTTLHours int    `env:"TOKEN_TTL_HOURS" envDefault:"72"`
	RateLimit     int    `env:"RATE_LIMIT" envDefault:"120"`
	RateWindowSec int    `env:"RATE_WINDOW_SEC" envDefault:"60"`
	CORSOrigins   string `env:"APP_CORS_ORIGINS" envDefault:"http://localhost:28504"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// CORSOriginsSlice 将逗号分隔的环境变量解析为 CORS 允许来源列表。
func (c *Config) CORSOriginsSlice() []string {
	return splitAndTrim(c.CORSOrigins)
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func splitAndTrim(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"http://localhost:28504"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return []string{"http://localhost:28504"}
	}
	return out
}
