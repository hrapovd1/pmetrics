package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func Test_runAnalyzer(t *testing.T) {
	t.Run("Check file with os.exit", func(t *testing.T) {
		_ = analysistest.Run(t, analysistest.TestData(), MainExitAnalyzer)
	})
}
