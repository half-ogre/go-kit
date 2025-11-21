package subcmd

import (
	"fmt"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available migrations",
	Long:  `List all migration files in the specified directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(migrationsDir)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&migrationsDir, "dir", "d", "migrations", "Directory containing migration files")
}

// runList contains the main logic for listing migrations from the directory
func runList(dir string) error {
	migrations, err := pgkit.ListMigrationsFromDir(dir)
	if err != nil {
		return fmt.Errorf("failed to list migrations: %w", err)
	}

	if len(migrations) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	fmt.Printf("Found %d migration(s):\n\n", len(migrations))
	for _, m := range migrations {
		fmt.Printf("Version %d: %s (%s)\n", m.Version, m.Description, m.Filename)
	}

	return nil
}
