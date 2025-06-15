package lemur_test

import (
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/ukiahsmith/lemur"
)

func TestLemur_Render(t *testing.T) {
	type testCase struct {
		Name           string
		TemplateFS     fs.FS // Changed from TmplDir string
		TmplName       string
		Data           interface{}
		ExpectedOutput string
		ExpectError    bool
		ErrorContains  string // Substring to check in error message
	}

	testCases := []testCase{
		{
			Name:           "Successful render",
			TemplateFS:     os.DirFS("testdata/minimal"),
			TmplName:       "_defaults",
			Data:           nil,
			ExpectedOutput: "This is the most simple template.\n",
			ExpectError:    false,
		},
		{
			Name:           "Empty template name",
			TemplateFS:     os.DirFS("testdata/minimal"),
			TmplName:       "", // Should default to _defaults
			Data:           nil,
			ExpectedOutput: "This is the most simple template.\n",
			ExpectError:    false,
		},
		{
			Name:          "Non-existent template name",
			TemplateFS:    os.DirFS("testdata/minimal"),
			TmplName:      "nonexistent",
			Data:          nil,
			ExpectError:   true,
			ErrorContains: `lemur Render: no template with name "nonexistent"`,
		},
		{
			Name:          "ExecuteTemplate error",
			TemplateFS:    os.DirFS("testdata/render_error"),
			TmplName:      "_defaults",
			Data:          map[string]int{"Num": 10, "Denom": 0},
			ExpectError:   true,
			ErrorContains: "lemur Render: could not render template",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			wh, err := lemur.New(tc.TemplateFS, nil)
			// All test cases in TestLemur_Render expect lemur.New to succeed.
			// If it fails, it's a setup issue for that specific test case.
			if err != nil {
				// Using tc.Name with TemplateFS might be confusing as TemplateFS is not a simple string.
				// For logging, it's better to refer to the test case name.
				t.Fatalf("For test case %q, lemur.New failed during setup: %v", tc.Name, err)
			}

			var buf strings.Builder
			err = wh.Render(&buf, tc.TmplName, tc.Data)

			if tc.ExpectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				} else if tc.ErrorContains != "" && !strings.Contains(err.Error(), tc.ErrorContains) {
					t.Errorf("Expected error message to contain %q, but got %q", tc.ErrorContains, err.Error())
				}
				// Specific check for the "mod by zero" error's underlying cause
				if tc.Name == "ExecuteTemplate error" {
					if !strings.Contains(err.Error(), "number can't be divided by zero") {
						t.Errorf("Expected underlying error 'number can't be divided by zero', got: %s", err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if buf.String() != tc.ExpectedOutput {
					t.Errorf("Expected output %q, but got %q", tc.ExpectedOutput, buf.String())
				}
			}
		})
	}
}

func TestLemur_Srender_Errors(t *testing.T) {
	type testCase struct {
		Name          string
		TemplateFS    fs.FS // Changed from TmplDir string
		TmplName      string
		Data          interface{}
		ErrorContains string // Substring to check in error message
	}

	testCases := []testCase{
		{
			Name:          "Non-existent template name",
			TemplateFS:    os.DirFS("testdata/minimal"),
			TmplName:      "nonexistent_srender",
			Data:          nil,
			ErrorContains: `lemur Render: no template with name "nonexistent_srender"`,
		},
		{
			Name:          "ExecuteTemplate error",
			TemplateFS:    os.DirFS("testdata/render_error"),
			TmplName:      "_defaults",
			Data:          map[string]int{"Num": 5, "Denom": 0},
			ErrorContains: "lemur Render: could not render template",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			wh, err := lemur.New(tc.TemplateFS, nil)
			// All test cases in TestLemur_Srender_Errors expect lemur.New to succeed.
			// If it fails, it's a setup issue for that specific test case.
			if err != nil {
				t.Fatalf("For test case %q, lemur.New failed during setup: %v", tc.Name, err)
			}

			_, err = wh.Srender(tc.TmplName, tc.Data)

			if err == nil {
				t.Errorf("Expected an error, but got nil")
			} else if tc.ErrorContains != "" && !strings.Contains(err.Error(), tc.ErrorContains) {
				t.Errorf("Expected error message to contain %q, but got %q", tc.ErrorContains, err.Error())
			}
			// Specific check for the "mod by zero" error's underlying cause for Srender
			if tc.Name == "ExecuteTemplate error" {
				if !strings.Contains(err.Error(), "number can't be divided by zero") {
					t.Errorf("Expected underlying error 'number can't be divided by zero', got: %s", err.Error())
				}
			}
		})
	}
}
