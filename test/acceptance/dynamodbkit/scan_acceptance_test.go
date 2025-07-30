//go:build acceptance

package dynamodbkit_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/half-ogre/go-kit/dynamodbkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUser is a test model for acceptance tests
type TestUser struct {
	ID    string `dynamodbav:"id"`
	Name  string `dynamodbav:"name"`
	Email string `dynamodbav:"email"`
}

func TestScanAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	t.Run("scan_empty_table_returns_empty_results", func(t *testing.T) {
		// Clear table first
		clearTestTable(t, ctx)

		// Scan empty table
		result, err := dynamodbkit.Scan[TestUser](ctx, "test_users")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Items)
		assert.Nil(t, result.LastEvaluatedKey)
	})

	t.Run("scan_single_item_returns_one_result", func(t *testing.T) {
		// Clear table and add one item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "single-user", Name: "SingleUser", Email: "single@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Scan the table
		result, err := dynamodbkit.Scan[TestUser](ctx, "test_users")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, testUser.ID, result.Items[0].ID)
		assert.Equal(t, testUser.Name, result.Items[0].Name)
		assert.Equal(t, testUser.Email, result.Items[0].Email)
		assert.Nil(t, result.LastEvaluatedKey)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("scan_multiple_items_without_pagination", func(t *testing.T) {
		// Clear table and add multiple items
		clearTestTable(t, ctx)
		testUsers := []TestUser{
			{ID: "user1", Name: "Alice", Email: "alice@example.com"},
			{ID: "user2", Name: "Bob", Email: "bob@example.com"},
			{ID: "user3", Name: "Charlie", Email: "charlie@example.com"},
		}

		// Put test data
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Scan all items
		result, err := dynamodbkit.Scan[TestUser](ctx, "test_users")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Items, 3)

		// Verify all items are returned (order may vary)
		returnedIDs := make([]string, len(result.Items))
		for i, item := range result.Items {
			returnedIDs[i] = item.ID
		}
		sort.Strings(returnedIDs)
		expectedIDs := []string{"user1", "user2", "user3"}
		sort.Strings(expectedIDs)
		assert.Equal(t, expectedIDs, returnedIDs)

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
		}
	})

	t.Run("scan_with_limit_returns_correct_number_of_items", func(t *testing.T) {
		// Clear table and add multiple items
		clearTestTable(t, ctx)
		testUsers := createTestUsers(5)

		// Put test data
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Scan with limit of 2
		result, err := dynamodbkit.Scan[TestUser](ctx, "test_users", dynamodbkit.WithScanLimit(2))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Items, 2)

		// Should have a LastEvaluatedKey for pagination
		assert.NotNil(t, result.LastEvaluatedKey)

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
		}
	})

	t.Run("scan_with_pagination_retrieves_all_items", func(t *testing.T) {
		// Clear table and add multiple items
		clearTestTable(t, ctx)
		testUsers := createTestUsers(7) // 7 items to ensure multiple pages

		// Put test data
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Scan with pagination using limit of 3
		allItems := []TestUser{}
		var exclusiveStartKey *string
		pageCount := 0

		for {
			pageCount++
			var result *dynamodbkit.ScanOutput[TestUser]
			var err error

			if exclusiveStartKey != nil {
				result, err = dynamodbkit.Scan[TestUser](ctx, "test_users",
					dynamodbkit.WithScanLimit(3),
					dynamodbkit.WithScanExclusiveStartKey(*exclusiveStartKey))
			} else {
				result, err = dynamodbkit.Scan[TestUser](ctx, "test_users",
					dynamodbkit.WithScanLimit(3))
			}

			require.NoError(t, err)
			assert.NotNil(t, result)

			t.Logf("Page %d: Retrieved %d items", pageCount, len(result.Items))
			allItems = append(allItems, result.Items...)

			// Check if there are more pages
			if result.LastEvaluatedKey == nil {
				break
			}
			exclusiveStartKey = result.LastEvaluatedKey
		}

		// Verify we got all items
		assert.Len(t, allItems, 7)
		assert.True(t, pageCount >= 2, "Should have required multiple pages")

		// Verify all unique IDs are present
		retrievedIDs := make(map[string]bool)
		for _, item := range allItems {
			retrievedIDs[item.ID] = true
		}
		assert.Len(t, retrievedIDs, 7, "All items should have unique IDs")

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
		}
	})

	t.Run("scan_with_exclusive_start_key_continues_from_correct_position", func(t *testing.T) {
		// Clear table and add test data
		clearTestTable(t, ctx)
		testUsers := createTestUsers(5)

		// Put test data
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// First scan with limit
		firstResult, err := dynamodbkit.Scan[TestUser](ctx, "test_users", dynamodbkit.WithScanLimit(2))
		require.NoError(t, err)
		require.NotNil(t, firstResult.LastEvaluatedKey)
		firstPageIDs := make(map[string]bool)
		for _, item := range firstResult.Items {
			firstPageIDs[item.ID] = true
		}

		// Second scan starting from LastEvaluatedKey
		secondResult, err := dynamodbkit.Scan[TestUser](ctx, "test_users",
			dynamodbkit.WithScanExclusiveStartKey(*firstResult.LastEvaluatedKey))
		require.NoError(t, err)

		// Verify no overlap between first and second page
		for _, item := range secondResult.Items {
			assert.False(t, firstPageIDs[item.ID], "Item %s should not appear in both pages", item.ID)
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
		}
	})

	t.Run("scan_with_limit_one_returns_single_item", func(t *testing.T) {
		// Clear table and add test data
		clearTestTable(t, ctx)
		testUsers := createTestUsers(3)

		// Put test data
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Scan with limit 1 (should return exactly 1 item)
		result, err := dynamodbkit.Scan[TestUser](ctx, "test_users", dynamodbkit.WithScanLimit(1))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Items, 1)            // Should get exactly 1 item
		assert.NotNil(t, result.LastEvaluatedKey) // Should have pagination key

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
		}
	})

	t.Run("scan_pagination_last_evaluated_key_format_verification", func(t *testing.T) {
		// Clear table and add test data to force pagination
		clearTestTable(t, ctx)
		testUsers := createTestUsers(5)

		// Put test data
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Scan with small limit to force pagination
		result, err := dynamodbkit.Scan[TestUser](ctx, "test_users", dynamodbkit.WithScanLimit(2))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Items, 2)

		// Verify LastEvaluatedKey format
		if result.LastEvaluatedKey != nil {
			t.Logf("LastEvaluatedKey (base64): %s", *result.LastEvaluatedKey)

			// Decode the base64 to see the actual JSON format
			decodedJson, err := base64.StdEncoding.DecodeString(*result.LastEvaluatedKey)
			require.NoError(t, err)
			t.Logf("LastEvaluatedKey (JSON): %s", string(decodedJson))

			// Verify the JSON contains the expected structure
			assert.Contains(t, string(decodedJson), "id")
			// Note: Local DynamoDB format is simpler: {"id":"user5"}
			// Real AWS DynamoDB format would be: {"id":{"S":"user5"}}
			// Both formats work with our SDK wrapper

			// Test that the key can be used for subsequent scans
			nextResult, err := dynamodbkit.Scan[TestUser](ctx, "test_users",
				dynamodbkit.WithScanExclusiveStartKey(*result.LastEvaluatedKey))
			assert.NoError(t, err, "Should be able to use real LastEvaluatedKey for next scan")
			assert.NotNil(t, nextResult)
			t.Logf("Next scan returned %d items", len(nextResult.Items))
		} else {
			t.Log("No LastEvaluatedKey returned (all items fit in one page)")
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
		}
	})

	t.Run("scan_with_large_limit_returns_all_available_items", func(t *testing.T) {
		// Clear table and add test data
		clearTestTable(t, ctx)
		testUsers := createTestUsers(3)

		// Put test data
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Scan with limit larger than available items
		result, err := dynamodbkit.Scan[TestUser](ctx, "test_users", dynamodbkit.WithScanLimit(100))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Items, 3)         // Should get all 3 items
		assert.Nil(t, result.LastEvaluatedKey) // No pagination needed

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
		}
	})
}

