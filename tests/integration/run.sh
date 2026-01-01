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
# New: Init with server flag
ALICE_OUTPUT=$(./go-send-client config init --user alice --server "$SERVER_URL" --config "$ALICE_CONFIG")
echo "$ALICE_OUTPUT"

echo "--- Scenario: Setup Bob ---"
# New: Init with server flag
BOB_OUTPUT=$(./go-send-client config init --user bob --server "$SERVER_URL" --config "$BOB_CONFIG")
echo "$BOB_OUTPUT"

# Extract keys from output (just for verification logging)
ALICE_ID_PUB=$(echo "$ALICE_OUTPUT" | grep "Identity Public Key:" | awk '{print $4}')
BOB_ID_PUB=$(echo "$BOB_OUTPUT" | grep "Identity Public Key:" | awk '{print $4}')

echo "Alice ID Pub: $ALICE_ID_PUB"
echo "Bob ID Pub: $BOB_ID_PUB"

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

echo "--- Scenario: Login ---"
./go-send-client login --config "$ALICE_CONFIG"
./go-send-client login --config "$BOB_CONFIG"

echo "--- Scenario: Alice Sends File to Bob (Discovery) ---"
echo "This is a secret message" > "$TEST_DIR/test_file.txt"
# Bob is NOT in Alice's config, so this tests discovery
SEND_OUTPUT=$(./go-send-client send-file bob "$TEST_DIR/test_file.txt" --config "$ALICE_CONFIG")
echo "$SEND_OUTPUT"

if echo "$SEND_OUTPUT" | grep -q "Found user 'bob'"; then
    echo "SUCCESS: User discovery worked"
else
    echo "FAILURE: User discovery failed"
    exit 1
fi

echo "--- Scenario: Bob Lists Files (Indices) ---"
LIST_OUTPUT=$(./go-send-client list-files --config "$BOB_CONFIG")
echo "$LIST_OUTPUT"

# Check for index format "1 - [ID] filename"
if echo "$LIST_OUTPUT" | grep -q "1 - \[.*\] test_file.txt"; then
    echo "SUCCESS: List files shows index and filename"
else
    echo "FAILURE: List files output format incorrect"
    exit 1
fi

# Extract File ID for verification (optional, since we'll use index)
FILE_ID=$(echo "$LIST_OUTPUT" | grep "test_file.txt" | awk -F'[][]' '{print $2}')
echo "File ID: $FILE_ID"

echo "--- Scenario: Bob Downloads File by Index ---"
# Download to test dir using Index 1
cd "$TEST_DIR"
"$PROJECT_ROOT/go-send-client" download-file 1 --config "$BOB_CONFIG"

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

# List to get index (should be 2 now, or 1 if previous was deleted? No, previous wasn't deleted)
# Actually, list-files caches the list.
LIST_OUTPUT_2=$(./go-send-client list-files --config "$BOB_CONFIG")
echo "$LIST_OUTPUT_2"

# Assuming it's the second file, or we grep for it.
# The index depends on sort order.
# Let's grep for the line with secret.txt and extract the index.
SECRET_INDEX=$(echo "$LIST_OUTPUT_2" | grep "secret.txt" | awk '{print $1}')
echo "Secret File Index: $SECRET_INDEX"

if [ -z "$SECRET_INDEX" ]; then
    echo "Secret file not found!"
    exit 1
fi

# Download (should succeed)
cd "$TEST_DIR"
"$PROJECT_ROOT/go-send-client" download-file "$SECRET_INDEX" --config "$BOB_CONFIG"

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

LIST_OUTPUT_3=$(./go-send-client list-files --config "$BOB_CONFIG")
DEL_INDEX=$(echo "$LIST_OUTPUT_3" | grep "del_recip.txt" | awk '{print $1}')
echo "File to delete (Recipient) Index: $DEL_INDEX"

# Delete by Index (assuming delete-file supports index? No, delete-file likely still takes ID based on current implementation plan/code)
# Wait, did I update delete-file to support indices?
# I updated download-file and list-files. I don't recall updating delete-file.
# Let's check delete_cmd.go.
# If not, I need to get the ID.
DEL_ID=$(echo "$LIST_OUTPUT_3" | grep "del_recip.txt" | awk -F'[][]' '{print $2}')
./go-send-client delete-file "$DEL_ID" --config "$BOB_CONFIG"

# Verify deletion
if ./go-send-client list-files --config "$BOB_CONFIG" | grep "$DEL_ID"; then
    echo "FAILURE: File not deleted by recipient!"
    exit 1
else
    echo "SUCCESS: File deleted by recipient!"
fi

echo "--- Scenario: User Deletion (Remote) ---"
# Create a test user Charlie
CHARLIE_CONFIG="$TEST_DIR/charlie_config.json"
CHARLIE_OUTPUT=$(./go-send-client config init --user charlie --server "$SERVER_URL" --config "$CHARLIE_CONFIG")
echo "$CHARLIE_OUTPUT"

# Register and login Charlie
./go-send-client register --token secret123 --config "$CHARLIE_CONFIG"
./go-send-client login --config "$CHARLIE_CONFIG"

# Delete Charlie's account from server
./go-send-client remove-user charlie --remote --config "$CHARLIE_CONFIG"

# Try to login again (should fail since account is deleted)
if ./go-send-client login --config "$CHARLIE_CONFIG" 2>&1 | grep -i "error"; then
    echo "SUCCESS: Charlie's account was deleted from server!"
else
    echo "FAILURE: Charlie can still login after deletion!"
    exit 1
fi

echo "Integration Test Passed!"
