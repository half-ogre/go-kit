package dynamodbkit

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/half-ogre/go-kit/kit"
)

type DeleteItemInputOption func(*dynamodb.DeleteItemInput) error

func WithDeleteItemSortKey[TSortKey string | int](sortKey string, sortKeyValue TSortKey) DeleteItemInputOption {
	return func(input *dynamodb.DeleteItemInput) error {
		sortKeyAttributeValue, err := getKeyAttributeValue(sortKeyValue)
		if err != nil {
			return err
		}

		input.Key[sortKey] = sortKeyAttributeValue

		return nil
	}
}

func WithDeleteItemReturnValues(returnValues types.ReturnValue) DeleteItemInputOption {
	return func(input *dynamodb.DeleteItemInput) error {
		input.ReturnValues = returnValues
		return nil
	}
}

func DeleteItem[TPartitionKey string | int](ctx context.Context, tableName string, partitionKey string, partitionKeyValue TPartitionKey, options ...DeleteItemInputOption) error {
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

	for _, option := range options {
		err := option(deleteItemInput)
		if err != nil {
			return kit.WrapError(err, "error processing option")
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
