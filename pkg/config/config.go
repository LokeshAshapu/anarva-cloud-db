package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string         `mapstructure:"ENVIRONMENT"`
	Server      ServerConfig   `mapstructure:"SERVER"`
	Database    DatabaseConfig `mapstructure:"DATABASE"`
	Redis       RedisConfig    `mapstructure:"REDIS"`
	JWT         JWTConfig      `mapstructure:"JWT"`
	Storage     StorageConfig  `mapstructure:"STORAGE"`
	Metrics     MetricsConfig  `mapstructure:"METRICS"`
	Provider    ProviderConfig `mapstructure:"PROVIDER"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"PORT"`
	GRPCPort     int           `mapstructure:"GRPC_PORT"`
	ReadTimeout  time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"WRITE_TIMEOUT"`
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"URL"`
	Host            string        `mapstructure:"HOST"`
	Port            int           `mapstructure:"PORT"`
	User            string        `mapstructure:"USER"`
	Password        string        `mapstructure:"PASSWORD"`
	DBName          string        `mapstructure:"DB_NAME"`
	SSLMode         string        `mapstructure:"SSL_MODE"`
	MaxOpenConns    int           `mapstructure:"MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"CONN_MAX_LIFETIME"`
}

func (db DatabaseConfig) DSN() string {
	urlStr := strings.TrimSpace(db.URL)
	if urlStr == "" {
		urlStr = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if urlStr == "" {
		urlStr = strings.TrimSpace(os.Getenv("POSTGRES_URL"))
	}
	if urlStr == "" {
		urlStr = strings.TrimSpace(os.Getenv("DB_URL"))
	}

	if urlStr != "" {
		return urlStr
	}

	env := strings.ToLower(strings.TrimSpace(os.Getenv("ANARVA_ENV")))
	if env == "" {
		env = strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	}

	// Production Fail-Closed Requirement: Never fall back to localhost in production
	if env == "production" {
		return ""
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.DBName, db.SSLMode)
}

type RedisConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	Password string `mapstructure:"PASSWORD"`
	DB       int    `mapstructure:"DB"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"SECRET"`
	AccessExpiry  time.Duration `mapstructure:"ACCESS_EXPIRY"`
	RefreshExpiry time.Duration `mapstructure:"REFRESH_EXPIRY"`
	Issuer        string        `mapstructure:"ISSUER"`
}

type StorageConfig struct {
	Driver          string `mapstructure:"DRIVER"` // local or s3
	LocalPath       string `mapstructure:"LOCAL_PATH"`
	S3Endpoint      string `mapstructure:"S3_ENDPOINT"`
	S3Region        string `mapstructure:"S3_REGION"`
	S3Bucket        string `mapstructure:"S3_BUCKET"`
	S3AccessKey     string `mapstructure:"S3_ACCESS_KEY"`
	S3SecretKey     string `mapstructure:"S3_SECRET_KEY"`
	S3UseSSL        bool   `mapstructure:"S3_USE_SSL"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"ENABLED"`
	Path    string `mapstructure:"PATH"`
	Port    int    `mapstructure:"PORT"`
}

type ProviderConfig struct {
	Mode               string `mapstructure:"MODE"` // local or real
	AWSAccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWSRegion          string `mapstructure:"AWS_REGION"`
}

