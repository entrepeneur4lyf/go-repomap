// types.go

package repomap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/css"
	"github.com/smacker/go-tree-sitter/dockerfile"
	"github.com/smacker/go-tree-sitter/elixir"
	"github.com/smacker/go-tree-sitter/elm"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/groovy"
	"github.com/smacker/go-tree-sitter/hcl"
	"github.com/smacker/go-tree-sitter/html"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/kotlin"
	"github.com/smacker/go-tree-sitter/ocaml"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/protobuf"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/scala"
	"github.com/smacker/go-tree-sitter/sql"
	"github.com/smacker/go-tree-sitter/svelte"
	"github.com/smacker/go-tree-sitter/swift"
	"github.com/smacker/go-tree-sitter/toml"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/smacker/go-tree-sitter/yaml"
)

var tsLanguages map[string]*tree_sitter.Language

func init() {
	tsLanguages = make(map[string]*tree_sitter.Language)

	// Initialize all supported languages
	tsLanguages["bash"] = bash.GetLanguage()
	tsLanguages["c"] = c.GetLanguage()
	tsLanguages["cpp"] = cpp.GetLanguage()
	tsLanguages["cs"] = csharp.GetLanguage()
	tsLanguages["csharp"] = csharp.GetLanguage()
	tsLanguages["css"] = css.GetLanguage()
	tsLanguages["dockerfile"] = dockerfile.GetLanguage()
	tsLanguages["elixir"] = elixir.GetLanguage()
	tsLanguages["elm"] = elm.GetLanguage()
	tsLanguages["go"] = golang.GetLanguage()
	tsLanguages["golang"] = golang.GetLanguage()
	tsLanguages["groovy"] = groovy.GetLanguage()
	tsLanguages["hcl"] = hcl.GetLanguage()
	tsLanguages["html"] = html.GetLanguage()
	tsLanguages["java"] = java.GetLanguage()
	tsLanguages["javascript"] = javascript.GetLanguage()
	tsLanguages["js"] = javascript.GetLanguage()
	tsLanguages["kt"] = kotlin.GetLanguage()
	tsLanguages["kotlin"] = kotlin.GetLanguage()
	tsLanguages["ml"] = ocaml.GetLanguage()
	tsLanguages["ocaml"] = ocaml.GetLanguage()
	tsLanguages["php"] = php.GetLanguage()
	tsLanguages["proto"] = protobuf.GetLanguage()
	tsLanguages["protobuf"] = protobuf.GetLanguage()
	tsLanguages["py"] = python.GetLanguage()
	tsLanguages["python"] = python.GetLanguage()
	tsLanguages["rb"] = ruby.GetLanguage()
	tsLanguages["ruby"] = ruby.GetLanguage()
	tsLanguages["rs"] = rust.GetLanguage()
	tsLanguages["rust"] = rust.GetLanguage()
	tsLanguages["scala"] = scala.GetLanguage()
	tsLanguages["sql"] = sql.GetLanguage()
	tsLanguages["svelte"] = svelte.GetLanguage()
	tsLanguages["swift"] = swift.GetLanguage()
	tsLanguages["toml"] = toml.GetLanguage()
	tsLanguages["ts"] = typescript.GetLanguage()
	tsLanguages["tsx"] = typescript.GetLanguage()
	tsLanguages["typescript"] = typescript.GetLanguage()
	tsLanguages["yaml"] = yaml.GetLanguage()
	tsLanguages["yml"] = yaml.GetLanguage()
}

const REPOMAP_DEFAULT_TOKENS = 1024

type RepoMap struct {
	MapTokens int
}

func NewRepoMap() *RepoMap {
	return &RepoMap{
		MapTokens: REPOMAP_DEFAULT_TOKENS,
	}
}

func (rm *RepoMap) WithMapTokens(mapTokens int) *RepoMap {
	rm.MapTokens = mapTokens
	return rm
}

func (rm *RepoMap) GetRepoMap(tagIndex *TagIndex) (string, error) {
	repomap, err := rm.getRankedTagsMap(rm.MapTokens, tagIndex)
	if err != nil {
		return "", err
	}

	if repomap == "" {
		return "", NewTreeGenerationError("No tree generated")
	}

	fmt.Printf("Repomap: %dk tokens\n", rm.getTokenCount(repomap)/1024)

	return repomap, nil
}

func (rm *RepoMap) getTokenCount(tree string) int {
	chars := len(tree)

	// https://platform.openai.com/tokenizer
	tokenPerCharRatio := 0.25

	tokenEstimate := int(float64(chars) * tokenPerCharRatio)

	return tokenEstimate
}

