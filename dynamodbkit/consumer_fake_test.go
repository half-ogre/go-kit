package dynamodbkit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test implementation of the consumer fake interface
type TestFakeConsumer struct {
	deleteItemResult error
	getItemResult    any
	getItemError     error
	listTablesResult *ListTablesOutput
	listTablesError  error
	putItemResult    error
	queryResult      any
	queryError       error
	scanResult       any
	scanError        error
}

func (f *TestFakeConsumer) DeleteItem(ctx context.Context, tableName string, partitionKey string, partitionKeyValue any, options ...DeleteItemOption) error {
	return f.deleteItemResult
}

func (f *TestFakeConsumer) GetItem(ctx context.Context, tableName string, partitionKey string, partitionKeyValue any, options ...GetItemOption) (any, error) {
	return f.getItemResult, f.getItemError
}

func (f *TestFakeConsumer) ListTables(ctx context.Context, options ...ListTablesOption) (*ListTablesOutput, error) {
	return f.listTablesResult, f.listTablesError
}

func (f *TestFakeConsumer) PutItem(ctx context.Context, tableName string, item any, options ...PutItemOption) error {
	return f.putItemResult
}

func (f *TestFakeConsumer) Query(ctx context.Context, tableName string, partitionKey string, partitionKeyValue any, options ...QueryOption) (any, error) {
	return f.queryResult, f.queryError
}

func (f *TestFakeConsumer) Scan(ctx context.Context, tableName string, options ...ScanOption) (any, error) {
	return f.scanResult, f.scanError
}

func TestConsumerFake(t *testing.T) {
	t.Run("delete_item_uses_consumer_fake_when_set", func(t *testing.T) {
		fake := &TestFakeConsumer{
			deleteItemResult: errors.New("the consumer fake error"),
		}
		SetFake(fake)
		t.Cleanup(func() { SetFake(nil) })

		err := DeleteItem(context.Background(), "aTable", "id", "aUserID")

		assert.EqualError(t, err, "the consumer fake error")
	})

	t.Run("get_item_uses_consumer_fake_when_set", func(t *testing.T) {
		expectedUser := &TestUser{ID: "theUserID", Name: "theUserName", Email: "theUserEmail"}
		fake := &TestFakeConsumer{
			getItemResult: expectedUser,
			getItemError:  nil,
		}
		SetFake(fake)
		t.Cleanup(func() { SetFake(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, result)
	})

	t.Run("get_item_returns_error_from_consumer_fake", func(t *testing.T) {
		fake := &TestFakeConsumer{
			getItemResult: nil,
			getItemError:  errors.New("the consumer fake error"),
		}
		SetFake(fake)
		t.Cleanup(func() { SetFake(nil) })

		result, err := GetItem[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.Nil(t, result)
		assert.EqualError(t, err, "the consumer fake error")
	})

	t.Run("list_tables_uses_consumer_fake_when_set", func(t *testing.T) {
		expectedOutput := &ListTablesOutput{
			TableNames: []string{"table1", "table2"},
		}
		fake := &TestFakeConsumer{
			listTablesResult: expectedOutput,
			listTablesError:  nil,
		}
		SetFake(fake)
		t.Cleanup(func() { SetFake(nil) })

		result, err := ListTables(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, result)
	})

	t.Run("put_item_uses_consumer_fake_when_set", func(t *testing.T) {
		fake := &TestFakeConsumer{
			putItemResult: errors.New("the consumer fake error"),
		}
		SetFake(fake)
		t.Cleanup(func() { SetFake(nil) })

		item := TestUser{ID: "aUserID", Name: "aUserName", Email: "aUserEmail"}
		err := PutItem(context.Background(), "aTable", item)

		assert.EqualError(t, err, "the consumer fake error")
	})

	t.Run("query_uses_consumer_fake_when_set", func(t *testing.T) {
		expectedOutput := &QueryOutput[TestUser]{
			Items: []TestUser{
				{ID: "theUserID", Name: "theUserName", Email: "theUserEmail"},
			},
		}
		fake := &TestFakeConsumer{
			queryResult: expectedOutput,
			queryError:  nil,
		}
		SetFake(fake)
		t.Cleanup(func() { SetFake(nil) })

		result, err := Query[TestUser](context.Background(), "aTable", "id", "aUserID")

		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, result)
	})

	t.Run("scan_uses_consumer_fake_when_set", func(t *testing.T) {
		expectedOutput := &ScanOutput[TestUser]{
			Items: []TestUser{
				{ID: "theUserID", Name: "theUserName", Email: "theUserEmail"},
			},
		}
		fake := &TestFakeConsumer{
			scanResult: expectedOutput,
			scanError:  nil,
		}
		SetFake(fake)
		t.Cleanup(func() { SetFake(nil) })

		result, err := Scan[TestUser](context.Background(), "aTable")

		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, result)
	})

	t.Run("consumer_fake_takes_precedence_over_sdk_fake", func(t *testing.T) {
		// Set up SDK fake that would return different result
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { 
			return nil, errors.New("sdk fake error") 
		})
		t.Cleanup(func() { setFakeSDK(nil) })

		// Set up consumer fake
		fake := &TestFakeConsumer{
			deleteItemResult: errors.New("consumer fake error"),
		}
		SetFake(fake)
		t.Cleanup(func() { SetFake(nil) })

		err := DeleteItem(context.Background(), "aTable", "id", "aUserID")

		// Should get the consumer fake error, not the SDK fake error
		assert.EqualError(t, err, "consumer fake error")
	})

	t.Run("clearing_consumer_fake_falls_back_to_sdk", func(t *testing.T) {
		// First set consumer fake
		fake := &TestFakeConsumer{
			deleteItemResult: errors.New("consumer fake error"),
		}
		SetFake(fake)

		// Verify consumer fake is used
		err1 := DeleteItem(context.Background(), "aTable", "id", "aUserID")
		assert.EqualError(t, err1, "consumer fake error")

		// Clear consumer fake
		SetFake(nil)

		// Set SDK fake
		setFakeSDK(func(ctx context.Context) (DynamoDB, error) { 
			return nil, errors.New("sdk fake error") 
		})
		t.Cleanup(func() { setFakeSDK(nil) })

		// Should now use SDK fake  
		err2 := DeleteItem(context.Background(), "aTable", "id", "aUserID")
		assert.EqualError(t, err2, "error creating DynamoDB client: sdk fake error")
	})
}