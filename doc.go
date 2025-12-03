/*
Package godotenv provides a lightweight, type-safe environment configuration
loader for Go applications.

# Overview

godotenv automatically maps environment variables from .env files to Go structs
using struct tags. It provides type-safe parsing, default values, required field
validation, and comprehensive type conversion — all with minimal setup.

# Installation

	go get github.com/go-fynx/godotenv

# Quick Start

Create a .env file:

	APP_NAME=MyAwesomeApp
	PORT=8080
	DEBUG=true
	TIMEOUT=30s
	DATABASE_URL=postgres://localhost/mydb

Define your configuration struct and load it:

	package main

	import (
		"log"
		"time"

		"github.com/go-fynx/godotenv"
	)

	type Config struct {
		AppName     string        `env:"APP_NAME" default:"DefaultApp"`
		Port        int           `env:"PORT" default:"3000"`
		Debug       bool          `env:"DEBUG" default:"false"`
		Timeout     time.Duration `env:"TIMEOUT" default:"10s"`
		DatabaseURL string        `env:"DATABASE_URL" required:"true"`
	}

	func main() {
		var cfg Config

		if err := godotenv.LoadAndParse(".env", &cfg); err != nil {
			log.Fatal("Failed to load config:", err)
		}

		log.Printf("App: %s running on port %d", cfg.AppName, cfg.Port)
	}

# Struct Tags

The following struct tags are supported:

	env      - Maps field to environment variable name
	         Example: `env:"PORT"`

	default  - Fallback value when env var is missing
	         Example: `default:"8080"`

	required - Fails if missing and no default
	         Example: `required:"true"`

Example usage:

	type Config struct {
		// Required field - fails if DATABASE_URL is missing
		DatabaseURL string `env:"DATABASE_URL" required:"true"`

		// Optional with default - uses 8080 if PORT is missing
		Port int `env:"PORT" default:"8080"`

		// Optional without default - empty string if missing
		LogPath string `env:"LOG_PATH"`

		// Required with default - never fails (default satisfies requirement)
		AppName string `env:"APP_NAME" required:"true" default:"MyApp"`
	}

# Supported Types

Basic Types:
  - string
  - int, int8, int16, int32, int64
  - uint, uint8, uint16, uint32, uint64
  - float32, float64
  - bool (accepts: true, false, 1, 0)
  - time.Duration (e.g., "5s", "2m", "1h30m")

Slices (comma-separated values):

	type Config struct {
		Tags    []string  `env:"TAGS" default:"web,api"`
		Ports   []int     `env:"PORTS" default:"8080,9090"`
		Enabled []bool    `env:"ENABLED"`
	}

Empty values in slices are automatically filtered.
For example, TAGS=web,,api results in ["web", "api"].

Maps (comma-separated key:value pairs):

	type Config struct {
		Labels   map[string]string `env:"LABELS"`
		Features map[string]bool   `env:"FEATURES"`
		Limits   map[string]int    `env:"LIMITS"`
	}

# Production Pattern

Use the singleton pattern for application-wide configuration:

	package config

	import (
		"sync"
		"time"

		"github.com/go-fynx/godotenv"
	)

	type Config struct {
		AppName     string            `env:"APP_NAME" default:"MyApp"`
		Port        int               `env:"PORT" default:"8080"`
		Debug       bool              `env:"DEBUG" default:"false"`
		DatabaseURL string            `env:"DATABASE_URL" required:"true"`
		DBTimeout   time.Duration     `env:"DB_TIMEOUT" default:"30s"`
		RedisHosts  []string          `env:"REDIS_HOSTS" default:"localhost:6379"`
		Features    map[string]bool   `env:"FEATURES"`
	}

	var (
		instance Config
		once     sync.Once
		loadErr  error
	)

	// Load initializes configuration (call once at startup)
	func Load(envPath string) error {
		once.Do(func() {
			loadErr = godotenv.LoadAndParse(envPath, &instance)
		})
		return loadErr
	}

	// Get returns the loaded configuration
	func Get() Config {
		return instance
	}

Usage:

	func main() {
		if err := config.Load(".env"); err != nil {
			log.Fatal("Config error:", err)
		}

		cfg := config.Get()
		log.Printf("Starting %s on port %d", cfg.AppName, cfg.Port)
	}

# Error Handling

godotenv provides descriptive errors for common issues:

  - Missing required field: "missing required field: field=DatabaseURL env=DATABASE_URL"
  - Invalid type conversion: "invalid int for field 'Port': strconv.ParseInt: parsing \"abc\": invalid syntax"
  - Invalid target: "target must be a pointer to struct"
  - Invalid duration: "invalid duration for field 'Timeout': time.ParseDuration: invalid duration \"xyz\""

If the .env file doesn't exist, godotenv logs a warning and continues
with default values only (graceful degradation).

# Limitations

  - Nested structs are not supported — use flat structures
  - Pointer fields are not supported — use value types
  - Map keys must be strings
  - Slice elements must be basic types (string, int, float, bool)

For more details and examples, see the README.md file at:
https://github.com/go-fynx/godotenv
*/
package godotenv
