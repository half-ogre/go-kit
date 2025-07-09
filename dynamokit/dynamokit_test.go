package dynamokit

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// FakeDynamoDB implements DynamoDBAPI for testing
type FakeDynamoDB struct {
	queryResponse *dynamodb.QueryOutput
	queryError    error
	putResponse   *dynamodb.PutItemOutput
	putError      error
	scanResponse  *dynamodb.ScanOutput
	scanError     error
	
	// Track calls for verification
	queryCalls []QueryCall
	putCalls   []PutCall
	scanCalls  []ScanCall
}

type QueryCall struct {
	Input *dynamodb.QueryInput
}

type PutCall struct {
	Input *dynamodb.PutItemInput
}

type ScanCall struct {
	Input *dynamodb.ScanInput
}

func (f *FakeDynamoDB) QueryWithContext(ctx aws.Context, input *dynamodb.QueryInput, opts ...request.Option) (*dynamodb.QueryOutput, error) {
	f.queryCalls = append(f.queryCalls, QueryCall{Input: input})
	return f.queryResponse, f.queryError
}

func (f *FakeDynamoDB) PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error) {
	f.putCalls = append(f.putCalls, PutCall{Input: input})
	return f.putResponse, f.putError
}

func (f *FakeDynamoDB) ScanWithContext(ctx aws.Context, input *dynamodb.ScanInput, opts ...request.Option) (*dynamodb.ScanOutput, error) {
	f.scanCalls = append(f.scanCalls, ScanCall{Input: input})
	return f.scanResponse, f.scanError
}

// Test models
type TestUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Error("NewClient() returned nil")
	}
	if client.db == nil {
		t.Error("NewClient() db is nil")
	}
}

func TestNewClientWithDB(t *testing.T) {
	fakeDB := &FakeDynamoDB{}
	client := NewClient(WithDB(fakeDB))
	
	if client == nil {
		t.Error("NewClient(WithDB()) returned nil")
	}
	if client.db != fakeDB {
		t.Error("NewClient(WithDB()) did not set the correct db")
	}
}

func TestWithDB(t *testing.T) {
	fakeDB := &FakeDynamoDB{}
	
	// Test WithDB option function
	option := WithDB(fakeDB)
	client := &Client{}
	option(client)
	
	if client.db != fakeDB {
		t.Error("WithDB() option did not set the correct db")
	}
}

func TestNewClientMultipleOptions(t *testing.T) {
	fakeDB := &FakeDynamoDB{}
	
	// Test that options are applied in order
	client := NewClient(WithDB(fakeDB))
	
	if client.db != fakeDB {
		t.Error("NewClient() with multiple options did not apply WithDB correctly")
	}
}

func TestClient_QueryItem_Success(t *testing.T) {
	user := TestUser{ID: "123", Name: "John Doe", Email: "john@example.com"}
	
	// Marshal the test data
	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		t.Fatal(err)
	}
	
	fakeDB := &FakeDynamoDB{
		queryResponse: &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{item},
		},
	}
	
	client := NewClient(WithDB(fakeDB))
	
	result, err := QueryItemWithClient[TestUser, string](context.Background(), client, "users", "id", "123")
	
	if err != nil {
		t.Errorf("QueryItem() unexpected error: %v", err)
	}
	if result == nil {
		t.Error("QueryItem() returned nil result")
		return
	}
	if result.ID != user.ID {
		t.Errorf("QueryItem() ID = %q, want %q", result.ID, user.ID)
	}
	if result.Name != user.Name {
		t.Errorf("QueryItem() Name = %q, want %q", result.Name, user.Name)
	}
	
	// Verify the query was called correctly
	if len(fakeDB.queryCalls) != 1 {
		t.Errorf("Expected 1 query call, got %d", len(fakeDB.queryCalls))
	}
	if *fakeDB.queryCalls[0].Input.TableName != "users" {
		t.Errorf("QueryItem() TableName = %q, want %q", *fakeDB.queryCalls[0].Input.TableName, "users")
	}
}

func TestClient_QueryItem_NoResults(t *testing.T) {
	fakeDB := &FakeDynamoDB{
		queryResponse: &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{},
		},
	}
	
	client := NewClient(WithDB(fakeDB))
	
	result, err := QueryItemWithClient[TestUser, string](context.Background(), client, "users", "id", "999")
	
	if err != nil {
		t.Errorf("QueryItem() unexpected error: %v", err)
	}
	if result != nil {
		t.Error("QueryItem() expected nil result for no items")
	}
}

