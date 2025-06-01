package diff

import (
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/stretchr/testify/require"
)

var diffStringLines = []string{
	"~ .people.name: John -> Bob",
	"~ .people.surname: Doe -> Rose",
	"~ .city.name: New York -> San Francisco",
	"~ .item.id: 124 -> 123",
	"~ .item.price: 10.9 -> 10.3",
}

func TestDiffsArray(t *testing.T) {
	arrayYamlDiffs := []struct {
		left                     []int
		right                    []int
		expectedDiffs            [][2]int
		expectedDiffsIgnoreIndex [][2]int
	}{
		{
			left:                     []int{1, 2, 3, 4, 5},
			right:                    []int{1, 2, 3, 4, 5},
			expectedDiffs:            [][2]int{},
			expectedDiffsIgnoreIndex: [][2]int{},
		},
		{
			left:  []int{1, 2, 3, 4, 5},
			right: []int{4, 5},
			expectedDiffs: [][2]int{
				{1, 4}, {2, 5}, {3, 0}, {4, 0}, {5, 0},
			},
			expectedDiffsIgnoreIndex: [][2]int{
				{1, 0}, {2, 0}, {3, 0},
			},
		},
		{
			left:  []int{4, 5},
			right: []int{1, 2, 3, 4, 5},
			expectedDiffs: [][2]int{
				{4, 1}, {5, 2}, {0, 3}, {0, 4}, {0, 5},
			},
			expectedDiffsIgnoreIndex: [][2]int{
				{0, 1}, {0, 2}, {0, 3},
			},
		},
		{
			left:  []int{1, 2, 3, 4, 5},
			right: []int{5, 4, 3, 2, 1},
			expectedDiffs: [][2]int{
				{1, 5}, {2, 4}, {4, 2}, {5, 1},
			},
			expectedDiffsIgnoreIndex: [][2]int{},
		},
		{
			left:  []int{1, 2, 3, 4, 5, 8},
			right: []int{9, 3, 2, 5, 4, 6, 7},
			expectedDiffs: [][2]int{
				{1, 9}, {2, 3}, {3, 2}, {4, 5}, {5, 4}, {8, 6}, {0, 7},
			},
			expectedDiffsIgnoreIndex: [][2]int{
				{1, 9}, {8, 6}, {0, 7},
			},
		},
	}

	compareIntegerDiff := func(t *testing.T, diff *Diff, expected [2]int) {
		if diff.leftNode == nil {
			require.EqualValues(t, 0, expected[0])
		} else {
			intNode := diff.leftNode.(*ast.IntegerNode)
			value := intNode.Value.(uint64)
			require.EqualValues(t, value, expected[0])
		}
		if diff.rightNode == nil {
			require.EqualValues(t, 0, expected[1])
		} else {
			intNode := diff.rightNode.(*ast.IntegerNode)
			value := intNode.Value.(uint64)
			require.EqualValues(t, value, expected[1])
		}
	}

	for _, arrayYamlDiff := range arrayYamlDiffs {
		leftYaml := toYaml(t, arrayYamlDiff.left)
		rightYaml := toYaml(t, arrayYamlDiff.right)

		fileDiffs, err := Compare(leftYaml, rightYaml)
		require.NoError(t, err)
		require.Len(t, fileDiffs, 1)

		docDiffs := fileDiffs[0]
		require.Len(t, docDiffs, len(arrayYamlDiff.expectedDiffs))

		for i, diff := range docDiffs {
			expectedDiff := arrayYamlDiff.expectedDiffs[i]
			compareIntegerDiff(t, diff, expectedDiff)
		}
	}

	t.Run("ignore sequence order", func(t *testing.T) {
		for _, arrayYamlDiff := range arrayYamlDiffs {
			leftYaml := toYaml(t, arrayYamlDiff.left)
			rightYaml := toYaml(t, arrayYamlDiff.right)
			fileDiffs, err := Compare(leftYaml, rightYaml, IgnoreSeqOrder)
			require.NoError(t, err)

			require.Len(t, fileDiffs, 1)
			docDiffs := fileDiffs[0]
			require.Len(t, docDiffs, len(arrayYamlDiff.expectedDiffsIgnoreIndex))

			for i, diff := range docDiffs {
				expectedDiff := arrayYamlDiff.expectedDiffsIgnoreIndex[i]
				compareIntegerDiff(t, diff, expectedDiff)
			}
		}
	})
}

func TestFormat(t *testing.T) {
	diffs, err := CompareFile(fileLeft, fileRight)
	require.NoError(t, err)

	output := diffs.Format(Plain)

	require.Equal(t, output, strings.Join(diffStringLines, "\n"))
}

func toYaml(t *testing.T, a any) []byte {
	b, err := yaml.Marshal(a)
	if err != nil {
		t.Error(err)
	}
	return b
}

func readFile(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
	}
	return data
}
