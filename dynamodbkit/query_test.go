package dynamodbkit

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	t.Run("returns_an_error_when_getting_a_new_dynamodb_connection_returns_an_error", func(t *testing.T) {
		setFake(func(ctx context.Context) (DynamoDB, error) { return nil, errors.New("the fake error") })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.Nil(t, result)
		assert.EqualError(t, err, "error creating DynamoDB client: the fake error")
	})

	t.Run("passes_the_table_name_to_query", func(t *testing.T) {
		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "theTableName", "id", "aUserID")

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.Equal(t, "theTableName", actualTableName)
	})

	t.Run("builds_key_condition_expression_for_string_partition_key", func(t *testing.T) {
		var actualInput *dynamodb.QueryInput
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				actualInput = params
				return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "userId", "theUserID")

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.NotNil(t, actualInput.KeyConditionExpression)
		assert.NotNil(t, actualInput.ExpressionAttributeNames)
		assert.NotNil(t, actualInput.ExpressionAttributeValues)
		// The expression should reference the partition key
		assert.Contains(t, *actualInput.KeyConditionExpression, "#0")
		assert.Contains(t, actualInput.ExpressionAttributeNames, "#0")
		assert.Equal(t, "userId", actualInput.ExpressionAttributeNames["#0"])
		assert.Contains(t, actualInput.ExpressionAttributeValues, ":0")
		assert.Equal(t, &types.AttributeValueMemberS{Value: "theUserID"}, actualInput.ExpressionAttributeValues[":0"])
	})

	t.Run("builds_key_condition_expression_for_integer_partition_key", func(t *testing.T) {
		var actualInput *dynamodb.QueryInput
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				actualInput = params
				return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", 12345)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.NotNil(t, actualInput.KeyConditionExpression)
		assert.NotNil(t, actualInput.ExpressionAttributeValues)
		// The expression should contain the integer value
		assert.Contains(t, actualInput.ExpressionAttributeValues, ":0")
		assert.Equal(t, &types.AttributeValueMemberN{Value: "12345"}, actualInput.ExpressionAttributeValues[":0"])
	})

	t.Run("returns_an_error_when_query_returns_an_error", func(t *testing.T) {
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				return nil, errors.New("the fake error")
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.Nil(t, result)
		assert.EqualError(t, err, "error querying table aTable: the fake error")
	})

	t.Run("returns_empty_results_when_query_returns_no_items", func(t *testing.T) {
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns_multiple_items_when_query_succeeds", func(t *testing.T) {
		user1 := TestUser{ID: "theUserID1", Name: "theUserName1", Email: "theUserEmail1"}
		user2 := TestUser{ID: "theUserID2", Name: "theUserName2", Email: "theUserEmail2"}
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						mustMarshalMap(t, user1),
						mustMarshalMap(t, user2),
					},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "theUserID1", result[0].ID)
		assert.Equal(t, "theUserName1", result[0].Name)
		assert.Equal(t, "theUserEmail1", result[0].Email)
		assert.Equal(t, "theUserID2", result[1].ID)
		assert.Equal(t, "theUserName2", result[1].Name)
		assert.Equal(t, "theUserEmail2", result[1].Email)
	})

	t.Run("returns_an_error_when_an_item_cannot_be_unmarshalled", func(t *testing.T) {
		invalidItem := map[string]types.AttributeValue{
			"id":   &types.AttributeValueMemberS{Value: "123"},
			"name": &types.AttributeValueMemberL{Value: []types.AttributeValue{}}, // Invalid for string field
		}
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{invalidItem}}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.Nil(t, result)
		assert.EqualError(t, err, "error unmarshalling queried item: unmarshal failed, cannot unmarshal list into Go value type string")
	})

	t.Run("applies_query_options_correctly", func(t *testing.T) {
		var actualInput *dynamodb.QueryInput
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				actualInput = params
				return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID",
			WithQueryProjectionExpression("id, #name"))

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.NotNil(t, actualInput)
		assert.Equal(t, "id, #name", *actualInput.ProjectionExpression)
	})

	t.Run("returns_an_error_when_query_option_processing_fails", func(t *testing.T) {
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				return &dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		failingOption := func(input *dynamodb.QueryInput) error {
			return errors.New("option processing failed")
		}

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID", failingOption)

		assert.Nil(t, result)
		assert.EqualError(t, err, "error processing option: option processing failed")
	})

	t.Run("succeeds_when_no_errors", func(t *testing.T) {
		user := TestUser{ID: "theUserID", Name: "theUserName", Email: "theUserEmail"}
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						mustMarshalMap(t, user),
					},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "theUserID", result[0].ID)
		assert.Equal(t, "theUserName", result[0].Name)
		assert.Equal(t, "theUserEmail", result[0].Email)
	})
}

func TestWithQueryProjectionExpression(t *testing.T) {
	t.Run("sets_projection_expression_when_given_string", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		option := WithQueryProjectionExpression("id, name, email")

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.ProjectionExpression)
		assert.Equal(t, "id, name, email", *input.ProjectionExpression)
	})

	t.Run("sets_projection_expression_when_given_complex_expression", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		option := WithQueryProjectionExpression("#id, #name, #ts")

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.ProjectionExpression)
		assert.Equal(t, "#id, #name, #ts", *input.ProjectionExpression)
	})
}
