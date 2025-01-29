// tag.go

package repomap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	tree_sitter "github.com/smacker/go-tree-sitter"
)

type Tag struct {
	RelFname string
	Fname    string
	Line     int
	Name     string
	Kind     TagKind
}

// String implements the Stringer interface for Tag
func (t Tag) String() string {
	return fmt.Sprintf("%s:%d - %s (%s)", t.RelFname, t.Line, t.Name, t.Kind)
}

type TagKind int

const (
	Definition TagKind = iota
	Reference
)

func (k TagKind) String() string {
	switch k {
	case Definition:
		return "Definition"
	case Reference:
		return "Reference"
	default:
		return "Unknown"
	}
}

type TagIndex struct {
	Defines     map[string]map[string]struct{}
	References  map[string][]string
	Definitions map[string][]Tag
	CommonTags  map[string]struct{}
	FileToTags  map[string]map[string]struct{}
	Path        string
	mu          sync.Mutex
}

func NewTagIndex(path string) *TagIndex {
	return &TagIndex{
		Defines:     make(map[string]map[string]struct{}),
		References:  make(map[string][]string),
		Definitions: make(map[string][]Tag),
		CommonTags:  make(map[string]struct{}),
		FileToTags:  make(map[string]map[string]struct{}),
		Path:        path,
	}
}

// GetFiles returns a map of file paths to their contents
func (ti *TagIndex) GetFiles(dir string) (map[string][]byte, error) {
	files := make(map[string][]byte)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[path] = content
		}
		return nil
	})
	return files, err
}

// Query patterns for different languages
const (
	goQuery = `
		(function_declaration 
			name: (identifier) @def.function)
		(method_declaration 
			receiver: (parameter_list) @method.receiver
			name: (field_identifier) @def.method)
		(type_declaration 
			(type_spec 
				name: (type_identifier) @def.type))
		(identifier) @ref.ident
		(field_identifier) @ref.field
	`
	jsQuery = `
		(function_declaration 
			name: (identifier) @def.function)
		(method_definition 
			name: (property_identifier) @def.method)
		(class_declaration 
			name: (identifier) @def.class)
		(identifier) @ref.ident
		(property_identifier) @ref.prop
	`
)

// GenerateFromFiles generates tags from the given files
func (ti *TagIndex) GenerateFromFiles(ctx context.Context, files map[string][]byte) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	for path, content := range files {
		// Skip non-source files
		ext := filepath.Ext(path)
		if _, ok := tsLanguages[strings.TrimPrefix(ext, ".")]; !ok {
			continue
		}

		parser := tree_sitter.NewParser()
		lang := tsLanguages[strings.TrimPrefix(ext, ".")]
		parser.SetLanguage(lang)

		tree, err := parser.ParseCtx(ctx, nil, content)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Select query based on file extension
		var queryStr string
		switch strings.TrimPrefix(ext, ".") {
		case "go":
			queryStr = goQuery
		case "js", "ts", "jsx", "tsx":
			queryStr = jsQuery
		default:
			continue
		}

		query, err := tree_sitter.NewQuery([]byte(queryStr), lang)
		if err != nil {
			return fmt.Errorf("failed to create query for %s: %w", path, err)
		}

		cursor := tree_sitter.NewQueryCursor()
		cursor.Exec(query, tree.RootNode())

		// Make path relative to the index path
		relPath, err := filepath.Rel(ti.Path, path)
		if err != nil {
			relPath = path
		}

		for {
			match, ok := cursor.NextMatch()
			if !ok {
				break
			}

			for _, capture := range match.Captures {
				patternName := query.CaptureNameForId(capture.Index)
				parts := strings.SplitN(patternName, ".", 2)
				if len(parts) != 2 {
					continue
				}

				kind := parts[0]
				name := string(capture.Node.Content(content))

				// Skip empty names and special characters
				if name == "" || strings.ContainsAny(name, "()[]{}") {
					continue
				}

				tag := Tag{
					RelFname: relPath,
					Fname:    path,
					Line:     int(capture.Node.StartPoint().Row) + 1, // Convert to 1-based line numbers
					Name:     name,
					Kind:     Definition,
				}

				if kind == "ref" {
					tag.Kind = Reference
				}

				ti.AddTag(tag, relPath)
			}
		}
	}

	// Process tags after all files have been processed
	ti.PostProcessTags()

	return nil
}

func (ti *TagIndex) AddTag(tag Tag, relPath string) {
	switch tag.Kind {
	case Definition:
		if _, ok := ti.Defines[tag.Name]; !ok {
			ti.Defines[tag.Name] = make(map[string]struct{})
		}
		ti.Defines[tag.Name][relPath] = struct{}{}
		ti.Definitions[filepath.Join(relPath, tag.Name)] = append(ti.Definitions[filepath.Join(relPath, tag.Name)], tag)

		if _, ok := ti.FileToTags[relPath]; !ok {
			ti.FileToTags[relPath] = make(map[string]struct{})
		}
		ti.FileToTags[relPath][tag.Name] = struct{}{}
	case Reference:
		ti.References[tag.Name] = append(ti.References[tag.Name], relPath)

		if _, ok := ti.FileToTags[relPath]; !ok {
			ti.FileToTags[relPath] = make(map[string]struct{})
		}
		ti.FileToTags[relPath][tag.Name] = struct{}{}
	}
}

func (ti *TagIndex) PostProcessTags() {
	ti.processEmptyReferences()
	ti.processCommonTags()
}

func (ti *TagIndex) processEmptyReferences() {
	if len(ti.References) == 0 {
		for k, v := range ti.Defines {
			ti.References[k] = make([]string, 0, len(v))
			for path := range v {
				ti.References[k] = append(ti.References[k], path)
			}
		}
	}
}

func (ti *TagIndex) processCommonTags() {
	for key := range ti.Defines {
		if _, ok := ti.References[key]; ok {
			ti.CommonTags[key] = struct{}{}
		}
	}
}
