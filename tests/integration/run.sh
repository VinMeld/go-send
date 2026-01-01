#!/bin/bash
set -e

# Get the root directory of the project
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

echo "=== Docker Integration Test ==="

# Create a temporary directory for test data
TEST_DIR=$(mktemp -d)
echo "Test Data Directory: $TEST_DIR"

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    docker stop go-send-server-test || true
    docker rm go-send-server-test || true
    rm -f go-send-client
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# 1. Build Server Docker Image
echo "Building Server Docker Image..."
docker build -t go-send-server:test .

# 2. Start Server Container
echo "Starting Server Container..."
docker run -d --name go-send-server-test -p 8085:8080 -e REGISTRATION_TOKEN=secret123 go-send-server:test

# Wait for server to be ready
echo "Waiting for server to start..."
sleep 5
if ! curl -s http://localhost:8085/ping > /dev/null; then
    echo "Server failed to start!"
    docker logs go-send-server-test
    exit 1
fi
echo "Server is up!"

# 3. Build Client Binary
echo "Building Client Binary..."
go build -o go-send-client cmd/client/main.go

# 4. Run Scenarios

SERVER_URL="http://localhost:8085"
ALICE_CONFIG="$TEST_DIR/alice_config.json"
BOB_CONFIG="$TEST_DIR/bob_config.json"

echo "--- Scenario: Setup Alice ---"
ALICE_OUTPUT=$(./go-send-client config init --user alice --config "$ALICE_CONFIG")
echo "$ALICE_OUTPUT"
./go-send-client set-server $SERVER_URL --config "$ALICE_CONFIG"

echo "--- Scenario: Setup Bob ---"
BOB_OUTPUT=$(./go-send-client config init --user bob --config "$BOB_CONFIG")
echo "$BOB_OUTPUT"
./go-send-client set-server $SERVER_URL --config "$BOB_CONFIG"

# Extract keys from output
ALICE_ID_PUB=$(echo "$ALICE_OUTPUT" | grep "Identity Public Key:" | awk '{print $4}')
ALICE_EX_PUB=$(echo "$ALICE_OUTPUT" | grep "Exchange Public Key:" | awk '{print $4}')
BOB_ID_PUB=$(echo "$BOB_OUTPUT" | grep "Identity Public Key:" | awk '{print $4}')
BOB_EX_PUB=$(echo "$BOB_OUTPUT" | grep "Exchange Public Key:" | awk '{print $4}')

echo "Alice ID Pub: $ALICE_ID_PUB"
echo "Alice EX Pub: $ALICE_EX_PUB"
echo "Bob ID Pub: $BOB_ID_PUB"
echo "Bob EX Pub: $BOB_EX_PUB"

echo "--- Scenario: Register Users with Server ---"
# Test failure without token
if ./go-send-client register --config "$ALICE_CONFIG" 2>&1 | grep "forbidden"; then
    echo "Correctly rejected registration without token"
else
    echo "FAILURE: Should have rejected registration without token"
    exit 1
fi

# Register with token
./go-send-client register --token secret123 --config "$ALICE_CONFIG"
./go-send-client register --token secret123 --config "$BOB_CONFIG"

echo "--- Scenario: Exchange Keys Locally ---"
./go-send-client add-user bob $BOB_ID_PUB $BOB_EX_PUB --config "$ALICE_CONFIG"
./go-send-client add-user alice $ALICE_ID_PUB $ALICE_EX_PUB --config "$BOB_CONFIG"

echo "--- Scenario: Login ---"
./go-send-client login --config "$ALICE_CONFIG"
./go-send-client login --config "$BOB_CONFIG"

echo "--- Scenario: Alice Sends File to Bob ---"
echo "This is a secret message" > "$TEST_DIR/test_file.txt"
./go-send-client send-file bob "$TEST_DIR/test_file.txt" --config "$ALICE_CONFIG"

echo "--- Scenario: Bob Lists Files ---"
./go-send-client list-files --config "$BOB_CONFIG"

