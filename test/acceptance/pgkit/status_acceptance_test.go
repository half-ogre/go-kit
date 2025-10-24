//go:build acceptance

package pgkit_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/half-ogre/go-kit/pgkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping acceptance test - DATABASE_URL not set")
	}

	// Build the pgkit binary
	buildPgkit(t)
	defer cleanupPgkit(t)

	dbURL := os.Getenv("DATABASE_URL")

	t.Run("displays_no_migrations_when_none_applied", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()

		cmd := exec.Command("./pgkit", "status", "--db", dbURL)
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "status command should succeed: %s", string(output))
		assert.Contains(t, string(output), "Applied migrations:")
		assert.Contains(t, string(output), "(none)")
	})

	t.Run("displays_all_migrations_after_running_migrations", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()
		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "testdata")
		require.NoError(t, err)

		cmd := exec.Command("./pgkit", "status", "--db", dbURL)
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "status command should succeed: %s", string(output))
		assert.Contains(t, string(output), "Applied migrations:")
		assert.Contains(t, string(output), "001_create_users.sql")
		assert.Contains(t, string(output), "002_add_email_to_users.sql")
		assert.Contains(t, string(output), "003_add_index_on_email.sql")
		assert.Contains(t, string(output), "004_add_status_column.sql")
	})

	t.Run("displays_migrations_in_order", func(t *testing.T) {
		db := setupTestDB(t, dbURL)
		defer db.Close()
		migrator := pgkit.NewMigrator()
		err := migrator.RunMigrations(db, "testdata")
		require.NoError(t, err)

		cmd := exec.Command("./pgkit", "status", "--db", dbURL)
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "status command should succeed: %s", string(output))
		outputStr := string(output)
		idx1 := strings.Index(outputStr, "001_create_users.sql")
		idx2 := strings.Index(outputStr, "002_add_email_to_users.sql")
		idx3 := strings.Index(outputStr, "003_add_index_on_email.sql")
		idx4 := strings.Index(outputStr, "004_add_status_column.sql")
		assert.True(t, idx1 < idx2, "001 should appear before 002")
		assert.True(t, idx2 < idx3, "002 should appear before 003")
		assert.True(t, idx3 < idx4, "003 should appear before 004")
	})

	t.Run("returns_error_when_database_url_not_set", func(t *testing.T) {
		cmd := exec.Command("./pgkit", "status")
		cmd.Env = []string{} // Clear environment

		output, err := cmd.CombinedOutput()

		assert.Error(t, err)
		assert.Contains(t, string(output), "database URL not provided")
	})
}
