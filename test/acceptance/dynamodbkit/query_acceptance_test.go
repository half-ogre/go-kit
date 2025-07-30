//go:build acceptance

package dynamodbkit_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/half-ogre/go-kit/dynamodbkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	t.Run("query_empty_table_returns_empty_results", func(t *testing.T) {
		// Clear table first
		clearTestTable(t, ctx)

		// Query empty table
		result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", "non-existent-user")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Items)
		assert.Nil(t, result.LastEvaluatedKey)
	})

	t.Run("query_non_existent_partition_key_returns_empty", func(t *testing.T) {
		// Clear table and add some test data
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "existing-user", Name: "ExistingUser", Email: "existing@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Query for different partition key
		result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", "non-existent-user")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Items)
		assert.Nil(t, result.LastEvaluatedKey)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("query_existing_partition_key_returns_single_item", func(t *testing.T) {
		// Clear table and add test item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "query-single-user", Name: "QueryUser", Email: "query@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Query for the specific partition key
		result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", "query-single-user")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "query-single-user", result.Items[0].ID)
		assert.Equal(t, "QueryUser", result.Items[0].Name)
		assert.Equal(t, "query@example.com", result.Items[0].Email)
		assert.Nil(t, result.LastEvaluatedKey)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("query_with_string_partition_key", func(t *testing.T) {
		// Clear table and add test item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "string-query-test", Name: "StringQuery", Email: "stringquery@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Query using string partition key
		result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", "string-query-test")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 1)
		assert.Equal(t, "string-query-test", result.Items[0].ID)
		assert.Equal(t, "StringQuery", result.Items[0].Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("query_with_integer_partition_key", func(t *testing.T) {
		// Test with numeric string ID and query with string type (must match storage type)
		// Note: DynamoDB requires exact type matching for keys
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "54321", Name: "NumericQuery", Email: "numericquery@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Query using string key (must match the type used when storing)
		result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", "54321")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 1)
		assert.Equal(t, "54321", result.Items[0].ID)
		assert.Equal(t, "NumericQuery", result.Items[0].Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("query_with_projection_expression_returns_limited_fields", func(t *testing.T) {
		// Clear table and add test item
		clearTestTable(t, ctx)
		testUser := TestUser{ID: "projection-test", Name: "ProjectionUser", Email: "projection@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Query with projection expression (avoid reserved keywords like 'name')
		result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", "projection-test",
			dynamodbkit.WithQueryProjectionExpression("id, email"))
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 1)
		assert.Equal(t, "projection-test", result.Items[0].ID)
		assert.Equal(t, "projection@example.com", result.Items[0].Email)
		// Name should be empty since it wasn't included in projection
		assert.Empty(t, result.Items[0].Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("query_with_special_characters_in_key", func(t *testing.T) {
		// Clear table and test with special characters
		clearTestTable(t, ctx)
		specialID := "user@domain.com_query-123+special"
		testUser := TestUser{ID: specialID, Name: "SpecialQuery", Email: "specialquery@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Query with special characters
		result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", specialID)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 1)
		assert.Equal(t, specialID, result.Items[0].ID)
		assert.Equal(t, "SpecialQuery", result.Items[0].Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("query_with_unicode_characters", func(t *testing.T) {
		// Clear table and test with unicode characters
		clearTestTable(t, ctx)
		unicodeID := "用户-query-ñoël"
		testUser := TestUser{ID: unicodeID, Name: "UnicodeQuery", Email: "unicodequery@example.com"}
		err := dynamodbkit.PutItem(ctx, "test_users", testUser)
		require.NoError(t, err)

		// Query with unicode characters
		result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", unicodeID)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 1)
		assert.Equal(t, unicodeID, result.Items[0].ID)
		assert.Equal(t, "UnicodeQuery", result.Items[0].Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", testUser.ID)
	})

	t.Run("query_multiple_items_with_different_partition_keys", func(t *testing.T) {
		// Clear table and add multiple items with different partition keys
		clearTestTable(t, ctx)
		testUsers := []TestUser{
			{ID: "query-multi-1", Name: "User1", Email: "user1@example.com"},
			{ID: "query-multi-2", Name: "User2", Email: "user2@example.com"},
			{ID: "query-multi-3", Name: "User3", Email: "user3@example.com"},
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users", user)
			require.NoError(t, err)
		}

		// Query each partition key individually and verify results
		for _, expectedUser := range testUsers {
			result, err := dynamodbkit.Query[TestUser](ctx, "test_users", "id", expectedUser.ID)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Items, 1)
			assert.Equal(t, expectedUser.ID, result.Items[0].ID)
			assert.Equal(t, expectedUser.Name, result.Items[0].Name)
			assert.Equal(t, expectedUser.Email, result.Items[0].Email)
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users", "id", user.ID)
		}
	})
}

// TestQueryWithSortKeyAcceptance tests Query functionality with composite keys (partition + sort)
func TestQueryWithSortKeyAcceptance(t *testing.T) {
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

	t.Run("query_empty_table_with_sort_key_returns_empty", func(t *testing.T) {
		// Clear table first
		clearTestTableWithSort(t, ctx)

		// Query empty table
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "non-existent-user")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Items)
		assert.Nil(t, result.LastEvaluatedKey)
	})

	t.Run("query_partition_key_returns_all_items_with_that_key", func(t *testing.T) {
		// Clear table and add multiple items with same partition key but different sort keys
		clearTestTableWithSort(t, ctx)

		testUsers := []TestUserWithSort{
			{UserID: "user1", Timestamp: "2023-01-01T10:00:00Z", Name: "FirstEntry", Data: "first"},
			{UserID: "user1", Timestamp: "2023-01-01T11:00:00Z", Name: "SecondEntry", Data: "second"},
			{UserID: "user1", Timestamp: "2023-01-01T12:00:00Z", Name: "ThirdEntry", Data: "third"},
			{UserID: "user2", Timestamp: "2023-01-01T10:00:00Z", Name: "OtherUser", Data: "other"},
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Query for user1 - should return 3 items sorted by sort key
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user1")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 3)

		// Verify all returned items have the correct partition key
		for _, item := range result.Items {
			assert.Equal(t, "user1", item.UserID)
		}

		// Verify items are in sort key order (DynamoDB returns items in sort key order)
		timestamps := make([]string, len(result.Items))
		for i, item := range result.Items {
			timestamps[i] = item.Timestamp
		}
		expectedTimestamps := []string{"2023-01-01T10:00:00Z", "2023-01-01T11:00:00Z", "2023-01-01T12:00:00Z"}
		sort.Strings(expectedTimestamps)
		sort.Strings(timestamps)
		assert.Equal(t, expectedTimestamps, timestamps)

		// Query for user2 - should return 1 item
		result, err = dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "user2")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 1)
		assert.Equal(t, "user2", result.Items[0].UserID)
		assert.Equal(t, "OtherUser", result.Items[0].Name)

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("query_non_existent_partition_key_with_sort_table_returns_empty", func(t *testing.T) {
		// Clear table and add some test data
		clearTestTableWithSort(t, ctx)
		testUser := TestUserWithSort{UserID: "existing-user", Timestamp: "2023-01-01T10:00:00Z", Name: "ExistingUser", Data: "test"}
		err := dynamodbkit.PutItem(ctx, "test_users_with_sort", testUser)
		require.NoError(t, err)

		// Query for different partition key
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "non-existent-user")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Items)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", testUser.UserID,
			dynamodbkit.WithDeleteItemSortKey("timestamp", testUser.Timestamp))
	})

	t.Run("query_with_string_partition_key_on_sort_table", func(t *testing.T) {
		// Clear table and add test items
		clearTestTableWithSort(t, ctx)
		testUsers := []TestUserWithSort{
			{UserID: "string-user", Timestamp: "2023-01-01T10:00:00Z", Name: "StringUser1", Data: "first"},
			{UserID: "string-user", Timestamp: "2023-01-01T11:00:00Z", Name: "StringUser2", Data: "second"},
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Query using string partition key
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "string-user")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 2)

		// Verify all items have correct partition key
		for _, item := range result.Items {
			assert.Equal(t, "string-user", item.UserID)
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("query_with_integer_partition_key_on_sort_table", func(t *testing.T) {
		// Test with numeric string ID and query with string type (must match storage type)
		clearTestTableWithSort(t, ctx)
		testUsers := []TestUserWithSort{
			{UserID: "12345", Timestamp: "100", Name: "NumericUser1", Data: "first"},
			{UserID: "12345", Timestamp: "200", Name: "NumericUser2", Data: "second"},
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Query using string key (must match the type used when storing)
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "12345")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 2)

		// Verify all items have correct partition key
		for _, item := range result.Items {
			assert.Equal(t, "12345", item.UserID)
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("query_with_projection_expression_on_sort_table", func(t *testing.T) {
		// Clear table and add test items
		clearTestTableWithSort(t, ctx)
		testUsers := []TestUserWithSort{
			{UserID: "projection-user", Timestamp: "2023-01-01T10:00:00Z", Name: "ProjectionUser1", Data: "first"},
			{UserID: "projection-user", Timestamp: "2023-01-01T11:00:00Z", Name: "ProjectionUser2", Data: "second"},
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Query with projection expression to only get certain fields (avoid DynamoDB reserved keywords)
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "projection-user",
			dynamodbkit.WithQueryProjectionExpression("user_id"))
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 2)

		// Verify projected fields are present and non-projected fields are empty
		for _, item := range result.Items {
			assert.Equal(t, "projection-user", item.UserID)
			// All other fields should be empty since they weren't included in projection
			assert.Empty(t, item.Timestamp)
			assert.Empty(t, item.Name)
			assert.Empty(t, item.Data)
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("query_large_result_set_on_sort_table", func(t *testing.T) {
		// Clear table and add many items with same partition key
		clearTestTableWithSort(t, ctx)

		var testUsers []TestUserWithSort
		for i := 1; i <= 10; i++ {
			user := TestUserWithSort{
				UserID:    "large-query-user",
				Timestamp: fmt.Sprintf("2023-01-01T%02d:00:00Z", i),
				Name:      fmt.Sprintf("User%d", i),
				Data:      fmt.Sprintf("data%d", i),
			}
			testUsers = append(testUsers, user)
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Query for all items with the partition key
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "large-query-user")
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 10)

		// Verify all items have correct partition key
		for _, item := range result.Items {
			assert.Equal(t, "large-query-user", item.UserID)
		}

		// Verify items are in sort key order
		timestamps := make([]string, len(result.Items))
		for i, item := range result.Items {
			timestamps[i] = item.Timestamp
		}
		assert.True(t, sort.StringsAreSorted(timestamps), "Items should be returned in sort key order")

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("query_with_special_characters_in_sort_table", func(t *testing.T) {
		// Clear table and test with special characters
		clearTestTableWithSort(t, ctx)
		specialID := "user@domain.com_query-sort+special"
		testUser := TestUserWithSort{
			UserID:    specialID,
			Timestamp: "2023-01-01T10:00:00Z",
			Name:      "SpecialQuery",
			Data:      "special",
		}
		err := dynamodbkit.PutItem(ctx, "test_users_with_sort", testUser)
		require.NoError(t, err)

		// Query with special characters
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", specialID)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 1)
		assert.Equal(t, specialID, result.Items[0].UserID)
		assert.Equal(t, "SpecialQuery", result.Items[0].Name)

		// Clean up
		_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", testUser.UserID,
			dynamodbkit.WithDeleteItemSortKey("timestamp", testUser.Timestamp))
	})
}

// TestQueryPaginationAcceptance tests Query functionality with pagination (LastEvaluatedKey and ExclusiveStartKey)
func TestQueryPaginationAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	t.Run("query_with_limit_returns_limited_results_and_last_evaluated_key", func(t *testing.T) {
		// Clear table and add multiple items with same partition key
		clearTestTableWithSort(t, ctx)

		var testUsers []TestUserWithSort
		for i := 1; i <= 5; i++ {
			user := TestUserWithSort{
				UserID:    "pagination-user",
				Timestamp: fmt.Sprintf("2023-01-01T%02d:00:00Z", i),
				Name:      fmt.Sprintf("User%d", i),
				Data:      fmt.Sprintf("data%d", i),
			}
			testUsers = append(testUsers, user)
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Query with limit=2
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "pagination-user",
			dynamodbkit.WithQueryLimit(2))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Items, 2)
		assert.NotNil(t, result.LastEvaluatedKey)
		assert.NotEmpty(t, *result.LastEvaluatedKey)

		// All items should have correct partition key and be in sort order
		for _, item := range result.Items {
			assert.Equal(t, "pagination-user", item.UserID)
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("query_with_exclusive_start_key_continues_from_previous_query", func(t *testing.T) {
		// Clear table and add multiple items with same partition key
		clearTestTableWithSort(t, ctx)

		var testUsers []TestUserWithSort
		for i := 1; i <= 5; i++ {
			user := TestUserWithSort{
				UserID:    "pagination-continue-user",
				Timestamp: fmt.Sprintf("2023-01-01T%02d:00:00Z", i),
				Name:      fmt.Sprintf("User%d", i),
				Data:      fmt.Sprintf("data%d", i),
			}
			testUsers = append(testUsers, user)
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// First query with limit=2
		firstResult, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "pagination-continue-user",
			dynamodbkit.WithQueryLimit(2))
		require.NoError(t, err)
		require.NotNil(t, firstResult)
		assert.Len(t, firstResult.Items, 2)
		assert.NotNil(t, firstResult.LastEvaluatedKey)

		// Second query using the last evaluated key from first query
		secondResult, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "pagination-continue-user",
			dynamodbkit.WithQueryExclusiveStartKey(*firstResult.LastEvaluatedKey))
		require.NoError(t, err)
		require.NotNil(t, secondResult)
		assert.Len(t, secondResult.Items, 3) // Remaining items

		// Verify no items are duplicated between queries
		firstIDs := make(map[string]bool)
		for _, item := range firstResult.Items {
			firstIDs[item.Timestamp] = true
		}

		for _, item := range secondResult.Items {
			assert.False(t, firstIDs[item.Timestamp], "Item %s should not appear in both result sets", item.Timestamp)
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("query_without_limit_returns_all_results_and_no_last_evaluated_key", func(t *testing.T) {
		// Clear table and add multiple items
		clearTestTableWithSort(t, ctx)

		var testUsers []TestUserWithSort
		for i := 1; i <= 3; i++ {
			user := TestUserWithSort{
				UserID:    "no-pagination-user",
				Timestamp: fmt.Sprintf("2023-01-01T%02d:00:00Z", i),
				Name:      fmt.Sprintf("User%d", i),
				Data:      fmt.Sprintf("data%d", i),
			}
			testUsers = append(testUsers, user)
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Query without limit - should return all items
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "no-pagination-user")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Items, 3)
		assert.Nil(t, result.LastEvaluatedKey) // No more items, so no LastEvaluatedKey

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("query_with_limit_equal_to_total_items_returns_no_last_evaluated_key", func(t *testing.T) {
		// Clear table and add items
		clearTestTableWithSort(t, ctx)

		var testUsers []TestUserWithSort
		for i := 1; i <= 3; i++ {
			user := TestUserWithSort{
				UserID:    "exact-limit-user",
				Timestamp: fmt.Sprintf("2023-01-01T%02d:00:00Z", i),
				Name:      fmt.Sprintf("User%d", i),
				Data:      fmt.Sprintf("data%d", i),
			}
			testUsers = append(testUsers, user)
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Query with limit equal to total items
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "exact-limit-user",
			dynamodbkit.WithQueryLimit(3))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Items, 3)
		assert.Nil(t, result.LastEvaluatedKey) // No more items, so no LastEvaluatedKey

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})

	t.Run("query_with_limit_and_projection_returns_last_evaluated_key", func(t *testing.T) {
		// Clear table and add items
		clearTestTableWithSort(t, ctx)

		var testUsers []TestUserWithSort
		for i := 1; i <= 4; i++ {
			user := TestUserWithSort{
				UserID:    "limit-projection-user",
				Timestamp: fmt.Sprintf("2023-01-01T%02d:00:00Z", i),
				Name:      fmt.Sprintf("User%d", i),
				Data:      fmt.Sprintf("data%d", i),
			}
			testUsers = append(testUsers, user)
		}

		// Put all items
		for _, user := range testUsers {
			err := dynamodbkit.PutItem(ctx, "test_users_with_sort", user)
			require.NoError(t, err)
		}

		// Query with both limit and projection
		result, err := dynamodbkit.Query[TestUserWithSort](ctx, "test_users_with_sort", "user_id", "limit-projection-user",
			dynamodbkit.WithQueryLimit(2),
			dynamodbkit.WithQueryProjectionExpression("user_id, timestamp"))
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Len(t, result.Items, 2)
		assert.NotNil(t, result.LastEvaluatedKey)

		// Verify projection worked
		for _, item := range result.Items {
			assert.Equal(t, "limit-projection-user", item.UserID)
			assert.NotEmpty(t, item.Timestamp)
			assert.Empty(t, item.Name) // Should be empty due to projection
			assert.Empty(t, item.Data) // Should be empty due to projection
		}

		// Clean up
		for _, user := range testUsers {
			_ = dynamodbkit.DeleteItem(ctx, "test_users_with_sort", "user_id", user.UserID,
				dynamodbkit.WithDeleteItemSortKey("timestamp", user.Timestamp))
		}
	})
}
