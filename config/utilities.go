package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// envVarGetter is a function that gets config value from environment variable.
type envVarGetter func(key string) string

// AppConfig simple interface for retrieving configuration values
type AppConfig interface {
	GetConfig(key, defaultValue string) string
	GetFlag(key string, defaultValue bool) bool
}

// EnvironmentAppConfig attaches the os environment
type EnvironmentAppConfig struct {
	env         envVarGetter
	warnMissing bool
}

// GetConfig returns a string value from the environment
func (e *EnvironmentAppConfig) GetConfig(key, defaultValue string) string {
	v := e.env(key)

	if v == "" {
		if e.warnMissing {
			logrus.Errorf("No value set for environment variable: [ %+v ]; Using default value: `%+v`", key, defaultValue)
		}

		return defaultValue
	}

	return v
}

// GetFlag returns a boolean flag value from the environment
func (e *EnvironmentAppConfig) GetFlag(key string, defaultValue bool) bool {
	v := e.env(key)

	if v == "" {
		if e.warnMissing {
			logrus.Errorf("No value set for environment variable: [ %+v ]; Using default value: `%+v`", key, defaultValue)
		}

		return defaultValue
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		if e.warnMissing {
			logrus.Errorf("Invalid boolean value set for environment variable: [ %+v ]; Using default value: `%+v`", key, defaultValue)
		}

		return defaultValue
	}

	return b
}

// envConfigGetter returns a configGetter that gets configs with the given envGetter.
func envConfigGetter(envGetter envVarGetter, warnMissing bool) AppConfig {
	return &EnvironmentAppConfig{env: envGetter, warnMissing: warnMissing}
}

// getEnv detects the current server environment.
func getEnv() Env {
	if flag.Lookup("test.v") != nil {
		return EnvTest
	}

	env := Env(os.Getenv("ENV"))

	if env == EnvTest {
		return EnvTest
	}

	if env == EnvProduction {
		return EnvProduction
	}

	if env == EnvStaging {
		return EnvStaging
	}

	return EnvDevelopment
}
