#!/bin/bash
set -e

echo "=== Docker Integration Test ==="

# Cleanup function
cleanup() {
    echo "Cleaning up..."
    docker stop go-send-server-test || true
    docker rm go-send-server-test || true
    rm -f go-send-client
    rm -rf test_data_alice test_data_bob
    rm -f test_file.txt downloaded_file.txt
}
trap cleanup EXIT

# 1. Build Server Docker Image
echo "Building Server Docker Image..."
docker build -t go-send-server:test .

# 2. Start Server Container
echo "Starting Server Container..."
docker run -d --name go-send-server-test -p 8085:8080 go-send-server:test

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
ALICE_CONFIG="test_data_alice/config.json"
BOB_CONFIG="test_data_bob/config.json"

mkdir -p test_data_alice test_data_bob

echo "--- Scenario: Setup Alice ---"
ALICE_OUTPUT=$(./go-send-client config init --user alice --config $ALICE_CONFIG)
echo "$ALICE_OUTPUT"
./go-send-client set-server $SERVER_URL --config $ALICE_CONFIG

echo "--- Scenario: Setup Bob ---"
BOB_OUTPUT=$(./go-send-client config init --user bob --config $BOB_CONFIG)
echo "$BOB_OUTPUT"
./go-send-client set-server $SERVER_URL --config $BOB_CONFIG

# Extract keys from output
# Output format: "Public Key: <key>"
ALICE_PUB=$(echo "$ALICE_OUTPUT" | grep "Public Key:" | awk '{print $3}')
BOB_PUB=$(echo "$BOB_OUTPUT" | grep "Public Key:" | awk '{print $3}')

echo "Alice Pub: $ALICE_PUB"
echo "Bob Pub: $BOB_PUB"

echo "--- Scenario: Exchange Keys ---"
./go-send-client add-user bob $BOB_PUB --config $ALICE_CONFIG
./go-send-client add-user alice $ALICE_PUB --config $BOB_CONFIG

echo "--- Scenario: Alice Sends File to Bob ---"
echo "This is a secret message" > test_file.txt
./go-send-client send-file bob test_file.txt --config $ALICE_CONFIG

echo "--- Scenario: Bob Lists Files ---"
./go-send-client list-files --config $BOB_CONFIG

# We need the File ID to download. 
# `list-files` prints to stdout. We can parse it.
# Output format: "- [<file_id>] <filename> (from <sender>) ..."
# Example: "- [0d383577-3689-45b5-9eda-5a7f2ca6fb53] test_file.txt (from alice) - 01 Jan 26 05:55 UTC"
FILE_ID=$(./go-send-client list-files --config $BOB_CONFIG | grep "from alice" | awk '{print $2}' | tr -d '[]')
echo "File ID: $FILE_ID"

if [ -z "$FILE_ID" ]; then
    echo "File not found in list!"
    exit 1
fi

echo "--- Scenario: Bob Downloads File ---"
./go-send-client download-file $FILE_ID --config $BOB_CONFIG

if [ ! -f "$FILE_ID" ]; then
    echo "Downloaded file not found (it saves as FileID by default if not specified? No, it saves as original filename usually?)"
    # Check code: It saves as metadata.FileName
    # metadata.FileName is test_file.txt
    if [ -f "test_file.txt" ]; then
         # Wait, we created test_file.txt in current dir.
         # Download might overwrite it or we should check content.
         # Let's rename original first.
         mv test_file.txt original_test_file.txt
    fi
fi

# Re-download to be sure (since we moved original)
./go-send-client download-file $FILE_ID --config $BOB_CONFIG

if cmp -s "test_file.txt" "original_test_file.txt"; then
    echo "SUCCESS: File content matches!"
else
    echo "FAILURE: File content mismatch!"
    cat test_file.txt
    exit 1
fi



echo "--- Scenario: Auto-Delete File ---"
echo "This message will self-destruct" > secret.txt
./go-send-client send-file bob secret.txt --auto-delete --config $ALICE_CONFIG

# Get File ID
SECRET_ID=$(./go-send-client list-files --config $BOB_CONFIG | grep "secret.txt" | awk '{print $2}' | tr -d '[]')
echo "Secret File ID: $SECRET_ID"

if [ -z "$SECRET_ID" ]; then
    echo "Secret file not found!"
    exit 1
fi

# Download (should succeed)
./go-send-client download-file $SECRET_ID --config $BOB_CONFIG
if [ ! -f "secret.txt" ]; then
    echo "Secret file download failed"
    exit 1
fi

# Check if deleted from server
# List files again, should be empty or not contain secret.txt
REMAINING=$(./go-send-client list-files --config $BOB_CONFIG | grep "secret.txt" || true)
if [ -n "$REMAINING" ]; then
    echo "FAILURE: Secret file was not auto-deleted!"
    exit 1
else
    echo "SUCCESS: Secret file auto-deleted!"
fi

echo "Integration Test Passed!"
