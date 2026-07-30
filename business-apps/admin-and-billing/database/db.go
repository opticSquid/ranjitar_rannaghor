package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
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

func formConnString() string {
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
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbName, sslMode)
}

func InitDB(ctx context.Context) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(formConnString())
	if err != nil {
		return nil, err
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		dataType, err := conn.LoadType(ctx, "MEAL_TYPE")
		if err != nil {
			return fmt.Errorf("failed to load MEAL_TYPE type:  %w", err)
		}
		conn.TypeMap().RegisterType(dataType)
		dataType, err = conn.LoadType(ctx, "MENU_ITEM_CATEGORY")
		if err != nil {
			return fmt.Errorf("failed to load MENU_ITEM_CATEGORY type:  %w", err)
		}
		conn.TypeMap().RegisterType(dataType)
		dataType, err = conn.LoadType(ctx, "USER_ROLE")
		if err != nil {
			return fmt.Errorf("failed to load USER_ROLE type:  %w", err)
		}
		conn.TypeMap().RegisterType(dataType)
		dataType, err = conn.LoadType(ctx, "SUBSCRIPTION_PLAN")
		if err != nil {
			return fmt.Errorf("failed to load SUBSCRIPTION_PLAN type:  %w", err)
		}
		conn.TypeMap().RegisterType(dataType)
		return nil
	}
	dbPool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("Unable to connect to database", "err", err)
		os.Exit(1)
	}

	err = dbPool.Ping(ctx)
	if err != nil {
		slog.Error("Unable to ping database", "err", err)
		os.Exit(1)
	}
	slog.Info("Connected to database successfully")
	return dbPool, nil
}
