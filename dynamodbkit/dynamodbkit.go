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

// FakeDynamoDB is the interface for consumer testing - allows faking all dynamodbkit operations
type FakeDynamoDB interface {
	DeleteItem(ctx context.Context, tableName string, partitionKey string, partitionKeyValue any, options ...DeleteItemOption) error
	GetItem(ctx context.Context, tableName string, partitionKey string, partitionKeyValue any, options ...GetItemOption) (any, error)
	ListTables(ctx context.Context, options ...ListTablesOption) (*ListTablesOutput, error)
	PutItem(ctx context.Context, tableName string, item any, options ...PutItemOption) error
	Query(ctx context.Context, tableName string, partitionKey string, partitionKeyValue any, options ...QueryOption) (any, error)
	Scan(ctx context.Context, tableName string, options ...ScanOption) (any, error)
}

// SetFake sets a fake implementation for all dynamodbkit operations (for consumer testing)
func SetFake(fake FakeDynamoDB) {
	fakeDynamoDBMu.Lock()
	defer fakeDynamoDBMu.Unlock()
	fakeDynamoDB = fake
}

var fakeDynamoDB FakeDynamoDB
var fakeDynamoDBMu sync.Mutex

func getFakeDynamoDB() FakeDynamoDB {
	fakeDynamoDBMu.Lock()
	defer fakeDynamoDBMu.Unlock()
	return fakeDynamoDB
}

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
	ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
}

func newDynamoDBSDK(ctx context.Context) (DynamoDB, error) {
	fakeSDKMu.Lock()
	defer fakeSDKMu.Unlock()
	if fakeNewDynamoDBSDK != nil {
		return fakeNewDynamoDBSDK(ctx)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, kit.WrapError(err, "error loading default AWS config")
	}

	return dynamodb.NewFromConfig(cfg), nil
}

var fakeNewDynamoDBSDK func(ctx context.Context) (DynamoDB, error)
var fakeSDKMu sync.Mutex

func setFakeSDK(fake func(ctx context.Context) (DynamoDB, error)) {
	fakeSDKMu.Lock()
	defer fakeSDKMu.Unlock()
	fakeNewDynamoDBSDK = fake
}

var tableNameSuffix string
var tableNameSuffixMu sync.Mutex

func getTableNameSuffix() string {
	tableNameSuffixMu.Lock()
	defer tableNameSuffixMu.Unlock()
	return tableNameSuffix
}
