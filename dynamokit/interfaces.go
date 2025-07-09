package dynamokit

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// DynamoDBAPI defines the interface for DynamoDB operations
type DynamoDBAPI interface {
	QueryWithContext(ctx aws.Context, input *dynamodb.QueryInput, opts ...request.Option) (*dynamodb.QueryOutput, error)
	PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error)
	ScanWithContext(ctx aws.Context, input *dynamodb.ScanInput, opts ...request.Option) (*dynamodb.ScanOutput, error)
}

// Client wraps the DynamoDB operations with dependency injection
type Client struct {
	db DynamoDBAPI
}

// ClientOption defines a function type for configuring the Client
type ClientOption func(*Client)

// WithDB sets a custom DynamoDB implementation for the client
func WithDB(db DynamoDBAPI) ClientOption {
	return func(c *Client) {
		c.db = db
	}
}

// NewClient creates a new DynamoDB client with optional configurations
func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		db: newDynamoDB(), // default implementation
	}
	
	// Apply options
	for _, opt := range opts {
		opt(client)
	}
	
	return client
}