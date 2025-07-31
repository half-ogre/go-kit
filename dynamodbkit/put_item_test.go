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

func TestPutItem(t *testing.T) {

	t.Run("returns_an_error_when_option_processing_fails", func(t *testing.T) {
		item := TestUser{ID: "aUserID", Name: "aUserName", Email: "aUserEmail"}
		failingOption := func(input *dynamodb.PutItemInput) error {
			return errors.New("option processing failed")
		}

		err := PutItem(context.Background(), "aTable", item, failingOption)

		assert.EqualError(t, err, "error processing option: option processing failed")
	})

	t.Run("returns_an_error_when_getting_a_new_dynamodb_connection_returns_an_error", func(t *testing.T) {
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return nil, errors.New("the fake error") })
		t.Cleanup(func() { setFakeSDK(nil) })

		item := TestUser{ID: "aUserID", Name: "aUserName", Email: "aUserEmail"}

		err := PutItem(context.Background(), "aTable", item)

		assert.EqualError(t, err, "error creating DynamoDB client: the fake error")
	})

	t.Run("passes_the_table_name_to_put_item", func(t *testing.T) {
		actualTableName := ""
		fakeDB := &FakeSDKDynamoDB{
			PutItemFake: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.PutItemOutput{}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		item := TestUser{ID: "aUserID", Name: "aUserName", Email: "aUserEmail"}

		err := PutItem(context.Background(), "theTableName", item)

		assert.NoError(t, err)
		assert.Equal(t, "theTableName", actualTableName)
	})

	t.Run("passes_the_marshalled_item_to_put_item", func(t *testing.T) {
		var actualItem map[string]types.AttributeValue
		fakeDB := &FakeSDKDynamoDB{
			PutItemFake: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
				actualItem = params.Item
				return &dynamodb.PutItemOutput{}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		item := TestUser{ID: "theUserID", Name: "theUserName", Email: "theUserEmail"}

		err := PutItem(context.Background(), "aTable", item)

		assert.NoError(t, err)
		assert.NotNil(t, actualItem)
		assert.Contains(t, actualItem, "id")
		assert.Contains(t, actualItem, "name")
		assert.Contains(t, actualItem, "email")
		assert.Equal(t, &types.AttributeValueMemberS{Value: "theUserID"}, actualItem["id"])
		assert.Equal(t, &types.AttributeValueMemberS{Value: "theUserName"}, actualItem["name"])
		assert.Equal(t, &types.AttributeValueMemberS{Value: "theUserEmail"}, actualItem["email"])
	})

	t.Run("returns_an_error_when_put_item_returns_an_error", func(t *testing.T) {
		fakeDB := &FakeSDKDynamoDB{
			PutItemFake: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
				return nil, errors.New("the fake error")
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		item := TestUser{ID: "aUserID", Name: "aUserName", Email: "aUserEmail"}

		err := PutItem(context.Background(), "aTable", item)

		assert.EqualError(t, err, "the fake error")
	})

	t.Run("applies_put_item_options_correctly", func(t *testing.T) {
		var actualInput *dynamodb.PutItemInput
		fakeDB := &FakeSDKDynamoDB{
			PutItemFake: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
				actualInput = params
				return &dynamodb.PutItemOutput{}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		item := TestUser{ID: "aUserID", Name: "aUserName", Email: "aUserEmail"}
		expressionValues := map[string]types.AttributeValue{
			":val": &types.AttributeValueMemberS{Value: "aValue"},
		}

		err := PutItem(context.Background(), "aTable", item,
			WithPutItemCondition("attribute_not_exists(id)"),
			WithPutItemExpressionAttributeValues(expressionValues))

		assert.NoError(t, err)
		assert.NotNil(t, actualInput)
		assert.Equal(t, "attribute_not_exists(id)", *actualInput.ConditionExpression)
		assert.Equal(t, expressionValues, actualInput.ExpressionAttributeValues)
	})

	t.Run("succeeds_when_no_errors", func(t *testing.T) {
		fakeDB := &FakeSDKDynamoDB{
			PutItemFake: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
				return &dynamodb.PutItemOutput{}, nil
			},
		}
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFakeSDK(nil) })

		item := TestUser{ID: "aUserID", Name: "aUserName", Email: "aUserEmail"}

		err := PutItem(context.Background(), "aTable", item)

		assert.NoError(t, err)
	})
}

func TestWithPutItemCondition(t *testing.T) {
	t.Run("sets_condition_expression_when_given_string", func(t *testing.T) {
		input := &dynamodb.PutItemInput{}
		option := WithPutItemCondition("attribute_not_exists(id)")

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.ConditionExpression)
		assert.Equal(t, "attribute_not_exists(id)", *input.ConditionExpression)
	})

	t.Run("sets_condition_expression_when_given_complex_condition", func(t *testing.T) {
		input := &dynamodb.PutItemInput{}
		option := WithPutItemCondition("attribute_not_exists(id) AND #name = :name")

		err := option(input)

		assert.NoError(t, err)
		assert.NotNil(t, input.ConditionExpression)
		assert.Equal(t, "attribute_not_exists(id) AND #name = :name", *input.ConditionExpression)
	})
}

func TestWithPutItemExpressionAttributeValues(t *testing.T) {
	t.Run("sets_expression_attribute_values_when_given_map", func(t *testing.T) {
		input := &dynamodb.PutItemInput{}
		values := map[string]types.AttributeValue{
			":name": &types.AttributeValueMemberS{Value: "aName"},
			":age":  &types.AttributeValueMemberN{Value: "25"},
		}
		option := WithPutItemExpressionAttributeValues(values)

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, values, input.ExpressionAttributeValues)
		assert.Contains(t, input.ExpressionAttributeValues, ":name")
		assert.Contains(t, input.ExpressionAttributeValues, ":age")
	})

	t.Run("sets_empty_map_when_given_empty_map", func(t *testing.T) {
		input := &dynamodb.PutItemInput{}
		values := map[string]types.AttributeValue{}
		option := WithPutItemExpressionAttributeValues(values)

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, values, input.ExpressionAttributeValues)
		assert.Len(t, input.ExpressionAttributeValues, 0)
	})
}

func TestWithPutItemTableNameSuffix(t *testing.T) {
	t.Run("appends_suffix_to_table_name", func(t *testing.T) {
		input := &dynamodb.PutItemInput{
			TableName: aws.String("theTableName"),
		}
		option := WithPutItemTableNameSuffix("theSuffix")

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, "theTableNametheSuffix", *input.TableName)
	})

	t.Run("appends_suffix_to_table_name_with_existing_suffix", func(t *testing.T) {
		input := &dynamodb.PutItemInput{
			TableName: aws.String("theTableName-existingSuffix"),
		}
		option := WithPutItemTableNameSuffix("newSuffix")

		err := option(input)

		assert.NoError(t, err)
		assert.Equal(t, "theTableName-existingSuffixnewSuffix", *input.TableName)
	})
}
