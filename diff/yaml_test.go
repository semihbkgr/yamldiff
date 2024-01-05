package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	dataFileLeft  = "testdata/data-left.yaml"
	dataFileRight = "testdata/data-right.yaml"
)

func TestNewDiffContext(t *testing.T) {
	diffCtx, err := NewDiffContext(dataFileLeft, dataFileRight)
	assert.NoError(t, err)
	assert.NotNil(t, diffCtx)
}

func TestDiffContextDiffs(t *testing.T) {
	diffCtx, err := NewDiffContext(dataFileLeft, dataFileRight)
	assert.NoError(t, err)

	fileDiffs := diffCtx.Diffs(DefaultDiffConfig)
	assert.Len(t, fileDiffs, 1)

	docDiffs := fileDiffs[0]
	assert.Len(t, docDiffs, 5)
}

func TestFileDiffsHasDifference(t *testing.T) {
	diffCtx, err := NewDiffContext(dataFileLeft, dataFileRight)
	assert.NoError(t, err)

	fileDiffs := diffCtx.Diffs(DefaultDiffConfig)
	assert.True(t, fileDiffs.HasDifference())
}
