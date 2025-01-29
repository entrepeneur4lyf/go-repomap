// helpers.go

package repomap

import (
	"sort"
	"strings"
)

// Checks if a path is a git directory, it looks for any commit hash present
// and gets the timestamp for it as a poor-man's check
func IsGitRepository(dir string) bool {
	// Placeholder implementation for checking if a directory is a git repository
	// This should be replaced with actual logic to check for a git repository
	return strings.Contains(dir, ".git")
}

// TODO(codestory): Improve the name over here
func CloseSmallGapsHelper(lines map[int]struct{}, codeSplitByLines []string, codeLen int) map[int]struct{} {
	// a "closing" operation on the integers in set.
	// if i and i+2 are in there but i+1 is not, I want to add i+1
	// Create a new set for the "closed" lines
	closedShow := make(map[int]struct{})
	for k := range lines {
		closedShow[k] = struct{}{}
	}
	sortedShow := make([]int, 0, len(lines))
	for k := range lines {
		sortedShow = append(sortedShow, k)
	}
	sort.Ints(sortedShow)

	for i := 0; i < len(sortedShow)-1; i++ {
		if sortedShow[i+1]-sortedShow[i] == 2 {
			closedShow[sortedShow[i]+1] = struct{}{}
		}
	}

	// pick up adjacent blank lines
	for i, line := range codeSplitByLines {
		if _, ok := closedShow[i]; !ok {
			continue
		}

		// looking at the current line and if its not empty
		// and we are 2 lines above the end and the next line is empty
		if strings.TrimSpace(line) != "" && i < codeLen-2 && strings.TrimSpace(codeSplitByLines[i+1]) == "" {
			closedShow[i+1] = struct{}{}
		}
	}

	closedClosedShow := make([]int, 0, len(closedShow))
	for k := range closedShow {
		closedClosedShow = append(closedClosedShow, k)
	}
	sort.Ints(closedClosedShow)

	return closedShow
}
