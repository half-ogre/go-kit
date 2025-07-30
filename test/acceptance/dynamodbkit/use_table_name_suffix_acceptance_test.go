//go:build acceptance

package dynamodbkit_test

import (
	"context"
	"os"
	"testing"

	"github.com/half-ogre/go-kit/dynamodbkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseTableNameSuffixAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	t.Run("global_suffix_with_nonexistent_table_returns_error", func(t *testing.T) {
		// Set global suffix to something that creates a non-existent table
		dynamodbkit.UseTableNameSuffix("nonexistent")
		t.Cleanup(func() { dynamodbkit.UseTableNameSuffix("") })

		// Try to get item from nonexistent table (test_usersnonexistent)
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "test-user")

		// Should get an error about the table not existing
		assert.Error(t, err)
		assert.Nil(t, result)
		// The error should be a ResourceNotFoundException from DynamoDB
		assert.Contains(t, err.Error(), "ResourceNotFoundException")
	})

	t.Run("get_item_with_global_suffix_returns_error_for_nonexistent_table", func(t *testing.T) {
		// Set up a global suffix that creates a non-existent table name
		dynamodbkit.UseTableNameSuffix("with_sort")
		t.Cleanup(func() { dynamodbkit.UseTableNameSuffix("") })

		// Try to retrieve from non-existent table (test_userswith_sort doesn't exist)
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users", "user_id", "global-get-user",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-01T10:00:00Z"))

		// Should get an error about the table not existing
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")
	})

	t.Run("put_item_with_global_suffix_returns_error_for_nonexistent_table", func(t *testing.T) {
		// Set up a global suffix that creates a non-existent table name
		dynamodbkit.UseTableNameSuffix("with_sort")
		t.Cleanup(func() { dynamodbkit.UseTableNameSuffix("") })

		// Try to put item to non-existent table (test_userswith_sort doesn't exist)
		testUser := TestUserWithSort{
			UserID:    "global-put-user",
			Timestamp: "2023-01-01T10:00:00Z",
			Name:      "GlobalPutUser",
			Data:      "put-data",
		}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)

		// Should get an error about the table not existing
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")
	})

	t.Run("query_with_global_suffix_returns_error_for_nonexistent_table", func(t *testing.T) {
		// Set up a global suffix that creates a non-existent table name
		dynamodbkit.UseTableNameSuffix("with_sort")
		t.Cleanup(func() { dynamodbkit.UseTableNameSuffix("") })

		// Try to query non-existent table (test_userswith_sort doesn't exist)
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users", "user_id", "global-query-user")

		// Should get an error about the table not existing
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")
	})

	t.Run("scan_with_global_suffix_returns_error_for_nonexistent_table", func(t *testing.T) {
		// Set up a global suffix that creates a non-existent table name
		dynamodbkit.UseTableNameSuffix("with_sort")
		t.Cleanup(func() { dynamodbkit.UseTableNameSuffix("") })

		// Try to scan non-existent table (test_userswith_sort doesn't exist)
		result, err := dynamodbkit.Scan[TestUserWithSort](ctx, "test_users")

		// Should get an error about the table not existing
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")
	})

	t.Run("delete_item_with_global_suffix_returns_error_for_nonexistent_table", func(t *testing.T) {
		// Set up a global suffix that creates a non-existent table name
		dynamodbkit.UseTableNameSuffix("with_sort")
		t.Cleanup(func() { dynamodbkit.UseTableNameSuffix("") })

		// Try to delete from non-existent table (test_userswith_sort doesn't exist)
		err := dynamodbkit.DeleteItem(ctx, "test_users", "user_id", "global-delete-user",
			dynamodbkit.WithDeleteItemSortKey("timestamp", "2023-01-01T10:00:00Z"))

		// Should get an error about the table not existing
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")
	})

	t.Run("option_suffix_overrides_global_suffix", func(t *testing.T) {
		// Clear the main table before setting global suffix
		clearTestTable(t, ctx)

		// First, add an item to the main table (test_users) without any global suffix
		testUser := TestUser{ID: "option-override-user", Name: "OptionUser", Email: "option@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Now set up a global suffix that would normally create non-existent table
		dynamodbkit.UseTableNameSuffix("nonexistent")
		t.Cleanup(func() { dynamodbkit.UseTableNameSuffix("") })

		// Query with option suffix that overrides global suffix
		// This should query test_users (base table) instead of test_usersnonexistent
		result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", "option-override-user",
			dynamodbkit.WithQueryTableNameSuffix(""))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "option-override-user", result.Items[0].ID)
		assert.Equal(t, "OptionUser", result.Items[0].Name)

		// Clean up (reset suffix to empty so delete works)
		dynamodbkit.UseTableNameSuffix("")
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("multiple_operations_with_global_suffix_all_return_same_error", func(t *testing.T) {
		// Set up a global suffix that creates a non-existent table name
		dynamodbkit.UseTableNameSuffix("with_sort")
		t.Cleanup(func() { dynamodbkit.UseTableNameSuffix("") })

		// Test multiple operations using the same global suffix all fail consistently
		testUser := TestUserWithSort{
			UserID:    "multi-op-user",
			Timestamp: "2023-01-01T10:00:00Z",
			Name:      "MultiOpUser",
			Data:      "multi-data",
		}

		// Put item using global suffix should fail
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")

		// Get item using global suffix should fail
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users", "user_id", "multi-op-user",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-01T10:00:00Z"))
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")

		// Query using global suffix should fail
		queryResult, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users", "user_id", "multi-op-user")
		assert.Error(t, err)
		assert.Nil(t, queryResult)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")

		// Delete using global suffix should fail
		err = dynamodbkit.DeleteItem(ctx, "test_users", "user_id", "multi-op-user",
			dynamodbkit.WithDeleteItemSortKey("timestamp", "2023-01-01T10:00:00Z"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")
	})

	t.Run("changing_global_suffix_affects_subsequent_operations", func(t *testing.T) {
		// Clear the main table before setting any global suffix
		clearTestTable(t, ctx)

		// Set first suffix (empty) and add item to main table
		dynamodbkit.UseTableNameSuffix("")
		testUser1 := TestUser{ID: "changing-suffix-user1", Name: "User1", Email: "user1@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser1)
		require.NoError(t, err)

		// Change to second suffix that creates non-existent table
		dynamodbkit.UseTableNameSuffix("nonexistent")
		testUser2 := TestUser{
			ID:    "changing-suffix-user2",
			Name:  "User2",
			Email: "user2@example.com",
		}
		// This should fail because test_usersnonexistent doesn't exist
		err = dynamodbkit.PutItem(ctx, "test_users", testUser2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")

		// Verify first item is still in main table (using no suffix)
		dynamodbkit.UseTableNameSuffix("")
		result1, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "changing-suffix-user1")
		require.NoError(t, err)
		require.NotNil(t, result1)
		assert.Equal(t, "changing-suffix-user1", result1.ID)

		// Verify second item operations fail with the problematic suffix
		dynamodbkit.UseTableNameSuffix("nonexistent")
		result2, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "changing-suffix-user2")
		assert.Error(t, err)
		assert.Nil(t, result2)
		assert.Contains(t, err.Error(), "ResourceNotFoundException")

		// Clean up
		dynamodbkit.UseTableNameSuffix("")
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser1.ID)

		t.Cleanup(func() { dynamodbkit.UseTableNameSuffix("") })
	})
}
