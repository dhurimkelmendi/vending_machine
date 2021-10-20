// Package config reads configuration values from environment variables. It
// exports each of these values as a global variable, available to the entire
// application.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Env is the overall environment to fetch configuration from.
type Env string

// The different overall environments that are supported.
const (
	EnvDevelopment Env = "development"
	EnvTest        Env = "test"
	EnvProduction  Env = "production"
	EnvStaging     Env = "staging"
)

// Config holds configuration values.
type Config struct {
	// Env is the current environment the configs were reading from.
	Env Env

	// APIOrigin is the publicly reachable origin of the API server.
	// It must include protocol, host and port (except 80 and 443).
	// It must not have trailing slash.
	// For example "https://tests.api.com:8080"
	APIOrigin string

	// AllowAllCORSOrigins determines if all cross origin requests from all origins should be allowed or not.
	AllowAllCORSOrigins bool

	// CORSOrigins is a comma separated list of origins (i.e. https://admin.api.com:2371)
	// that're allowed to make cross-origin requests to the API server.
	CORSOrigins string

	// DatabaseHost is the host of the Postgres database the application will connect to.
	DatabaseHost string

	// DatabasePort is the port of the Postgres database the application will connect to.
	DatabasePort string

	// DatabaseName is the name of the Postgres database the application will connect to.
	DatabaseName string

	// DatabaseUsername is the Postgres username of the user who is connecting to the database.
	DatabaseUsername string

	// DatabasePassword is the Postgres password of the user who is connecting to the database.
	DatabasePassword string

	// DebugDatabase if enabled, will display all queries sent to database
	DebugDatabase bool

	// HTTPAddr is the port to start the web server on.
	HTTPAddr string

	// JWTSecret is the JWT secret used to generate tokens - must be at least 64 bytes long!
	JWTSecret string

	// APISecret is the JWT secret used to generate service-to-service tokens - must be at least 64 bytes long!
	APISecret string

	// APIHost is the host (with protocol) to the  API without trailing slash, eg: https://staging.api.com
	APIHost string

	// RespondWithInnerError determines if API error response should include inner error messages.
	RespondWithInnerError bool

	// AcceptableDepositAmountValues specifies the amounts acceptable for deposit
	AcceptableDepositAmountValues []int32
}

var defaultInstance *Config

// GetDefaultInstance returns the default instance of Config.
func GetDefaultInstance() *Config {
	if defaultInstance == nil {
		defaultInstance = &Config{
			Env: getEnv(),
		}
		defaultInstance.readConfigs()
	}
	return defaultInstance
}

// SetLogLevel checks LOG_LEVEL; in case of not set, default to debug, except we are on production, where the default must be info level
func (c *Config) SetLogLevel() {
	var el string
	if el = os.Getenv("LOG_LEVEL"); el == "" {
		el = getDefaultLevel(string(c.Env))
	}

	ll, err := logrus.ParseLevel(el)

	switch {
	case err == nil:
		logrus.SetLevel(ll)
	case c.Env == EnvProduction:
		logrus.SetLevel(logrus.InfoLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Warnf("LOG_LEVEL set to %+v", logrus.GetLevel())
}

func getDefaultLevel(env string) string {
	if env == string(EnvProduction) {
		return "info"
	}

	return "debug"
}

func (c *Config) readConfigs() {
	appConfig := envConfigGetter(os.Getenv, c.Env == EnvProduction || c.Env == EnvStaging)

	c.APIOrigin = appConfig.GetConfig("API_ORIGIN", "")
	c.CORSOrigins = appConfig.GetConfig("CORS_ORIGINS", "")
	c.DatabaseHost = appConfig.GetConfig("DB_HOST", "localhost")
	c.DatabasePort = appConfig.GetConfig("DB_PORT", "5432")
	c.DatabaseName = appConfig.GetConfig("DB_NAME", "vending_machine_db")
	c.DatabaseUsername = appConfig.GetConfig("DB_USERNAME", "vending_machine")
	c.DatabasePassword = appConfig.GetConfig("DB_PASSWORD", "vending_machine_pass")
	c.HTTPAddr = appConfig.GetConfig("HTTP_ADDR", ":8080")
	c.JWTSecret = appConfig.GetConfig("JWT_SECRET", "jwt_secret_signing_key")
	c.APISecret = appConfig.GetConfig("API_SECRET", "app_secret_signing_key")
	c.APIHost = appConfig.GetConfig("API_HOST", "http://localhost:8080")
	c.AcceptableDepositAmountValues = []int32{5, 10, 20, 50, 100}

	// Set flags
	c.DebugDatabase = appConfig.GetFlag("DEBUG_DATABASE", false)
	c.AllowAllCORSOrigins = c.Env == EnvDevelopment
	c.RespondWithInnerError = c.Env != EnvProduction
}

// LogConfigs logs the config values.
func (c *Config) LogConfigs() {
	logrus.Warn("[Config] Values:")
	logrus.Warn(fmt.Sprintf("  * Environment: %+v", c.Env))
	logrus.Warn(fmt.Sprintf("  * APIOrigin: %+v", c.APIOrigin))
	logrus.Warn(fmt.Sprintf("  * CORSOrigins: %+v", c.CORSOrigins))
	logrus.Warn(fmt.Sprintf("  * DebugDatabase: %+v", c.DebugDatabase))
	logrus.Warn(fmt.Sprintf("  * DatabaseHost: %+v", c.DatabaseHost))
	logrus.Warn(fmt.Sprintf("  * DatabasePort: %+v", c.DatabasePort))
	logrus.Warn(fmt.Sprintf("  * DatabaseName: %+v", c.DatabaseName))
	logrus.Warn(fmt.Sprintf("  * DatabaseUsername: %+v", c.DatabaseUsername))
	logrus.Warn(fmt.Sprintf("  * DatabasePassword: %+v", strings.Repeat("*", len(c.DatabasePassword))))
	logrus.Warn(fmt.Sprintf("  * HTTPAddr: %+v", c.HTTPAddr))
	logrus.Warn(fmt.Sprintf("  * JWTSecret: %+v", c.JWTSecret))
	logrus.Warn(fmt.Sprintf("  * APISecret: %+v", c.APISecret))
	logrus.Warn(fmt.Sprintf("  * APIHost: %+v", c.APIHost))
	logrus.Warn(fmt.Sprintf("  * AcceptableDepositAmountValues: %+v", c.AcceptableDepositAmountValues))
}
