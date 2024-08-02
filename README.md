# Deadigations Tool

This guide provides instructions on setting up and using the deadigations tool as a wrapper for the Gormigrate package. This tool simplifies managing database schema changes in your Go applications.

## Table of Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
- [Creating Migrations](#creating-migrations)
- [Running Migrations](#running-migrations)
- [Best Practices](#best-practices)
- [References](#references)

## Overview

The Deadigation Tool allows you to manage database schema changes in a Go application efficiently. It leverages Gormigrate for handling migrations and provides a command-line interface for creating, applying, and rolling back migrations.

## Getting Started

To start using Deadigation, follow these steps to integrate it into your Go project.

### Step 1: Setup Your Project

#### Import the Necessary Packages

First, ensure that your project imports the necessary packages:

```go
import (
    "fmt"
    "log"
    "os"

    "github.com/Bparsons0904/deadigations"
    "your_project/config"
    _ "your_project/migrator/migrations" // Import the migrations to register them
)
```

#### Initialize the Migration Tool

Use the following code snippet in your main application file (e.g., `migrator.go`) to initialize and run the migration tool:

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/Bparsons0904/deadigations"
    "your_project/config"
    _ "your_project/migrator/migrations" // Import the migrations to register them
)

func main() {
    // Load environment configuration
    config, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Could not load environment variables:", err)
    }

    // Define the Data Source Name (DSN) for PostgreSQL connection
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)

    // Initialize the migration tool
    migrationTool := deadigations.NewMigrationTool(dsn)

    // Run the migration tool with command-line arguments
    migrationTool.Run(os.Args)
}
```

#### Environment Configuration

Ensure your config package loads the necessary environment variables for database connection:

```go
package config

import (
    "github.com/joho/godotenv"
    "log"
    "os"
)

type Config struct {
    DBHost     string
    DBUser     string
    DBPassword string
    DBName     string
    DBPort     string
}

func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    config := &Config{
        DBHost:     os.Getenv("DB_HOST"),
        DBUser:     os.Getenv("DB_USER"),
        DBPassword: os.Getenv("DB_PASSWORD"),
        DBName:     os.Getenv("DB_NAME"),
        DBPort:     os.Getenv("DB_PORT"),
    }

    return config, nil
}
```

### Step 2: Prepare Migration Files

#### Create a Migrations Directory

Ensure your migrations are stored in a dedicated directory:

```bash
mkdir -p migrator/migrations
```

#### Migration File Template

When you create a new migration, the tool generates a file like this:

```go
package migrations

import (
    "github.com/Bparsons0904/deadigations"
    "gorm.io/gorm"
)

func init() {
    deadigations.RegisterMigration(deadigations.Migration{
        ID: "20240803120000",
        Description: "Add users table",
        Migrate: func(tx *gorm.DB) error {
            type User struct {
                ID    uint   `gorm:"primaryKey"`
                Name  string `gorm:"unique;not null"`
                Email string `gorm:"unique;not null"`
            }
            return tx.AutoMigrate(&User{})
        },
        Rollback: func(tx *gorm.DB) error {
            return tx.Migrator().DropTable("users")
        },
    })
}
```

## Creating Migrations

To create a new migration file, use the command line tool as follows:

#### Create a Basic Migration

```bash
go run ./migrator/migrator.go -create <migration_description>
```

This will generate a new migration file with the timestamped ID and description.

#### Create a Transaction Migration

For more complex migrations involving transactions, use:

```bash
go run ./migrator/migrator.go -create-tx <migration_description>
```

This template supports multiple operations within a transaction.

## Running Migrations

With your migrations set up, execute them using the command-line interface:

#### Apply Migrations

To apply all pending migrations, run:

```bash
go run ./migrator/migrator.go -up
```

#### Rollback Migrations

To rollback the last applied migration, use:

```bash
go run ./migrator/migrator.go -down
```

## Best Practices

- **Unique IDs:** Ensure each migration has a unique timestamp-based ID (YYYYMMDDHHMMSS).
- **Local Testing:** Test migrations locally before applying them to production environments.
- **Schema Reflection:** Keep your models updated to reflect schema changes.
- **Careful Rollbacks:** Implement rollback logic to safely reverse changes.

## References

- [Gormigrate](https://github.com/go-gormigrate/gormigrate)
- [Gorm Migrations Documentation](https://gorm.io/docs/migration.html)

```

```