func (rm *RepoMap) findBestTree(rankedTags []Tag, maxMapTokens int) string {
	numTags := len(rankedTags)
	fmt.Println("Initial conditions:")
	fmt.Printf("  Number of tags: %d\n", numTags)
	fmt.Printf("  Max map tokens: %d\n", maxMapTokens)

	if numTags == 0 {
		return ""
	}

	lowerBound := 0
	upperBound := numTags
	var bestTree string
	bestTreeTokens := 0
	middle := min(maxMapTokens/25, numTags)
	iteration := 0

	for lowerBound <= upperBound {
		iteration++
		fmt.Printf("\nIteration %d:\n", iteration)
		fmt.Printf("  Bounds: [%d, %d]\n", lowerBound, upperBound)
		fmt.Printf("  Middle: %d\n", middle)

		if middle == 0 {
			middle = 1
		}

		tree := rm.toTree(rankedTags[:middle])
		numTokens := rm.getTokenCount(tree)

		fmt.Printf("  Tree tokens: %d\n", numTokens)

		if numTokens < maxMapTokens && numTokens > bestTreeTokens {
			fmt.Println("  New best tree found!")
			fmt.Printf("    Previous best: %d tokens\n", bestTreeTokens)
			fmt.Printf("    New best: %d tokens\n", numTokens)
			bestTree = tree
			bestTreeTokens = numTokens
		}

		if numTokens < maxMapTokens {
			fmt.Println("  Increasing lower bound")
			lowerBound = middle + 1
		} else {
			fmt.Println("  Decreasing upper bound")
			upperBound = middle - 1
		}

		middle = (lowerBound + upperBound) / 2

		fmt.Printf("  Next middle: %d\n", middle)
	}

	fmt.Println("\nSearch completed:")
	fmt.Printf("  Best tree tokens: %d\n", bestTreeTokens)
	fmt.Printf("  Final bounds: [%d, %d]\n", lowerBound, upperBound)

	return bestTree
}

func (rm *RepoMap) getRankedTagsMap(maxMapTokens int, tagIndex *TagIndex) (string, error) {
	analyser := NewTagAnalyzer(tagIndex)

	fmt.Println("[Analyser] Ranking tags...")
	rankedTags := analyser.GetRankedTags()
	fmt.Printf("[Analyser] tags::len(%d)\n", len(rankedTags))

	fmt.Println("[Tree] Finding best tree...")
	tree := rm.findBestTree(rankedTags, maxMapTokens)

	if tree == "" && len(rankedTags) > 0 {
		// If findBestTree failed but we have tags, return a tree with all tags
		tree = rm.toTree(rankedTags)
	}

	return tree, nil
}

func (rm *RepoMap) toTree(tags []Tag) string {
	if len(tags) == 0 {
		return ""
	}

	var output strings.Builder
	var curFname string
	var lois []int

	for _, tag := range tags {
		thisRelFname := tag.RelFname

		if thisRelFname != curFname {
			if len(lois) > 0 {
				output.WriteString("\n")
				output.WriteString(tag.Fname)
				output.WriteString(":\n")
				fileContent, err := os.ReadFile(tag.Fname)
				if err != nil {
					continue
				}
				output.WriteString(rm.renderTree(tag.Fname, fileContent, lois))
			} else if curFname != "" {
				output.WriteString("\n")
				output.WriteString(tag.Fname)
				output.WriteString("\n")
			}

			lois = []int{}
			curFname = thisRelFname
		}

		if tag.Line > 0 {
			lois = append(lois, tag.Line-1) // Convert to 0-based line numbers
		}
	}

	// Handle the last file
	if len(lois) > 0 && len(tags) > 0 {
		lastTag := tags[len(tags)-1]
		output.WriteString("\n")
		output.WriteString(lastTag.Fname)
		output.WriteString(":\n")
		fileContent, err := os.ReadFile(lastTag.Fname)
		if err == nil {
			output.WriteString(rm.renderTree(lastTag.Fname, fileContent, lois))
		}
	}

	outputString := output.String()
	if outputString == "" {
		return ""
	}

	outputString = strings.Join(strings.Split(outputString, "\n"), "\n")
	outputString += "\n"

	return outputString
}

func (rm *RepoMap) renderTree(absFname string, fileContent []byte, lois []int) string {
	code := string(fileContent)
	if !strings.HasSuffix(code, "\n") {
		code += "\n"
	}

	parser := tree_sitter.NewParser()
	ext := strings.ToLower(strings.TrimPrefix(strings.ToLower(filepath.Ext(absFname)), "."))
	if lang, ok := tsLanguages[ext]; ok {
		parser.SetLanguage(lang)
	} else {
		parser.SetLanguage(tsLanguages["javascript"]) // fallback
	}

	tree, err := parser.ParseCtx(context.Background(), nil, []byte(code))
	if err != nil {
		return ""
	}

	rootNode := tree.RootNode()
	cursor := tree_sitter.NewTreeCursor(rootNode)

	context := NewTreeContext(code, absFname)
	context.Init(cursor)

	context.AddLois(lois)
	context.AddContext()

	return context.Format()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
