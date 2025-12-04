# envload

<div align="center">

![envload](env.png)

**A lightweight, type-safe environment configuration loader for Go**

[![Go Reference](https://pkg.go.dev/badge/github.com/go-fynx/envload.svg)](https://pkg.go.dev/github.com/go-fynx/envload)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-fynx/envload)](https://goreportcard.com/report/github.com/go-fynx/envload)

</div>

---

## Overview

**envload** automatically maps environment variables from `.env` files to Go structs using struct tags. It provides type-safe parsing, default values, required field validation, and comprehensive type conversion — all with minimal setup.

```go
type Config struct {
    AppName string        `env:"APP_NAME" default:"MyApp"`
    Port    int           `env:"PORT" default:"8080"`
    Debug   bool          `env:"DEBUG" default:"false"`
    Timeout time.Duration `env:"TIMEOUT" default:"30s"`
}

var cfg Config
envload.LoadAndParse(".env", &cfg)
```

---

## Features

| Feature | Description |
|---------|-------------|
| **Type-Safe Parsing** | Automatic conversion with validation for strings, numbers, booleans, durations, slices, and maps |
| **Struct Tags** | Simple configuration via `env`, `default`, and `required` tags |
| **Graceful Degradation** | Falls back to defaults when `.env` file is missing |
| **Zero Config** | Works out of the box with sensible defaults |
| **Comprehensive Errors** | Descriptive error messages for debugging |

---

## Installation

```bash
go get github.com/go-fynx/envload
```

---

## Quick Start

### 1. Create a `.env` file

```env
APP_NAME=MyAwesomeApp
PORT=8080
DEBUG=true
TIMEOUT=30s
DATABASE_URL=postgres://localhost/mydb
```

### 2. Define your configuration struct

```go
package main

import (
    "log"
    "time"

    "github.com/go-fynx/envload"
)

type Config struct {
    // Application settings
    AppName string        `env:"APP_NAME" default:"DefaultApp"`
    Port    int           `env:"PORT" default:"3000"`
    Debug   bool          `env:"DEBUG" default:"false"`
    Timeout time.Duration `env:"TIMEOUT" default:"10s"`

    // Database (required)
    DatabaseURL string `env:"DATABASE_URL" required:"true"`
}

func main() {
    var cfg Config

    if err := envload.LoadAndParse(".env", &cfg); err != nil {
        log.Fatal("Failed to load config:", err)
    }

    log.Printf("App: %s running on port %d", cfg.AppName, cfg.Port)
}
```

---

## Struct Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `env` | Maps field to environment variable | `env:"PORT"` |
| `default` | Fallback value when env var is missing | `default:"8080"` |
| `required` | Fails if missing and no default | `required:"true"` |

```go
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
```

---

## Supported Types

### Basic Types

| Type | Example Env Value | Go Type |
|------|-------------------|---------|
| String | `APP_NAME=MyApp` | `string` |
| Integer | `PORT=8080` | `int`, `int8`, `int16`, `int32`, `int64` |
| Unsigned | `MAX_CONN=100` | `uint`, `uint8`, `uint16`, `uint32`, `uint64` |
| Float | `RATE=3.14` | `float32`, `float64` |
| Boolean | `DEBUG=true` | `bool` (accepts: `true`, `false`, `1`, `0`) |
| Duration | `TIMEOUT=30s` | `time.Duration` (e.g., `5s`, `2m`, `1h30m`) |

### Collection Types

#### Slices (comma-separated values)

```env
TAGS=web,api,service
PORTS=8080,9090,3000
ENABLED=true,false,true
```

```go
type Config struct {
    Tags    []string  `env:"TAGS" default:"web,api"`
    Ports   []int     `env:"PORTS" default:"8080,9090"`
    Enabled []bool    `env:"ENABLED"`
}
```

> **Note:** Empty values are automatically filtered. `TAGS=web,,api` results in `["web", "api"]`

#### Maps (comma-separated key:value pairs)

```env
LABELS=env:prod,team:backend
FEATURES=cache:true,debug:false
LIMITS=cpu:80,memory:512
```

```go
type Config struct {
    Labels   map[string]string `env:"LABELS"`
    Features map[string]bool   `env:"FEATURES"`
    Limits   map[string]int    `env:"LIMITS"`
}
```

---

## Production Pattern

Use the singleton pattern for application-wide configuration:

```go
package config

import (
    "sync"

    "github.com/go-fynx/envload"
)

type Config struct {
    AppName     string        `env:"APP_NAME" default:"MyApp"`
    Port        int           `env:"PORT" default:"8080"`
    Debug       bool          `env:"DEBUG" default:"false"`
    DatabaseURL string        `env:"DATABASE_URL" required:"true"`
    DBTimeout   time.Duration `env:"DB_TIMEOUT" default:"30s"`
    RedisHosts  []string      `env:"REDIS_HOSTS" default:"localhost:6379"`
    Features    map[string]bool `env:"FEATURES"`
}

var (
    instance Config
    once     sync.Once
    loadErr  error
)

// Load initializes configuration (call once at startup)
func Load(envPath string) error {
    once.Do(func() {
        loadErr = envload.LoadAndParse(envPath, &instance)
    })
    return loadErr
}

// Get returns the loaded configuration
func Get() Config {
    return instance
}
```

**Usage:**

```go
func main() {
    if err := config.Load(".env"); err != nil {
        log.Fatal("Config error:", err)
    }

    cfg := config.Get()
    log.Printf("Starting %s on port %d", cfg.AppName, cfg.Port)
}
```

---

## Error Handling

**envload** provides descriptive errors for common issues:

| Scenario | Error Message |
|----------|---------------|
| Missing required field | `missing required field: field=DatabaseURL env=DATABASE_URL` |
| Invalid type conversion | `invalid int for field 'Port': strconv.ParseInt: parsing "abc": invalid syntax` |
| Invalid target | `target must be a pointer to struct` |
| Invalid duration | `invalid duration for field 'Timeout': time.ParseDuration: invalid duration "xyz"` |

**Graceful degradation:** If the `.env` file doesn't exist, envload logs a warning and continues with default values only.

---

## Limitations

- **Nested structs** are not supported — use flat structures
- **Pointer fields** are not supported — use value types
- **Map keys** must be strings
- **Slice elements** must be basic types (string, int, float, bool)

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

---

<div align="center">

**Made with L❤️VE for the Go community**

</div>
