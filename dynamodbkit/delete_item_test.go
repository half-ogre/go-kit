package dynamodbkit

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestDeleteItem(t *testing.T) {
	t.Run("returns_an_error_when_getting_a_new_dynamodb_connection_returns_an_error", func(t *testing.T) {
		setFake(func(ctx context.Context) (DynamoDB, error) { return nil, errors.New("the fake error") })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "aTable", "id", "aUserID")

		assert.EqualError(t, err, "error creating DynamoDB client: the fake error")
	})

	t.Run("passes_the_table_name_to_delete_item", func(t *testing.T) {
		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "theTableName", "id", "aUserID")

		assert.NoError(t, err)
		assert.Equal(t, "theTableName", actualTableName)
	})

	t.Run("passes_the_partition_key_and_value_to_delete_item", func(t *testing.T) {
		var actualKey map[string]types.AttributeValue
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				actualKey = params.Key
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "aTable", "userId", "theUserID")

		assert.NoError(t, err)
		assert.NotNil(t, actualKey)
		assert.Contains(t, actualKey, "userId")
		assert.Equal(t, &types.AttributeValueMemberS{Value: "theUserID"}, actualKey["userId"])
	})

	t.Run("passes_integer_partition_key_value_correctly", func(t *testing.T) {
		var actualKey map[string]types.AttributeValue
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				actualKey = params.Key
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "aTable", "id", 12345)

		assert.NoError(t, err)
		assert.NotNil(t, actualKey)
		assert.Contains(t, actualKey, "id")
		assert.Equal(t, &types.AttributeValueMemberN{Value: "12345"}, actualKey["id"])
	})

	t.Run("returns_an_error_when_delete_item_returns_an_error", func(t *testing.T) {
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				return nil, errors.New("the fake error")
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "aTable", "id", "aUserID")

		assert.EqualError(t, err, "error deleting item: the fake error")
	})

	t.Run("applies_delete_item_options_correctly", func(t *testing.T) {
		var actualInput *dynamodb.DeleteItemInput
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				actualInput = params
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "aTable", "userId", "aUserID",
			WithDeleteItemSortKey("timestamp", "2023-01-01"),
			WithDeleteItemReturnValues(types.ReturnValueAllOld))

		assert.NoError(t, err)
		assert.NotNil(t, actualInput)
		assert.Contains(t, actualInput.Key, "timestamp")
		assert.Equal(t, &types.AttributeValueMemberS{Value: "2023-01-01"}, actualInput.Key["timestamp"])
		assert.Equal(t, types.ReturnValueAllOld, actualInput.ReturnValues)
	})

	t.Run("returns_an_error_when_delete_item_option_processing_fails", func(t *testing.T) {
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		failingOption := func(input *dynamodb.DeleteItemInput) error {
			return errors.New("option processing failed")
		}

		err := DeleteItem(context.Background(), "aTable", "id", "aUserID", failingOption)

		assert.EqualError(t, err, "error processing option: option processing failed")
	})

	t.Run("succeeds_when_no_errors", func(t *testing.T) {
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
	})
}

func TestWithDeleteItemSortKey(t *testing.T) {
	t.Run("sets_string_sort_key_when_given_string_value", func(t *testing.T) {
		input := &dynamodb.DeleteItemInput{
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: "aUserID"},
			},
		}
		option := WithDeleteItemSortKey("timestamp", "2023-01-01")

		err := option(input)

		assert.NoError(t, err)
		assert.Contains(t, input.Key, "timestamp")
		assert.Equal(t, &types.AttributeValueMemberS{Value: "2023-01-01"}, input.Key["timestamp"])
	})

	t.Run("sets_integer_sort_key_when_given_integer_value", func(t *testing.T) {
		input := &dynamodb.DeleteItemInput{
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: "aUserID"},
			},
		}
		option := WithDeleteItemSortKey("score", 98765)

		err := option(input)

		assert.NoError(t, err)
		assert.Contains(t, input.Key, "score")
		assert.Equal(t, &types.AttributeValueMemberN{Value: "98765"}, input.Key["score"])
	})
}

func TestWithDeleteItemReturnValues(t *testing.T) {
	t.Run("sets_return_values_when_given_all_old", func(t *testing.T) {
		input := &dynamodb.DeleteItemInput{}
		option := WithDeleteItemReturnValues(types.ReturnValueAllOld)

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, types.ReturnValueAllOld, input.ReturnValues)
	})

	t.Run("sets_return_values_when_given_none", func(t *testing.T) {
		input := &dynamodb.DeleteItemInput{}
		option := WithDeleteItemReturnValues(types.ReturnValueNone)

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, types.ReturnValueNone, input.ReturnValues)
	})
}
