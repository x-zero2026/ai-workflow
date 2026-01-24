package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
)

// Connect creates a database connection to Supabase PostgreSQL
func Connect(supabaseURL, password string) (*sql.DB, error) {
	// Extract project reference from Supabase URL
	// Format: https://xxx.supabase.co -> xxx
	projectRef := extractProjectRef(supabaseURL)
	
	if projectRef == "" {
		return nil, fmt.Errorf("invalid SUPABASE_URL format: %s", supabaseURL)
	}

	// URL encode the password to handle special characters
	encodedPassword := url.QueryEscape(password)

	// Build connection string using Supabase Pooler (port 6543)
	// IMPORTANT: prefer_simple_protocol=yes disables prepared statements
	// This is required for Supabase Pooler in Transaction Mode
	// Format: postgresql://postgres.PROJECT_REF:PASSWORD@aws-1-ap-south-1.pooler.supabase.com:6543/postgres
	connStr := fmt.Sprintf(
		"postgresql://postgres.%s:%s@aws-1-ap-south-1.pooler.supabase.com:6543/postgres?sslmode=require&prefer_simple_protocol=yes",
		projectRef,
		encodedPassword,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings for Lambda
	// Keep these low to avoid exhausting connections
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return db, nil
}

// extractProjectRef extracts the project reference from Supabase URL
func extractProjectRef(url string) string {
	// Remove https:// or http://
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Extract project ref (before .supabase.co)
	parts := strings.Split(url, ".supabase.co")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}

	return ""
}

// CheckProjectAccess checks if a user has access to a project
func CheckProjectAccess(db *sql.DB, userDID, projectID string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_projects
			WHERE user_did = $1 AND project_id = $2
		)
	`
	err := db.QueryRow(query, userDID, projectID).Scan(&exists)
	return exists, err
}

// CheckProjectAdmin checks if a user is an admin of a project
func CheckProjectAdmin(db *sql.DB, userDID, projectID string) (bool, error) {
	var isAdmin bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_projects
			WHERE user_did = $1 AND project_id = $2 AND role = 'admin'
		)
	`
	err := db.QueryRow(query, userDID, projectID).Scan(&isAdmin)
	return isAdmin, err
}
