package psql

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file loaded: %v", err)
	}

	ctx := context.Background()

	databaseURL := getEnv("DATABASE_URL", "")

	host := getEnv("POSTGRES_HOST", "")
	port := getEnv("POSTGRES_PORT", "")
	user := getEnv("POSTGRES_USER", "")
	password := getEnv("POSTGRES_PASSWORD", "")
	dbname := getEnv("POSTGRES_DB", "")
	sslmode := getEnv("POSTGRES_SSLMODE", "")

	var dsn string
	if databaseURL != "" {
		dsn = databaseURL
	} else {
		dsn = fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, sslmode)
	}

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Printf("Failed to ping database: %v", err)
	}
	log.Printf("Connected to database successfully")

	DB = db

	DB.SetMaxOpenConns(5)
	DB.SetMaxIdleConns(2)
	DB.SetConnMaxLifetime(60 * time.Minute)
	DB.SetConnMaxIdleTime(30 * time.Minute)
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
func GetDB() *sqlx.DB {
	return DB
}
