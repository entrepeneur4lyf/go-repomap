// analyser.go

package repomap

import (
	"fmt"
	"sort"
)

type TagAnalyzer struct {
	tagIndex *TagIndex
	tagGraph *TagGraph
}

func NewTagAnalyzer(tagIndex *TagIndex) *TagAnalyzer {
	// Create a map of mentioned identifiers from the references
	mentionedIdents := make(map[string]struct{})
	for ident := range tagIndex.References {
		mentionedIdents[ident] = struct{}{}
	}

	tagGraph := NewTagGraphFromTagIndex(tagIndex, mentionedIdents)
	return &TagAnalyzer{
		tagIndex: tagIndex,
		tagGraph: tagGraph,
	}
}

func (ta *TagAnalyzer) GetRankedTags() []Tag {
	ta.tagGraph.CalculateAndDistributeRanks()

	sortedDefinitions := ta.tagGraph.GetSortedDefinitions()
	if len(sortedDefinitions) == 0 {
		// If no sorted definitions, try to return all definitions
		var allTags []Tag
		for _, tags := range ta.tagIndex.Definitions {
			allTags = append(allTags, tags...)
		}
		return allTags
	}

	var tags []Tag
	for _, def := range sortedDefinitions {
		node := def.Key
		if int(node) >= len(ta.tagGraph.GetGraph().Nodes) {
			continue
		}

		nodePath := ta.tagGraph.GetGraph().Nodes[node]

		// Collect all definitions for this file
		var fileTags []Tag
		for _, defs := range ta.tagIndex.Definitions {
			for _, tag := range defs {
				if tag.RelFname == nodePath {
					fileTags = append(fileTags, tag)
				}
			}
		}

		// Add sorted by line number
		sort.Slice(fileTags, func(i, j int) bool { return fileTags[i].Line < fileTags[j].Line })
		tags = append(tags, fileTags...)
	}

	return tags
}

func (ta *TagAnalyzer) DebugPrintRankedTags() {
	rankedTags := ta.GetRankedTags()
	for _, tag := range rankedTags {
		fmt.Printf("%s:%d - %s (%s)\n", tag.RelFname, tag.Line, tag.Name, tag.Kind)
	}
}
