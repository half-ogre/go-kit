package dynamodbkit

import (
	"context"

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

func Query[TItem any, TPartitionKey string | int](ctx context.Context, tableName string, partitionKey string, partitionKeyValue TPartitionKey, options ...QueryOption) ([]TItem, error) {
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
		return nil, kit.WrapError(err, "error querying table %v", *queryInput.TableName)
	}

	items := make([]TItem, 0)

	for _, i := range output.Items {
		var item TItem

		err = attributevalue.UnmarshalMap(i, &item)
		if err != nil {
			return nil, kit.WrapError(err, "error unmarshalling queried item")
		}

		items = append(items, item)
	}

	return items, nil
}
