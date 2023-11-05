package infrastructure

import (
	"log"
	"os"

	"invisible-tir-go/cmd/user"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDb() (db *gorm.DB, err error) {
	dsn := "host=localhost user=invisibleprogrammer password=invisiblepassword dbname=tir-db port=5432 sslmode=disable TimeZone=Europe/Budapest"

	// Initialize a new GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,        // Disable color
		},
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalln("Failed to connect to database")
		panic("Failed to connect to database")
	}
	log.Println("Database connected")
	db.AutoMigrate(&user.User{})
	log.Println("Running migration scripts")
	db.Exec("CREATE UNIQUE INDEX idx_email ON users(email) WHERE deleted_at IS NULL;")
	log.Println("Database Migrated")

	return db, err
}
