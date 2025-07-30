//go:build acceptance

package dynamodbkit_test

import (
	"context"
	"os"
	"slices"
	"testing"

	"github.com/half-ogre/go-kit/dynamodbkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTablesAcceptance(t *testing.T) {
	// Skip if not running against local DynamoDB
	if os.Getenv("AWS_ENDPOINT_URL") == "" {
		t.Skip("Skipping acceptance test - AWS_ENDPOINT_URL not set")
	}

	ctx := context.Background()

	t.Run("list_tables_returns_expected_tables", func(t *testing.T) {
		// List all tables
		result, err := dynamodbkit.ListTables(ctx)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should contain the test tables that exist for other acceptance tests
		expectedTables := []string{"test_users", "test_users_with_sort"}

		for _, expectedTable := range expectedTables {
			assert.Contains(t, result.TableNames, expectedTable, "Expected table %s to be in the list", expectedTable)
		}

		// All returned table names should be non-empty strings
		for _, tableName := range result.TableNames {
			assert.NotEmpty(t, tableName, "Table name should not be empty")
		}
	})

	t.Run("list_tables_with_limit_respects_limit", func(t *testing.T) {
		// List tables with a limit of 1
		result, err := dynamodbkit.ListTables(ctx, dynamodbkit.WithListTablesLimit(1))
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should return at most 1 table
		assert.LessOrEqual(t, len(result.TableNames), 1, "Should return at most 1 table when limit is 1")

		// If there are more tables than the limit, LastEvaluatedTableName should be set
		if len(result.TableNames) == 1 {
			// Get total count to see if pagination is needed
			totalResult, err := dynamodbkit.ListTables(ctx)
			require.NoError(t, err)

			if len(totalResult.TableNames) > 1 {
				assert.NotNil(t, result.LastEvaluatedTableName, "LastEvaluatedTableName should be set when there are more results")
			}
		}
	})

	t.Run("list_tables_with_pagination_works_correctly", func(t *testing.T) {
		// First, get all tables to see if we have enough for pagination
		allTablesResult, err := dynamodbkit.ListTables(ctx)
		require.NoError(t, err)

		if len(allTablesResult.TableNames) <= 1 {
			t.Skip("Not enough tables for pagination test")
		}

		// Get first page with limit 1
		firstPage, err := dynamodbkit.ListTables(ctx, dynamodbkit.WithListTablesLimit(1))
		require.NoError(t, err)
		require.NotNil(t, firstPage)
		require.Len(t, firstPage.TableNames, 1)
		require.NotNil(t, firstPage.LastEvaluatedTableName)

		// Get second page using the LastEvaluatedTableName
		secondPage, err := dynamodbkit.ListTables(ctx,
			dynamodbkit.WithListTablesLimit(1),
			dynamodbkit.WithListTablesExclusiveStartTableName(*firstPage.LastEvaluatedTableName))
		require.NoError(t, err)
		require.NotNil(t, secondPage)

		// The first table from second page should be different from first page
		if len(secondPage.TableNames) > 0 {
			assert.NotEqual(t, firstPage.TableNames[0], secondPage.TableNames[0],
				"Second page should return different table than first page")
		}

		// Collect all tables from pagination and compare with direct call
		var paginatedTables []string
		paginatedTables = append(paginatedTables, firstPage.TableNames...)
		paginatedTables = append(paginatedTables, secondPage.TableNames...)

		// Continue pagination if needed
		currentPage := secondPage
		for currentPage.LastEvaluatedTableName != nil {
			nextPage, err := dynamodbkit.ListTables(ctx,
				dynamodbkit.WithListTablesLimit(1),
				dynamodbkit.WithListTablesExclusiveStartTableName(*currentPage.LastEvaluatedTableName))
			require.NoError(t, err)
			paginatedTables = append(paginatedTables, nextPage.TableNames...)
			currentPage = nextPage
		}

		// Sort both slices for comparison
		slices.Sort(paginatedTables)
		slices.Sort(allTablesResult.TableNames)

		assert.Equal(t, allTablesResult.TableNames, paginatedTables,
			"Paginated results should match direct call results")
	})
}
