// checker пакет содержит переменные и функции для сбора
// аргументов типа *analysis.Analizer в один слайс, который
// передается в main функции multichecker.Main
package checker

import (
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/hrapovd1/pmetrics/internal/analyzer"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Статичечкие анализаторы из пакета
// golang.org/x/tools/go/analysis/passes
var standardCheckers = []*analysis.Analyzer{
	asmdecl.Analyzer,
	assign.Analyzer,
	atomic.Analyzer,
	bools.Analyzer,
	buildtag.Analyzer,
	cgocall.Analyzer,
	composite.Analyzer,
	copylock.Analyzer,
	ctrlflow.Analyzer,
	errorsas.Analyzer,
	framepointer.Analyzer,
	httpresponse.Analyzer,
	ifaceassert.Analyzer,
	inspect.Analyzer,
	loopclosure.Analyzer,
	lostcancel.Analyzer,
	nilfunc.Analyzer,
	printf.Analyzer,
	shift.Analyzer,
	sigchanyzer.Analyzer,
	stdmethods.Analyzer,
	stringintconv.Analyzer,
	structtag.Analyzer,
	testinggoroutine.Analyzer,
	tests.Analyzer,
	unmarshal.Analyzer,
	unreachable.Analyzer,
	unsafeptr.Analyzer,
	unusedresult.Analyzer,
}

// CollectAnalyzers собирает все желаемые анализаторы в один
// слайс, который выступает аргументом функции multichecker.Main
func CollectAnalyzers() []*analysis.Analyzer {
	out := make([]*analysis.Analyzer, len(standardCheckers)-1)
	// Добавляю стандартные статические анализаторы
	// из golang.org/x/tools/go/analysis/passes
	copy(out, standardCheckers)
	// Добавляю анализаторы класса SA пакета staticcheck.io
	for _, sa := range staticcheck.Analyzers {
		out = append(out, sa.Analyzer)
	}
	// Добавляю анализатор S1005 пакета staticcheck.io
	for _, s := range simple.Analyzers {
		if s.Analyzer.Name == "S1005" {
			out = append(out, s.Analyzer)
		}
	}
	// Добавляю анализатор ST1013 пакета staticcheck.io
	for _, s := range stylecheck.Analyzers {
		if s.Analyzer.Name == "ST1013" {
			out = append(out, s.Analyzer)
		}
	}
	// Добавляю анализатор QF1003 пакета staticcheck.io
	for _, s := range quickfix.Analyzers {
		if s.Analyzer.Name == "QF1003" {
			out = append(out, s.Analyzer)
		}
	}

	// Добавляю анализатор проверки пропущенных ошибок из проекта
	// https://github.com/kisielk/errcheck
	out = append(out, errcheck.Analyzer)

	// Добавляю анализатор проверки неэффективных присваиваний из проекта
	// https://github.com/gordonklaus/ineffassign
	out = append(out, ineffassign.Analyzer)

	// Добавляю анализатор прямого вызова os.Exit в main
	out = append(out, analyzer.MainExitAnalyzer)

	return out
}
