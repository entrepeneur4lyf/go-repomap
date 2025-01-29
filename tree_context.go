// tree_context.go

package repomap

import (
	"fmt"
	"sort"
	"strings"

	tree_sitter "github.com/smacker/go-tree-sitter"
)

type TreeContext struct {
	Filename                 string
	ParentContext            bool
	ChildContext             bool
	LastLine                 bool
	Margin                   int
	MarkLois                 bool
	HeaderMax                int
	ShowTopOfFileParentScope bool
	LoiPad                   int
	Lois                     map[int]struct{}
	ShowLines                map[int]struct{}
	NumLines                 int
	Lines                    []string
	LineNumber               bool
	DoneParentScopes         map[int]struct{}
	Nodes                    [][]*tree_sitter.Node
	Scopes                   []map[int]struct{}
	Header                   [][][3]int
}

func NewTreeContext(code string, fsFilePath string) *TreeContext {
	lines := strings.Split(code, "\n")
	numLines := len(lines) + 1

	// Initialize all slices and maps
	nodes := make([][]*tree_sitter.Node, numLines)
	scopes := make([]map[int]struct{}, numLines)
	header := make([][][3]int, numLines)

	// Initialize maps in scopes
	for i := range scopes {
		scopes[i] = make(map[int]struct{})
	}

	return &TreeContext{
		Filename:                 fsFilePath,
		ParentContext:            true,
		ChildContext:             false,
		LastLine:                 false,
		Margin:                   0,
		MarkLois:                 false,
		HeaderMax:                10,
		ShowTopOfFileParentScope: false,
		LoiPad:                   0,
		Lois:                     make(map[int]struct{}),
		ShowLines:                make(map[int]struct{}),
		NumLines:                 numLines,
		Lines:                    lines,
		LineNumber:               false,
		DoneParentScopes:         make(map[int]struct{}),
		Nodes:                    nodes,
		Scopes:                   scopes,
		Header:                   header,
	}
}

func (tc *TreeContext) Init(cursor *tree_sitter.TreeCursor) {
	if cursor == nil {
		return
	}
	tc.Walk(cursor)
	tc.ArrangeHeaders()
}

func (tc *TreeContext) Walk(cursor *tree_sitter.TreeCursor) {
	if cursor == nil {
		return
	}

	for {
		node := cursor.CurrentNode()
		if node == nil {
			return
		}

		startLine := int(node.StartPoint().Row)
		endLine := int(node.EndPoint().Row)
		size := endLine - startLine

		// Bounds check
		if startLine < 0 || startLine >= tc.NumLines {
			return
		}

		tc.Nodes[startLine] = append(tc.Nodes[startLine], node)

		if size > 0 {
			tc.Header[startLine] = append(tc.Header[startLine], [3]int{size, startLine, endLine})
		}

		for i := startLine; i <= endLine && i < tc.NumLines; i++ {
			tc.Scopes[i][startLine] = struct{}{}
		}

		if cursor.GoToFirstChild() {
			continue
		}

		if cursor.GoToNextSibling() {
			continue
		}

		for {
			if !cursor.GoToParent() {
				return
			}

			if cursor.GoToNextSibling() {
				break
			}
		}
	}
}

func (tc *TreeContext) GetLois() map[int]struct{} {
	return tc.Lois
}

func (tc *TreeContext) AddLois(lois []int) {
	for _, loi := range lois {
		if loi >= 0 && loi < tc.NumLines {
			tc.Lois[loi] = struct{}{}
		}
	}
}

func (tc *TreeContext) PrintState() {
	for lineNumber, values := range tc.Scopes {
		if len(values) > 0 {
			var scopeValues []string
			for value := range values {
				scopeValues = append(scopeValues, fmt.Sprintf("%d", value+1))
			}
			sort.Strings(scopeValues)
			fmt.Printf("scope::(%d)::(%s)\n", lineNumber+1, strings.Join(scopeValues, ","))
		}
	}
}

func (tc *TreeContext) AddContext() {
	if len(tc.Lois) == 0 {
		return
	}

	// Copy Lois to ShowLines
	for k := range tc.Lois {
		tc.ShowLines[k] = struct{}{}
	}

	if tc.LoiPad > 0 {
		for line := range tc.ShowLines {
			start := max(0, line-tc.LoiPad)
			end := min(tc.NumLines-1, line+tc.LoiPad)
			for newLine := start; newLine <= end; newLine++ {
				tc.ShowLines[newLine] = struct{}{}
			}
		}
	}

	if tc.LastLine && tc.NumLines > 2 {
		bottomLine := tc.NumLines - 2
		tc.ShowLines[bottomLine] = struct{}{}
		tc.AddParentScopes(bottomLine, nil)
	}

	if tc.ParentContext {
		for index := range tc.Lois {
			tc.AddParentScopes(index, nil)
		}
	}

	if tc.ChildContext {
		for index := range tc.Lois {
			tc.AddChildContext(index)
		}
	}

	if tc.Margin > 0 {
		for i := 0; i < min(tc.Margin, tc.NumLines); i++ {
			tc.ShowLines[i] = struct{}{}
		}
	}

	tc.CloseSmallGaps()
}

func (tc *TreeContext) CloseSmallGaps() {
	tc.ShowLines = CloseSmallGapsHelper(tc.ShowLines, tc.Lines, tc.NumLines)
}

