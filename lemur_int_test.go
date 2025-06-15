package lemur

import (
	"errors"
	"fmt"
	"io/fs" // Added import for fs.ModeDir
	"strings"
	"testing"
	"testing/fstest"
	"time" // Added for mockFileInfo
)

func TestValidateTemplateDirectory_LayoutsIsFile(t *testing.T) {
	testCases := []struct {
		name                string
		templateFS          fstest.MapFS
		expectedErr         error
		expectedMsgContains string
	}{
		{
			name: "Layouts path is a file",
			templateFS: fstest.MapFS{
				"layouts": &fstest.MapFile{Data: []byte("this is a file"), Mode: 0o644}, // Mode 0o644 makes it a regular file
			},
			expectedErr:         ErrTemplateDir,
			expectedMsgContains: "path layouts is not a directory in the filesystem",
		},
		{
			name: "Layouts path exists, but _defaults is a file",
			templateFS: fstest.MapFS{
				"layouts/":          &fstest.MapFile{Mode: fs.ModeDir},                            // layouts is a directory
				"layouts/_defaults": &fstest.MapFile{Data: []byte("this is a file"), Mode: 0o644}, // _defaults is a file
			},
			expectedErr:         ErrTemplateDir,
			expectedMsgContains: "path layouts/_defaults is not a directory in the filesystem",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTemplateDirectory(tc.templateFS)

			if err == nil {
				t.Fatalf("Expected an error, but got nil")
			}

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error to be %v, but got %v", tc.expectedErr, err)
			}

			if tc.expectedMsgContains != "" && !strings.Contains(err.Error(), tc.expectedMsgContains) {
				t.Errorf("Expected error message to contain %q, but got %q", tc.expectedMsgContains, err.Error())
			}
		})
	}
}

func TestValidateTemplateDirectory_DefaultsNotExists(t *testing.T) {
	// Create a test filesystem with a layouts directory but no _defaults directory
	testFS := fstest.MapFS{
		"layouts/": &fstest.MapFile{Mode: fs.ModeDir}, // Only the layouts directory exists
	}

	// Call the function under test
	err := validateTemplateDirectory(testFS)

	// Assert that an error was returned
	if err == nil {
		t.Fatalf("Expected an error when _defaults directory does not exist, but got nil")
	}

	// Assert that the error is of the expected type
	if !errors.Is(err, ErrTemplateDir) {
		t.Errorf("Expected error to be %s, but got %s", ErrTemplateDir, err)
	}

	// Assert that the error message contains the expected text
	expectedErrMsg := "layouts/_defaults directory does not exist in the filesystem"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error message to contain %q, but got %q", expectedErrMsg, err.Error())
	}
}

// mockFileSystem implements fs.FS and returns custom errors for specific paths
type mockFileSystem struct {
	directoryMap map[string]bool                    // map of path to is-directory flag
	errorMap     map[string]error                   // map of path to custom error
	fileInfoMap  map[string]fs.FileInfo             // map of path to file info
	openFuncMap  map[string]func() (fs.File, error) // map of path to open function
}

// Open implements fs.FS
func (m mockFileSystem) Open(name string) (fs.File, error) {
	if f, ok := m.openFuncMap[name]; ok {
		return f()
	}
	return nil, fmt.Errorf("mockFileSystem: file %s not found", name)
}

// mockFileInfo implements fs.FileInfo
type mockFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() fs.FileMode  { return m.mode }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return m.sys }

// mockStatFS implements fs.StatFS for custom stat responses
type mockStatFS struct {
	fs.FS
	statFunc func(name string) (fs.FileInfo, error)
}

func (m mockStatFS) Stat(name string) (fs.FileInfo, error) {
	return m.statFunc(name)
}

func TestValidateTemplateDirectory_DefaultsNonExistError(t *testing.T) {
	// Create a custom error that is different from fs.ErrNotExist
	customErr := fmt.Errorf("custom error: permission denied")

	// Create a mock filesystem that returns our custom error for _defaults
	mockFS := mockStatFS{
		FS: fstest.MapFS{
			"layouts/": &fstest.MapFile{Mode: fs.ModeDir},
		},
		statFunc: func(name string) (fs.FileInfo, error) {
			if name == "layouts" {
				// Return valid directory info for layouts
				return mockFileInfo{
					name:  "layouts",
					mode:  fs.ModeDir,
					isDir: true,
				}, nil
			} else if name == "layouts/_defaults" {
				// Return our custom error for _defaults, not fs.ErrNotExist
				return nil, customErr
			}
			return nil, fs.ErrNotExist
		},
	}

	// Call the function under test
	err := validateTemplateDirectory(mockFS)

	// Assert that an error was returned
	if err == nil {
		t.Fatalf("Expected an error when accessing _defaults directory fails, but got nil")
	}

	// Assert that the error is wrapped with ErrTemplateDir
	if !errors.Is(err, ErrTemplateDir) {
		t.Errorf("Expected error to be %v, but got %v", ErrTemplateDir, err)
	}

	// Verify that the original error is not fs.ErrNotExist
	var notExistErr *fs.PathError
	if errors.As(err, &notExistErr) && errors.Is(notExistErr.Err, fs.ErrNotExist) {
		t.Errorf("Expected error not to be fs.ErrNotExist, but it was")
	}

	// Assert that the error message contains reference to the custom error
	if !strings.Contains(err.Error(), customErr.Error()) {
		t.Errorf("Expected error message to contain %q, but got %q", customErr.Error(), err.Error())
	}
}
