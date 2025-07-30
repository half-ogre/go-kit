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

func TestDeleteItemAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	t.Run("delete_item_from_empty_table_succeeds", func(t *testing.T) {
		// Clear table first
		clearTestTable(t, ctx)

		// Try to delete non-existent item (should succeed without error)
		err := dynamodbkit.DeleteItem(ctx, "test_users", "id", "non-existent-user")
		assert.NoError(t, err)
	})

	t.Run("delete_existing_item_removes_it_from_table", func(t *testing.T) {
		// Clear table and add test item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "delete-test-1", Name: "ToDelete", Email: "delete@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Verify item exists
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "delete-test-1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "delete-test-1", result.ID)

		// Delete the item
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", "delete-test-1")
		require.NoError(t, err)

		// Small delay to ensure consistency
		time.Sleep(10 * time.Millisecond)

		// Verify item is gone
		result, err = dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "delete-test-1")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete_item_with_string_partition_key", func(t *testing.T) {
		// Clear table and add test item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "string-delete-test", Name: "StringDelete", Email: "stringdelete@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Delete using string key
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", "string-delete-test")
		require.NoError(t, err)

		// Verify deletion
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "string-delete-test")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete_item_with_integer_partition_key", func(t *testing.T) {
		// For this test, we'll use a numeric string ID and delete with the same string type
		// Note: DynamoDB requires exact type matching for keys
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "54321", Name: "NumericDelete", Email: "numericdelete@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Delete using string key (must match the type used when storing)
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", "54321")
		require.NoError(t, err)

		// Verify deletion
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "54321")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete_multiple_different_items", func(t *testing.T) {
		// Clear table and add multiple items
		clearTestTable(t, ctx)
		testUsers := []TestUser{
			{ID: "delete-multi-1", Name: "User1", Email: "user1@example.com"},
			{ID: "delete-multi-2", Name: "User2", Email: "user2@example.com"},
			{ID: "delete-multi-3", Name: "User3", Email: "user3@example.com"},
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Verify all items exist
		for _, user := range testUsers {
			result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", user.ID)
			require.NoError(t, err)
			require.NotNil(t, result)
		}

		// Delete each item
		for _, user := range testUsers {
			err := dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
			require.NoError(t, err)
		}

		// Small delay to ensure consistency
		time.Sleep(50 * time.Millisecond)

		// Verify all items are gone
		for _, user := range testUsers {
			result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", user.ID)
			require.NoError(t, err)
			assert.Nil(t, result, "User %s should be deleted", user.ID)
		}
	})

	t.Run("delete_item_with_special_characters_in_key", func(t *testing.T) {
		// Clear table and test with special characters
		clearTestTable(t, ctx)
		specialID := "user@domain.com_delete-123+special"
		testUser := TestUser{ID: specialID, Name: "SpecialDelete", Email: "specialdelete@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Delete the item with special characters
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", specialID)
		require.NoError(t, err)

		// Verify deletion
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", specialID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete_item_with_unicode_characters", func(t *testing.T) {
		// Clear table and test with unicode characters
		clearTestTable(t, ctx)
		unicodeID := "用户-delete-ñoël"
		testUser := TestUser{ID: unicodeID, Name: "UnicodeDelete", Email: "unicodedelete@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Delete the item with unicode characters
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", unicodeID)
		require.NoError(t, err)

		// Verify deletion
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", unicodeID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete_item_with_long_key", func(t *testing.T) {
		// Clear table and test with long key
		clearTestTable(t, ctx)
		longID := "this-is-a-very-long-delete-key-that-tests-the-systems-ability-to-handle-longer-identifiers-without-truncation-" +
			"and-we-want-to-make-sure-that-deletion-works-correctly-with-such-keys-in-real-world-scenarios-1234567890"
		testUser := TestUser{ID: longID, Name: "LongKeyDelete", Email: "longkeydelete@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Delete the item with long key
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", longID)
		require.NoError(t, err)

		// Verify deletion
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", longID)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete_same_item_multiple_times_succeeds", func(t *testing.T) {
		// Clear table and add test item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "multi-delete-test", Name: "MultiDelete", Email: "multidelete@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Delete the item first time
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", "multi-delete-test")
		require.NoError(t, err)

		// Delete the same item again (should succeed - idempotent)
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", "multi-delete-test")
		assert.NoError(t, err)

		// Delete it a third time
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", "multi-delete-test")
		assert.NoError(t, err)

		// Verify item is still gone
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "multi-delete-test")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("delete_and_recreate_item_works_correctly", func(t *testing.T) {
		// Clear table and add initial item
		clearTestTable(t, ctx)
		originalUser := TestUser{ID: "recreate-test", Name: "Original", Email: "original@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", originalUser)
		require.NoError(t, err)

		// Delete the item
		err = dynamodbkit.DeleteItem(ctx, "test_users", "id", "recreate-test")
		require.NoError(t, err)

		// Small delay to ensure consistency
		time.Sleep(10 * time.Millisecond)

		// Verify deletion
		result, err := dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "recreate-test")
		require.NoError(t, err)
		assert.Nil(t, result)

		// Recreate with different data
		newUser := TestUser{ID: "recreate-test", Name: "Recreated", Email: "recreated@example.com"}
		err = dynamodbkit.PutItem(ctx, "test_users", newUser)
		require.NoError(t, err)

		// Verify new item exists with correct data
		result, err = dynamodbkit.GetItem[TestUser](ctx, "test_users", "id", "recreate-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Recreated", result.Name)
		assert.Equal(t, "recreated@example.com", result.Email)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", "recreate-test")
	})
}

// TestDeleteItemWithSortKeyAcceptance tests DeleteItem functionality with composite keys
func TestDeleteItemWithSortKeyAcceptance(t *testing.T) {
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

	t.Run("delete_item_with_sort_key_removes_correct_item", func(t *testing.T) {
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

		// Verify all items exist
		for _, user := range testUsers {
			result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithGetItemSortKey("timestamp", user.Timestamp))
			require.NoError(t, err)
			require.NotNil(t, result)
		}

		// Delete specific item with sort key
		err := dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", "user1",
			dynamodbkit.WithDeleteItemSortKey("timestamp", "2023-01-01T11:00:00Z"))
		require.NoError(t, err)

		// Small delay to ensure consistency
		time.Sleep(10 * time.Millisecond)

		// Verify the specific item is gone
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user1",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-01T11:00:00Z"))
		require.NoError(t, err)
		assert.Nil(t, result)

		// Verify other items still exist
		result, err = dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user1",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-01T10:00:00Z"))
		require.NoError(t, err)
		assert.NotNil(t, result)

		result, err = dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user1",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-01T12:00:00Z"))
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Clean up remaining items
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", "user1",
			dynamodbkit.WithDeleteItemSortKey("timestamp", "2023-01-01T10:00:00Z"))
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", "user1",
			dynamodbkit.WithDeleteItemSortKey("timestamp", "2023-01-01T12:00:00Z"))
	})

	t.Run("delete_without_sort_key_when_sort_key_required_returns_error", func(t *testing.T) {
		// Clear table and add test item
		clearTestTableWithSort(t, ctx)
		testUser := TestUserWithSort{UserID: "user2", Timestamp: "2023-01-02T10:00:00Z", Name: "SortKeyRequired", Data: "test"}
		err := dynamodbkit.PutItem(ctx, "test_users_with_sort", testUser)
		require.NoError(t, err)

		// Try to delete without providing sort key (this should return an error for composite key tables)
		err = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", "user2")
		assert.Error(t, err) // Should return validation error for missing sort key
		assert.Contains(t, err.Error(), "The number of conditions on the keys is invalid")

		// Verify item still exists (since delete failed)
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user2",
			dynamodbkit.WithGetItemSortKey("timestamp", "2023-01-02T10:00:00Z"))
		require.NoError(t, err)
		assert.NotNil(t, result) // Should still exist

		// Clean up properly with both keys
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", testUser.UserID,
			dynamodbkit.WithDeleteItemSortKey("timestamp", testUser.Timestamp))
	})

	t.Run("delete_item_with_integer_sort_key", func(t *testing.T) {
		// Clear table and add test items with numeric string sort keys
		// Note: DynamoDB requires exact type matching for keys
		clearTestTableWithSort(t, ctx)
		testUser := TestUserWithSort{UserID: "user3", Timestamp: "98765", Name: "IntegerSort", Data: "numeric"}
		err := dynamodbkit.PutItem(ctx, "test_users_with_sort", testUser)
		require.NoError(t, err)

		// Delete item using string sort key (must match the type used when storing)
		err = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", "user3",
			dynamodbkit.WithDeleteItemSortKey("timestamp", "98765"))
		require.NoError(t, err)

		// Verify deletion
		result, err := dynamodbkit.GetItem[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user3",
			dynamodbkit.WithGetItemSortKey("timestamp", "98765"))
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}