// LoadConfig loads configuration from path or environment variables.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	v.SetDefault("ENVIRONMENT", "development")
	v.SetDefault("SERVER.PORT", 8080)
	v.SetDefault("SERVER.GRPC_PORT", 9090)
	v.SetDefault("SERVER.READ_TIMEOUT", 15*time.Second)
	v.SetDefault("SERVER.WRITE_TIMEOUT", 15*time.Second)

	v.BindEnv("DATABASE.DATABASE_URL", "DATABASE_URL")
	v.BindEnv("DATABASE.DATABASE_URL", "POSTGRES_URL")
	v.BindEnv("DATABASE.DATABASE_URL", "DB_URL")
	v.SetDefault("DATABASE.DATABASE_URL", "")

	v.SetDefault("DATABASE.HOST", "localhost")
	v.SetDefault("DATABASE.PORT", 5432)
	v.SetDefault("DATABASE.USER", "anarva_admin")
	v.SetDefault("DATABASE.PASSWORD", "anarva_password")
	v.SetDefault("DATABASE.DB_NAME", "anarva_cloud_db")
	v.SetDefault("DATABASE.SSL_MODE", "disable")
	v.SetDefault("DATABASE.MAX_OPEN_CONNS", 25)
	v.SetDefault("DATABASE.MAX_IDLE_CONNS", 5)
	v.SetDefault("DATABASE.CONN_MAX_LIFETIME", 15*time.Minute)

	v.SetDefault("REDIS.HOST", "localhost")
	v.SetDefault("REDIS.PORT", 6379)
	v.SetDefault("REDIS.PASSWORD", "")
	v.SetDefault("REDIS.DB", 0)

	v.SetDefault("JWT.SECRET", "anarva_cloud_db_super_secret_jwt_key_2026")
	v.SetDefault("JWT.ACCESS_EXPIRY", 24*time.Hour)
	v.SetDefault("JWT.REFRESH_EXPIRY", 7*24*time.Hour)
	v.SetDefault("JWT.ISSUER", "anarva-cloud-db")

	v.BindEnv("STORAGE.DRIVER", "STORAGE_PROVIDER", "STORAGE_DRIVER")
	v.BindEnv("STORAGE.S3_BUCKET", "STORAGE_S3_BUCKET")
	v.BindEnv("STORAGE.S3_REGION", "STORAGE_S3_REGION")
	v.BindEnv("STORAGE.S3_ACCESS_KEY", "STORAGE_S3_ACCESS_KEY")
	v.BindEnv("STORAGE.S3_SECRET_KEY", "STORAGE_S3_SECRET_KEY")
	v.BindEnv("STORAGE.S3_ENDPOINT", "STORAGE_S3_ENDPOINT")

	v.SetDefault("STORAGE.DRIVER", "local")
	v.SetDefault("STORAGE.LOCAL_PATH", "./data/storage")
	v.SetDefault("STORAGE.S3_REGION", "auto")

	v.SetDefault("METRICS.ENABLED", true)
	v.SetDefault("METRICS.PATH", "/metrics")
	v.SetDefault("METRICS.PORT", 9091)

	v.SetDefault("PROVIDER.MODE", "local")
	v.SetDefault("PROVIDER.AWS_REGION", "us-east-1")

	if path != "" {
		v.AddConfigPath(path)
		v.SetConfigName("config")
		v.SetConfigType("yaml")

		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into config struct: %w", err)
	}

	if dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); dbURL != "" {
		cfg.Database.URL = dbURL
	} else if dbURL := strings.TrimSpace(os.Getenv("POSTGRES_URL")); dbURL != "" {
		cfg.Database.URL = dbURL
	} else if dbURL := strings.TrimSpace(os.Getenv("DB_URL")); dbURL != "" {
		cfg.Database.URL = dbURL
	}

	if prov := strings.TrimSpace(os.Getenv("STORAGE_PROVIDER")); prov != "" {
		cfg.Storage.Driver = prov
	} else if prov := strings.TrimSpace(os.Getenv("STORAGE_DRIVER")); prov != "" {
		cfg.Storage.Driver = prov
	}
	if bkt := strings.TrimSpace(os.Getenv("STORAGE_S3_BUCKET")); bkt != "" {
		cfg.Storage.S3Bucket = bkt
	}
	if reg := strings.TrimSpace(os.Getenv("STORAGE_S3_REGION")); reg != "" {
		cfg.Storage.S3Region = reg
	}
	if key := strings.TrimSpace(os.Getenv("STORAGE_S3_ACCESS_KEY")); key != "" {
		cfg.Storage.S3AccessKey = key
	}
	if sec := strings.TrimSpace(os.Getenv("STORAGE_S3_SECRET_KEY")); sec != "" {
		cfg.Storage.S3SecretKey = sec
	}
	if ep := strings.TrimSpace(os.Getenv("STORAGE_S3_ENDPOINT")); ep != "" {
		cfg.Storage.S3Endpoint = ep
	}

	if strings.ToLower(cfg.Environment) == "production" {
		if cfg.JWT.Secret == "" || cfg.JWT.Secret == "anarva_cloud_db_super_secret_jwt_key_2026" {
			cfg.JWT.Secret = generateSecureProductionJWTSecret()
		}
	}

	if err := ValidateProductionConfig(&cfg); err != nil && strings.ToLower(cfg.Environment) == "production" {
		return nil, err
	}

	return &cfg, nil
}

// ValidateProductionConfig validates configuration rules. In production, it fails closed.
func ValidateProductionConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("CONFIG_VALIDATION_FAILURE: configuration object is nil")
	}

	env := strings.ToLower(cfg.Environment)
	if env == "production" {
		if cfg.Server.Port <= 0 {
			return fmt.Errorf("CONFIG_VALIDATION_FAILURE: SERVER_PORT is required in production")
		}
		if cfg.JWT.Secret == "" || cfg.JWT.Secret == "anarva_cloud_db_super_secret_jwt_key_2026" {
			return fmt.Errorf("CONFIG_VALIDATION_FAILURE: Production environment requires a strong non-default JWT_SECRET")
		}
		if strings.ToLower(cfg.Storage.Driver) == "local" || strings.ToLower(cfg.Storage.Driver) == "" {
			return fmt.Errorf("CONFIG_VALIDATION_FAILURE: Production environment forbids LocalStorageProvider (STORAGE_PROVIDER=local). STORAGE_PROVIDER=s3 with valid S3/R2 credentials is required.")
		}
		if strings.ToLower(cfg.Storage.Driver) == "s3" {
			if cfg.Storage.S3Bucket == "" || cfg.Storage.S3AccessKey == "" || cfg.Storage.S3SecretKey == "" {
				return fmt.Errorf("CONFIG_VALIDATION_FAILURE: STORAGE_PROVIDER=s3 in production requires STORAGE_S3_BUCKET, STORAGE_S3_ACCESS_KEY, and STORAGE_S3_SECRET_KEY")
			}
		}
		if strings.ToLower(cfg.Provider.Mode) == "real" {
			if cfg.Provider.AWSAccessKeyID == "" || cfg.Provider.AWSSecretAccessKey == "" {
				return fmt.Errorf("CONFIG_VALIDATION_FAILURE: Production real provider mode requires active AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY credentials")
			}
		}
	}

	return nil
}

func generateSecureProductionJWTSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("anarva_prod_sec_rand_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("anarva_prod_sec_%s", hex.EncodeToString(b))
}
