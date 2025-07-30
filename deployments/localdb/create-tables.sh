#!/bin/bash

# Set fake AWS credentials for DynamoDB Local
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy
export AWS_DEFAULT_REGION=us-east-1

# Wait for DynamoDB to be ready (up to 1 minute)
echo "Waiting for DynamoDB to be ready..."
TIMEOUT=60
ELAPSED=0
while ! curl -s http://localhost:8000 >/dev/null 2>&1; do
    if [ $ELAPSED -ge $TIMEOUT ]; then
        echo "Timeout: DynamoDB did not become ready within $TIMEOUT seconds"
        exit 1
    fi
    echo "DynamoDB not ready, waiting... (${ELAPSED}s/${TIMEOUT}s)"
    sleep 2
    ELAPSED=$((ELAPSED + 2))
done
echo "DynamoDB is ready!"

# Read table definitions from JSON file
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TABLES_JSON="$SCRIPT_DIR/tables.json"

if [ ! -f "$TABLES_JSON" ]; then
    echo "Error: tables.json not found at $TABLES_JSON"
    exit 1
fi

# Extract table names and create each table
jq -r '.tables[].TableName' "$TABLES_JSON" | while read -r table_name; do
    echo "Checking if table '$table_name' exists..."
    
    # Check if table already exists
    if aws dynamodb describe-table --table-name "$table_name" --endpoint-url http://localhost:8000 --region us-east-1 --no-cli-pager >/dev/null 2>&1; then
        echo "Table '$table_name' already exists, skipping creation"
    else
        echo "Creating table '$table_name'..."
        
        # Extract table components from JSON
        key_schema=$(jq -r ".tables[] | select(.TableName == \"$table_name\") | .KeySchema | map(\"AttributeName=\" + .AttributeName + \",KeyType=\" + .KeyType) | join(\" \")" "$TABLES_JSON")
        attr_definitions=$(jq -r ".tables[] | select(.TableName == \"$table_name\") | .AttributeDefinitions | map(\"AttributeName=\" + .AttributeName + \",AttributeType=\" + .AttributeType) | join(\" \")" "$TABLES_JSON")
        billing_mode=$(jq -r ".tables[] | select(.TableName == \"$table_name\") | .BillingMode" "$TABLES_JSON")
        
        # Create the table using AWS CLI arguments
        if aws dynamodb create-table \
            --table-name "$table_name" \
            --key-schema $key_schema \
            --attribute-definitions $attr_definitions \
            --billing-mode "$billing_mode" \
            --endpoint-url http://localhost:8000 \
            --region us-east-1 \
            --no-cli-pager >/dev/null; then
            echo "Successfully created table '$table_name'"
        else
            echo "Failed to create table '$table_name'"
            exit 1
        fi
    fi
done

echo "Table creation process completed!"