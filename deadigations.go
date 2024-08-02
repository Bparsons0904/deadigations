package deadigations

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MigrationTool struct {
	db                   *gorm.DB
	registeredMigrations []*gormigrate.Migration
	options              *gormigrate.Options
}

func NewMigrationTool(dsn string, registeredMigrations []*gormigrate.Migration, args []string) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	migrationTool := &MigrationTool{
		db:                   db,
		registeredMigrations: registeredMigrations,
		options: &gormigrate.Options{
			TableName:                 "migrations",
			IDColumnName:              "id",
			IDColumnSize:              255,
			UseTransaction:            true,
			ValidateUnknownMigrations: false,
		},
	}

	if len(args) > 1 {
		switch args[1] {
		case "-up":
			if err := migrationTool.MigrateUp(); err != nil {
				log.Fatalf("Migration failed: %v", err)
			}
		case "-down":
			if err := migrationTool.MigrateDown(); err != nil {
				log.Fatalf("Rollback failed: %v", err)
			}
		case "-create":
			if len(args) < 3 {
				log.Fatal("Please provide a name for the migration")
			}
			migrationName := args[2]
			if err := migrationTool.CreateMigrationFile(migrationName); err != nil {
				log.Fatalf("Failed to create migration file: %v", err)
			}
		default:
			log.Fatal("Invalid command. Use -up, -down, or -create")
		}
	} else {
		log.Println("No command provided. Use -up, -down, or -create")
	}
}

func (m *MigrationTool) MigrateUp() error {
	if len(m.registeredMigrations) == 0 {
		log.Println("No migrations registered")
		return nil
	}

	sort.SliceStable(m.registeredMigrations, func(i, j int) bool {
		return m.registeredMigrations[i].ID < m.registeredMigrations[j].ID
	})

	migrator := gormigrate.New(m.db, m.options, m.registeredMigrations)

	if err := migrator.Migrate(); err != nil {
		return err
	}
	log.Println("Migrations applied successfully!")
	return nil
}

func (m *MigrationTool) MigrateDown() error {
	if len(m.registeredMigrations) == 0 {
		log.Println("No migrations registered")
		return nil
	}

	sort.SliceStable(m.registeredMigrations, func(i, j int) bool {
		return m.registeredMigrations[i].ID > m.registeredMigrations[j].ID
	})

	migrator := gormigrate.New(m.db, m.options, m.registeredMigrations)

	if err := migrator.RollbackLast(); err != nil {
		return err
	}
	log.Println("Last migration rolled back successfully!")
	return nil
}

func (m *MigrationTool) CreateMigrationFile(name string) error {
	timestamp := time.Now().Format("20060102150405") // YYYYMMDDHHMMSS
	filename := fmt.Sprintf("%s_%s.go", timestamp, strings.Replace(name, " ", "_", -1))
	filePath := fmt.Sprintf("./migrations/%s", filename)

	if err := os.MkdirAll("./migrations", os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	migrationTemplate := `package migrations

import (
	"gorm.io/gorm"
)

func init() {
	RegisterMigration(Migration{
		ID:          "%s",
		Description: "Add description of changes",
		Migrate: func(tx *gorm.DB) error {
			// Your migration logic goes here.
			return nil // Replace with actual code
		},
		Rollback: func(tx *gorm.DB) error {
			// Your rollback logic goes here.
			return nil // Replace with actual code
		},
	})
}`

	migrationContent := fmt.Sprintf(migrationTemplate, timestamp)
	_, err = file.WriteString(migrationContent)
	if err != nil {
		return err
	}

	log.Printf("Migration file created: %s", filePath)
	return nil
}
