package subcmd

import (
	"fmt"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [database-name]",
	Short: "Create a new database",
	Long:  `Create a new PostgreSQL database. If no name is provided, uses the database from the connection string.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return withAdminDBConnection(cmd, args, func(db pgkit.DB, dbName string) error {
			return runCreate(db, dbName)
		})
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

// runCreate contains the main logic for creating a database
func runCreate(adminDB pgkit.DB, dbName string) error {
	// Check if database already exists
	var exists bool
	err := adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if exists {
		return fmt.Errorf("database '%s' already exists", dbName)
	}

	// Create the database
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", quoteIdentifier(dbName)))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	fmt.Printf("Database '%s' created successfully\n", dbName)
	return nil
}
