package subcmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Show all applied migrations and when they were applied.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return withDBConnection(cmd, func(db pgkit.DB) error {
			return runStatus(db)
		})
	},
}

// runStatus contains the main logic for showing migration status
func runStatus(db pgkit.DB) error {
	rows, err := db.Query("SELECT filename, applied_at FROM pgkit_migrations ORDER BY applied_at")
	if err != nil {
		// If the migrations table doesn't exist yet, just show no migrations
		if strings.Contains(err.Error(), "relation \"pgkit_migrations\" does not exist") {
			fmt.Println("Applied migrations:")
			fmt.Println("  (none)")
			return nil
		}
		return fmt.Errorf("failed to query pgkit_migrations: %w", err)
	}
	defer rows.Close()

	fmt.Println("Applied migrations:")
	hasRows := false
	for rows.Next() {
		hasRows = true
		var filename string
		var appliedAt time.Time
		if err := rows.Scan(&filename, &appliedAt); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		fmt.Printf("  %s (applied at %s)\n", filename, appliedAt.Format(time.RFC3339))
	}

	if !hasRows {
		fmt.Println("  (none)")
	}

	return rows.Err()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
