package dynamodbkit

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/half-ogre/go-kit/kit"
)

func ListTables(ctx context.Context, options ...ListTablesOption) (*ListTablesOutput, error) {
	if ctx == nil {
		return nil, kit.WrapError(nil, "context cannot be nil")
	}

	db, err := newDynamoDB(ctx)
	if err != nil {
		return nil, kit.WrapError(err, "error creating DynamoDB client")
	}

	listTablesInput := &dynamodb.ListTablesInput{}

	for _, option := range options {
		err := option(listTablesInput)
		if err != nil {
			return nil, kit.WrapError(err, "error processing option")
		}
	}

	output, err := db.ListTables(ctx, listTablesInput)
	if err != nil {
		return nil, kit.WrapError(err, "error listing tables")
	}

	result := &ListTablesOutput{
		TableNames: output.TableNames,
	}

	if output.LastEvaluatedTableName != nil {
		result.LastEvaluatedTableName = output.LastEvaluatedTableName
	}

	return result, nil
}

type ListTablesOutput struct {
	LastEvaluatedTableName *string
	TableNames             []string
}

type ListTablesOption func(*dynamodb.ListTablesInput) error

func WithListTablesLimit(limit int32) ListTablesOption {
	return func(input *dynamodb.ListTablesInput) error {
		if limit < 0 {
			return kit.WrapError(nil, "limit must be non-negative, got %d", limit)
		}
		input.Limit = aws.Int32(limit)
		return nil
	}
}

func WithListTablesExclusiveStartTableName(tableName string) ListTablesOption {
	return func(input *dynamodb.ListTablesInput) error {
		if tableName == "" {
			return kit.WrapError(nil, "exclusive start table name cannot be empty")
		}
		input.ExclusiveStartTableName = aws.String(tableName)
		return nil
	}
}
