package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	dataFile    = "testdata/data.yaml"
	dataNewFile = "testdata/data-new.yaml"
)

func TestNewDiffContext(t *testing.T) {
	diffCtx, err := NewDiffContext(dataFile, dataNewFile)
	assert.NoError(t, err)
	assert.NotNil(t, diffCtx)
}

func TestDiffContextDiffs(t *testing.T) {
	diffCtx, err := NewDiffContext(dataFile, dataNewFile)
	assert.NoError(t, err)

	fileDiffs := diffCtx.Diffs(DefaultDiffConfig)
	assert.Len(t, fileDiffs, 1)

	docDiffs := fileDiffs[0]
	assert.Len(t, docDiffs, 5)
}

func TestFileDiffsHasDifference(t *testing.T) {
	diffCtx, err := NewDiffContext(dataFile, dataNewFile)
	assert.NoError(t, err)

	fileDiffs := diffCtx.Diffs(DefaultDiffConfig)
	assert.True(t, fileDiffs.HasDifference())
}
