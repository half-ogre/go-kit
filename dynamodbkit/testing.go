package dynamodbkit

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func mustMarshalMap(t *testing.T, v any) map[string]types.AttributeValue {
	m, err := attributevalue.MarshalMap(v)
	if err != nil {
		t.Logf("failed to marhsal %v to map", v)
		t.FailNow()
	}
	return m
}

type FakeSDKDynamoDB struct {
	DeleteItemFake func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	GetItemFake    func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	ListTablesFake func(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
	PutItemFake    func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	QueryFake      func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	ScanFake       func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func (f *FakeSDKDynamoDB) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	if f.DeleteItemFake != nil {
		return f.DeleteItemFake(ctx, params, optFns...)
	} else {
		panic("DeleteItem fake not implemented")
	}
}

func (f *FakeSDKDynamoDB) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	if f.GetItemFake != nil {
		return f.GetItemFake(ctx, params, optFns...)
	} else {
		panic("GetItem fake not implemented")
	}
}

func (f *FakeSDKDynamoDB) ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
	if f.ListTablesFake != nil {
		return f.ListTablesFake(ctx, params, optFns...)
	} else {
		panic("ListTables fake not implemented")
	}
}

func (f *FakeSDKDynamoDB) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if f.PutItemFake != nil {
		return f.PutItemFake(ctx, params, optFns...)
	} else {
		panic("PutItem fake not implemented")
	}
}

func (f *FakeSDKDynamoDB) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	if f.QueryFake != nil {
		return f.QueryFake(ctx, params, optFns...)
	} else {
		panic("Query fake not implemented")
	}
}

func (f *FakeSDKDynamoDB) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	if f.ScanFake != nil {
		return f.ScanFake(ctx, params, optFns...)
	} else {
		panic("Scan fake not implemented")
	}
}

// TestUser is a common test model used across test files
type TestUser struct {
	ID    string `dynamodbav:"id"`
	Name  string `dynamodbav:"name"`
	Email string `dynamodbav:"email"`
}

// TestUserWithSort is a test model with composite key (partition + sort key)
type TestUserWithSort struct {
	UserID    string `dynamodbav:"user_id"`
	Timestamp string `dynamodbav:"timestamp"`
	Name      string `dynamodbav:"name"`
	Data      string `dynamodbav:"data"`
}
