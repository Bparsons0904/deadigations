package deadigations

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Migration struct {
	ID          string
	Description string
	Migrate     func(tx *gorm.DB) error
	Rollback    func(tx *gorm.DB) error
}

var (
	once                 sync.Once
	instance             *MigrationTool
	registeredMigrations []*gormigrate.Migration
)

func RegisterMigration(migration Migration) {
	gormMigration := &gormigrate.Migration{
		ID:       migration.ID,
		Migrate:  migration.Migrate,
		Rollback: migration.Rollback,
	}
	registeredMigrations = append(registeredMigrations, gormMigration)
}

type MigrationTool struct {
	db      *gorm.DB
	options *gormigrate.Options
}

func NewMigrationTool(dsn string) *MigrationTool {
	once.Do(func() {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to the database: %v", err)
		}

		instance = &MigrationTool{
			db: db,
			options: &gormigrate.Options{
				TableName:                 "migrations",
				IDColumnName:              "id",
				IDColumnSize:              255,
				UseTransaction:            true,
				ValidateUnknownMigrations: false,
			},
		}
	})

	return instance
}

func (m *MigrationTool) Run(args []string) {
	if len(args) > 1 {
		switch args[1] {
		case "-up":
			if err := m.MigrateUp(); err != nil {
				log.Fatalf("Migration failed: %v", err)
			}
		case "-down":
			if err := m.MigrateDown(); err != nil {
				log.Fatalf("Rollback failed: %v", err)
			}
		case "-create":
			if len(args) < 3 {
				log.Fatal("Please provide a name for the migration")
			}
			migrationName := args[2]
			if err := m.CreateMigrationFile(migrationName); err != nil {
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
	if len(registeredMigrations) == 0 {
		log.Println("No migrations registered")
		return nil
	}

	sort.SliceStable(registeredMigrations, func(i, j int) bool {
		return registeredMigrations[i].ID < registeredMigrations[j].ID
	})

	migrator := gormigrate.New(m.db, m.options, registeredMigrations)

	if err := migrator.Migrate(); err != nil {
		return err
	}
	log.Println("Migrations applied successfully!")
	return nil
}

func (m *MigrationTool) MigrateDown() error {
	if len(registeredMigrations) == 0 {
		log.Println("No migrations registered")
		return nil
	}

	sort.SliceStable(registeredMigrations, func(i, j int) bool {
		return registeredMigrations[i].ID > registeredMigrations[j].ID
	})

	migrator := gormigrate.New(m.db, m.options, registeredMigrations)

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
	"github.com/Bparsons0904/deadigations"
	"gorm.io/gorm"
)

func init() {
	deadigations.RegisterMigration(deadigations.Migration{
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
