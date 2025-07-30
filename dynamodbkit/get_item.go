package dynamodbkit

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/half-ogre/go-kit/kit"
)

func GetItem[TItem any, TPartitionKey string | int](ctx context.Context, tableName string, partitionKey string, partitionKeyValue TPartitionKey, options ...GetItemInputOption) (*TItem, error) {
	db, err := newDynamoDB(ctx)
	if err != nil {
		return nil, kit.WrapError(err, "error creating DynamoDB client")
	}

	partitionKeyAttributeValue, err := getKeyAttributeValue(partitionKeyValue)
	if err != nil {
		return nil, err
	}

	getItemInput := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			partitionKey: partitionKeyAttributeValue,
		},
	}

	for _, option := range options {
		err := option(getItemInput)
		if err != nil {
			return nil, kit.WrapError(err, "error processing option")
		}
	}

	output, err := db.GetItem(ctx, getItemInput)
	if err != nil {
		return nil, kit.WrapError(err, "error getting item %s=%v from table %v", partitionKey, partitionKeyValue, *getItemInput.TableName)
	}

	if output.Item == nil {
		return nil, nil
	}

	var item TItem
	err = attributevalue.UnmarshalMap(output.Item, &item)
	if err != nil {
		return nil, kit.WrapError(err, "failed to unmarshal item")
	}

	return &item, nil
}

type GetItemInputOption func(*dynamodb.GetItemInput) error

func WithGetItemSortKey[TSortKey string | int](sortKey string, sortKeyValue TSortKey) GetItemInputOption {
	return func(input *dynamodb.GetItemInput) error {
		sortKeyAttributeValue, err := getKeyAttributeValue(sortKeyValue)
		if err != nil {
			return err
		}

		input.Key[sortKey] = sortKeyAttributeValue

		return nil
	}
}

func WithGetItemTableNameSuffix(suffix string) GetItemInputOption {
	return func(input *dynamodb.GetItemInput) error {
		input.TableName = aws.String(fmt.Sprintf("%s-%s", *input.TableName, suffix))
		return nil
	}
}
