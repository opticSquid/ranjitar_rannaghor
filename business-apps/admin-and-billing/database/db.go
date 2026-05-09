package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func GetDbConn() *pgxpool.Pool {
	return dbPool
}

// SetDbConn allows tests to inject a test database pool.
func SetDbConn(pool *pgxpool.Pool) {
	dbPool = pool
}

func InitDB() *pgxpool.Pool {
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		user = "postgres" // fallback for local dev if not set
	}
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		password = "Password" // fallback for local dev if not set
	}
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost" // fallback for local dev if not set
	}
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432" // fallback for local dev if not set
	}
	dbName := os.Getenv("POSTGRES_DB")
	if dbName == "" {
		dbName = "postgres" // fallback for local dev if not set
	}
	sslMode := os.Getenv("POSTGRES_SSLMODE")
	if sslMode == "" {
		sslMode = "disable" // fallback for local dev if not set
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbName, sslMode)

	var err error
	dbPool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	err = dbPool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}
	// Create tables if they don't exist
	_, err = dbPool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS EXPENSES (
            EXPENSE_ID SERIAL PRIMARY KEY,
            EXPENSE_DATE DATE NOT NULL,
            REASON TEXT NOT NULL,
            AMOUNT DECIMAL(10,2) NOT NULL,
            CREATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
		CREATE TABLE IF NOT EXISTS MEAL_PRICES (
            ITEM_ID VARCHAR(50) PRIMARY KEY,
            ITEM_NAME VARCHAR(100)  UNIQUE NOT NULL,
            PRICE DECIMAL(10,2) NOT NULL,
            UPDATED_AT TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
	`)
	if err != nil {
		log.Fatalf("Unable to create tables: %v\n", err)
	}

	// Initialize default meal prices if the table is empty
	var count int
	err = dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM MEAL_PRICES").Scan(&count)
	if err == nil && count == 0 {
		_, err = dbPool.Exec(context.Background(), `
			INSERT INTO MEAL_PRICES (ITEM_ID, ITEM_NAME, PRICE) VALUES 
			('standard', 'Standard Meal', 52.5),
			('special', 'Special Meal', 120.0),
			('rice', 'Extra Rice', 10.0),
			('roti', 'Extra Roti', 4.0),
			('chicken', 'Extra Chicken', 30.0),
			('fish', 'Extra Fish', 20.0),
			('egg', 'Extra Egg', 10.0),
			('vegetable', 'Extra Vegetable', 15.0);
		`)
		if err != nil {
			log.Printf("Unable to insert default meal prices: %v\n", err)
		}
	}
	if err != nil {
		log.Fatalf("Unable to create tables: %v\n", err)
	}

	log.Println("Connected to database successfully")
	return dbPool
}
