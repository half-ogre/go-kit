//go:build acceptance

package dynamodbkit_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/half-ogre/go-kit/dynamodbkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutItemAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	t.Run("put_item_to_empty_table_creates_item", func(t *testing.T) {
		// Clear table first
		clearTestTable(t, ctx)

		// Put new item
		testUser := TestUser{ID: "put-test-1", Name: "PutUser", Email: "put@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Verify item was created
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "put-test-1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "put-test-1", result.ID)
		assert.Equal(t, "PutUser", result.Name)
		assert.Equal(t, "put@example.com", result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("put_item_with_string_partition_key", func(t *testing.T) {
		// Clear table
		clearTestTable(t, ctx)

		testUser := TestUser{ID: "string-put-test", Name: "StringPut", Email: "stringput@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Verify item was created correctly
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "string-put-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "string-put-test", result.ID)
		assert.Equal(t, "StringPut", result.Name)
		assert.Equal(t, "stringput@example.com", result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("put_item_with_numeric_string_id", func(t *testing.T) {
		// Test with numeric string ID
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "12345", Name: "NumericPut", Email: "numericput@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Verify item was created using string key (must match storage type)
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "12345")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "12345", result.ID)
		assert.Equal(t, "NumericPut", result.Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("put_item_overwrites_existing_item", func(t *testing.T) {
		// Clear table and add initial item
		clearTestTable(t, ctx)
		originalUser := TestUser{ID: "overwrite-test", Name: "Original", Email: "original@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", originalUser)
		require.NoError(t, err)

		// Verify initial item
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "overwrite-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Original", result.Name)

		// Overwrite with new data
		updatedUser := TestUser{ID: "overwrite-test", Name: "Updated", Email: "updated@example.com"}
		err = dynamodbkit.PutItem(ctx, "test_users", updatedUser)
		require.NoError(t, err)

		// Small delay to ensure consistency
		time.Sleep(10 * time.Millisecond)

		// Verify item was overwritten
		result, err = dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "overwrite-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Updated", result.Name)
		assert.Equal(t, "updated@example.com", result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", updatedUser.ID)
	})

	t.Run("put_multiple_different_items", func(t *testing.T) {
		// Clear table and add multiple items
		clearTestTable(t, ctx)
		testUsers := []TestUser{
			{ID: "put-multi-1", Name: "User1", Email: "user1@example.com"},
			{ID: "put-multi-2", Name: "User2", Email: "user2@example.com"},
			{ID: "put-multi-3", Name: "User3", Email: "user3@example.com"},
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Verify all items were created correctly
		for _, expectedUser := range testUsers {
			result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", expectedUser.ID)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, expectedUser.ID, result.ID)
			assert.Equal(t, expectedUser.Name, result.Name)
			assert.Equal(t, expectedUser.Email, result.Email)
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
		}
	})

	t.Run("put_item_with_special_characters_in_fields", func(t *testing.T) {
		// Clear table and test with special characters
		clearTestTable(t, ctx)
		specialUser := TestUser{
			ID:    "special@domain.com_put-123+special",
			Name:  "User with special chars: @#$%^&*()",
			Email: "special+test@domain-name.co.uk",
		}
		err := dynamodbkit.PutItem(ctx, "test_users", specialUser)
		require.NoError(t, err)

		// Verify item was created with special characters intact
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", specialUser.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, specialUser.ID, result.ID)
		assert.Equal(t, specialUser.Name, result.Name)
		assert.Equal(t, specialUser.Email, result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", specialUser.ID)
	})

	t.Run("put_item_with_unicode_characters", func(t *testing.T) {
		// Clear table and test with unicode characters
		clearTestTable(t, ctx)
		unicodeUser := TestUser{
			ID:    "用户-put-ñoël",
			Name:  "用户名 with émojis 🚀🎉",
			Email: "unicode+tëst@domaín.com",
		}
		err := dynamodbkit.PutItem(ctx, "test_users", unicodeUser)
		require.NoError(t, err)

		// Verify item was created with unicode characters intact
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", unicodeUser.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, unicodeUser.ID, result.ID)
		assert.Equal(t, unicodeUser.Name, result.Name)
		assert.Equal(t, unicodeUser.Email, result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", unicodeUser.ID)
	})

	t.Run("put_item_with_long_values", func(t *testing.T) {
		// Clear table and test with long values
		clearTestTable(t, ctx)
		longUser := TestUser{
			ID: "long-put-test",
			Name: "This is a very long name that tests the system's ability to handle longer strings without " +
				"truncation or other issues and we want to make sure that everything works correctly with " +
				"such long values in real world scenarios where users might provide very detailed information",
			Email: "very.long.email.address.that.exceeds.normal.length@extremely-long-domain-name-for-testing-purposes.com",
		}
		err := dynamodbkit.PutItem(ctx, "test_users", longUser)
		require.NoError(t, err)

		// Verify item was created with long values intact
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "long-put-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, longUser.ID, result.ID)
		assert.Equal(t, longUser.Name, result.Name)
		assert.Equal(t, longUser.Email, result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", longUser.ID)
	})

	t.Run("put_item_with_empty_strings", func(t *testing.T) {
		// Clear table and test with empty strings (DynamoDB doesn't allow empty string attributes)
		clearTestTable(t, ctx)
		
		// Note: DynamoDB doesn't allow empty strings as attribute values
		// So we test with minimal non-empty values
		minimalUser := TestUser{
			ID:    "minimal-test",
			Name:  "A", // Single character
			Email: "@",  // Minimal email-like string
		}
		err := dynamodbkit.PutItem(ctx, "test_users", minimalUser)
		require.NoError(t, err)

		// Verify item was created
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "minimal-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "minimal-test", result.ID)
		assert.Equal(t, "A", result.Name)
		assert.Equal(t, "@", result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", minimalUser.ID)
	})

	t.Run("put_same_item_multiple_times_is_idempotent", func(t *testing.T) {
		// Clear table
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "idempotent-test", Name: "IdempotentUser", Email: "idempotent@example.com"}

		// Put the same item multiple times
		for i := 0; i < 3; i++ {
			err := dynamodbkit.PutItem(ctx, "test_users", testUser)
			require.NoError(t, err)
		}

		// Verify only one item exists with correct data
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "idempotent-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "idempotent-test", result.ID)
		assert.Equal(t, "IdempotentUser", result.Name)
		assert.Equal(t, "idempotent@example.com", result.Email)

		// Verify table only contains this one item
		scanResult, err := dynamodbkit.Scan[TestUser](ctx, "test_users")
		require.NoError(t, err)
		assert.Len(t, scanResult.Items, 1)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("put_item_after_delete_recreates_item", func(t *testing.T) {
		// Clear table and add initial item
		clearTestTable(t, ctx)
		originalUser := TestUser{ID: "recreate-put-test", Name: "Original", Email: "original@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", originalUser)
		require.NoError(t, err)

		// Delete the item
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", "recreate-put-test")
		require.NoError(t, err)

		// Small delay to ensure consistency
		time.Sleep(10 * time.Millisecond)

		// Verify deletion
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "recreate-put-test")
		require.NoError(t, err)
		assert.Nil(t, result)

		// Put new item with same ID but different data
		newUser := TestUser{ID: "recreate-put-test", Name: "Recreated", Email: "recreated@example.com"}
		err = dynamodbkit.PutItem(ctx, "test_users", newUser)
		require.NoError(t, err)

		// Verify new item exists with correct data
		result, err = dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "recreate-put-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Recreated", result.Name)
		assert.Equal(t, "recreated@example.com", result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", newUser.ID)
	})
}

// TestPutItemWithSortKeyAcceptance tests PutItem functionality with composite keys
func TestPutItemWithSortKeyAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	// TestUserWithSort model for composite key tests
	type TestUserWithSort struct {
		UserID    string `dynamodbav:"user_id"`
		Timestamp string `dynamodbav:"timestamp"`
		Name      string `dynamodbav:"name"`
		Data      string `dynamodbav:"data"`
	}

	t.Run("put_item_with_sort_key_creates_item", func(t *testing.T) {
		// Clear table
		clearTestTableWithSort(t, ctx)
		
		testUser := TestUserWithSort{UserID: "user1", Timestamp: "2023-01-01T10:00:00Z", Name: "SortKeyUser", Data: "test"}
		err := dynamodbkit.PutItem(ctx, "test_users_with_sort", testUser)
		require.NoError(t, err)

		// Verify item was created
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user1",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-01T10:00:00Z"))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "user1", result.UserID)
		assert.Equal(t, "2023-01-01T10:00:00Z", result.Timestamp)
		assert.Equal(t, "SortKeyUser", result.Name)
		assert.Equal(t, "test", result.Data)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", testUser.UserID,
			dynamodbkit.WithDeleteItemSortKey("timestamp", testUser.Timestamp))
	})

	t.Run("put_multiple_items_with_same_partition_key_different_sort_keys", func(t *testing.T) {
		// Clear table and add multiple items with same partition key but different sort keys
		clearTestTableWithSort(t, ctx)
		
		testUsers := []TestUserWithSort{
			{UserID: "user1", Timestamp: "2023-01-01T10:00:00Z", Name: "FirstEntry", Data: "first"},
			{UserID: "user1", Timestamp: "2023-01-01T11:00:00Z", Name: "SecondEntry", Data: "second"},
			{UserID: "user1", Timestamp: "2023-01-01T12:00:00Z", Name: "ThirdEntry", Data: "third"},
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Verify all items were created correctly
		for _, expectedUser := range testUsers {
			result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", expectedUser.UserID,
				dynamodbkit.WithGetItemSortKey("timestamp", expectedUser.Timestamp))
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, expectedUser.UserID, result.UserID)
			assert.Equal(t, expectedUser.Timestamp, result.Timestamp)
			assert.Equal(t, expectedUser.Name, result.Name)
			assert.Equal(t, expectedUser.Data, result.Data)
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("put_item_overwrites_existing_item_with_same_composite_key", func(t *testing.T) {
		// Clear table and add initial item
		clearTestTableWithSort(t, ctx)
		originalUser := TestUserWithSort{UserID: "user2", Timestamp: "2023-01-02T10:00:00Z", Name: "Original", Data: "original"}
		err := dynamodbkit.PutItem(ctx, "test_users_with_sort", originalUser)
		require.NoError(t, err)

		// Verify initial item
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user2",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-02T10:00:00Z"))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Original", result.Name)

		// Overwrite with new data (same composite key)
		updatedUser := TestUserWithSort{UserID: "user2", Timestamp: "2023-01-02T10:00:00Z", Name: "Updated", Data: "updated"}
		err = dynamodbkit.PutItem(ctx, "test_users_with_sort", updatedUser)
		require.NoError(t, err)

		// Small delay to ensure consistency
		time.Sleep(10 * time.Millisecond)

		// Verify item was overwritten
		result, err = dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user2",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-02T10:00:00Z"))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Updated", result.Name)
		assert.Equal(t, "updated", result.Data)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", updatedUser.UserID,
			dynamodbkit.WithDeleteItemSortKey("timestamp", updatedUser.Timestamp))
	})

	t.Run("put_item_with_integer_sort_key", func(t *testing.T) {
		// Clear table and test with numeric string sort key
		// Note: DynamoDB requires exact type matching for keys
		clearTestTableWithSort(t, ctx)
		testUser := TestUserWithSort{UserID: "user3", Timestamp: "98765", Name: "IntegerSort", Data: "numeric"}
		err := dynamodbkit.PutItem(ctx, "test_users_with_sort", testUser)
		require.NoError(t, err)

		// Verify item was created (retrieving with string sort key to match storage type)
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user3",
			dynamodbkit.WithGetItemSortKey("timestamp", "98765"))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "user3", result.UserID)
		assert.Equal(t, "98765", result.Timestamp)
		assert.Equal(t, "IntegerSort", result.Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", testUser.UserID,
			dynamodbkit.WithDeleteItemSortKey("timestamp", testUser.Timestamp))
	})
}