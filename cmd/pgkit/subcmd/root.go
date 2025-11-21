package subcmd

import (
	"context"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
)

var (
	dbURL string
)

var rootCmd = &cobra.Command{
	Use:   "pgkit",
	Short: "PostgreSQL toolkit",
	Long:  `A toolkit for managing PostgreSQL databases including migrations, creation, and deletion.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip DB URL requirement for commands that don't need a database
		if cmd.Name() == "version" || cmd.Name() == "list" {
			return nil
		}

		// Get DB URL from flag or environment
		url, err := getDBURL()
		if err != nil {
			return err
		}

		// Store in context for subcommands
		ctx := context.WithValue(cmd.Context(), dbURLKey, url)
		cmd.SetContext(ctx)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbURL, "db", "", "Database connection string (or use DATABASE_URL env var)")
}

// getDBURL returns the database URL from flag or environment variable
func getDBURL() (string, error) {
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		return "", fmt.Errorf("database URL not provided (use --db flag or DATABASE_URL environment variable)")
	}
	return dbURL, nil
}