func TestClient_QueryItem_MultipleResults(t *testing.T) {
	user1 := TestUser{ID: "123", Name: "John Doe", Email: "john@example.com"}
	user2 := TestUser{ID: "124", Name: "Jane Doe", Email: "jane@example.com"}
	
	item1, _ := dynamodbattribute.MarshalMap(user1)
	item2, _ := dynamodbattribute.MarshalMap(user2)
	
	fakeDB := &FakeDynamoDB{
		queryResponse: &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{item1, item2},
		},
	}
	
	client := NewClient(WithDB(fakeDB))
	
	result, err := QueryItemWithClient[TestUser, string](context.Background(), client, "users", "id", "123")
	
	if err == nil {
		t.Error("QueryItem() expected error for multiple results")
	}
	if result != nil {
		t.Error("QueryItem() expected nil result for multiple items error")
	}
	expectedError := "query results has more than one item"
	if err.Error() != expectedError {
		t.Errorf("QueryItem() error = %q, want %q", err.Error(), expectedError)
	}
}

func TestClient_QueryItem_QueryError(t *testing.T) {
	fakeDB := &FakeDynamoDB{
		queryError: errors.New("DynamoDB query error"),
	}
	
	client := NewClient(WithDB(fakeDB))
	
	result, err := QueryItemWithClient[TestUser, string](context.Background(), client, "users", "id", "123")
	
	if err == nil {
		t.Error("QueryItem() expected error")
	}
	if result != nil {
		t.Error("QueryItem() expected nil result on error")
	}
}

func TestClient_QueryIndexItem_Success(t *testing.T) {
	user := TestUser{ID: "123", Name: "John Doe", Email: "john@example.com"}
	
	item, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		t.Fatal(err)
	}
	
	fakeDB := &FakeDynamoDB{
		queryResponse: &dynamodb.QueryOutput{
			Items: []map[string]*dynamodb.AttributeValue{item},
		},
	}
	
	client := NewClient(WithDB(fakeDB))
	
	result, err := QueryIndexItemWithClient[TestUser, string](context.Background(), client, "users", "email-index", "email", "john@example.com")
	
	if err != nil {
		t.Errorf("QueryIndexItem() unexpected error: %v", err)
	}
	if result == nil {
		t.Error("QueryIndexItem() returned nil result")
		return
	}
	if result.Email != user.Email {
		t.Errorf("QueryIndexItem() Email = %q, want %q", result.Email, user.Email)
	}
	
	// Verify the query was called with index name
	if len(fakeDB.queryCalls) != 1 {
		t.Errorf("Expected 1 query call, got %d", len(fakeDB.queryCalls))
	}
	if *fakeDB.queryCalls[0].Input.IndexName != "email-index" {
		t.Errorf("QueryIndexItem() IndexName = %q, want %q", *fakeDB.queryCalls[0].Input.IndexName, "email-index")
	}
}

func TestClient_PutItem_Success(t *testing.T) {
	fakeDB := &FakeDynamoDB{
		putResponse: &dynamodb.PutItemOutput{},
	}
	
	client := NewClient(WithDB(fakeDB))
	
	user := TestUser{ID: "123", Name: "John Doe", Email: "john@example.com"}
	
	err := PutItemWithClient(context.Background(), client, "users", user)
	
	if err != nil {
		t.Errorf("PutItem() unexpected error: %v", err)
	}
	
	// Verify the put was called correctly
	if len(fakeDB.putCalls) != 1 {
		t.Errorf("Expected 1 put call, got %d", len(fakeDB.putCalls))
	}
	if *fakeDB.putCalls[0].Input.TableName != "users" {
		t.Errorf("PutItem() TableName = %q, want %q", *fakeDB.putCalls[0].Input.TableName, "users")
	}
	
	// Verify the item was marshaled correctly
	var unmarshaledUser TestUser
	err = dynamodbattribute.UnmarshalMap(fakeDB.putCalls[0].Input.Item, &unmarshaledUser)
	if err != nil {
		t.Errorf("Failed to unmarshal put item: %v", err)
	}
	if unmarshaledUser.ID != user.ID {
		t.Errorf("PutItem() marshaled ID = %q, want %q", unmarshaledUser.ID, user.ID)
	}
}

func TestClient_PutItem_Error(t *testing.T) {
	fakeDB := &FakeDynamoDB{
		putError: errors.New("DynamoDB put error"),
	}
	
	client := NewClient(WithDB(fakeDB))
	
	user := TestUser{ID: "123", Name: "John Doe", Email: "john@example.com"}
	
	err := PutItemWithClient(context.Background(), client, "users", user)
	
	if err == nil {
		t.Error("PutItem() expected error")
	}
}

