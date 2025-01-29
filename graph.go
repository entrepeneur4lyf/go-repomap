// graph.go

package repomap

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type NodeIndex int
type EdgeIndex int

type Edge struct {
	Source NodeIndex
	Target NodeIndex
	Weight float64
	Index  EdgeIndex
}

type DiGraph struct {
	Nodes []string
	Edges map[NodeIndex][]Edge
}

func NewDiGraph() *DiGraph {
	return &DiGraph{
		Nodes: make([]string, 0),
		Edges: make(map[NodeIndex][]Edge),
	}
}

func (g *DiGraph) AddNode(label string) NodeIndex {
	g.Nodes = append(g.Nodes, label)
	return NodeIndex(len(g.Nodes) - 1)
}

func (g *DiGraph) AddEdge(source, target NodeIndex, weight float64) EdgeIndex {
	edge := Edge{
		Source: source,
		Target: target,
		Weight: weight,
		Index:  EdgeIndex(len(g.Edges[source])),
	}
	g.Edges[source] = append(g.Edges[source], edge)
	return edge.Index
}

func (g *DiGraph) NumNodes() int {
	return len(g.Nodes)
}

func (g *DiGraph) NumEdges() int {
	count := 0
	for _, edges := range g.Edges {
		count += len(edges)
	}
	return count
}

type RankedDefinition struct {
	Key   NodeIndex
	Value float64
}

type RankedDefinitionsMap map[NodeIndex]float64

type TagGraph struct {
	graph             *DiGraph
	nodeIndices       map[string]NodeIndex
	edgeToIdent       map[EdgeIndex]string
	rankedDefinitions RankedDefinitionsMap
	sortedDefinitions []RankedDefinition
}

func NewTagGraph() *TagGraph {
	return &TagGraph{
		graph:             NewDiGraph(),
		nodeIndices:       make(map[string]NodeIndex),
		edgeToIdent:       make(map[EdgeIndex]string),
		rankedDefinitions: make(RankedDefinitionsMap),
		sortedDefinitions: []RankedDefinition{},
	}
}

func (tg *TagGraph) GetGraph() *DiGraph {
	return tg.graph
}

func NewTagGraphFromTagIndex(tagIndex *TagIndex, mentionedIdents map[string]struct{}) *TagGraph {
	tagGraph := NewTagGraph()
	tagGraph.PopulateFromTagIndex(tagIndex, mentionedIdents)
	return tagGraph
}

func (tg *TagGraph) PopulateFromTagIndex(tagIndex *TagIndex, mentionedIdents map[string]struct{}) {
	if mentionedIdents == nil {
		mentionedIdents = make(map[string]struct{})
	}

	// First, create nodes for all files that contain definitions or references
	for path := range tagIndex.FileToTags {
		tg.getOrCreateNode(path)
	}

	// Then create edges based on references
	for ident := range tagIndex.CommonTags {
		mul := tg.calculateMultiplier(ident, mentionedIdents)
		numRefs := float64(len(tagIndex.References[ident]))
		scaledRefs := math.Sqrt(numRefs)

		// For each reference to this identifier
		for _, referencer := range tagIndex.References[ident] {
			// For each file that defines this identifier
			if defines, ok := tagIndex.Defines[ident]; ok {
				for definer := range defines {
					// Skip self-references
					if referencer == definer {
						continue
					}

					referencerIdx := tg.getOrCreateNode(referencer)
					definerIdx := tg.getOrCreateNode(definer)

					// Create an edge from the referencer to the definer
					edgeIndex := tg.graph.AddEdge(referencerIdx, definerIdx, mul*scaledRefs)
					tg.edgeToIdent[edgeIndex] = ident
				}
			}
		}
	}
}

func (tg *TagGraph) CalculatePageRanks() []float64 {
	numNodes := tg.graph.NumNodes()
	if numNodes == 0 {
		return nil
	}

	ranks := make([]float64, numNodes)
	for i := range ranks {
		ranks[i] = 1.0 / float64(numNodes)
	}

	damping := 0.85
	for i := 0; i < 100; i++ {
		newRanks := make([]float64, numNodes)
		for node := range ranks {
			newRanks[node] = (1 - damping) / float64(numNodes)
			for _, edge := range tg.graph.Edges[NodeIndex(node)] {
				newRanks[node] += damping * ranks[edge.Target] * edge.Weight
			}
		}
		ranks = newRanks
	}

	return ranks
}

