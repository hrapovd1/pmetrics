// staticlint утилита статического анализа go кода.
// для получения справки необходимо запустить с параметром help
package main

import (
	"github.com/hrapovd1/pmetrics/internal/checker"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(checker.CollectAnalyzers()...)
}
