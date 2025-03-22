# Configuration Files

This directory contains configuration files for the application.

## Usage

1. Copy `../app.toml.example` to `app.toml` in this directory
2. Modify the settings as needed for your environment

## Environment-specific Configurations

You can create multiple configuration files for different environments:

- `app.toml` - Default configuration
- `app.dev.toml` - Development configuration
- `app.test.toml` - Testing configuration
- `app.prod.toml` - Production configuration

To use a specific configuration file, use the `--config` flag:

```
go-template daemon --config configs/app.prod.toml
```

## Note

All configuration files in this directory are ignored by git to avoid
committing sensitive information and to allow for environment-specific settings.