func (tg *TagGraph) GetRankedDefinitions() RankedDefinitionsMap {
	return tg.rankedDefinitions
}

func (tg *TagGraph) DebugRankedDefinitions() {
	for nodeIndex, rank := range tg.rankedDefinitions {
		fmt.Printf("%s: %f\n", tg.graph.Nodes[nodeIndex], rank)
	}
}

func (tg *TagGraph) DebugSortedDefinitions() {
	for _, def := range tg.sortedDefinitions {
		fmt.Printf("%s: %f\n", tg.graph.Nodes[def.Key], def.Value)
	}
}

func (tg *TagGraph) GetSortedDefinitions() []RankedDefinition {
	return tg.sortedDefinitions
}

func (tg *TagGraph) CalculateAndDistributeRanks() {
	ranks := tg.CalculatePageRanks()
	if ranks == nil {
		return
	}
	tg.distributeRank(ranks)
	tg.sortByRank()
}

func (tg *TagGraph) sortByRank() {
	var vec []RankedDefinition
	for k, v := range tg.rankedDefinitions {
		vec = append(vec, RankedDefinition{Key: k, Value: v})
	}

	sort.Slice(vec, func(i, j int) bool {
		return vec[i].Value > vec[j].Value
	})

	tg.sortedDefinitions = vec
}

func (tg *TagGraph) distributeRank(ranks []float64) {
	for src := range tg.graph.Nodes {
		srcRank := ranks[src]
		totalOutgoingWeights := 0.0
		for _, edge := range tg.graph.Edges[NodeIndex(src)] {
			totalOutgoingWeights += edge.Weight
		}

		if totalOutgoingWeights == 0 {
			// If a node has no outgoing edges, distribute its rank to itself
			tg.rankedDefinitions[NodeIndex(src)] = srcRank
			continue
		}

		for _, edge := range tg.graph.Edges[NodeIndex(src)] {
			destination := edge.Target
			weight := edge.Weight
			newWeight := srcRank * weight / totalOutgoingWeights

			if _, ok := tg.rankedDefinitions[destination]; !ok {
				tg.rankedDefinitions[destination] = 0.0
			}
			tg.rankedDefinitions[destination] += newWeight
		}
	}

	// Ensure all nodes have a rank, even if they have no incoming edges
	for i := range tg.graph.Nodes {
		if _, ok := tg.rankedDefinitions[NodeIndex(i)]; !ok {
			tg.rankedDefinitions[NodeIndex(i)] = 0.0
		}
	}
}

func (tg *TagGraph) GenerateDotRepresentation() string {
	var dot strings.Builder
	dot.WriteString("digraph {\n")

	for nodeIndex, nodeLabel := range tg.graph.Nodes {
		dot.WriteString(fmt.Sprintf("    %d [ label = \"%s\" ]\n", nodeIndex, nodeLabel))
	}

	for _, edges := range tg.graph.Edges {
		for _, edge := range edges {
			dot.WriteString(fmt.Sprintf("    %d -> %d [ label = \"%f\" ]\n", edge.Source, edge.Target, edge.Weight))
		}
	}

	dot.WriteString("}\n")
	return dot.String()
}

func (tg *TagGraph) PrintDot() {
	fmt.Println(tg.GenerateDotRepresentation())
}

func (tg *TagGraph) getOrCreateNode(name string) NodeIndex {
	if idx, ok := tg.nodeIndices[name]; ok {
		return idx
	}
	idx := tg.graph.AddNode(name)
	tg.nodeIndices[name] = idx
	return idx
}

func (tg *TagGraph) calculateMultiplier(tag string, mentionedIdents map[string]struct{}) float64 {
	if _, ok := mentionedIdents[tag]; ok {
		return 10.0
	} else if strings.HasPrefix(tag, "_") {
		return 0.1
	} else {
		return 1.0
	}
}
