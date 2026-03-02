package psql

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func Init() {
	ctx := context.Background()
	dsn := "user=postgres password=821100 dbname=shop_system_test host=localhost port=5432 sslmode=disable"

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)

	if err != nil {
		log.Println("Failed to connect to database: ", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database: ", err)
	}
	log.Println("Connected to database successfully")

	DB = db

	DB.SetMaxOpenConns(5)
	DB.SetMaxIdleConns(2)
	DB.SetConnMaxLifetime(60 * time.Minute)
	DB.SetConnMaxIdleTime(30 * time.Minute)
}

func GetDB() *sqlx.DB {
	return DB
}
