package dynamodbkit

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/half-ogre/go-kit/kit"
)

func Scan[TItem any](ctx context.Context, tableName string, options ...ScanOption) (*ScanOutput[TItem], error) {
	if ctx == nil {
		return nil, kit.WrapError(nil, "context cannot be nil")
	}

	if tableName == "" {
		return nil, kit.WrapError(nil, "table name cannot be empty")
	}

	db, err := newDynamoDB(ctx)
	if err != nil {
		return nil, kit.WrapError(err, "error creating DynamoDB client")
	}

	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	for _, option := range options {
		err := option(scanInput)
		if err != nil {
			return nil, kit.WrapError(err, "error processing option")
		}
	}

	output, err := db.Scan(ctx, scanInput)
	if err != nil {
		return nil, kit.WrapError(err, "error scanning table %s", *scanInput.TableName)
	}

	result := &ScanOutput[TItem]{
		Items: make([]TItem, 0),
	}

	for _, i := range output.Items {
		var item TItem

		err = attributevalue.UnmarshalMap(i, &item)
		if err != nil {
			return nil, kit.WrapError(err, "error unmarshalling scanned item")
		}

		result.Items = append(result.Items, item)
	}

	if output.LastEvaluatedKey != nil {
		var lastEvaluatedKey any
		err := attributevalue.UnmarshalMap(output.LastEvaluatedKey, &lastEvaluatedKey)
		if err != nil {
			return nil, kit.WrapError(err, "failed to unmarshal LastEvaluatedKey map %v", output.LastEvaluatedKey)
		}

		jsonBytes, err := json.Marshal(lastEvaluatedKey)
		if err != nil {
			return nil, kit.WrapError(err, "failed to marshal LastEvaluatedKey %v to JSON", output.LastEvaluatedKey)
		}

		encodedJson := base64.StdEncoding.EncodeToString(jsonBytes)

		result.LastEvaluatedKey = &encodedJson
	}

	return result, nil
}

func WithScanExclusiveStartKey(exclusiveStartKey string) ScanOption {
	return func(input *dynamodb.ScanInput) error {
		decodedJson, err := base64.StdEncoding.DecodeString(exclusiveStartKey)
		if err != nil {
			return kit.WrapError(err, "failed to decode exclusiveStartKey %s", exclusiveStartKey)
		}

		var v interface{}
		err = json.Unmarshal(decodedJson, &v)
		if err != nil {
			return kit.WrapError(err, "failed to unmarshal exclusiveStartKey JSON %s", decodedJson)
		}

		k, err := attributevalue.MarshalMap(v)
		if err != nil {
			return kit.WrapError(err, "failed to unmarshal exclusiveStartKey JSON %s", decodedJson)
		}

		input.ExclusiveStartKey = k
		return nil
	}
}

func WithScanLimit(limit int64) ScanOption {
	return func(input *dynamodb.ScanInput) error {
		if limit < 0 {
			return kit.WrapError(nil, "limit must be non-negative, got %d", limit)
		}
		if limit > 2147483647 { // int32 max
			return kit.WrapError(nil, "limit exceeds maximum allowed value, got %d", limit)
		}
		input.Limit = aws.Int32(int32(limit))
		return nil
	}
}

type ScanOption func(*dynamodb.ScanInput) error

type ScanOutput[TItem any] struct {
	LastEvaluatedKey *string
	Items            []TItem
}
