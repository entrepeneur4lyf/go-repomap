// tree_walker.go

package repomap

import (
	tree_sitter "github.com/smacker/go-tree-sitter"
)

type TreeWalker2 struct {
	Nodes  [][]*tree_sitter.Node
	Scopes []map[int]struct{}
	Header [][][3]int
}

func NewTreeWalker2(numLines int) *TreeWalker2 {
	nodes := make([][]*tree_sitter.Node, numLines)
	scopes := make([]map[int]struct{}, numLines)
	header := make([][][3]int, numLines)

	// Initialize maps in scopes
	for i := range scopes {
		scopes[i] = make(map[int]struct{})
	}

	return &TreeWalker2{
		Nodes:  nodes,
		Scopes: scopes,
		Header: header,
	}
}

func (tw *TreeWalker2) GetAllTrueNodes() [][]*tree_sitter.Node {
	return tw.Nodes
}

func (tw *TreeWalker2) GetNodesForLine(line int) []*tree_sitter.Node {
	if line < 0 || line >= len(tw.Nodes) {
		return nil
	}
	return tw.Nodes[line]
}

func (tw *TreeWalker2) Walk(cursor *tree_sitter.TreeCursor) {
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
		if startLine < 0 || startLine >= len(tw.Nodes) {
			return
		}

		tw.Nodes[startLine] = append(tw.Nodes[startLine], node)

		if size > 0 {
			tw.Header[startLine] = append(tw.Header[startLine], [3]int{size, startLine, endLine})
		}

		for i := startLine; i <= endLine && i < len(tw.Nodes); i++ {
			tw.Scopes[i][startLine] = struct{}{}
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

type TreeWalker struct {
	Scopes []map[int]struct{}
	Header [][][3]int
	Nodes  [][]*tree_sitter.Node
	Tree   *tree_sitter.Tree
}

func NewTreeWalker(tree *tree_sitter.Tree, numLines int) *TreeWalker {
	nodes := make([][]*tree_sitter.Node, numLines)
	scopes := make([]map[int]struct{}, numLines)
	header := make([][][3]int, numLines)

	// Initialize maps in scopes
	for i := range scopes {
		scopes[i] = make(map[int]struct{})
	}

	return &TreeWalker{
		Scopes: scopes,
		Header: header,
		Nodes:  nodes,
		Tree:   tree,
	}
}

func (tw *TreeWalker) GetTree() *tree_sitter.Tree {
	return tw.Tree
}

func (tw *TreeWalker) WalkTree(node *tree_sitter.Node) {
	if node == nil {
		return
	}

	startLine := int(node.StartPoint().Row)
	endLine := int(node.EndPoint().Row)
	size := endLine - startLine

	// Bounds check
	if startLine < 0 || startLine >= len(tw.Nodes) {
		return
	}

	tw.Nodes[startLine] = append(tw.Nodes[startLine], node)

	if size > 0 {
		tw.Header[startLine] = append(tw.Header[startLine], [3]int{size, startLine, endLine})
	}

	for i := startLine; i <= endLine && i < len(tw.Nodes); i++ {
		tw.Scopes[i][startLine] = struct{}{}
	}

	cursor := tree_sitter.NewTreeCursor(node)
	if cursor == nil {
		return
	}

	if cursor.GoToFirstChild() {
		for {
			tw.WalkTree(cursor.CurrentNode())
			if !cursor.GoToNextSibling() {
				break
			}
		}
		cursor.GoToParent()
	}
}

func (tw *TreeWalker) GetScopes() []map[int]struct{} {
	return tw.Scopes
}

func (tw *TreeWalker) GetHeaders() [][][3]int {
	return tw.Header
}

func (tw *TreeWalker) GetNodes() [][]*tree_sitter.Node {
	return tw.Nodes
}
