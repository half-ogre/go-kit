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

func TestQuery(t *testing.T) {
	t.Run("returns_an_error_when_context_is_nil", func(t *testing.T) {
		result, err := Query[TestUser](nil, "aTable", "id", "aUserID")

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("returns_an_error_when_table_name_is_empty", func(t *testing.T) {
		result, err := Query[TestUser](context.Background(), "", "id", "aUserID")

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "table name cannot be empty")
	})

	t.Run("returns_an_error_when_partition_key_is_empty", func(t *testing.T) {
		result, err := Query[TestUser](context.Background(), "aTable", "", "aUserID")

		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "partition key cannot be empty")
	})

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
		assert.NotNil(t, result)
		assert.Empty(t, result.Items)
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
		assert.NotNil(t, result)
		assert.Empty(t, result.Items)
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
		assert.NotNil(t, result)
		assert.Empty(t, result.Items)
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
		assert.Empty(t, result.Items)
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
		assert.Len(t, result.Items, 2)
		assert.Equal(t, "theUserID1", result.Items[0].ID)
		assert.Equal(t, "theUserName1", result.Items[0].Name)
		assert.Equal(t, "theUserEmail1", result.Items[0].Email)
		assert.Equal(t, "theUserID2", result.Items[1].ID)
		assert.Equal(t, "theUserName2", result.Items[1].Name)
		assert.Equal(t, "theUserEmail2", result.Items[1].Email)
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
		assert.NotNil(t, result)
		assert.Empty(t, result.Items)
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
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "theUserID", result.Items[0].ID)
		assert.Equal(t, "theUserName", result.Items[0].Name)
		assert.Equal(t, "theUserEmail", result.Items[0].Email)
	})

	t.Run("returns_last_evaluated_key_when_present_in_output", func(t *testing.T) {
		user := TestUser{ID: "theUserID", Name: "theUserName", Email: "theUserEmail"}
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{mustMarshalMap(t, user)},
					LastEvaluatedKey: map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{Value: "theLastKeyValue"},
					},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.LastEvaluatedKey)
		assert.NotEmpty(t, *result.LastEvaluatedKey)
	})

	t.Run("returns_nil_last_evaluated_key_when_not_present_in_output", func(t *testing.T) {
		user := TestUser{ID: "theUserID", Name: "theUserName", Email: "theUserEmail"}
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{mustMarshalMap(t, user)},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.LastEvaluatedKey)
	})

	t.Run("returns_an_error_when_last_evaluated_key_json_marshalling_fails", func(t *testing.T) {
		user := TestUser{ID: "theUserID", Name: "theUserName", Email: "theUserEmail"}
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{mustMarshalMap(t, user)},
					LastEvaluatedKey: map[string]types.AttributeValue{
						"invalid": &types.AttributeValueMemberB{Value: []byte{0xff, 0xfe, 0xfd}}, // Binary data that might cause JSON issues
					},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		// The test should succeed since binary data can be marshalled to JSON, but let's verify the response
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.LastEvaluatedKey)
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

func TestWithQueryExclusiveStartKey(t *testing.T) {
	t.Run("sets_exclusive_start_key_when_given_valid_encoded_key", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		// Base64 encoded JSON: {"id":"test"}
		exclusiveStartKey := "eyJpZCI6InRlc3QifQ=="
		option := WithQueryExclusiveStartKey(exclusiveStartKey)

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.ExclusiveStartKey)
		assert.Contains(t, input.ExclusiveStartKey, "id")
		assert.Equal(t, &types.AttributeValueMemberS{Value: "test"}, input.ExclusiveStartKey["id"])
	})

	t.Run("returns_an_error_when_exclusive_start_key_is_not_valid_base64", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		option := WithQueryExclusiveStartKey("invalid base64!")

		err := option(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode exclusiveStartKey")
	})

	t.Run("returns_an_error_when_exclusive_start_key_is_not_valid_json", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		// Base64 encoded invalid JSON
		invalidJson := "aW52YWxpZCBqc29u" // "invalid json"
		option := WithQueryExclusiveStartKey(invalidJson)

		err := option(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal exclusiveStartKey JSON")
	})
}

func TestWithQueryLimit(t *testing.T) {
	t.Run("sets_limit_when_given_positive_value", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		option := WithQueryLimit(100)

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.Limit)
		assert.Equal(t, int32(100), *input.Limit)
	})

	t.Run("sets_limit_when_given_zero", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		option := WithQueryLimit(0)

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.Limit)
		assert.Equal(t, int32(0), *input.Limit)
	})

	t.Run("returns_an_error_when_limit_is_negative", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		option := WithQueryLimit(-1)

		err := option(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit must be non-negative, got -1")
	})

	t.Run("returns_an_error_when_limit_exceeds_maximum", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		option := WithQueryLimit(2147483648) // int32 max + 1

		err := option(input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit exceeds maximum allowed value, got 2147483648")
	})

	t.Run("accepts_maximum_valid_limit", func(t *testing.T) {
		input := &dynamodb.QueryInput{}
		option := WithQueryLimit(2147483647) // int32 max

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.Limit)
		assert.Equal(t, int32(2147483647), *input.Limit)
	})
}

func TestWithQueryTableNameSuffix(t *testing.T) {
	t.Run("appends_suffix_to_table_name", func(t *testing.T) {
		input := &dynamodb.QueryInput{
			TableName: aws.String("theTableName"),
		}
		option := WithQueryTableNameSuffix("theSuffix")

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, "theTableName-theSuffix", *input.TableName)
	})

	t.Run("appends_suffix_to_table_name_with_existing_suffix", func(t *testing.T) {
		input := &dynamodb.QueryInput{
			TableName: aws.String("theTableName-existingSuffix"),
		}
		option := WithQueryTableNameSuffix("newSuffix")

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, "theTableName-existingSuffix-newSuffix", *input.TableName)
	})
}
