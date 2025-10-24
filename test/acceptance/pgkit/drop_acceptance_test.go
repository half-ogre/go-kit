//go:build acceptance

package pgkit_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDrop(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping acceptance test - DATABASE_URL not set")
	}

	// Build the pgkit binary
	buildPgkit(t)
	defer cleanupPgkit(t)

	dbURL := os.Getenv("DATABASE_URL")
	adminDBURL := strings.Replace(dbURL, "/testdb", "/postgres", 1)

	t.Run("successfully_drops_database", func(t *testing.T) {
		testDBName := "test_drop_db"
		createTestDatabase(t, adminDBURL, testDBName)

		cmd := exec.Command("./pgkit", "drop", testDBName, "--db", adminDBURL, "--force")
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "drop command should succeed: %s", string(output))
		assert.Contains(t, string(output), fmt.Sprintf("Database '%s' dropped successfully", testDBName))
		assertDatabaseExists(t, adminDBURL, testDBName, false)
	})

	t.Run("returns_error_when_database_does_not_exist", func(t *testing.T) {
		testDBName := "nonexistent_db"

		cmd := exec.Command("./pgkit", "drop", testDBName, "--db", adminDBURL, "--force")
		output, err := cmd.CombinedOutput()

		assert.Error(t, err, "drop command should fail when database does not exist")
		assert.Contains(t, string(output), "does not exist")
	})

	t.Run("returns_error_when_database_url_not_set", func(t *testing.T) {
		cmd := exec.Command("./pgkit", "drop", "somedb")
		cmd.Env = []string{} // Clear environment

		output, err := cmd.CombinedOutput()

		assert.Error(t, err)
		assert.Contains(t, string(output), "database URL not provided")
	})
}
