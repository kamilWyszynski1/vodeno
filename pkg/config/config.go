package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port    int           `json:"port" mapstructure:"port"`
	DB      DBConfig      `json:"db" mapstructure:"db"`
	Watcher WatcherConfig `json:"watcher" mapstructure:"watcher"`
}

type DBConfig struct {
	Host                string `json:"host" mapstructure:"host"`
	Port                int    `json:"port" mapstructure:"port"`
	DbName              string `json:"name" mapstructure:"name"`
	User                string `json:"user" mapstructure:"user"`
	Password            string `json:"password" mapstructure:"password"`
	SSLEnabled          bool   `json:"ssl_enabled" mapstructure:"ssl_enabled"`
	ConnMaxLifetimeSecs int    `json:"conn_max_lifetime_secs" mapstructure:"conn_max_lifetime_secs"`
	MaxOpenConns        int    `json:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns        int    `json:"max_idle_conns" mapstructure:"max_idle_conns"`
}

type WatcherConfig struct {
	TickPeriod time.Duration `json:"tick_period" mapstructure:"tick_period"`
}

// ConnectionString returns database connection string.
func (c DBConfig) ConnectionString() string {
	sslMode := "disable"

	if c.SSLEnabled {
		sslMode = "require"
	}

	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		c.User,
		c.Password,
		c.DbName,
		c.Host,
		c.Port,
		sslMode,
	)
}

const (
	configPathEnv = "CONFIG_PATH"
	// defaultConnMaxLifetimeSecs is the default maximum amount of time a connection may be reused.
	defaultConnMaxLifetimeSecs = 30
	// defaultMaxOpenConns is the default maximum number of open connections to the database.
	defaultMaxOpenConns = 3
	// defaultMaxIdleConns is the default maximum number of connections in the idle connection pool.
	defaultMaxIdleConns = 1
)

// Load loads configuration from the specified file.
func Load() (*Config, error) {
	viper := viper.New()
	viper.SetConfigName("config")

	configPath := os.Getenv(configPathEnv)
	if configPath != "" {
		viper.SetConfigFile(configPath)
	}

	viper.SetDefault("db.conn_max_lifetime_secs", defaultConnMaxLifetimeSecs)
	viper.SetDefault("db.max_open_conns", defaultMaxOpenConns)
	viper.SetDefault("db.max_idle_conns", defaultMaxIdleConns)

	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