func (tc *TreeContext) AddChildContext(index int) {
	if index < 0 || index >= tc.NumLines || len(tc.Nodes[index]) == 0 {
		return
	}

	lastLine := tc.GetLastLineOfScope(index)
	size := lastLine - index

	if size < 5 {
		for i := index; i <= lastLine && i < tc.NumLines; i++ {
			tc.ShowLines[i] = struct{}{}
		}
		return
	}

	var children []*tree_sitter.Node
	for _, node := range tc.Nodes[index] {
		if node != nil {
			children = append(children, tc.FindAllChildren(node)...)
		}
	}

	sort.Slice(children, func(i, j int) bool {
		if children[i] == nil || children[j] == nil {
			return false
		}
		return (children[i].EndPoint().Row - children[i].StartPoint().Row) >
			(children[j].EndPoint().Row - children[j].StartPoint().Row)
	})

	currentlyShowing := len(tc.ShowLines)
	maxToShow := max(min(int(float64(size)*0.10), 25), 5)

	for _, child := range children {
		if child == nil {
			continue
		}
		if len(tc.ShowLines) > currentlyShowing+maxToShow {
			return
		}
		childStartLine := int(child.StartPoint().Row)
		if childStartLine >= 0 && childStartLine < tc.NumLines {
			tc.AddParentScopes(childStartLine, nil)
		}
	}
}

func (tc *TreeContext) FindAllChildren(node *tree_sitter.Node) []*tree_sitter.Node {
	if node == nil {
		return nil
	}

	var children []*tree_sitter.Node
	cursor := tree_sitter.NewTreeCursor(node)
	if cursor == nil {
		return nil
	}

	if cursor.GoToFirstChild() {
		children = append(children, cursor.CurrentNode())
		for cursor.GoToNextSibling() {
			if n := cursor.CurrentNode(); n != nil {
				children = append(children, n)
			}
		}
	}

	return children
}

func (tc *TreeContext) GetLastLineOfScope(index int) int {
	if index < 0 || index >= tc.NumLines {
		return index
	}

	lastLine := index
	for _, node := range tc.Nodes[index] {
		if node != nil {
			endLine := int(node.EndPoint().Row)
			if endLine > lastLine {
				lastLine = endLine
			}
		}
	}
	return min(lastLine, tc.NumLines-1)
}

func (tc *TreeContext) Format() string {
	if len(tc.ShowLines) == 0 {
		return ""
	}

	var output strings.Builder
	dots := false
	if _, ok := tc.ShowLines[0]; !ok {
		dots = true
	}

	for index, line := range tc.Lines {
		if _, ok := tc.ShowLines[index]; !ok {
			if dots {
				if tc.LineNumber {
					output.WriteString("...⋮...\n")
				} else {
					output.WriteString("⋮...\n")
				}
				dots = false
			}
			continue
		}

		spacer := "|"
		if _, ok := tc.Lois[index]; ok && tc.MarkLois {
			spacer = "█"
		}

		output.WriteString(fmt.Sprintf("%s%s\n", spacer, line))
		dots = true
	}

	return output.String()
}

func (tc *TreeContext) AddParentScopes(index int, recurseDepth []int) {
	if index < 0 || index >= tc.NumLines {
		return
	}

	if _, ok := tc.DoneParentScopes[index]; ok {
		return
	}

	tc.DoneParentScopes[index] = struct{}{}

	for lineNum := range tc.Scopes[index] {
		if lineNum >= len(tc.Header) || len(tc.Header[lineNum]) == 0 {
			continue
		}

		headStart, headEnd := tc.Header[lineNum][0][1], tc.Header[lineNum][0][2]
		if headStart >= 0 && headEnd < tc.NumLines && (headStart > 0 || tc.ShowTopOfFileParentScope) {
			for i := headStart; i < headEnd; i++ {
				tc.ShowLines[i] = struct{}{}
			}
		}

		if tc.LastLine {
			newRecurseDepth := make([]int, len(recurseDepth)+1)
			copy(newRecurseDepth, recurseDepth)
			newRecurseDepth[len(recurseDepth)] = index
			lastLine := tc.GetLastLineOfScope(lineNum)
			tc.AddParentScopes(lastLine, newRecurseDepth)
		}
	}
}

func (tc *TreeContext) ArrangeHeaders() {
	for lineNumber := 0; lineNumber < tc.NumLines; lineNumber++ {
		if len(tc.Header[lineNumber]) == 0 {
			tc.Header[lineNumber] = [][3]int{{0, lineNumber, lineNumber + 1}}
			continue
		}

		sort.Slice(tc.Header[lineNumber], func(i, j int) bool {
			return tc.Header[lineNumber][i][0] < tc.Header[lineNumber][j][0]
		})

		var startEnd [2]int
		if len(tc.Header[lineNumber]) > 1 {
			size := tc.Header[lineNumber][0][0]
			start := tc.Header[lineNumber][0][1]
			end := tc.Header[lineNumber][0][2]

			if size > tc.HeaderMax {
				startEnd = [2]int{start, start + tc.HeaderMax}
			} else {
				startEnd = [2]int{start, end}
			}
		} else {
			startEnd = [2]int{lineNumber, lineNumber + 1}
		}

		tc.Header[lineNumber] = [][3]int{{0, startEnd[0], startEnd[1]}}
	}
}
