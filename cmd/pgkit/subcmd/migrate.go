package subcmd

import (
	"fmt"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/spf13/cobra"
)

var (
	migrationsDir string
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run SQL migration files from a directory against the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDBConnection(cmd, func(db pgkit.DB) error {
			return runMigrate(db, migrationsDir, pgkit.NewMigrator())
		})
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&migrationsDir, "dir", "d", "migrations", "Directory containing migration files")
}

// runMigrate contains the main logic for running database migrations
func runMigrate(db pgkit.DB, dir string, migrator pgkit.Migrator) error {
	fmt.Printf("Running migrations from %s...\n", dir)
	if err := migrator.RunMigrations(db, dir); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("All migrations completed successfully")
	return nil
}
