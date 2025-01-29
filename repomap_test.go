package repomap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tree_sitter "github.com/smacker/go-tree-sitter"
)

func TestRepoMap(t *testing.T) {
	// Get the test data directory
	testDataDir := filepath.Join("testdata", "web")

	// Create tag index
	tagIndex := NewTagIndex(testDataDir)

	// Get all Go files in the test data directory
	files := make(map[string][]byte)
	err := filepath.Walk(testDataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files[path] = content
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to read test files: %v", err)
	}

	// Generate tags
	ctx := context.Background()
	if err := tagIndex.GenerateFromFiles(ctx, files); err != nil {
		t.Fatalf("Failed to generate tags: %v", err)
	}

	// Debug print the tag index state
	t.Logf("Defines: %+v", tagIndex.Defines)
	t.Logf("References: %+v", tagIndex.References)
	t.Logf("Definitions: %+v", tagIndex.Definitions)

	// Test definitions
	expectedDefs := map[string]bool{
		"main":           false,
		"HandleUsers":    false,
		"HandlePosts":    false,
		"HandleHealth":   false,
		"NewUserService": false,
		"NewPostService": false,
		"NewAuthService": false,
		"User":           false,
		"Post":           false,
		"PostComment":    false,
		"AuthMiddleware": false,
	}

	for key := range tagIndex.Defines {
		if _, ok := expectedDefs[key]; ok {
			expectedDefs[key] = true
			t.Logf("Found definition: %s", key)
		}
	}

	for name, found := range expectedDefs {
		if !found {
			t.Errorf("Expected definition not found: %s", name)
		}
	}

	// Test references
	expectedRefs := map[string]bool{
		"http":        false,
		"log":         false,
		"json":        false,
		"time":        false,
		"sync":        false,
		"HandleUsers": false,
		"HandlePosts": false,
		"User":        false,
		"Post":        false,
	}

	for key := range tagIndex.References {
		if _, ok := expectedRefs[key]; ok {
			expectedRefs[key] = true
			t.Logf("Found reference: %s", key)
		}
	}

	// Create analyzer
	analyzer := NewTagAnalyzer(tagIndex)

	// Get ranked tags
	rankedTags := analyzer.GetRankedTags()
	if len(rankedTags) == 0 {
		t.Error("Expected ranked tags got none")
	} else {
		t.Logf("Got %d ranked tags", len(rankedTags))
		for _, tag := range rankedTags {
			t.Logf("Ranked tag: %s", tag)
		}
	}

	// Test repo map generation
	rm := NewRepoMap()
	repomap, err := rm.GetRepoMap(tagIndex)
	if err != nil {
		t.Fatalf("Failed to get repo map: %v", err)
	}
	if repomap == "" {
		t.Error("Expected non-empty repo map")
	} else {
		t.Logf("Generated repo map:\n%s", repomap)
	}
}

func TestLanguageSupport(t *testing.T) {
	// Test that we support all advertised languages
	expectedLanguages := []string{
		"go", "javascript", "typescript", "python", "java",
		"ruby", "rust", "cpp", "csharp", "php",
	}

	for _, lang := range expectedLanguages {
		if _, ok := tsLanguages[lang]; !ok {
			t.Errorf("Expected support for %s language", lang)
		}
	}
}

func TestFileExtensionMapping(t *testing.T) {
	testCases := []struct {
		filename string
		ext      string
	}{
		{"test.go", "go"},
		{"test.js", "js"},
		{"test.ts", "ts"},
		{"test.py", "py"},
		{"test.java", "java"},
		{"test.rb", "rb"},
		{"test.rs", "rs"},
		{"test.cpp", "cpp"},
		{"test.cs", "cs"},
		{"test.php", "php"},
	}

	parser := tree_sitter.NewParser()

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			ext := strings.TrimPrefix(filepath.Ext(tc.filename), ".")
			lang, ok := tsLanguages[ext]
			if !ok {
				t.Errorf("No language support for extension %s", ext)
				return
			}

			parser.SetLanguage(lang)
			tree, err := parser.ParseCtx(context.Background(), nil, []byte("// Test file"))
			if err != nil {
				t.Errorf("Failed to parse %s: %v", tc.filename, err)
			}
			if tree == nil {
				t.Errorf("Expected parse tree for %s, got nil", tc.filename)
			}
		})
	}
}

func TestQueryParsing(t *testing.T) {
	// Test that our query patterns are valid
	testCases := []struct {
		name  string
		query string
		lang  string
	}{
		{"Go", goQuery, "go"},
		{"JavaScript", jsQuery, "javascript"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lang := tsLanguages[tc.lang]
			if lang == nil {
				t.Fatalf("Language %s not supported", tc.lang)
			}

			_, err := tree_sitter.NewQuery([]byte(tc.query), lang)
			if err != nil {
				t.Errorf("Failed to parse query for %s: %v", tc.name, err)
			}
		})
	}
}
