package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
    Server       ServerConfig
    Database     DatabaseConfig
    JWT          JWTConfig
    CORS         CORSConfig
    Logging      LoggingConfig
    Meal         MealConfig
    Cleanup      CleanupConfig
    RateLimit    RateLimitConfig
    WorkLocation WorkLocationConfig
    Headcount HeadcountConfig
}

type ServerConfig struct {
    Host string
    Port int
    Env  string
}

type DatabaseConfig struct {
    Host            string
    Port            string
    User            string
    Password        string
    Name            string
    SSLMode         string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
}

type JWTConfig struct {
    Secret     string
    Expiration time.Duration
}

type CORSConfig struct {
    AllowedOrigins []string
}

type LoggingConfig struct {
    Level  string
    Format string
}

type MealConfig struct {
    CutoffTime     string
    CutoffTimezone string
    WeekendDays    []string
    ForwardWindowDays int
}

type CleanupConfig struct {
    RetentionMonths int
    CronSchedule    string
}

type RateLimitConfig struct {
    Enabled           bool
    RequestsPerMinute int
}

type WorkLocationConfig struct {
    MonthlyWFHAllowance int
}

type HeadcountConfig struct {
    MaxForecastDays int
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName(".env")
    viper.SetConfigType("env")
    viper.AddConfigPath(".")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("error reading config file: %w", err)
        }
        fmt.Println("Warning: .env file not found, using environment variables")
    }

    setDefaults()

    config := &Config{
        Server: ServerConfig{
            Host: viper.GetString("SERVER_HOST"),
            Port: viper.GetInt("SERVER_PORT"),
            Env:  viper.GetString("ENV"),
        },
        Database: DatabaseConfig{
            Host:            viper.GetString("DB_HOST"),
            Port:            viper.GetString("DB_PORT"),
            User:            viper.GetString("DB_USER"),
            Password:        viper.GetString("DB_PASSWORD"),
            Name:            viper.GetString("DB_NAME"),
            SSLMode:         viper.GetString("DB_SSLMODE"),
            MaxOpenConns:    viper.GetInt("DB_MAX_OPEN_CONNS"),
            MaxIdleConns:    viper.GetInt("DB_MAX_IDLE_CONNS"),
            ConnMaxLifetime: viper.GetDuration("DB_CONN_MAX_LIFETIME"),
        },
        JWT: JWTConfig{
            Secret:     viper.GetString("JWT_SECRET"),
            Expiration: viper.GetDuration("JWT_EXPIRATION"),
        },
        CORS: CORSConfig{
            AllowedOrigins: parseCommaSeparated(viper.GetString("CORS_ALLOWED_ORIGINS")),
        },
        Logging: LoggingConfig{
            Level:  viper.GetString("LOG_LEVEL"),
            Format: viper.GetString("LOG_FORMAT"),
        },
        Meal: MealConfig{
            CutoffTime:     viper.GetString("MEAL_CUTOFF_TIME"),
            CutoffTimezone: viper.GetString("MEAL_CUTOFF_TIMEZONE"),
            WeekendDays:    parseCommaSeparated(viper.GetString("MEAL_WEEKEND_DAYS")),
            ForwardWindowDays: viper.GetInt("MEAL_FORWARD_WINDOW_DAYS"),
        },
        Cleanup: CleanupConfig{
            RetentionMonths: viper.GetInt("HISTORY_RETENTION_MONTHS"),
            CronSchedule:    viper.GetString("CLEANUP_CRON"),
        },
        RateLimit: RateLimitConfig{
            Enabled:           viper.GetBool("RATE_LIMIT_ENABLED"),
            RequestsPerMinute: viper.GetInt("RATE_LIMIT_REQUESTS_PER_MINUTE"),
        },
        WorkLocation: WorkLocationConfig{
            MonthlyWFHAllowance: viper.GetInt("WORK_LOCATION_MONTHLY_WFH_ALLOWANCE"),
        },
        Headcount: HeadcountConfig{
            MaxForecastDays: viper.GetInt("HEADCOUNT_MAX_FORECAST_DAYS"),
        },
    }

    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    return config, nil
}

func setDefaults() {
    viper.SetDefault("SERVER_HOST", "localhost")
    viper.SetDefault("SERVER_PORT", 8080)
    viper.SetDefault("ENV", "development")

    viper.SetDefault("DB_HOST", "localhost")
    viper.SetDefault("DB_PORT", "5432")
    viper.SetDefault("DB_USER", "postgres")
    viper.SetDefault("DB_NAME", "craftsbite_db")
    viper.SetDefault("DB_SSLMODE", "disable")
    viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
    viper.SetDefault("DB_MAX_IDLE_CONNS", 5)
    viper.SetDefault("DB_CONN_MAX_LIFETIME", "5m")

    viper.SetDefault("JWT_EXPIRATION", "24h")

    viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

    viper.SetDefault("LOG_LEVEL", "info")
    viper.SetDefault("LOG_FORMAT", "json")

    viper.SetDefault("MEAL_CUTOFF_TIME", "21:00")
    viper.SetDefault("MEAL_CUTOFF_TIMEZONE", "Asia/Dhaka")
    viper.SetDefault("MEAL_WEEKEND_DAYS", "Saturday,Sunday")
    viper.SetDefault("MEAL_FORWARD_WINDOW_DAYS", 7)

    viper.SetDefault("HISTORY_RETENTION_MONTHS", 3)
    viper.SetDefault("CLEANUP_CRON", "0 2 * * *")

    viper.SetDefault("RATE_LIMIT_ENABLED", true)
    viper.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)

    viper.SetDefault("WORK_LOCATION_MONTHLY_WFH_ALLOWANCE", 5)
    viper.SetDefault("HEADCOUNT_MAX_FORECAST_DAYS", 14)
}   

func (c *Config) Validate() error {
    if c.JWT.Secret == "" {
        return fmt.Errorf("JWT_SECRET is required")
    }
    if len(c.JWT.Secret) < 32 {
        return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
    }
    if c.Database.Password == "" {
        return fmt.Errorf("DB_PASSWORD is required")
    }

    validEnvs := map[string]bool{
        "development": true,
        "staging":     true,
        "production":  true,
        "test":        true,
    }
    if !validEnvs[c.Server.Env] {
        return fmt.Errorf("ENV must be one of: development, staging, production, test")
    }

    return nil
}

func (c *DatabaseConfig) GetDSN() string {
    return fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
    )
}

func (c *Config) IsProduction() bool  { return c.Server.Env == "production" }
func (c *Config) IsDevelopment() bool { return c.Server.Env == "development" }

func (c *ServerConfig) GetServerAddress() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func parseCommaSeparated(s string) []string {
    if s == "" {
        return []string{}
    }
    parts := strings.Split(s, ",")
    result := make([]string, 0, len(parts))
    for _, part := range parts {
        if trimmed := strings.TrimSpace(part); trimmed != "" {
            result = append(result, trimmed)
        }
    }
    return result
}
