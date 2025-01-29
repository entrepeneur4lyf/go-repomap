package main

import (
	"context"
	"fmt"
	"os"

	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

func main() {
	// Example Go code to parse
	code := []byte(`
package main

func add(a, b int) int {
    return a + b
}

type User struct {
    Name string
    Age  int
}

func (u *User) GetName() string {
    return u.Name
}
`)

	// Initialize parser
	parser := tree_sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

	// Parse the code
	tree, err := parser.ParseCtx(context.Background(), nil, code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing code: %v\n", err)
		os.Exit(1)
	}

	// Example 1: Find all function definitions
	functionQuery := `(function_declaration) @function`
	findFunctions(tree, code, functionQuery)

	// Example 2: Find method definitions
	methodQuery := `(method_declaration) @method`
	findMethods(tree, code, methodQuery)

	// Example 3: Find struct definitions with fields
	structQuery := `
		(type_declaration 
			(type_spec 
				name: (type_identifier) @struct_name
				type: (struct_type 
					(field_declaration_list) @fields)))
	`
	findStructs(tree, code, structQuery)

	// Example 4: Find function parameters
	paramsQuery := `
		(function_declaration
			name: (identifier) @func_name
			parameters: (parameter_list
				(parameter_declaration
					name: (identifier) @param_name
					type: (type_identifier) @param_type)))
	`
	findFunctionParams(tree, code, paramsQuery)
}

func findFunctions(tree *tree_sitter.Tree, code []byte, query string) {
	q, err := tree_sitter.NewQuery([]byte(query), golang.GetLanguage())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating query: %v\n", err)
		return
	}

	qc := tree_sitter.NewQueryCursor()
	qc.Exec(q, tree.RootNode())

	fmt.Println("\nFunctions found:")
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, c := range m.Captures {
			fmt.Printf("- %s\n", c.Node.Content(code))
		}
	}
}

func findMethods(tree *tree_sitter.Tree, code []byte, query string) {
	q, err := tree_sitter.NewQuery([]byte(query), golang.GetLanguage())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating query: %v\n", err)
		return
	}

	qc := tree_sitter.NewQueryCursor()
	qc.Exec(q, tree.RootNode())

	fmt.Println("\nMethods found:")
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, c := range m.Captures {
			fmt.Printf("- %s\n", c.Node.Content(code))
		}
	}
}

func findStructs(tree *tree_sitter.Tree, code []byte, query string) {
	q, err := tree_sitter.NewQuery([]byte(query), golang.GetLanguage())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating query: %v\n", err)
		return
	}

	qc := tree_sitter.NewQueryCursor()
	qc.Exec(q, tree.RootNode())

	fmt.Println("\nStructs found:")
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, c := range m.Captures {
			// Get the pattern index that matched this capture
			patternIdx := q.CaptureNameForId(c.Index)
			switch patternIdx {
			case "struct_name":
				fmt.Printf("Struct: %s\n", c.Node.Content(code))
			case "fields":
				fmt.Printf("Fields: %s\n", c.Node.Content(code))
			}
		}
	}
}

func findFunctionParams(tree *tree_sitter.Tree, code []byte, query string) {
	q, err := tree_sitter.NewQuery([]byte(query), golang.GetLanguage())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating query: %v\n", err)
		return
	}

	qc := tree_sitter.NewQueryCursor()
	qc.Exec(q, tree.RootNode())

	fmt.Println("\nFunction parameters:")
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		var funcName, paramName, paramType string
		for _, c := range m.Captures {
			// Get the pattern index that matched this capture
			patternIdx := q.CaptureNameForId(c.Index)
			switch patternIdx {
			case "func_name":
				funcName = string(c.Node.Content(code))
			case "param_name":
				paramName = string(c.Node.Content(code))
			case "param_type":
				paramType = string(c.Node.Content(code))
			}
		}
		if funcName != "" && paramName != "" && paramType != "" {
			fmt.Printf("- Function %s has parameter %s of type %s\n", funcName, paramName, paramType)
		}
	}
}
