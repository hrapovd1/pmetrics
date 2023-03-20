package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectAnalyzers(t *testing.T) {
	t.Run("Count analyzers", func(t *testing.T) {
		assert.Equal(t, 124, len(CollectAnalyzers()))
	})
}
