package lemur

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"

	"github.com/ukiahsmith/lemur/funcs"
)

type Lemur struct {
	layouts map[string]*template.Template
	funcs   template.FuncMap
}

func New(templateFS fs.FS, userFuncs template.FuncMap) (Lemur, error) {
	var wh Lemur

	// Initialize the Lemur instance with function maps
	wh.initializeFuncMaps(userFuncs)

	// Validate the template directory structure
	if err := validateTemplateDirectory(templateFS); err != nil {
		return Lemur{}, err
	}

	// Create the base template with function map
	tmpl := template.New("lemur").Funcs(wh.funcs)

	// Process the _defaults directory first
	tmpl, err := processDefaultsDirectory(templateFS, tmpl)
	if err != nil {
		return Lemur{}, err
	}

	// Process all layout directories
	wh.layouts, err = processLayoutDirectories(templateFS, tmpl)
	if err != nil {
		return Lemur{}, err
	}

	return wh, nil
}

// initializeFuncMaps sets up the template function maps
func (wh *Lemur) initializeFuncMaps(userFuncs template.FuncMap) {
	wh.layouts = make(map[string]*template.Template)
	wh.funcs = funcs.DefaultFuncMap()

	// Merge userFuncs, user-defined funcs take precedence
	for k, v := range userFuncs {
		wh.funcs[k] = v
	}
}

// processDefaultsDirectory handles the special _defaults directory
func processDefaultsDirectory(templateFS fs.FS, tmpl *template.Template) (*template.Template, error) {
	layoutsDirPath := "layouts"
	defaultsDirFullPath := filepath.Join(layoutsDirPath, "_defaults")

	// Read the defaults directory
	defaultEntries, err := fs.ReadDir(templateFS, defaultsDirFullPath)
	if err != nil {
		return nil, fmt.Errorf("error reading _defaults directory %s from filesystem: %w", defaultsDirFullPath, err)
	}

	// Parse _defaults/_index.html.tmpl first, if it exists, into the base tmpl
	tmpl, err = processDefaultsIndexTemplate(templateFS, tmpl, layoutsDirPath)
	if err != nil {
		return nil, err
	}

	// Process all other files in _defaults
	for _, de := range defaultEntries {
		if de.IsDir() {
			continue
		}
		fileName := de.Name()
		if fileName == "_index.html.tmpl" || fileName[0] == '.' {
			continue
		}

		defaultFileRelPath := filepath.Join(layoutsDirPath, "_defaults", fileName)
		tmpl, err = parseTemplateFile(templateFS, tmpl, defaultFileRelPath, filepath.Base(defaultFileRelPath))
		if err != nil {
			return nil, fmt.Errorf("failed to process default template file %s: %w", defaultFileRelPath, err)
		}
	}

	return tmpl, nil
}

// processDefaultsIndexTemplate handles the special _defaults/_index.html.tmpl file
func processDefaultsIndexTemplate(templateFS fs.FS, tmpl *template.Template, layoutsDirPath string) (*template.Template, error) {
	defaultsIndexRelPath := filepath.Join(layoutsDirPath, "_defaults", "_index.html.tmpl")

	if _, statErr := fs.Stat(templateFS, defaultsIndexRelPath); statErr != nil {
		return nil, fmt.Errorf("error stating _defaults/_index.html.tmpl %q: %w", defaultsIndexRelPath, statErr)
	}

	return parseTemplateFile(templateFS, tmpl, defaultsIndexRelPath, filepath.Base(defaultsIndexRelPath))
	// return tmpl, nil
}

// parseTemplateFile reads and parses a template file
func parseTemplateFile(templateFS fs.FS, tmpl *template.Template, filePath string, templateName string) (*template.Template, error) {
	content, readErr := fs.ReadFile(templateFS, filePath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read template file %q: %w", filePath, readErr)
	}

	// Parse content into a new template named by its base filename, associated with 'tmpl'
	parsedTmpl, err := tmpl.New(templateName).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template file %q: %w", filePath, err)
	}

	return parsedTmpl, nil
}

// processLayoutDirectories handles all layout directories and their templates
func processLayoutDirectories(templateFS fs.FS, baseTmpl *template.Template) (map[string]*template.Template, error) {
	layoutMap := make(map[string]*template.Template)
	layoutsDirPath := "layouts"

	// Get all entries in the layouts directory
	entries, err := fs.ReadDir(templateFS, layoutsDirPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: supplied filesystem does not contain directory %s", ErrTemplateDir, layoutsDirPath)
		}
		return nil, fmt.Errorf("%w: reading %s from filesystem: %s", ErrTemplateDir, layoutsDirPath, err)
	}

	// Process each layout directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name() // This is the layout name, e.g., "_defaults", "mytemplate"
		if name[0] == '.' {
			continue
		}

		// Process a single layout directory
		tmpl, err := processLayoutDirectory(templateFS, baseTmpl, layoutsDirPath, name)
		if err != nil {
			return nil, err
		}

		layoutMap[name] = tmpl
	}

	return layoutMap, nil
}

