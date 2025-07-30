package dynamodbkit

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/half-ogre/go-kit/kit"
)

type QueryOption func(*dynamodb.QueryInput) error

func WithQueryProjectionExpression(projectionExpression string) QueryOption {
	return func(input *dynamodb.QueryInput) error {
		input.ProjectionExpression = aws.String(projectionExpression)
		return nil
	}
}

func WithQueryExclusiveStartKey(exclusiveStartKey string) QueryOption {
	return func(input *dynamodb.QueryInput) error {
		decodedJson, err := base64.StdEncoding.DecodeString(exclusiveStartKey)
		if err != nil {
			return kit.WrapError(err, "failed to decode exclusiveStartKey %s", exclusiveStartKey)
		}

		var v interface{}
		err = json.Unmarshal(decodedJson, &v)
		if err != nil {
			return kit.WrapError(err, "failed to unmarshal exclusiveStartKey JSON %s", decodedJson)
		}

		k, err := attributevalue.MarshalMap(v)
		if err != nil {
			return kit.WrapError(err, "failed to unmarshal exclusiveStartKey JSON %s", decodedJson)
		}

		input.ExclusiveStartKey = k
		return nil
	}
}

func WithQueryLimit(limit int64) QueryOption {
	return func(input *dynamodb.QueryInput) error {
		if limit < 0 {
			return kit.WrapError(nil, "limit must be non-negative, got %d", limit)
		}
		if limit > 2147483647 { // int32 max
			return kit.WrapError(nil, "limit exceeds maximum allowed value, got %d", limit)
		}
		input.Limit = aws.Int32(int32(limit))
		return nil
	}
}

func Query[TItem any, TPartitionKey string | int](ctx context.Context, tableName string, partitionKey string, partitionKeyValue TPartitionKey, options ...QueryOption) (*QueryOutput[TItem], error) {
	if ctx == nil {
		return nil, kit.WrapError(nil, "context cannot be nil")
	}

	if tableName == "" {
		return nil, kit.WrapError(nil, "table name cannot be empty")
	}

	if partitionKey == "" {
		return nil, kit.WrapError(nil, "partition key cannot be empty")
	}

	db, err := newDynamoDB(ctx)
	if err != nil {
		return nil, kit.WrapError(err, "error creating DynamoDB client")
	}

	keyConditionExpr := expression.Key(partitionKey).Equal(expression.Value(partitionKeyValue))
	expr, err := expression.NewBuilder().
		WithKeyCondition(keyConditionExpr).
		Build()

	if err != nil {
		return nil, kit.WrapError(err, "error building expression")
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	for _, option := range options {
		err = option(queryInput)
		if err != nil {
			return nil, kit.WrapError(err, "error processing option")
		}
	}

	output, err := db.Query(ctx, queryInput)
	if err != nil {
		return nil, kit.WrapError(err, "error querying table %s", *queryInput.TableName)
	}

	result := &QueryOutput[TItem]{
		Items: make([]TItem, 0),
	}

	for _, i := range output.Items {
		var item TItem

		err = attributevalue.UnmarshalMap(i, &item)
		if err != nil {
			return nil, kit.WrapError(err, "error unmarshalling queried item")
		}

		result.Items = append(result.Items, item)
	}

	if output.LastEvaluatedKey != nil {
		var lastEvaluatedKey any
		err := attributevalue.UnmarshalMap(output.LastEvaluatedKey, &lastEvaluatedKey)
		if err != nil {
			return nil, kit.WrapError(err, "failed to unmarshal LastEvaluatedKey map %v", output.LastEvaluatedKey)
		}

		jsonBytes, err := json.Marshal(lastEvaluatedKey)
		if err != nil {
			return nil, kit.WrapError(err, "failed to marshal LastEvaluatedKey %v to JSON", output.LastEvaluatedKey)
		}

		encodedJson := base64.StdEncoding.EncodeToString(jsonBytes)

		result.LastEvaluatedKey = &encodedJson
	}

	return result, nil
}

type QueryOutput[TItem any] struct {
	LastEvaluatedKey *string
	Items            []TItem
}
