package subcmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/spf13/cobra"
)

type contextKey string

const (
	dbURLKey contextKey = "dbURL"
)

// getDBURLFromContext retrieves the database URL from the command context
func getDBURLFromContext(cmd *cobra.Command) (string, error) {
	url, ok := cmd.Context().Value(dbURLKey).(string)
	if !ok {
		return "", fmt.Errorf("database URL not found in context")
	}
	return url, nil
}

// withDBConnection handles the connection lifecycle and executes the provided function
// It gets the URL from the command context and connects to the database
func withDBConnection(cmd *cobra.Command, fn func(pgkit.DB) error) error {
	url, err := getDBURLFromContext(cmd)
	if err != nil {
		return err
	}

	db, err := pgkit.NewDB(url)
	if err != nil {
		return err
	}
	defer db.Close()
	return fn(db)
}

// withAdminDBConnection handles connection to the 'postgres' admin database for drop/create operations
// It parses the target database name from args and passes both to the callback
func withAdminDBConnection(cmd *cobra.Command, args []string, fn func(pgkit.DB, string) error) error {
	dbURLStr, err := getDBURLFromContext(cmd)
	if err != nil {
		return err
	}

	dbName, adminURL, err := parseDatabaseParams(args, dbURLStr)
	if err != nil {
		return err
	}

	db, err := pgkit.NewDB(adminURL)
	if err != nil {
		return err
	}
	defer db.Close()
	return fn(db, dbName)
}

// quoteIdentifier properly quotes a PostgreSQL identifier
func quoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

// parseDatabaseParams extracts the database name and admin URL for database management commands
// It returns the database name (from args or URL) and the admin URL (connecting to 'postgres' database)
func parseDatabaseParams(args []string, dbURLStr string) (dbName string, adminURL string, err error) {
	// Parse the connection string
	parsedURL, err := url.Parse(dbURLStr)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Get database name from args or URL path
	if len(args) > 0 {
		dbName = args[0]
	} else {
		// Extract database name from path
		dbName = strings.TrimPrefix(parsedURL.Path, "/")
		if dbName == "" {
			return "", "", fmt.Errorf("no database name provided and none found in connection string")
		}
	}

	// Create admin URL (connect to 'postgres' database)
	parsedURL.Path = "/postgres"
	adminURL = parsedURL.String()

	return dbName, adminURL, nil
}