// processLayoutDirectory handles a single layout directory and its templates
func processLayoutDirectory(templateFS fs.FS, baseTmpl *template.Template, layoutsDirPath string, layoutName string) (*template.Template, error) {
	currentLayoutPathRel := filepath.Join(layoutsDirPath, layoutName)

	// Read all files in this layout directory
	tmplEntries, err := fs.ReadDir(templateFS, currentLayoutPathRel)
	if err != nil {
		return nil, fmt.Errorf("error reading directory for template set %s from filesystem: %w", layoutName, err)
	}

	// Clone the base template to use for this layout
	ctmpl, err := baseTmpl.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone base template for set %s: %w", layoutName, err)
	}

	// Set up default block structure if needed
	ctmpl, err = ensureDefaultBlockStructure(ctmpl, layoutName)
	if err != nil {
		return nil, err
	}

	// Process index template if it exists
	ctmpl, err = processLayoutIndexTemplate(templateFS, ctmpl, layoutsDirPath, layoutName)
	if err != nil {
		return nil, err
	}

	// Process all other template files in this layout
	for _, tmplEntry := range tmplEntries {
		tmplFileName := tmplEntry.Name()
		if tmplFileName[0] == '.' || tmplFileName == "_index.html.tmpl" {
			continue
		}

		filePathToParseRel := filepath.Join(layoutsDirPath, layoutName, tmplFileName)
		ctmpl, err = parseTemplateFile(templateFS, ctmpl, filePathToParseRel, filepath.Base(filePathToParseRel))
		if err != nil {
			return nil, fmt.Errorf("error processing file %s in template set %s: %w", tmplFileName, layoutName, err)
		}
	}

	return ctmpl, nil
}

// ensureDefaultBlockStructure ensures the template has the required block structure
func ensureDefaultBlockStructure(tmpl *template.Template, layoutName string) (*template.Template, error) {
	if tmpl.Lookup("_index.html.tmpl") == nil {
		// This parses directly into tmpl and defines templates named "_index.html.tmpl" and "_main.html.tmpl"
		// if they are not already defined at the top level of tmpl.
		parsedTmpl, err := tmpl.Parse(`{{- block "_index.html.tmpl" . -}}{{- block "_main.html.tmpl" . -}}{{- end -}}{{- end }}`)
		if err != nil {
			return nil, fmt.Errorf("error parsing fallback template for set %s: %w", layoutName, err)
		}
		return parsedTmpl, nil
	}
	return tmpl, nil
}

// processLayoutIndexTemplate handles the _index.html.tmpl file for a layout
func processLayoutIndexTemplate(templateFS fs.FS, tmpl *template.Template, layoutsDirPath string, layoutName string) (*template.Template, error) {
	namedTmplIndexRelPath := filepath.Join(layoutsDirPath, layoutName, "_index.html.tmpl")

	if _, statErr := fs.Stat(templateFS, namedTmplIndexRelPath); statErr == nil {
		return parseTemplateFile(templateFS, tmpl, namedTmplIndexRelPath, filepath.Base(namedTmplIndexRelPath))
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return nil, fmt.Errorf("error stating _index.html.tmpl for template set %s: %w", layoutName, statErr)
	}

	return tmpl, nil
}

// validateTemplateDirectory checks if the template filesystem has the correct structure
func validateTemplateDirectory(templateFS fs.FS) error {
	layoutsDir := "layouts"
	layoutsInfo, err := fs.Stat(templateFS, layoutsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: layouts directory %s does not exist in the filesystem", ErrTemplateDir, layoutsDir)
		}
		return fmt.Errorf("%w: checking layouts directory %s in the filesystem: %s", ErrTemplateDir, layoutsDir, err)
	}
	if !layoutsInfo.IsDir() {
		return fmt.Errorf("%w: path %s is not a directory in the filesystem", ErrTemplateDir, layoutsDir)
	}

	defaultsDir := filepath.Join(layoutsDir, "_defaults")
	defaultsInfo, err := fs.Stat(templateFS, defaultsDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: %s directory does not exist in the filesystem", ErrTemplateDir, defaultsDir)
		}
		return fmt.Errorf("%w: checking _defaults directory %s in the filesystem: %s", ErrTemplateDir, defaultsDir, err)
	}
	if !defaultsInfo.IsDir() {
		return fmt.Errorf("%w: path %s is not a directory in the filesystem", ErrTemplateDir, defaultsDir)
	}

	return nil
}
