package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Database struct {
	Connection  *sql.DB
	RateLimiter chan int
}

// Start initializes the database connection
func Start() (*Database, error) {
	// Get database configuration from environment variables
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}

	// Check if we're in debug mode
	isDebug := os.Getenv("IS_DEBUG")
	var dbname string
	if isDebug == "true" {
		dbname = os.Getenv("DB_NAME_DEV")
		if dbname == "" {
			dbname = "ai_agent_check_dev"
		}
	} else {
		dbname = os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "ai_agent_check"
		}
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	if isDebug == "true" {
		log.Printf("Connecting to DEBUG database: %s on %s:%s", dbname, host, port)
	} else {
		log.Printf("Connecting to PRODUCTION database: %s on %s:%s", dbname, host, port)
	}

	dbString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	connection, err := sql.Open("postgres", dbString)
	if err != nil {
		return &Database{}, fmt.Errorf("failed to open database connection: %v", err)
	}

	// Test the connection
	if err := connection.Ping(); err != nil {
		return &Database{}, fmt.Errorf("failed to ping database: %v", err)
	}

	// Set connection pool settings
	connection.SetMaxOpenConns(25)
	connection.SetMaxIdleConns(5)

	rateLimiter := make(chan int, 50)

	log.Println("Database connection established successfully")
	return &Database{
		Connection:  connection,
		RateLimiter: rateLimiter,
	}, nil
}

// CreateUsersTable creates the users table if it doesn't exist
func (d *Database) CreateUsersTable() error {
	query := `CREATE TABLE IF NOT EXISTS public.users (
		user_id TEXT PRIMARY KEY,
		first_name TEXT DEFAULT '',
		last_name TEXT DEFAULT '',
		full_name TEXT DEFAULT '',
		email TEXT DEFAULT '' UNIQUE,
		subscription_plan TEXT DEFAULT 'SPARK',
		account_validated BOOLEAN DEFAULT false,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		last_login TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		company_id TEXT DEFAULT '',
		company_name TEXT DEFAULT '',
		credits INTEGER DEFAULT 20,
		is_active BOOLEAN DEFAULT true
	)`

	_, err := d.Connection.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	log.Println("Users table created/verified successfully")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.Connection != nil {
		return d.Connection.Close()
	}
	return nil
}
