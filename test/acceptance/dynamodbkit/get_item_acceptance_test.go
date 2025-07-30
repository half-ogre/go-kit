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

func TestGetItemAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	t.Run("get_item_from_empty_table_returns_nil", func(t *testing.T) {
		// Clear table first
		clearTestTable(t, ctx)

		// Try to get non-existent item
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "non-existent-user")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("get_existing_item_returns_correct_data", func(t *testing.T) {
		// Clear table and add test item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "test-user-123", Name: "TestUser", Email: "test@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Get the item
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "test-user-123")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testUser.ID, result.ID)
		assert.Equal(t, testUser.Name, result.Name)
		assert.Equal(t, testUser.Email, result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("get_item_with_string_partition_key", func(t *testing.T) {
		// Clear table and add test item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "string-key-test", Name: "StringKeyUser", Email: "stringkey@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Get the item using string key
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "string-key-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "string-key-test", result.ID)
		assert.Equal(t, "StringKeyUser", result.Name)
		assert.Equal(t, "stringkey@example.com", result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("get_item_with_integer_partition_key", func(t *testing.T) {
		// For this test, we'll use a numeric string ID and retrieve with string type
		// Note: DynamoDB requires exact type matching for keys
		clearTestTable(t, ctx)
		
		// Put an item using PutItem first (which handles string internally)
		testUser := TestUser{ID: "12345", Name: "NumericUser", Email: "numeric@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Get it using string key (must match the type used when storing)
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "12345")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "12345", result.ID)
		assert.Equal(t, "NumericUser", result.Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("get_item_after_update_returns_latest_data", func(t *testing.T) {
		// Clear table and add initial item
		clearTestTable(t, ctx)
		originalUser := TestUser{ID: "update-test", Name: "OriginalName", Email: "original@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", originalUser)
		require.NoError(t, err)

		// Verify initial data
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "update-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "OriginalName", result.Name)

		// Update the item
		updatedUser := TestUser{ID: "update-test", Name: "UpdatedName", Email: "updated@example.com"}
		err = dynamodbkit.PutItem(ctx, "test_users", updatedUser)
		require.NoError(t, err)

		// Small delay to ensure consistency
		time.Sleep(10 * time.Millisecond)

		// Get the updated data
		result, err = dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "update-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "UpdatedName", result.Name)
		assert.Equal(t, "updated@example.com", result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", updatedUser.ID)
	})

	t.Run("get_item_after_delete_returns_nil", func(t *testing.T) {
		// Clear table and add test item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "delete-test", Name: "ToBeDeleted", Email: "delete@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Verify item exists
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "delete-test")
		require.NoError(t, err)
		require.NotNil(t, result)

		// Delete the item
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", "delete-test")
		require.NoError(t, err)

		// Small delay to ensure consistency
		time.Sleep(10 * time.Millisecond)

		// Verify item is gone
		result, err = dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "delete-test")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("get_multiple_different_items", func(t *testing.T) {
		// Clear table and add multiple items
		clearTestTable(t, ctx)
		testUsers := []TestUser{
			{ID: "multi-1", Name: "User1", Email: "user1@example.com"},
			{ID: "multi-2", Name: "User2", Email: "user2@example.com"},
			{ID: "multi-3", Name: "User3", Email: "user3@example.com"},
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Get each item individually and verify
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

	t.Run("get_item_with_special_characters_in_key", func(t *testing.T) {
		// Clear table and test with special characters
		clearTestTable(t, ctx)
		specialID := "user@domain.com_test-123+special"
		testUser := TestUser{ID: specialID, Name: "SpecialUser", Email: "special@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Get the item with special characters
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", specialID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, specialID, result.ID)
		assert.Equal(t, "SpecialUser", result.Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("get_item_with_unicode_characters", func(t *testing.T) {
		// Clear table and test with unicode characters
		clearTestTable(t, ctx)
		unicodeID := "用户-123-ñoël"
		testUser := TestUser{ID: unicodeID, Name: "UnicodeUser", Email: "unicode@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Get the item with unicode characters
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", unicodeID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, unicodeID, result.ID)
		assert.Equal(t, "UnicodeUser", result.Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("get_item_with_long_key", func(t *testing.T) {
		// Clear table and test with long key (but within DynamoDB limits)
		clearTestTable(t, ctx)
		// DynamoDB partition key limit is 2048 bytes, so we'll use a reasonably long key
		longID := "this-is-a-very-long-key-that-tests-the-systems-ability-to-handle-longer-identifiers-without-truncation-or-other-issues-" +
			"and-we-want-to-make-sure-that-everything-works-correctly-with-such-keys-in-real-world-scenarios-where-users-might-" +
			"generate-longer-identifiers-for-their-data-items-1234567890"
		testUser := TestUser{ID: longID, Name: "LongKeyUser", Email: "longkey@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Get the item with long key
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", longID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, longID, result.ID)
		assert.Equal(t, "LongKeyUser", result.Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("get_item_consistency_check", func(t *testing.T) {
		// Clear table and add item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "consistency-test", Name: "ConsistencyUser", Email: "consistency@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Get the item multiple times to verify consistency
		for i := 0; i < 5; i++ {
			result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "consistency-test")
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, "consistency-test", result.ID)
			assert.Equal(t, "ConsistencyUser", result.Name)
			assert.Equal(t, "consistency@example.com", result.Email)
		}

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})
}

// TestGetItemWithSortKeyAcceptance tests GetItem functionality with composite keys
// Note: This would require the test_users_with_sort table
func TestGetItemWithSortKeyAcceptance(t *testing.T) {
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

	t.Run("get_item_with_sort_key_returns_correct_item", func(t *testing.T) {
		// Clear table and add test items with same partition key but different sort keys
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

		// Get specific item with sort key
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user1",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-01T11:00:00Z"))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "user1", result.UserID)
		assert.Equal(t, "2023-01-01T11:00:00Z", result.Timestamp)
		assert.Equal(t, "SecondEntry", result.Name)
		assert.Equal(t, "second", result.Data)

		// Verify we get the correct specific item, not any of the others
		assert.NotEqual(t, "FirstEntry", result.Name)
		assert.NotEqual(t, "ThirdEntry", result.Name)

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("get_item_without_sort_key_when_sort_key_required_returns_error", func(t *testing.T) {
		// Clear table and add test item
		clearTestTableWithSort(t, ctx)
		testUser := TestUserWithSort{UserID: "user2", Timestamp: "2023-01-02T10:00:00Z", Name: "SortKeyRequired", Data: "test"}
		err := dynamodbkit.PutItem(ctx, "test_users_with_sort", testUser)
		require.NoError(t, err)

		// Try to get without providing sort key (should return an error for composite key tables)
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user2")
		assert.Error(t, err) // Should return validation error for missing sort key
		assert.Contains(t, err.Error(), "The number of conditions on the keys is invalid")
		assert.Nil(t, result)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", testUser.UserID,
			dynamodbkit.WithDeleteItemSortKey("timestamp", testUser.Timestamp))
	})

	t.Run("get_item_with_integer_sort_key", func(t *testing.T) {
		// Clear table and add test items with numeric string sort keys
		// Note: DynamoDB requires exact type matching for keys
		clearTestTableWithSort(t, ctx)
		testUser := TestUserWithSort{UserID: "user3", Timestamp: "12345", Name: "IntegerSort", Data: "numeric"}
		err := dynamodbkit.PutItem(ctx, "test_users_with_sort", testUser)
		require.NoError(t, err)

		// Get item using string sort key (must match the type used when storing)
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user3",
			dynamodbkit.WithGetItemSortKey("timestamp", "12345"))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "user3", result.UserID)
		assert.Equal(t, "12345", result.Timestamp)
		assert.Equal(t, "IntegerSort", result.Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", testUser.UserID,
			dynamodbkit.WithDeleteItemSortKey("timestamp", testUser.Timestamp))
	})
}

// Helper function to clear the test table with sort key
func clearTestTableWithSort(t *testing.T, ctx context.Context) {
	// Define the model for scanning
	type TestUserWithSort struct {
		UserID    string `dynamodbav:"user_id"`
		Timestamp string `dynamodbav:"timestamp"`
		Name      string `dynamodbav:"name"`
		Data      string `dynamodbav:"data"`
	}

	// Scan all items with pagination to ensure we get everything
	var allItems []TestUserWithSort
	var exclusiveStartKey *string
	
	for {
		var result *dynamodbkit.ScanOutput[TestUserWithSort]
		var err error
		
		if exclusiveStartKey != nil {
			result, err = dynamodbkit.Scan[TestUserWithSort](ctx, "test_users_with_sort", 
				dynamodbkit.WithScanExclusiveStartKey(*exclusiveStartKey))
		} else {
			result, err = dynamodbkit.Scan[TestUserWithSort](ctx, "test_users_with_sort")
		}
		
		require.NoError(t, err)
		allItems = append(allItems, result.Items...)
		
		if result.LastEvaluatedKey == nil {
			break
		}
		exclusiveStartKey = result.LastEvaluatedKey
	}

	// Delete each item
	for _, item := range allItems {
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", item.UserID,
			dynamodbkit.WithDeleteItemSortKey("timestamp", item.Timestamp))
	}
	
	// Small delay to ensure deletions are complete
	if len(allItems) > 0 {
		time.Sleep(100 * time.Millisecond)
	}
}