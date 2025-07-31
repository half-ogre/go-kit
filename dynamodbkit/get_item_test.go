package dynamodbkit

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestGetItem(t *testing.T) {
	t.Run("returns_an_error_when_getting_a_new_dynamodb_connection_returns_an_error", func(t *testing.T) {
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return nil, errors.New("the fake error") })
		t.Cleanup(func() { setFakeSDK(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.Nil(t, result)
		assert.EqualError(t, err, "error creating DynamoDB client: the fake error")
	})

	t.Run("passes_the_table_name_to_get_item", func(t *testing.T) {
		actualTableName := ""
		fakeDB := &FakeSDKDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.GetItemOutput{}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		result, err := GetItem[TestUser](context.Background(), "theTableName", "id", "aUserID")

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "theTableName", actualTableName)
	})

	t.Run("passes_the_partition_key_and_value_to_get_item", func(t *testing.T) {
		var actualKey map[string]types.AttributeValue
		fakeDB := &FakeSDKDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				actualKey = params.Key
				return &dynamodb.GetItemOutput{}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "userId", "theUserID")

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NotNil(t, actualKey)
		assert.Contains(t, actualKey, "userId")
		assert.Equal(t, &types.AttributeValueMemberS{Value: "theUserID"}, actualKey["userId"])
	})

	t.Run("passes_integer_partition_key_value_correctly", func(t *testing.T) {
		var actualKey map[string]types.AttributeValue
		fakeDB := &FakeSDKDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				actualKey = params.Key
				return &dynamodb.GetItemOutput{}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "id", 12345)

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NotNil(t, actualKey)
		assert.Contains(t, actualKey, "id")
		assert.Equal(t, &types.AttributeValueMemberN{Value: "12345"}, actualKey["id"])
	})

	t.Run("returns_an_error_when_get_item_returns_an_error", func(t *testing.T) {
		fakeDB := &FakeSDKDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				return nil, errors.New("the fake error")
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.Nil(t, result)
		assert.EqualError(t, err, "error getting item id=aUserID from table aTable: the fake error")
	})

	t.Run("returns_nil_when_item_not_found", func(t *testing.T) {
		fakeDB := &FakeSDKDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				return &dynamodb.GetItemOutput{Item: nil}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns_the_item_when_found", func(t *testing.T) {
		user := TestUser{ID: "theUserID", Name: "theUserName", Email: "theUserEmail"}
		fakeDB := &FakeSDKDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				return &dynamodb.GetItemOutput{
					Item: mustMarshalMap(t, user),
				}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "theUserID", result.ID)
		assert.Equal(t, "theUserName", result.Name)
		assert.Equal(t, "theUserEmail", result.Email)
	})

	t.Run("returns_an_error_when_item_cannot_be_unmarshalled", func(t *testing.T) {
		invalidItem := map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: "123"},
			"name": &types.AttributeValueMemberL{Value: []types.AttributeValue{}}, // Invalid for string field
		}
		fakeDB := &FakeSDKDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				return &dynamodb.GetItemOutput{Item: invalidItem}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.Nil(t, result)
		assert.EqualError(t, err, "failed to unmarshal item: unmarshal failed, cannot unmarshal list into Go value type string")
	})

	t.Run("applies_get_item_options_correctly", func(t *testing.T) {
		var actualInput *dynamodb.GetItemInput
		fakeDB := &FakeSDKDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				actualInput = params
				return &dynamodb.GetItemOutput{}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "userId", "aUserID",
			WithGetItemSortKey("timestamp", "2023-01-01"))

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.NotNil(t, actualInput)
		assert.Contains(t, actualInput.Key, "timestamp")
		assert.Equal(t, &types.AttributeValueMemberS{Value: "2023-01-01"}, actualInput.Key["timestamp"])
	})

	t.Run("returns_an_error_when_get_item_option_processing_fails", func(t *testing.T) {
		fakeDB := &FakeSDKDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				return &dynamodb.GetItemOutput{}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		failingOption := func(input *dynamodb.GetItemInput) error {
			return errors.New("option processing failed")
		}

		result, err := GetItem[TestUser](context.Background(), "aTable", "id", "aUserID", failingOption)

		assert.Nil(t, result)
		assert.EqualError(t, err, "error processing option: option processing failed")
	})
}

func TestWithGetItemSortKey(t *testing.T) {
	t.Run("sets_string_sort_key_when_given_string_value", func(t *testing.T) {
		input := &dynamodb.GetItemInput{
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: "aUserID"},
			},
		}
		option := WithGetItemSortKey("timestamp", "2023-01-01")

		err := option(input)

		assert.NoError(t, err)
		assert.Contains(t, input.Key, "timestamp")
		assert.Equal(t, &types.AttributeValueMemberS{Value: "2023-01-01"}, input.Key["timestamp"])
	})

	t.Run("sets_integer_sort_key_when_given_integer_value", func(t *testing.T) {
		input := &dynamodb.GetItemInput{
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: "aUserID"},
			},
		}
		option := WithGetItemSortKey("score", 98765)

		err := option(input)

		assert.NoError(t, err)
		assert.Contains(t, input.Key, "score")
		assert.Equal(t, &types.AttributeValueMemberN{Value: "98765"}, input.Key["score"])
	})
}

func TestWithGetItemTableNameSuffix(t *testing.T) {
	t.Run("appends_suffix_to_table_name", func(t *testing.T) {
		input := &dynamodb.GetItemInput{
			TableName: aws.String("theTableName"),
		}
		option := WithGetItemTableNameSuffix("theSuffix")

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, "theTableNametheSuffix", *input.TableName)
	})

	t.Run("appends_suffix_to_table_name_with_existing_suffix", func(t *testing.T) {
		input := &dynamodb.GetItemInput{
			TableName: aws.String("theTableName-existingSuffix"),
		}
		option := WithGetItemTableNameSuffix("newSuffix")

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, "theTableName-existingSuffixnewSuffix", *input.TableName)
	})
}
