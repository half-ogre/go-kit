package dynamodbkit

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestUseTableNameSuffix(t *testing.T) {
	t.Run("delete_item_applies_global_suffix_when_no_option_suffix_provided", func(t *testing.T) {
		UseTableNameSuffix("theSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "theTableName", "id", "aUserID")

		assert.NoError(t, err)
		assert.Equal(t, "theTableNametheSuffix", actualTableName)
	})

	t.Run("delete_item_option_suffix_takes_precedence_over_global_suffix", func(t *testing.T) {
		UseTableNameSuffix("theGlobalSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "theTableName", "id", "aUserID",
			WithDeleteItemTableNameSuffix("theOptionSuffix"))

		assert.NoError(t, err)
		assert.Equal(t, "theTableNametheOptionSuffix", actualTableName)
	})

	t.Run("delete_item_does_not_apply_global_suffix_when_empty", func(t *testing.T) {
		UseTableNameSuffix("")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err := DeleteItem(context.Background(), "theTableName", "id", "aUserID")

		assert.NoError(t, err)
		assert.Equal(t, "theTableName", actualTableName)
	})

	t.Run("get_item_applies_global_suffix_when_no_option_suffix_provided", func(t *testing.T) {
		UseTableNameSuffix("theSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.GetItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := GetItem[TestUser](context.Background(), "theTableName", "id", "aUserID")

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "theTableNametheSuffix", actualTableName)
	})

	t.Run("get_item_option_suffix_takes_precedence_over_global_suffix", func(t *testing.T) {
		UseTableNameSuffix("theGlobalSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.GetItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := GetItem[TestUser](context.Background(), "theTableName", "id", "aUserID",
			WithGetItemTableNameSuffix("theOptionSuffix"))

		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "theTableNametheOptionSuffix", actualTableName)
	})

	t.Run("put_item_applies_global_suffix_when_no_option_suffix_provided", func(t *testing.T) {
		UseTableNameSuffix("theSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			PutItemFake: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.PutItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		testUser := TestUser{ID: "0", Name: "A Name", Email: "anEmail@anAddress.com"}
		err := PutItem(context.Background(), "theTableName", testUser)

		assert.NoError(t, err)
		assert.Equal(t, "theTableNametheSuffix", actualTableName)
	})

	t.Run("put_item_option_suffix_takes_precedence_over_global_suffix", func(t *testing.T) {
		UseTableNameSuffix("theGlobalSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			PutItemFake: func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.PutItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		testUser := TestUser{ID: "0", Name: "A Name", Email: "anEmail@anAddress.com"}
		err := PutItem(context.Background(), "theTableName", testUser,
			WithPutItemTableNameSuffix("theOptionSuffix"))

		assert.NoError(t, err)
		assert.Equal(t, "theTableNametheOptionSuffix", actualTableName)
	})

	t.Run("query_applies_global_suffix_when_no_option_suffix_provided", func(t *testing.T) {
		UseTableNameSuffix("theSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "theTableName", "id", "aUserID")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "theTableNametheSuffix", actualTableName)
	})

	t.Run("query_option_suffix_takes_precedence_over_global_suffix", func(t *testing.T) {
		UseTableNameSuffix("theGlobalSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			QueryFake: func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Query[TestUser](context.Background(), "theTableName", "id", "aUserID",
			WithQueryTableNameSuffix("theOptionSuffix"))

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "theTableNametheOptionSuffix", actualTableName)
	})

	t.Run("scan_applies_global_suffix_when_no_option_suffix_provided", func(t *testing.T) {
		UseTableNameSuffix("theSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			ScanFake: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Scan[TestUser](context.Background(), "theTableName")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "theTableNametheSuffix", actualTableName)
	})

	t.Run("scan_option_suffix_takes_precedence_over_global_suffix", func(t *testing.T) {
		UseTableNameSuffix("theGlobalSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		actualTableName := ""
		fakeDB := &FakeDynamoDB{
			ScanFake: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				actualTableName = *params.TableName
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		result, err := Scan[TestUser](context.Background(), "theTableName",
			WithScanTableNameSuffix("theOptionSuffix"))

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "theTableNametheOptionSuffix", actualTableName)
	})

	t.Run("multiple_operations_use_same_global_suffix", func(t *testing.T) {
		UseTableNameSuffix("theSharedSuffix")
		t.Cleanup(func() { UseTableNameSuffix("") })

		var tableNames []string
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				tableNames = append(tableNames, *params.TableName)
				return &dynamodb.DeleteItemOutput{}, nil
			},
			GetItemFake: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
				tableNames = append(tableNames, *params.TableName)
				return &dynamodb.GetItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		err1 := DeleteItem(context.Background(), "users", "id", "user1")
		result, err2 := GetItem[TestUser](context.Background(), "posts", "id", "post1")

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Nil(t, result)
		assert.Equal(t, []string{"userstheSharedSuffix", "poststheSharedSuffix"}, tableNames)
	})

	t.Run("changing_global_suffix_affects_subsequent_operations", func(t *testing.T) {
		var tableNames []string
		fakeDB := &FakeDynamoDB{
			DeleteItemFake: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				tableNames = append(tableNames, *params.TableName)
				return &dynamodb.DeleteItemOutput{}, nil
			},
		}
		setFake(func(ctx context.Context) (DynamoDB, error) { return fakeDB, nil })
		t.Cleanup(func() { setFake(nil) })

		UseTableNameSuffix("firstSuffix")
		err1 := DeleteItem(context.Background(), "users", "id", "user1")

		UseTableNameSuffix("secondSuffix")
		err2 := DeleteItem(context.Background(), "users", "id", "user2")

		UseTableNameSuffix("")
		t.Cleanup(func() { UseTableNameSuffix("") })

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, []string{"usersfirstSuffix", "userssecondSuffix"}, tableNames)
	})
}