// Helper function to create test users
func createTestUsers(count int) []TestUser {
	users := make([]TestUser, count)
	for i := 0; i < count; i++ {
		users[i] = TestUser{
			ID:    fmt.Sprintf("user%d", i+1),
			Name:  fmt.Sprintf("User%d", i+1),
			Email: fmt.Sprintf("user%d@example.com", i+1),
		}
	}
	return users
}

// Helper function to clear the test table
func clearTestTable(t *testing.T, ctx context.Context) {
	// Scan all items with pagination to ensure we get everything
	var allItems []TestUser
	var exclusiveStartKey *string

	for {
		var result *dynamodbkit.ScanOutput[TestUser]
		var err error

		if exclusiveStartKey != nil {
			result, err = dynamodbkit.Scan[TestUser](ctx, "test_users",
				dynamodbkit.WithScanExclusiveStartKey(*exclusiveStartKey))
		} else {
			result, err = dynamodbkit.Scan[TestUser](ctx, "test_users")
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
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", item.ID)
	}

	// Small delay to ensure deletions are complete
	if len(allItems) > 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func TestScanTableNameSuffixAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	t.Run("scan_with_table_name_suffix_modifies_table_name", func(t *testing.T) {
		// This test uses a table that doesn't exist (since suffix would make it invalid)
		// We expect this to fail with the modified table name
		result, err := dynamodbkit.Scan[TestUser](ctx, "test_users",
			dynamodbkit.WithScanTableNameSuffix("nonexistent"))

		// Should get an error about the table not existing
		assert.Error(t, err)
		assert.Nil(t, result)
		// The error should be a ResourceNotFoundException from DynamoDB
		assert.Contains(t, err.Error(), "ResourceNotFoundException")
	})
}
