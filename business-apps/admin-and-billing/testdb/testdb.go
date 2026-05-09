package testdb

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/soumalya/food-delivery-admin/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	postgresContainer *postgres.PostgresContainer
	DbPool            *pgxpool.Pool
)

// Setup creates a new PostgreSQL container and initializes the database schema
func Setup() {
	ctx := context.Background()

	// Spin up PostgreSQL container
	var err error
	postgresContainer, err = postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %s", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	// Connect to database
	DbPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}

	// Inject the test pool into the database package
	database.SetDbConn(DbPool)

	// Create Enum types
	_, err = DbPool.Exec(ctx, `
		CREATE TYPE shift AS ENUM ('lunch', 'dinner');
		CREATE TYPE user_type AS ENUM ('normal', 'admin');
		CREATE TYPE subscription_type AS ENUM ('standard', 'special');
		CREATE TYPE txn_type AS ENUM ('recharge', 'delivery', 'refund');
		CREATE TYPE txn_status AS ENUM ('confirmed', 'pending_acknowledgement');
	`)
	if err != nil {
		log.Fatalf("failed to create enums: %s", err)
	}

	// Create Tables
	_, err = DbPool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS public.users
		(
			user_id serial NOT NULL,
			name text,
			mobile_no text,
			building_no text,
			room_no text,
			role user_type DEFAULT 'normal'::user_type,
			plan subscription_type NOT NULL,
			CONSTRAINT users_pkey PRIMARY KEY (user_id)
		);

		CREATE TABLE IF NOT EXISTS public.daily_logs
		(
			log_id serial NOT NULL,
			user_id integer NOT NULL,
			log_date date NOT NULL DEFAULT CURRENT_DATE,
			meal_type shift NOT NULL,
			is_special boolean NOT NULL DEFAULT false,
			special_dish_name text,
			extra_rice_qty integer NOT NULL DEFAULT 0,
			extra_roti_qty integer NOT NULL DEFAULT 0,
			total_cost numeric(10, 2) NOT NULL,
			created_at timestamp with time zone DEFAULT now(),
			has_main_meal boolean DEFAULT true,
			extra_chicken_qty integer DEFAULT 0,
			extra_fish_qty integer DEFAULT 0,
			extra_egg_qty integer DEFAULT 0,
			extra_vegetable_qty integer DEFAULT 0,
			CONSTRAINT daily_logs_pkey PRIMARY KEY (log_id),
			CONSTRAINT daily_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users (user_id)
		);

		CREATE TABLE IF NOT EXISTS public.expenses
		(
			expense_id serial NOT NULL,
			expense_date date NOT NULL,
			reason text NOT NULL,
			amount numeric(10, 2) NOT NULL,
			created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT expenses_pkey PRIMARY KEY (expense_id)
		);

		CREATE TABLE IF NOT EXISTS public.meal_prices
		(
			item_id character varying(50) NOT NULL,
			item_name character varying(100) NOT NULL,
			price numeric(10, 2) NOT NULL,
			updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT meal_prices_pkey PRIMARY KEY (item_id),
			CONSTRAINT meal_prices_item_name_key UNIQUE (item_name)
		);

		CREATE TABLE IF NOT EXISTS public.wallet_transactions
		(
			txn_id serial NOT NULL,
			user_id integer NOT NULL,
			txn_type txn_type NOT NULL,
			status txn_status NOT NULL DEFAULT 'pending_acknowledgement'::txn_status,
			amount numeric(10, 2) NOT NULL,
			balance_after numeric(10, 2),
			reference_id text,
			created_at timestamp with time zone DEFAULT now(),
			updated_at timestamp with time zone DEFAULT now(),
			CONSTRAINT wallet_transactions_pkey PRIMARY KEY (txn_id),
			CONSTRAINT wallet_transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users (user_id)
		);
	`)
	if err != nil {
		log.Fatalf("failed to create tables: %s", err)
	}

	// Seed default meal prices
	_, err = DbPool.Exec(ctx, `
		INSERT INTO public.meal_prices (item_id, item_name, price) VALUES 
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
		log.Fatalf("failed to insert default meal prices: %s", err)
	}
}

// Teardown shuts down the PostgreSQL container
func Teardown() {
	ctx := context.Background()
	if DbPool != nil {
		DbPool.Close()
	}
	if postgresContainer != nil {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate postgres container: %s", err)
		}
	}
}

// ResetData clears all tables except meal_prices
func ResetData() {
	ctx := context.Background()
	_, err := DbPool.Exec(ctx, `
		TRUNCATE TABLE public.wallet_transactions CASCADE;
		TRUNCATE TABLE public.daily_logs CASCADE;
		TRUNCATE TABLE public.expenses CASCADE;
		TRUNCATE TABLE public.users CASCADE;
	`)
	if err != nil {
		log.Fatalf("failed to reset data: %s", err)
	}
}
