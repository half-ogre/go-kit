package subcmd

import (
	"fmt"
	"strings"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/spf13/cobra"
)

var (
	force bool
)

var dropCmd = &cobra.Command{
	Use:   "drop [database-name]",
	Short: "Drop a database",
	Long:  `Drop (delete) a PostgreSQL database. If no name is provided, uses the database from the connection string. Use --force to skip confirmation.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return withAdminDBConnection(cmd, args, func(db pgkit.DB, dbName string) error {
			return runDrop(db, dbName, force)
		})
	},
}

func init() {
	rootCmd.AddCommand(dropCmd)
	dropCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
}

// runDrop contains the main logic for dropping a database
func runDrop(adminDB pgkit.DB, dbName string, forceFlag bool) error {
	// Confirm deletion unless --force is used
	if !forceFlag {
		fmt.Printf("Are you sure you want to drop database '%s'? (yes/no): ", dbName)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "yes" {
			fmt.Println("Drop cancelled")
			return nil
		}
	}

	// Check if database exists
	var exists bool
	err := adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("database '%s' does not exist", dbName)
	}

	// Drop the database
	_, err = adminDB.Exec(fmt.Sprintf("DROP DATABASE %s", quoteIdentifier(dbName)))
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	fmt.Printf("Database '%s' dropped successfully\n", dbName)
	return nil
}
