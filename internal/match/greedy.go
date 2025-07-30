package match

import (
	"sort"
	"strings"

	"github.com/tychonis/cyanotype/internal/distance"
)

// sortTokenSlices sorts [][]string by (len, joined string)
func sortTokenSlices(slices [][]string) {
	sort.Slice(slices, func(i, j int) bool {
		if len(slices[i]) != len(slices[j]) {
			return len(slices[i]) < len(slices[j])
		}
		return strings.Join(slices[i], ".") < strings.Join(slices[j], ".")
	})
}

func GreedyMatch(srcTokens [][]string, dstTokens [][]string) map[string]string {
	sortTokenSlices(srcTokens)
	sortTokenSlices(dstTokens)

	usedDst := make([]bool, len(dstTokens))
	result := make(map[string]string)

	for i := range len(srcTokens) {
		minDist := -1
		minJ := -1

		for j := range len(dstTokens) {
			if usedDst[j] {
				continue
			}
			d := distance.EditDistance(srcTokens[i], dstTokens[j])
			if minDist == -1 || d < minDist {
				minDist = d
				minJ = j
			}
		}

		srcKey := strings.Join(srcTokens[i], ".")

		if minJ != -1 {
			dstKey := strings.Join(dstTokens[minJ], ".")
			result[srcKey] = dstKey
			usedDst[minJ] = true
		}
	}

	return result
}
