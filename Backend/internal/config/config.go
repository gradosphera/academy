package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

func NewConfig() *Config {
	config := env.Must(env.ParseAs[Config]())

	return &config
}

const (
	EnvironmentProduction = "prod"
	EnvironmentStage      = "stage"
)

type Config struct {
	App   AppConfig
	HTTP  HTTPConfig
	Auth  AuthConfig
	Mux   MuxConfig
	PG    DBConfig
	Redis RedisConfig
	TON   TONConfig
}

type AppConfig struct {
	Environment         string `env:"ENVIRONMENT,required"`
	UploadDirectory     string `env:"UPLOAD_DIRECTORY,required"`
	TempUploadDirectory string `env:"TEMP_UPLOAD_DIRECTORY,required"`
	TelegramBotConfig   string `env:"TELEGRAM_BOT_CONFIG"`
	TelegramBotImage    string `env:"TELEGRAM_BOT_IMAGE"`

	EnableDemoProduct bool `env:"ENABLE_DEMO_PRODUCT"`
}

type HTTPConfig struct {
	Host             string        `env:"HTTP_HOST,required"`
	Port             string        `env:"HTTP_PORT,required"`
	AllowOrigins     string        `env:"ALLOW_ORIGINS,required"`
	AllowCredentials bool          `env:"ALLOW_CREDENTIALS,required"`
	RangeLimit       int64         `env:"HTTP_RANGE_LIMIT,required"`
	Timeout          time.Duration `env:"HTTP_TIMEOUT,required"`
	WayForPayWebhook string        `env:"WAYFORPAY_WEBHOOK,required"`
}

type AuthConfig struct {
	RefreshTokenSignKey string        `env:"REFRESH_TOKEN_SIGN_KEY,required"`
	AccessTokenSignKey  string        `env:"ACCESS_TOKEN_SIGN_KEY,required"`
	RefreshTokenTTL     time.Duration `env:"REFRESH_TOKEN_TTL,required"`
	AccessTokenTTL      time.Duration `env:"ACCESS_TOKEN_TTL,required"`
	TelegramTokenTTL    time.Duration `env:"TELEGRAM_TOKEN_TTL,required"`
	TelegramBotToken    string        `env:"TELEGRAM_BOT_TOKEN,required"`
	EncryptionKey       string        `env:"ENCRYPTION_KEY,required"`

	SkipSecurityKey        string `env:"SKIP_SECURITY_KEY"`
	WayForPayDisableRefund bool   `env:"WAYFORPAY_DISABLE_REFUND"`
}

type MuxConfig struct {
	MuxTokenID           string `env:"MUX_TOKEN_ID"`
	MuxTokenSecret       string `env:"MUX_TOKEN_SECRET"`
	MuxSigningKey        string `env:"MUX_SIGNING_KEY"`
	MuxSigningPrivateKey string `env:"MUX_SIGNING_PRIVATE_KEY"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST,required"`
	DB       int    `env:"REDIS_DB,required"`
	Port     string `env:"REDIS_PORT,required"`
	Password string `env:"REDIS_PASSWORD,required"`
}

type TONConfig struct {
	Network         string `env:"TON_NETWORK,required"`
	ConfigURL       string `env:"TON_CONFIG_URL,required"`
	AcceptedJettons string `env:"TON_ACCEPTED_JETTONS"`

	TONAPIKey string `env:"TONAPI_KEY"`

	WayForPayLogin     string `env:"WAYFORPAY_LOGIN,required"`
	WayForPaySecretKey string `env:"WAYFORPAY_SECRET_KEY,required"`
}

type DBConfig struct {
	User     string `env:"POSTGRES_USER,required"`
	Password string `env:"POSTGRES_PASSWORD,required"`
	Port     string `env:"POSTGRES_PORT,required"`
	Host     string `env:"POSTGRES_HOST,required"`
	Database string `env:"POSTGRES_DATABASE,required"`
	SSLMode  string `env:"POSTGRES_SSLMODE,required"`
}

func (c *DBConfig) GetConnectString() string {
	info := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Database,
		c.SSLMode,
	)

	return info
}
