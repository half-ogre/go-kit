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

func GetItem[TItem any, TPartitionKey string | int](ctx context.Context, tableName string, partitionKey string, partitionKeyValue TPartitionKey, options ...GetItemOption) (*TItem, error) {
	// Check if consumer fake is set
	fake := getFakeDynamoDB()
	if fake != nil {
		result, err := fake.GetItem(ctx, tableName, partitionKey, partitionKeyValue, options...)
		if err != nil {
			return nil, err
		}
		// Type assert the result to the expected type
		item, ok := result.(*TItem)
		if !ok {
			return nil, kit.WrapError(nil, "fake returned unexpected type for GetItem")
		}
		return item, nil
	}

	db, err := newDynamoDBSDK(ctx)
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

	originalTableNamePtr := getItemInput.TableName

	for _, option := range options {
		err := option(getItemInput)
		if err != nil {
			return nil, kit.WrapError(err, "error processing option")
		}
	}

	// Apply global table name suffix if table name pointer wasn't changed by options
	if getItemInput.TableName == originalTableNamePtr {
		globalSuffix := getTableNameSuffix()
		if globalSuffix != "" {
			getItemInput.TableName = aws.String(fmt.Sprintf("%s%s", *getItemInput.TableName, globalSuffix))
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

type GetItemOption func(*dynamodb.GetItemInput) error

func WithGetItemSortKey[TSortKey string | int](sortKey string, sortKeyValue TSortKey) GetItemOption {
	return func(input *dynamodb.GetItemInput) error {
		sortKeyAttributeValue, err := getKeyAttributeValue(sortKeyValue)
		if err != nil {
			return err
		}

		input.Key[sortKey] = sortKeyAttributeValue

		return nil
	}
}

func WithGetItemTableNameSuffix(suffix string) GetItemOption {
	return func(input *dynamodb.GetItemInput) error {
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
