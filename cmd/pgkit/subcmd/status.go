package subcmd

import (
	"fmt"
	"time"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Show migration status - which migrations are available, which are applied, and when they were applied.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDBConnection(cmd, func(db pgkit.DB) error {
			return runStatus(db, migrationsDir, pgkit.NewMigrator())
		})
	},
}

// runStatus contains the main logic for showing migration status
func runStatus(db pgkit.DB, dir string, migrator pgkit.Migrator) error {
	migrations, err := migrator.ListMigrations(db, dir)
	if err != nil {
		return fmt.Errorf("failed to list migrations: %w", err)
	}

	if len(migrations) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	fmt.Printf("Migration status:\n\n")

	appliedCount := 0
	for _, m := range migrations {
		if m.Applied {
			appliedCount++
			fmt.Printf("✓ Version %d: %s (%s) - applied at %s\n",
				m.Version, m.Description, m.Filename, m.AppliedAt.Format(time.RFC3339))
		} else {
			fmt.Printf("  Version %d: %s (%s) - not applied\n",
				m.Version, m.Description, m.Filename)
		}
	}

	fmt.Printf("\n%d of %d migrations applied\n", appliedCount, len(migrations))
	return nil
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().StringVarP(&migrationsDir, "dir", "d", "migrations", "Directory containing migration files")
}
