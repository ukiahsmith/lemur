package lemur_test

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/ukiahsmith/lemur"
)

func Test_New(t *testing.T) {
	dataTemplate := []struct {
		Name     string
		TmplFS   fs.FS
		TmplName string
		Expected string
	}{
		{
			"Minimal",
			os.DirFS("testdata/minimal"),
			"_defaults",
			"This is the most simple template.\n",
		},
		{
			"Minimal empty tmplName",
			os.DirFS("testdata/minimal"),
			"", // Should default to _defaults
			"This is the most simple template.\n",
		},
		{
			"w_index",
			os.DirFS("testdata/w_index"),
			"mytemplate",
			"This is the w_index test.\n",
		},
		{
			"default_index_w_tmpl",
			os.DirFS("testdata/default_index_w_tmpl"),
			"mytemplate",
			`This is the default _index.html.tmpl file.
This is a_template.
`,
		},
		{
			"defaults_w_tmpl_w_index",
			os.DirFS("testdata/defaults_w_tmpl_w_index"),
			"mytemplate",
			`This is the defaults_w_tmpl_w_index mytemplate _index.html.tmpl.
defaults_w_tmpl_w_index atmpl.
`,
		},
		{
			"full_dir mytemplate",
			os.DirFS("testdata/full_dir"),
			"mytemplate",
			`full_dir _defaults _index.html.tmpl
full_dir mytemplate content.
`,
		},
		{
			"full_dir other_tmpl",
			os.DirFS("testdata/full_dir"),
			"other_tmpl",
			`full_dir other_tmpl _index.html.tmpl
full_dir atmpl.html.tmpl
`,
		},
		{
			"full_dir other_tmpl_2",
			os.DirFS("testdata/full_dir"),
			"other_tmpl_2",
			`full_dir other_tmpl _index.html.tmpl
full_dir atmpl.html.tmpl
Author: Name Name
`,
		},
	}

	for _, d := range dataTemplate {
		t.Run(d.Name, func(t *testing.T) {
			// Check if os.DirFS itself might have issues with the path (e.g. permissions, though unlikely for testdata)
			// For non-existent paths, os.DirFS doesn't error, but operations on the FS will.
			// This is handled by Test_New_ShouldErr. Here we assume valid testdata paths.
			tmplObj, err := lemur.New(d.TmplFS, nil)
			if err != nil {
				t.Fatalf("Test_New, %s: lemur.New failed: %s", d.Name, err)
			}

			maybe, err := tmplObj.Srender(d.TmplName, nil)
			if err != nil {
				t.Fatalf("Test_New, %s: Srender failed: %s", d.Name, err)
			}

			if maybe != d.Expected {
				t.Errorf("Test_New, %s: failed, maybe expected %q but got %q", d.Name, d.Expected, maybe)
			}
		})
	}
}

func Test_New_ShouldErr(t *testing.T) {
	data_New_ShouldErr := []struct {
		Name                string
		TemplateFS          fs.FS
		ExpectedErr         error
		ExpectedMsgContains string
	}{
		{
			"FS for non-existent dir",
			os.DirFS("this/directory/should/not/exist"), // os.DirFS itself doesn't error here.
			lemur.ErrTemplateDir,
			"layouts directory layouts does not exist in the filesystem",
		},
		{
			"FS for a file path",
			os.DirFS("testdata/notdirectory"), // os.DirFS itself doesn't error here.
			lemur.ErrTemplateDir,
			"not a directory",
		},
		// Example of a more controlled FS for missing _defaults (requires a helper or a real dir structure)
		// For now, the above cover the direct os.DirFS changes.
		// To test missing "layouts/_defaults", you'd need an FS where "layouts" exists
		// but "layouts/_defaults" does not. This can be done with a temporary directory
		// or a mock FS like testing/fstest.MapFS if we want to avoid disk I/O.
	}

	for _, d := range data_New_ShouldErr {
		t.Run(d.Name, func(t *testing.T) {
			_, err := lemur.New(d.TemplateFS, nil)
			if err == nil {
				t.Fatalf("Test_New_ShouldErr, %s: expected error, received no error", d.Name)
			}

			if !errors.Is(err, d.ExpectedErr) {
				t.Errorf("Test_New_ShouldErr, %s: expected error to wrap %v, but it did not. Got: %v", d.Name, d.ExpectedErr, err)
			}

			if d.ExpectedMsgContains != "" && !strings.Contains(err.Error(), d.ExpectedMsgContains) {
				t.Errorf("Test_New_ShouldErr, %s: expected error message to contain %q, received %q", d.Name, d.ExpectedMsgContains, err.Error())
			}
		})
	}
}
