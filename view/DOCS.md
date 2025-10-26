# View Package Documentation

## Overview

The View package provides a simple and efficient view management system for the Goravel framework. It offers functionality for sharing data across views, checking view existence, and generating view files through console commands.

## Features

- **View Existence Checking**: Verify if view files exist in the resources directory
- **Data Sharing**: Share data across views using key-value pairs
- **Console Commands**: Generate view files with customizable templates
- **Thread-Safe**: Uses sync.Map for concurrent access safety
- **Configurable**: Supports custom view paths, extensions, and templates

## Installation

The View package is included by default in the Goravel framework. To use it in your application, ensure the service provider is registered in your `config/app.go`:

```go
import "github.com/goravel/framework/view"

// In your providers array
&view.ServiceProvider{}
```

## Basic Usage

### Using the Facade

```go
import "github.com/goravel/framework/facades"

// Check if a view exists
exists := facades.View().Exists("welcome.tmpl")

// Share data with views
facades.View().Share("title", "Welcome to Goravel")
facades.View().Share("user", map[string]string{
    "name": "John Doe",
    "email": "john@example.com",
})

// Retrieve shared data
title := facades.View().Shared("title", "Default Title")
user := facades.View().Shared("user")

// Get all shared data
allShared := facades.View().GetShared()
```

### Direct Usage

```go
import "github.com/goravel/framework/view"

// Create a new view instance
viewInstance := view.NewView()

// Check if view exists
exists := viewInstance.Exists("dashboard.tmpl")

// Share data
viewInstance.Share("app_name", "My Application")

// Retrieve shared data
appName := viewInstance.Shared("app_name", "Default App")
```

## Console Commands

### Creating Views

Generate new view files using the `make:view` command:

```bash
# Create a basic view
go run . artisan make:view welcome

# Create a view with custom path
go run . artisan make:view dashboard --path=resources/views/admin

# Force create (overwrite existing)
go run . artisan make:view profile --force

# Create with specific extension
go run . artisan make:view home
```

### Command Options

- `--path`: Specify custom view directory (default: `resources/views`)
- `--force` or `-f`: Overwrite existing files

## Configuration

The view package can be configured through the `config/http.go` file:

```go
// config/http.go
package config

import (
	"github.com/gin-gonic/gin/render"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	"github.com/goravel/gin"
	ginfacades "github.com/goravel/gin/facades"
)

func init() {
	config := facades.Config()
	config.Add("http", map[string]any{
		// HTTP Driver
		"default": "gin",
		// HTTP Drivers
		"drivers": map[string]any{
			"gin": map[string]any{
				// Optional, default is 4096 KB
				"body_limit":   4096,
				"header_limit": 4096,
				"route": func() (route.Route, error) {
					return ginfacades.Route("gin"), nil
				},
				// Optional, default is http/template
				"template": func() (render.HTMLRender, error) {
					return gin.DefaultTemplate()
				},
			},
		},
		// HTTP URL
		"url": config.Env("APP_URL", "http://localhost"),
		// HTTP Host
		"host": config.Env("APP_HOST", "127.0.0.1"),
		// HTTP Port
		"port": config.Env("APP_PORT", "3000"),
		// HTTP Timeout, default is 3 seconds
		"request_timeout": 3,
		// HTTPS Configuration
		"tls": map[string]any{
			// HTTPS Host
			"host": config.Env("APP_HOST", "127.0.0.1"),
			// HTTPS Port
			"port": config.Env("APP_PORT", "3000"),
			// SSL Certificate, you can put the certificate in /public folder
			"ssl": map[string]any{
				// ca.pem
				"cert": "",
				// ca.key
				"key": "",
			},
		},
		// HTTP Client Configuration
		"client": map[string]any{
			"base_url":                config.GetString("HTTP_CLIENT_BASE_URL"),
			"timeout":                 config.GetDuration("HTTP_CLIENT_TIMEOUT"),
			"max_idle_conns":          config.GetInt("HTTP_CLIENT_MAX_IDLE_CONNS"),
			"max_idle_conns_per_host": config.GetInt("HTTP_CLIENT_MAX_IDLE_CONNS_PER_HOST"),
			"max_conns_per_host":      config.GetInt("HTTP_CLIENT_MAX_CONN_PER_HOST"),
			"idle_conn_timeout":       config.GetDuration("HTTP_CLIENT_IDLE_CONN_TIMEOUT"),
		},
        // View Custom
        "view": map[string]any{
            "path": "resources/views", // Custom view directory path
            "extension": ".tmpl", // View file extension (default: .tmpl)
            "stub": "..." // Custom template stub [DummyPathName, DummyPathDefinition, DummyViewName]
        }
	})
}
```

### Configuration Options

- **Path**: Custom directory for view files (default: `resources/views`)
- **Extension**: File extension for views (default: `.tmpl`)
- **Stub**: Custom template for generated views

## View File Structure

By default, views are stored in the `resources/views` directory with `.tmpl` extension. The generated view files use Go's template syntax:

```html
<!-- resources/views/welcome.tmpl -->
{{ define "welcome" }}
<h1>Welcome to welcome.tmpl</h1>
{{ end }}
```

## API Reference

### View Interface

```go
type View interface {
    // Exists checks if a view with the specified name exists
    Exists(view string) bool
    
    // Share associates a key-value pair with the current view context
    Share(key string, value any)
    
    // Shared retrieves the value associated with the given key
    // Returns default value if key doesn't exist
    Shared(key string, def ...any) any
    
    // GetShared returns all shared data as a map
    GetShared() map[string]any
}
```

### Methods

#### `Exists(view string) bool`
Checks if a view file exists in the resources/views directory.

**Parameters:**
- `view`: The view file name (e.g., "welcome.tmpl")

**Returns:**
- `bool`: True if the view exists, false otherwise

#### `Share(key string, value any)`
Stores a key-value pair that can be accessed by views.

**Parameters:**
- `key`: The key to store the value under
- `value`: The value to store (can be any type)

#### `Shared(key string, def ...any) any`
Retrieves a value from the shared data.

**Parameters:**
- `key`: The key to retrieve
- `def`: Optional default value if key doesn't exist

**Returns:**
- `any`: The stored value or default value

#### `GetShared() map[string]any`
Returns all shared data as a map.

**Returns:**
- `map[string]any`: All shared key-value pairs

## Examples

### Basic View Management

```go
package main

import (
    "github.com/goravel/framework/facades"
)

func main() {
    // Check if view exists
    if facades.View().Exists("welcome.tmpl") {
        // Share data with the view
        facades.View().Share("title", "Welcome Page")
        facades.View().Share("users", []string{"Alice", "Bob", "Charlie"})
        
        // Retrieve shared data
        title := facades.View().Shared("title", "Default Title")
        users := facades.View().Shared("users", []string{})
        
        // Get all shared data
        allData := facades.View().GetShared()
    }
}
```