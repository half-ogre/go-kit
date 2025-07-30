package dynamodbkit

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/half-ogre/go-kit/kit"
)

func DeleteItem[TPartitionKey string | int](ctx context.Context, tableName string, partitionKey string, partitionKeyValue TPartitionKey, options ...DeleteItemOption) error {
	db, err := newDynamoDB(ctx)
	if err != nil {
		return kit.WrapError(err, "error creating DynamoDB client")
	}

	partitionKeyAttributeValue, err := getKeyAttributeValue(partitionKeyValue)
	if err != nil {
		return err
	}

	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			partitionKey: partitionKeyAttributeValue,
		},
	}

	originalTableNamePtr := deleteItemInput.TableName

	for _, option := range options {
		err := option(deleteItemInput)
		if err != nil {
			return kit.WrapError(err, "error processing option")
		}
	}

	// Apply global table name suffix if table name pointer wasn't changed by options
	if deleteItemInput.TableName == originalTableNamePtr {
		globalSuffix := getTableNameSuffix()
		if globalSuffix != "" {
			deleteItemInput.TableName = aws.String(fmt.Sprintf("%s-%s", *deleteItemInput.TableName, globalSuffix))
		}
	}

	slog.Debug("deleting DynamoDB item", "input", deleteItemInput)

	output, err := db.DeleteItem(ctx, deleteItemInput)
	if err != nil {
		return kit.WrapError(err, "error deleting item")
	}

	slog.Info("delete-item", "attributes", output.Attributes)

	return nil
}

type DeleteItemOption func(*dynamodb.DeleteItemInput) error

func WithDeleteItemReturnValues(returnValues types.ReturnValue) DeleteItemOption {
	return func(input *dynamodb.DeleteItemInput) error {
		input.ReturnValues = returnValues
		return nil
	}
}

func WithDeleteItemSortKey[TSortKey string | int](sortKey string, sortKeyValue TSortKey) DeleteItemOption {
	return func(input *dynamodb.DeleteItemInput) error {
		sortKeyAttributeValue, err := getKeyAttributeValue(sortKeyValue)
		if err != nil {
			return err
		}

		input.Key[sortKey] = sortKeyAttributeValue

		return nil
	}
}

func WithDeleteItemTableNameSuffix(suffix string) DeleteItemOption {
	return func(input *dynamodb.DeleteItemInput) error {
		// Always create a new string to ensure pointer comparison detects change
		if suffix == "" {
			// Create new string with same content to mark as modified
			newTableName := *input.TableName
			input.TableName = &newTableName
		} else {
			input.TableName = aws.String(fmt.Sprintf("%s-%s", *input.TableName, suffix))
		}
		return nil
	}
}
