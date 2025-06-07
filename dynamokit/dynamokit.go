package dynamokit

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/half-ogre/go-kit/kit"
)

func QueryItem[TItem any, TPartitionKey string | int](ctx context.Context, tableName string, partitionKey string, partitionKeyValue TPartitionKey) (*TItem, error) {
	db := newDynamoDB()

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

	output, err := db.QueryWithContext(ctx, queryInput)
	if err != nil {
		return nil, kit.WrapError(err, "error querying table %v", queryInput.TableName)
	}

	items := make([]TItem, 0)

	for _, i := range output.Items {
		var item TItem

		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return nil, kit.WrapError(err, "error unmarshalling queried item")
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, nil
	} else if len(items) > 1 {
		return nil, errors.New("query results has more than one item")
	} else {
		return &items[0], nil
	}
}

func QueryIndexItem[TItem any, TPartitionKey string | int](ctx context.Context, tableName string, indexName string, partitionKey string, partitionKeyValue TPartitionKey) (*TItem, error) {
	db := newDynamoDB()

	keyConditionExpr := expression.Key(partitionKey).Equal(expression.Value(partitionKeyValue))
	expr, err := expression.NewBuilder().
		WithKeyCondition(keyConditionExpr).
		Build()

	if err != nil {
		return nil, kit.WrapError(err, "error building expression")
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String(indexName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	output, err := db.QueryWithContext(ctx, queryInput)
	if err != nil {
		return nil, kit.WrapError(err, "error querying table %v", queryInput.TableName)
	}

	items := make([]TItem, 0)

	for _, i := range output.Items {
		var item TItem

		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return nil, kit.WrapError(err, "error unmarshalling queried item")
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, nil
	} else if len(items) > 1 {
		return nil, errors.New("query results has more than one item")
	} else {
		return &items[0], nil
	}
}

func PutItem[T any](ctx context.Context, tableName string, item T) error {
	i, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	putInput := &dynamodb.PutItemInput{
		Item:      i,
		TableName: aws.String(tableName),
	}

	db := newDynamoDB()

	_, err = db.PutItemWithContext(ctx, putInput)
	if err != nil {
		return err
	}

	return nil
}

func ScanAllItems[TItem any](ctx context.Context, tableName string) ([]TItem, error) {
	db := newDynamoDB()

	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	output, err := db.ScanWithContext(ctx, scanInput)
	if err != nil {
		return nil, kit.WrapError(err, "error scanning table %s", *scanInput.TableName)
	}

	items := make([]TItem, 0)

	for _, i := range output.Items {
		var item TItem

		err = dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			return nil, kit.WrapError(err, "error unmarshalling queried item")
		}

		items = append(items, item)
	}

	return items, nil
}

// func QueryItems[TItem any, TPartitionKey string | int](ctx context.Context, tableName string, partitionKey string, partitionKeyValues []TPartitionKey) (*TItem, error) {
// 	db := newDynamoDB()

// 	var keys []map[string]*dynamodb.AttributeValue
// 	for _, v := range partitionKeyValues {
// 		//keyExpr := expression.Key(partitionKey).Equal(expression.Value(v))
// 		e := expression.Name(partitionKey).Equal(expression.Value(v))
// 		expr, err := expression.NewBuilder().
// 			WithCondition(e).
// 			// WithKeyCondition(keyExpr).
// 			Build()

// 		if err != nil {
// 			return nil, kit.WrapError(err, "error building expression")
// 		}

// 		keys = append(keys, expr.Values())
// 	}

// 	input := &dynamodb.BatchGetItemInput{
// 		RequestItems: map[string]*dynamodb.KeysAndAttributes{
// 			tableName: {
// 				Keys: keys,
// 			},
// 		},
// 	}

// 	output, err := db.BatchGetItemWithContext(ctx, input)
// 	if err != nil {
// 		return nil, kit.WrapError(err, "error getting batch items %v", input)
// 	}

// 	items := make([]TItem, 0)

// 	for _, i := range output.Responses[tableName] {
// 		var item TItem

// 		err = dynamodbattribute.UnmarshalMap(i, &item)
// 		if err != nil {
// 			return nil, kit.WrapError(err, "error unmarshalling queried item")
// 		}

// 		items = append(items, item)
// 	}

// 	if len(items) == 0 {
// 		return nil, nil
// 	} else {
// 		return &items[0], nil
// 	}
// }

func newDynamoDB() *dynamodb.DynamoDB {
	awsSession := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
		SharedConfigState: session.SharedConfigEnable,
	}))

	return dynamodb.New(awsSession)
}
