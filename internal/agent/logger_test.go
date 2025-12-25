package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGetCurrentSessionNumber(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "logger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		setupFunc    func(string) error
		expectedNum  int
		expectError  bool
	}{
		{
			name: "no sessions.json - returns 1",
			setupFunc: func(dir string) error {
				// Don't create sessions.json
				return nil
			},
			expectedNum: 1,
			expectError: false,
		},
		{
			name: "empty sessions - returns 1",
			setupFunc: func(dir string) error {
				data := SessionsData{Sessions: []struct{ ID int `json:"id"` }{}}
				bytes, _ := json.Marshal(data)
				return os.WriteFile(filepath.Join(dir, "sessions.json"), bytes, 0644)
			},
			expectedNum: 1,
			expectError: false,
		},
		{
			name: "sessions with IDs 1,2,3 - returns 4",
			setupFunc: func(dir string) error {
				data := SessionsData{
					Sessions: []struct{ ID int `json:"id"` }{
						{ID: 1},
						{ID: 2},
						{ID: 3},
					},
				}
				bytes, _ := json.Marshal(data)
				return os.WriteFile(filepath.Join(dir, "sessions.json"), bytes, 0644)
			},
			expectedNum: 4,
			expectError: false,
		},
		{
			name: "non-sequential IDs - returns max+1",
			setupFunc: func(dir string) error {
				data := SessionsData{
					Sessions: []struct{ ID int `json:"id"` }{
						{ID: 1},
						{ID: 5},
						{ID: 3},
					},
				}
				bytes, _ := json.Marshal(data)
				return os.WriteFile(filepath.Join(dir, "sessions.json"), bytes, 0644)
			},
			expectedNum: 6,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Execute
			num, err := getCurrentSessionNumber(tmpDir)

			// Verify
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if num != tt.expectedNum {
				t.Errorf("getCurrentSessionNumber() = %v, want %v", num, tt.expectedNum)
			}

			// Cleanup for next iteration
			os.Remove(filepath.Join(tmpDir, "sessions.json"))
		})
	}
}

func TestArchiveOldLogIfNeeded(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "logger-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name          string
		createLog     bool
		logSize       int64
		expectArchive bool
	}{
		{
			name:          "no log file - no archive",
			createLog:     false,
			expectArchive: false,
		},
		{
			name:          "small log file - no archive",
			createLog:     true,
			logSize:       100,
			expectArchive: false,
		},
		{
			name:          "large log file - archive",
			createLog:     true,
			logSize:       2 * 1024 * 1024, // 2MB
			expectArchive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logPath := filepath.Join(tmpDir, "test.log")

			// Setup
			if tt.createLog {
				// Create log file with specified size
				f, err := os.Create(logPath)
				if err != nil {
					t.Fatalf("Failed to create log file: %v", err)
				}
				if tt.logSize > 0 {
					data := make([]byte, tt.logSize)
					f.Write(data)
				}
				f.Close()
			}

			// Execute
			archiveOldLogIfNeeded(logPath)

			// Verify
			_, err := os.Stat(logPath)
			logExists := err == nil

			// Check if archived file exists
			archivedFiles, _ := filepath.Glob(logPath + ".archived-*")
			wasArchived := len(archivedFiles) > 0

			if tt.expectArchive {
				if !wasArchived {
					t.Errorf("Expected log to be archived but it wasn't")
				}
				if logExists {
					t.Errorf("Expected original log to be removed but it still exists")
				}
			} else {
				if wasArchived {
					t.Errorf("Expected log NOT to be archived but it was")
				}
			}

			// Cleanup
			os.Remove(logPath)
			for _, archived := range archivedFiles {
				os.Remove(archived)
			}
		})
	}
}
