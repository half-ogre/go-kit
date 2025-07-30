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

func PutItem[T any](ctx context.Context, tableName string, item T, options ...PutItemOption) error {
	i, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	putItemInput := &dynamodb.PutItemInput{
		Item:      i,
		TableName: aws.String(tableName),
	}

	originalTableNamePtr := putItemInput.TableName

	for _, option := range options {
		err = option(putItemInput)
		if err != nil {
			return kit.WrapError(err, "error processing option")
		}
	}

	// Apply global table name suffix if table name pointer wasn't changed by options
	if putItemInput.TableName == originalTableNamePtr {
		globalSuffix := getTableNameSuffix()
		if globalSuffix != "" {
			putItemInput.TableName = aws.String(fmt.Sprintf("%s%s", *putItemInput.TableName, globalSuffix))
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

type PutItemOption func(*dynamodb.PutItemInput) error

func WithPutItemCondition(conditionExpression string) PutItemOption {
	return func(input *dynamodb.PutItemInput) error {
		input.ConditionExpression = aws.String(conditionExpression)
		return nil
	}
}

func WithPutItemExpressionAttributeValues(expressionAttributeValues map[string]types.AttributeValue) PutItemOption {
	return func(input *dynamodb.PutItemInput) error {
		input.ExpressionAttributeValues = expressionAttributeValues
		return nil
	}
}

func WithPutItemTableNameSuffix(suffix string) PutItemOption {
	return func(input *dynamodb.PutItemInput) error {
		// Always create a new string to ensure pointer comparison detects change
		if suffix == "" {
			// Create new string with same content to mark as modified
			newTableName := *input.TableName
			input.TableName = &newTableName
		} else {
			input.TableName = aws.String(fmt.Sprintf("%s%s", *input.TableName, suffix))
		}
		return nil
	}
}
