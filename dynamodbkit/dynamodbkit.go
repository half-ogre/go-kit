package dynamodbkit

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/half-ogre/go-kit/kit"
)

func UseTableNameSuffix(suffix string) {
	tableNameSuffixMu.Lock()
	defer tableNameSuffixMu.Unlock()
	tableNameSuffix = suffix
}

// func QueryIndexItem[TItem any, TPartitionKey string | int](ctx context.Context, tableName string, indexName string, partitionKey string, partitionKeyValue TPartitionKey) ([]TItem, error) {
// 	db, err := newDynamoDB(ctx)
// 	if err != nil {
// 		return nil, kit.WrapError(err, "error creating DynamoDB client")
// 	}

// 	keyConditionExpr := expression.Key(partitionKey).Equal(expression.Value(partitionKeyValue))
// 	expr, err := expression.NewBuilder().
// 		WithKeyCondition(keyConditionExpr).
// 		Build()

// 	if err != nil {
// 		return nil, kit.WrapError(err, "error building expression")
// 	}

// 	queryInput := &dynamodb.QueryInput{
// 		TableName:                 aws.String(tableName),
// 		IndexName:                 aws.String(indexName),
// 		KeyConditionExpression:    expr.KeyCondition(),
// 		ExpressionAttributeNames:  expr.Names(),
// 		ExpressionAttributeValues: expr.Values(),
// 	}

// 	output, err := db.Query(ctx, queryInput)
// 	if err != nil {
// 		return nil, kit.WrapError(err, "error querying table %v", *queryInput.TableName)
// 	}

// 	items := make([]TItem, 0)

// 	for _, i := range output.Items {
// 		var item TItem

// 		err = attributevalue.UnmarshalMap(i, &item)
// 		if err != nil {
// 			return nil, kit.WrapError(err, "error unmarshalling queried item")
// 		}

// 		items = append(items, item)
// 	}

// 	if len(items) == 0 {
// 		return nil, nil
// 	} else {
// 		return items, nil
// 	}
// }

func getKeyAttributeValue[TKey string | int](keyValue TKey) (types.AttributeValue, error) {
	var keyAttributeValue types.AttributeValue
	switch t := any(keyValue).(type) {
	case int:
		keyAttributeValue = &types.AttributeValueMemberN{
			Value: fmt.Sprintf("%v", keyValue),
		}
	case string:
		keyAttributeValue = &types.AttributeValueMemberS{
			Value: fmt.Sprintf("%v", keyValue),
		}
	default:
		return nil, fmt.Errorf("impossible type %v for key value", t)
	}

	return keyAttributeValue, nil
}

type DynamoDB interface {
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
}

func newDynamoDB(ctx context.Context) (DynamoDB, error) {
	fakeMu.Lock()
	defer fakeMu.Unlock()
	if fakeNewDynamoDB != nil {
		return fakeNewDynamoDB(ctx)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, kit.WrapError(err, "error loading default AWS config")
	}

	return dynamodb.NewFromConfig(cfg), nil
}

var fakeNewDynamoDB func(ctx context.Context) (DynamoDB, error)
var fakeMu sync.Mutex

func setFake(fake func(ctx context.Context) (DynamoDB, error)) {
	fakeMu.Lock()
	defer fakeMu.Unlock()
	fakeNewDynamoDB = fake
}

var tableNameSuffix string
var tableNameSuffixMu sync.Mutex

func getTableNameSuffix() string {
	tableNameSuffixMu.Lock()
	defer tableNameSuffixMu.Unlock()
	return tableNameSuffix
}
