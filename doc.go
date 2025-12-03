/*
Package envload provides a lightweight, zero-dependency solution for loading
environment variables from .env files into Go structs with type-safe parsing.

# Overview

envload automatically maps environment variables to struct fields using struct tags,
with support for default values, required field validation, and comprehensive type conversion.

# Features

• Zero Dependencies - Uses only Go standard library
• Type-Safe Parsing - Automatic conversion with validation
• Rich Type Support - Strings, numbers, booleans, durations, slices, maps
• Struct Tags - Simple configuration via env, default, and required tags
• Error Handling - Comprehensive validation and helpful error messages
• Production Ready - Optimized for application startup patterns

# Basic Usage

Create a .env file:

	APP_NAME=MyApp
	PORT=8080
	DEBUG=true
	TIMEOUT=30s

Define your configuration struct:

	type Config struct {
		AppName string        `env:"APP_NAME" default:"DefaultApp"`
		Port    int           `env:"PORT" default:"3000"`
		Debug   bool          `env:"DEBUG" default:"false"`
		Timeout time.Duration `env:"TIMEOUT" default:"10s"`
	}

Load the configuration:

	var config Config
	err := envload.LoadAndParse(".env", &config)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

# Struct Tags

env - Maps struct field to environment variable name

	Field string `env:"ENV_VAR_NAME"`

default - Provides fallback value when environment variable is missing

	Port int `env:"PORT" default:"8080"`

required - Marks field as mandatory (fails if missing and no default)

	APIKey string `env:"API_KEY" required:"true"`

# Supported Data Types

Basic Types:
• Strings: string
• Integers: int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64
• Floats: float32, float64
• Booleans: bool (accepts true/false, 1/0, yes/no, on/off)
• Durations: time.Duration (e.g., "30s", "5m", "1h")

Collection Types:
• Slices: []string, []int, []float64, []bool (comma-separated values)
• Maps: map[string]T (comma-separated key:value pairs, string keys only)

# Examples

String and basic types:

	type Config struct {
		AppName     string        `env:"APP_NAME" default:"MyApp"`
		Port        int           `env:"PORT" default:"8080"`
		EnableSSL   bool          `env:"ENABLE_SSL" default:"true"`
		Timeout     time.Duration `env:"TIMEOUT" default:"30s"`
	}

Slices (comma-separated values, empty values automatically filtered):

	type Config struct {
		Tags    []string  `env:"TAGS" default:"web,api,service"`
		Ports   []int     `env:"PORTS" default:"8080,9090,3000"`
		Enabled []bool    `env:"ENABLED" default:"true,false,true"`
	}
	# Environment: TAGS=frontend,,backend,mobile
	# Result: ["frontend", "backend", "mobile"] (empty value skipped)

Maps (comma-separated key:value pairs):

	type Config struct {
		Labels   map[string]string `env:"LABELS" default:"env:prod,team:backend"`
		Features map[string]bool   `env:"FEATURES" default:"cache:true,debug:false"`
		Limits   map[string]int    `env:"LIMITS" default:"cpu:80,memory:512"`
	}

Required fields:

	type Config struct {
		DatabaseURL string `env:"DATABASE_URL" required:"true"`
		APIKey      string `env:"API_KEY" required:"true"`

		# This works - has default even though required
		AppName     string `env:"APP_NAME" required:"true" default:"MyApp"`
	}

# Complete Example

	package config

	import (
		"bitbucket.org/bookingkoala/genes/envload"
		"sync"
		"time"
	)

	type Config struct {
		# Application settings
		AppName string `env:"APP_NAME" default:"MyApp"`
		Port    int    `env:"PORT" default:"8080"`
		Debug   bool   `env:"DEBUG" default:"false"`

		# Database settings
		DatabaseURL string        `env:"DATABASE_URL" required:"true"`
		DBTimeout   time.Duration `env:"DB_TIMEOUT" default:"30s"`

		# Feature configuration
		Features   map[string]bool `env:"FEATURES"`
		Tags       []string        `env:"TAGS" default:"web,api"`

		# Optional settings
		RedisHosts []string `env:"REDIS_HOSTS" default:"localhost:6379"`
	}

	var (
		instance Config
		once     sync.Once
	)

	func Load() error {
		var err error
		once.Do(func() {
			err = envload.LoadAndParse(".env", &instance)
		})
		return err
	}

	func Get() Config {
		return instance
	}

# Error Handling

LoadAndParse returns descriptive errors for various failure cases:

• Missing required fields: "required field 'APIKey' (env: API_KEY) is missing and has no default value"
• Invalid type conversion: "invalid int for field 'Port': strconv.ParseInt: parsing \"abc\": invalid syntax"
• Invalid target: "target must be a pointer to struct"

If the .env file doesn't exist, envload logs a warning and continues with default values,
allowing for graceful degradation.

# Performance

envload is optimized for typical application startup patterns where configuration
is loaded once during initialization. Memory usage is minimal (~2-10KB per load)
and reflection overhead is negligible for typical configuration structs.

For production applications, use the sync.Once pattern to load configuration
once and reuse throughout the application lifecycle.

# Limitations

• Nested structs are not supported - use flat structures
• Pointer fields are not supported - use value types
• Map keys must be strings
• Slice elements must be basic types (string, int, float, bool)

For detailed documentation and examples, see the README.md file.
*/
package envload
