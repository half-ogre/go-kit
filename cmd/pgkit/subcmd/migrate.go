package subcmd

import (
	"fmt"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/spf13/cobra"
)

var (
	migrationsDir string
	toVersion     int
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run SQL migration files from a directory against the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDBConnection(cmd, func(db pgkit.DB) error {
			return runMigrate(db, migrationsDir, toVersion, pgkit.NewMigrator())
		})
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&migrationsDir, "dir", "d", "migrations", "Directory containing migration files")
	migrateCmd.Flags().IntVar(&toVersion, "to-version", 0, "Migrate up to and including this version number (e.g., 2)")
}

// runMigrate contains the main logic for running database migrations
func runMigrate(db pgkit.DB, dir string, toVersion int, migrator pgkit.Migrator) error {
	if toVersion > 0 {
		fmt.Printf("Running migrations from %s up to version %d...\n", dir, toVersion)
		if err := migrator.RunMigrationsToVersion(db, dir, toVersion); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	} else {
		fmt.Printf("Running migrations from %s...\n", dir)
		if err := migrator.RunMigrations(db, dir); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	fmt.Println("All migrations completed successfully")
	return nil
}
