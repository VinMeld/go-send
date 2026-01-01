package e2e

import (
	"bytes"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/VinMeld/go-send/internal/client"
	"github.com/VinMeld/go-send/internal/server"
)

func TestEndToEnd(t *testing.T) {
	// 1. Setup Server
	serverDir, err := os.MkdirTemp("", "go-send-e2e-server")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(serverDir) }()

	blobStore := server.NewLocalBlobStore(serverDir)
	storage, err := server.NewStorage(serverDir, blobStore)
	if err != nil {
		t.Fatal(err)
	}
	handler := server.NewHandler(storage)
	handler.SetRegistrationToken("secret-token")

	ts := httptest.NewServer(handler)
	defer ts.Close()

	// 2. Setup Client Configs
	aliceDir, err := os.MkdirTemp("", "go-send-e2e-alice")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(aliceDir) }()

	bobDir, err := os.MkdirTemp("", "go-send-e2e-bob")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(bobDir) }()

	// Helper to run client commands
	runCmd := func(configDir string, args ...string) (string, error) {
		// Reset client state if possible, or just set flags
		// Since cobra commands use global flags, we need to be careful.
		// We'll use the --config flag to point to the specific config file.
		configFile := filepath.Join(configDir, "config.json")

		// We need to capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Reset root command flags?
		// Cobra flags persist, so we just append --config
		cmd := client.GetRootCmd()
		cmd.SetArgs(append(args, "--config", configFile))

		// We also need to ensure the config is reloaded for each run
		// The client.Execute() calls initConfig() via OnInitialize,
		// but OnInitialize only registers the function.
		// initConfig() is called when Execute() runs.
		// However, client.cfg is global. We need to reset it or ensure it's reloaded.
		// The initConfig function checks if cfgFile is set.
		// If we pass --config, it should reload.
		// But we might need to reset the global 'cfg' variable in client package if it's not exported.
		// It is exported via GetConfig() but not settable?
		// Actually, initConfig() overwrites 'cfg'.
		// So passing --config should work.

		err := cmd.Execute()

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		output := buf.String()

		return output, err
	}

	// 3. Alice Register & Login
	// Init
	_, err = runCmd(aliceDir, "config", "init", "--user", "alice", "--server", ts.URL)
	if err != nil {
		t.Fatalf("Alice init failed: %v", err)
	}

	// Register
	_, err = runCmd(aliceDir, "register", "--token", "secret-token")
	if err != nil {
		t.Fatalf("Alice register failed: %v", err)
	}

	// Login
	_, err = runCmd(aliceDir, "login")
	if err != nil {
		t.Fatalf("Alice login failed: %v", err)
	}

	// 4. Bob Register & Login
	// Init
	_, err = runCmd(bobDir, "config", "init", "--user", "bob", "--server", ts.URL)
	if err != nil {
		t.Fatalf("Bob init failed: %v", err)
	}

	// Register
	_, err = runCmd(bobDir, "register", "--token", "secret-token")
	if err != nil {
		t.Fatalf("Bob register failed: %v", err)
	}

	// Login
	_, err = runCmd(bobDir, "login")
	if err != nil {
		t.Fatalf("Bob login failed: %v", err)
	}

	// 5. Alice Send File to Bob
	testFile := filepath.Join(aliceDir, "hello.txt")
	if err := os.WriteFile(testFile, []byte("Hello Bob!"), 0644); err != nil {
		t.Fatal(err)
	}

	// Bob is not in Alice's address book, so this tests discovery too
	output, err := runCmd(aliceDir, "send-file", "bob", testFile)
	if err != nil {
		t.Fatalf("Alice send-file failed: %v", err)
	}
	if !strings.Contains(output, "File sent successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}
	if !strings.Contains(output, "Found user 'bob'") {
		t.Errorf("Expected user discovery message, got: %s", output)
	}

	// 6. Bob List Files
	output, err = runCmd(bobDir, "list-files")
	if err != nil {
		t.Fatalf("Bob list-files failed: %v", err)
	}
	if !strings.Contains(output, "1 - [") {
		t.Errorf("Expected file list with index 1, got: %s", output)
	}
	if !strings.Contains(output, "hello.txt") {
		t.Errorf("Expected hello.txt in list, got: %s", output)
	}

	// 7. Bob Download File
	// We use index 1
	// We need to be in bobDir so the file is saved there?
	// The download command saves to current working directory.
	// We should change Cwd for the test or just check where it saves.
	// It saves to filepath.Base(filename).
	// Let's change Cwd temporarily.
	oldWd, _ := os.Getwd()
	if err := os.Chdir(bobDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	output, err = runCmd(bobDir, "download-file", "1")
	if err != nil {
		t.Fatalf("Bob download-file failed: %v", err)
	}
	if !strings.Contains(output, "File downloaded and decrypted") {
		t.Errorf("Expected download success, got: %s", output)
	}

	// 8. Verify Content
	content, err := os.ReadFile(filepath.Join(bobDir, "hello.txt"))
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}
	if string(content) != "Hello Bob!" {
		t.Errorf("Expected 'Hello Bob!', got '%s'", string(content))
	}
}