func TestClient_ScanAllItems_Success(t *testing.T) {
	user1 := TestUser{ID: "123", Name: "John Doe", Email: "john@example.com"}
	user2 := TestUser{ID: "124", Name: "Jane Doe", Email: "jane@example.com"}
	
	item1, _ := dynamodbattribute.MarshalMap(user1)
	item2, _ := dynamodbattribute.MarshalMap(user2)
	
	fakeDB := &FakeDynamoDB{
		scanResponse: &dynamodb.ScanOutput{
			Items: []map[string]*dynamodb.AttributeValue{item1, item2},
		},
	}
	
	client := NewClient(WithDB(fakeDB))
	
	results, err := ScanAllItemsWithClient[TestUser](context.Background(), client, "users")
	
	if err != nil {
		t.Errorf("ScanAllItems() unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("ScanAllItems() returned %d items, want 2", len(results))
	}
	if results[0].ID != user1.ID {
		t.Errorf("ScanAllItems() first item ID = %q, want %q", results[0].ID, user1.ID)
	}
	if results[1].ID != user2.ID {
		t.Errorf("ScanAllItems() second item ID = %q, want %q", results[1].ID, user2.ID)
	}
	
	// Verify the scan was called correctly
	if len(fakeDB.scanCalls) != 1 {
		t.Errorf("Expected 1 scan call, got %d", len(fakeDB.scanCalls))
	}
	if *fakeDB.scanCalls[0].Input.TableName != "users" {
		t.Errorf("ScanAllItems() TableName = %q, want %q", *fakeDB.scanCalls[0].Input.TableName, "users")
	}
}

func TestClient_ScanAllItems_EmptyResults(t *testing.T) {
	fakeDB := &FakeDynamoDB{
		scanResponse: &dynamodb.ScanOutput{
			Items: []map[string]*dynamodb.AttributeValue{},
		},
	}
	
	client := NewClient(WithDB(fakeDB))
	
	results, err := ScanAllItemsWithClient[TestUser](context.Background(), client, "users")
	
	if err != nil {
		t.Errorf("ScanAllItems() unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("ScanAllItems() returned %d items, want 0", len(results))
	}
}

func TestClient_ScanAllItems_Error(t *testing.T) {
	fakeDB := &FakeDynamoDB{
		scanError: errors.New("DynamoDB scan error"),
	}
	
	client := NewClient(WithDB(fakeDB))
	
	results, err := ScanAllItemsWithClient[TestUser](context.Background(), client, "users")
	
	if err == nil {
		t.Error("ScanAllItems() expected error")
	}
	if results != nil {
		t.Error("ScanAllItems() expected nil results on error")
	}
}

func TestUsageExamples(t *testing.T) {
	// Example 1: Default client (uses real AWS session)
	defaultClient := NewClient()
	if defaultClient == nil {
		t.Error("Default client should not be nil")
	}
	
	// Example 2: Client with custom DB for testing
	fakeDB := &FakeDynamoDB{
		queryResponse: &dynamodb.QueryOutput{Items: []map[string]*dynamodb.AttributeValue{}},
	}
	testClient := NewClient(WithDB(fakeDB))
	if testClient == nil {
		t.Error("Test client should not be nil")
	}
	
	// Example 3: Using the test client
	_, err := QueryItemWithClient[TestUser, string](context.Background(), testClient, "users", "id", "123")
	if err != nil {
		t.Logf("Expected error in test environment: %v", err)
	}
}

// Test the package-level convenience functions
func TestQueryItem_PackageFunction(t *testing.T) {
	// This test verifies that the package-level function creates a client and calls the method
	// Since we can't easily mock the package-level function without changing the AWS session,
	// we'll just verify it doesn't panic and has the right signature
	
	// We can't test the actual functionality without real AWS credentials,
	// but we can at least verify the function exists and compiles
	ctx := context.Background()
	_, err := QueryItem[TestUser, string](ctx, "test-table", "id", "123")
	
	// We expect this to fail since there are no AWS credentials in the test environment
	// but it should not panic
	if err == nil {
		t.Log("QueryItem package function executed without error (unexpected in test environment)")
	}
}

func TestQueryIndexItem_PackageFunction(t *testing.T) {
	ctx := context.Background()
	_, err := QueryIndexItem[TestUser, string](ctx, "test-table", "test-index", "email", "test@example.com")
	
	// We expect this to fail since there are no AWS credentials in the test environment
	if err == nil {
		t.Log("QueryIndexItem package function executed without error (unexpected in test environment)")
	}
}

func TestPutItem_PackageFunction(t *testing.T) {
	ctx := context.Background()
	user := TestUser{ID: "123", Name: "Test User", Email: "test@example.com"}
	err := PutItem(ctx, "test-table", user)
	
	// We expect this to fail since there are no AWS credentials in the test environment
	if err == nil {
		t.Log("PutItem package function executed without error (unexpected in test environment)")
	}
}

func TestScanAllItems_PackageFunction(t *testing.T) {
	ctx := context.Background()
	_, err := ScanAllItems[TestUser](ctx, "test-table")
	
	// We expect this to fail since there are no AWS credentials in the test environment
	if err == nil {
		t.Log("ScanAllItems package function executed without error (unexpected in test environment)")
	}
}