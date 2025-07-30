package dynamodbkit

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/half-ogre/go-kit/kit"
)

func PutItem[T any](ctx context.Context, tableName string, item T, options ...PutItemInputOption) error {
	i, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	putItemInput := &dynamodb.PutItemInput{
		Item:      i,
		TableName: aws.String(tableName),
	}

	for _, option := range options {
		err = option(putItemInput)
		if err != nil {
			return kit.WrapError(err, "error processing option")
		}
	}

	db, err := newDynamoDB(ctx)
	if err != nil {
		return kit.WrapError(err, "error creating DynamoDB client")
	}

	slog.Info("putting item into DynamoDB", "item", item, "table", tableName, "input", putItemInput)

	_, err = db.PutItem(ctx, putItemInput)
	if err != nil {
		return err
	}

	return nil
}

type PutItemInputOption func(*dynamodb.PutItemInput) error

func WithPutItemCondition(conditionExpression string) PutItemInputOption {
	return func(input *dynamodb.PutItemInput) error {
		input.ConditionExpression = aws.String(conditionExpression)
		return nil
	}
}

func WithPutItemExpressionAttributeValues(expressionAttributeValues map[string]types.AttributeValue) PutItemInputOption {
	return func(input *dynamodb.PutItemInput) error {
		input.ExpressionAttributeValues = expressionAttributeValues
		return nil
	}
}

func WithPutItemTableNameSuffix(suffix string) PutItemInputOption {
	return func(input *dynamodb.PutItemInput) error {
		input.TableName = aws.String(fmt.Sprintf("%s-%s", *input.TableName, suffix))
		return nil
	}
}