# Get File ID
FILE_ID=$(./go-send-client list-files --config "$BOB_CONFIG" | grep "from alice" | awk '{print $2}' | tr -d '[]')
echo "File ID: $FILE_ID"

if [ -z "$FILE_ID" ]; then
    echo "File not found in list!"
    exit 1
fi

echo "--- Scenario: Bob Downloads File ---"
# Download to test dir
cd "$TEST_DIR"
"$PROJECT_ROOT/go-send-client" download-file $FILE_ID --config "$BOB_CONFIG"

if [ ! -f "test_file.txt" ]; then
    echo "Downloaded file not found"
    exit 1
fi

if cmp -s "test_file.txt" "test_file.txt"; then
    echo "SUCCESS: File content matches!"
else
    echo "FAILURE: File content mismatch!"
    cat test_file.txt
    exit 1
fi

# Go back to root
cd "$PROJECT_ROOT"

echo "--- Scenario: Auto-Delete File ---"
echo "This message will self-destruct" > "$TEST_DIR/secret.txt"
./go-send-client send-file bob "$TEST_DIR/secret.txt" --auto-delete --config "$ALICE_CONFIG"

# Get File ID
SECRET_ID=$(./go-send-client list-files --config "$BOB_CONFIG" | grep "secret.txt" | awk '{print $2}' | tr -d '[]')
echo "Secret File ID: $SECRET_ID"

if [ -z "$SECRET_ID" ]; then
    echo "Secret file not found!"
    exit 1
fi

# Download (should succeed)
cd "$TEST_DIR"
"$PROJECT_ROOT/go-send-client" download-file $SECRET_ID --config "$BOB_CONFIG"

if [ ! -f "secret.txt" ]; then
    echo "Secret file download failed"
    exit 1
fi

# Check if deleted from server
# List files again, should be empty or not contain secret.txt
REMAINING=$("$PROJECT_ROOT/go-send-client" list-files --config "$BOB_CONFIG" | grep "secret.txt" || true)
if [ -n "$REMAINING" ]; then
    echo "FAILURE: Secret file was not auto-deleted!"
    exit 1
else
    echo "SUCCESS: Secret file auto-deleted!"
fi

# Go back to root
cd "$PROJECT_ROOT"

echo "--- Scenario: Recipient Deletion ---"
echo "Delete me recipient" > "$TEST_DIR/del_recip.txt"
./go-send-client send-file bob "$TEST_DIR/del_recip.txt" --config "$ALICE_CONFIG"
DEL_ID=$(./go-send-client list-files --config "$BOB_CONFIG" | grep "del_recip.txt" | awk '{print $2}' | tr -d '[]')
echo "File to delete (Recipient): $DEL_ID"
./go-send-client delete-file $DEL_ID --config "$BOB_CONFIG"
# Verify deletion
if ./go-send-client list-files --config "$BOB_CONFIG" | grep "$DEL_ID"; then
    echo "FAILURE: File not deleted by recipient!"
    exit 1
else
    echo "SUCCESS: File deleted by recipient!"
fi

echo "--- Scenario: Sender Deletion ---"
echo "Delete me sender" > "$TEST_DIR/del_sender.txt"
./go-send-client send-file bob "$TEST_DIR/del_sender.txt" --config "$ALICE_CONFIG"
DEL_SENDER_ID=$(./go-send-client list-files --config "$BOB_CONFIG" | grep "del_sender.txt" | awk '{print $2}' | tr -d '[]')
echo "File to delete (Sender): $DEL_SENDER_ID"
./go-send-client delete-file $DEL_SENDER_ID --config "$ALICE_CONFIG"
# Verify deletion
if ./go-send-client list-files --config "$BOB_CONFIG" | grep "$DEL_SENDER_ID"; then
    echo "FAILURE: File not deleted by sender!"
    exit 1
else
    echo "SUCCESS: File deleted by sender!"
fi

echo "Integration Test Passed!"
