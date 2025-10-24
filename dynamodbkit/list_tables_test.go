package dynamodbkit

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
)

func TestListTables(t *testing.T) {
	t.Run("returns_an_error_when_context_is_nil", func(t *testing.T) {
		//lint:ignore SA1012 intentionally testing nil context handling
		result, err := ListTables(nil)

		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("returns_an_error_when_getting_a_new_dynamodb_connection_returns_an_error", func(t *testing.T) {
		setFake(func(ctx context.Context) (DynamoDB, error) { return nil, errors.New("the fake error") })
		t.Cleanup(func() { setFake(nil) })

		result, err := ListTables(context.Background())

		assert.Nil(t, result)
		assert.EqualError(t, err, "error creating DynamoDB client: the fake error")
	})

	t.Run("returns_the_expected_output_when_no_errors_and_no_last_evaluated_table_name", func(t *testing.T) {
		tableNames := []string{"theFirstTable", "theSecondTable"}
		fakeDB := &FakeDynamoDB{
			ListTablesFake: func(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
				return &dynamodb.ListTablesOutput{
					TableNames: tableNames,
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := ListTables(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, tableNames, result.TableNames)
		assert.Nil(t, result.LastEvaluatedTableName)
	})

	t.Run("returns_the_expected_output_when_no_errors_and_has_last_evaluated_table_name", func(t *testing.T) {
		tableNames := []string{"theFirstTable", "theSecondTable"}
		lastEvaluatedTableName := "theLastTable"
		fakeDB := &FakeDynamoDB{
			ListTablesFake: func(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
				return &dynamodb.ListTablesOutput{
					TableNames:             tableNames,
					LastEvaluatedTableName: aws.String(lastEvaluatedTableName),
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := ListTables(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, tableNames, result.TableNames)
		assert.NotNil(t, result.LastEvaluatedTableName)
		assert.Equal(t, lastEvaluatedTableName, *result.LastEvaluatedTableName)
	})

	t.Run("returns_an_error_when_list_tables_returns_an_error", func(t *testing.T) {
		fakeDB := &FakeDynamoDB{
			ListTablesFake: func(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
				return nil, errors.New("the fake error")
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := ListTables(context.Background())

		assert.Nil(t, result)
		assert.EqualError(t, err, "error listing tables: the fake error")
	})

	t.Run("returns_empty_slice_when_no_tables_exist", func(t *testing.T) {
		fakeDB := &FakeDynamoDB{
			ListTablesFake: func(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
				return &dynamodb.ListTablesOutput{
					TableNames: []string{},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := ListTables(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.TableNames)
		assert.Nil(t, result.LastEvaluatedTableName)
	})

	t.Run("returns_an_error_when_option_returns_an_error", func(t *testing.T) {
		fakeDB := &FakeDynamoDB{}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		badOption := func(input *dynamodb.ListTablesInput) error {
			return errors.New("the option error")
		}

		result, err := ListTables(context.Background(), badOption)

		assert.Nil(t, result)
		assert.EqualError(t, err, "error processing option: the option error")
	})
}

func TestWithListTablesLimit(t *testing.T) {
	t.Run("sets_limit_on_input", func(t *testing.T) {
		input := &dynamodb.ListTablesInput{}
		option := WithListTablesLimit(10)

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.Limit)
		assert.Equal(t, int32(10), *input.Limit)
	})

	t.Run("accepts_zero_limit", func(t *testing.T) {
		input := &dynamodb.ListTablesInput{}
		option := WithListTablesLimit(0)

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.Limit)
		assert.Equal(t, int32(0), *input.Limit)
	})

	t.Run("returns_an_error_when_limit_is_negative", func(t *testing.T) {
		input := &dynamodb.ListTablesInput{}
		option := WithListTablesLimit(-1)

		err := option(input)

		assert.Contains(t, err.Error(), "limit must be non-negative, got -1")
		assert.Nil(t, input.Limit)
	})

	t.Run("passes_limit_to_list_tables", func(t *testing.T) {
		actualLimit := int32(0)
		fakeDB := &FakeDynamoDB{
			ListTablesFake: func(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
				if params.Limit != nil {
					actualLimit = *params.Limit
				}
				return &dynamodb.ListTablesOutput{
					TableNames: []string{},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := ListTables(context.Background(), WithListTablesLimit(25))

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(25), actualLimit)
	})
}

func TestWithListTablesExclusiveStartTableName(t *testing.T) {
	t.Run("sets_exclusive_start_table_name_on_input", func(t *testing.T) {
		input := &dynamodb.ListTablesInput{}
		option := WithListTablesExclusiveStartTableName("theStartTable")

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.ExclusiveStartTableName)
		assert.Equal(t, "theStartTable", *input.ExclusiveStartTableName)
	})

	t.Run("returns_an_error_when_table_name_is_empty", func(t *testing.T) {
		input := &dynamodb.ListTablesInput{}
		option := WithListTablesExclusiveStartTableName("")

		err := option(input)

		assert.Contains(t, err.Error(), "exclusive start table name cannot be empty")
		assert.Nil(t, input.ExclusiveStartTableName)
	})

	t.Run("passes_exclusive_start_table_name_to_list_tables", func(t *testing.T) {
		actualExclusiveStartTableName := ""
		fakeDB := &FakeDynamoDB{
			ListTablesFake: func(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
				if params.ExclusiveStartTableName != nil {
					actualExclusiveStartTableName = *params.ExclusiveStartTableName
				}
				return &dynamodb.ListTablesOutput{
					TableNames: []string{},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := ListTables(context.Background(), WithListTablesExclusiveStartTableName("theStartTable"))

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "theStartTable", actualExclusiveStartTableName)
	})
}
