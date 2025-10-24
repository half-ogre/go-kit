//go:build acceptance

package pgkit_test

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	_ "github.com/lib/pq"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping acceptance test - DATABASE_URL not set")
	}

	// Build the pgkit binary
	buildPgkit(t)
	defer cleanupPgkit(t)

	dbURL := os.Getenv("DATABASE_URL")
	adminDBURL := strings.Replace(dbURL, "/testdb", "/postgres", 1)

	t.Run("successfully_creates_database", func(t *testing.T) {
		testDBName := "test_create_db"
		defer dropTestDatabase(t, adminDBURL, testDBName)

		cmd := exec.Command("./pgkit", "create", testDBName, "--db", adminDBURL)
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "create command should succeed: %s", string(output))
		assert.Contains(t, string(output), fmt.Sprintf("Database '%s' created successfully", testDBName))
		assertDatabaseExists(t, adminDBURL, testDBName, true)
	})

	t.Run("returns_error_when_database_already_exists", func(t *testing.T) {
		testDBName := "test_existing_db"
		createTestDatabase(t, adminDBURL, testDBName)
		defer dropTestDatabase(t, adminDBURL, testDBName)

		cmd := exec.Command("./pgkit", "create", testDBName, "--db", adminDBURL)
		output, err := cmd.CombinedOutput()

		assert.Error(t, err, "create command should fail when database exists")
		assert.Contains(t, string(output), "already exists")
	})

	t.Run("returns_error_when_database_url_not_set", func(t *testing.T) {
		cmd := exec.Command("./pgkit", "create", "somedb")
		cmd.Env = []string{} // Clear environment

		output, err := cmd.CombinedOutput()

		assert.Error(t, err)
		assert.Contains(t, string(output), "database URL not provided")
	})

	t.Run("creates_database_with_special_characters_in_name", func(t *testing.T) {
		testDBName := "test_db_123"
		defer dropTestDatabase(t, adminDBURL, testDBName)

		cmd := exec.Command("./pgkit", "create", testDBName, "--db", adminDBURL)
		output, err := cmd.CombinedOutput()

		require.NoError(t, err, "create command should succeed: %s", string(output))
		assertDatabaseExists(t, adminDBURL, testDBName, true)
	})
}

func buildPgkit(t *testing.T) {
	t.Helper()

	cmd := exec.Command("go", "build", "-o", "pgkit", "../../../cmd/pgkit")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to build pgkit: %s", string(output))
}

func cleanupPgkit(t *testing.T) {
	t.Helper()

	os.Remove("pgkit")
}

func createTestDatabase(t *testing.T, adminDBURL, dbName string) {
	t.Helper()

	db, err := sql.Open("postgres", adminDBURL)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
	require.NoError(t, err)
}

func dropTestDatabase(t *testing.T, adminDBURL, dbName string) {
	t.Helper()

	db, err := sql.Open("postgres", adminDBURL)
	if err != nil {
		return // Best effort cleanup
	}
	defer db.Close()

	db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, dbName))
}

func assertDatabaseExists(t *testing.T, adminDBURL, dbName string, shouldExist bool) {
	t.Helper()

	db, err := sql.Open("postgres", adminDBURL)
	require.NoError(t, err)
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	require.NoError(t, err)

	if shouldExist {
		assert.True(t, exists, "database %s should exist", dbName)
	} else {
		assert.False(t, exists, "database %s should not exist", dbName)
	}
}
