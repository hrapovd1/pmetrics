package analyzer

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

var MainExitAnalyzer = &analysis.Analyzer{
	Name: "mainexitcheck",
	Doc:  "check for os.Exit in main function of main package",
	Run:  runAnalyzer,
}

func runAnalyzer(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			return nil, nil
		}
		var mainStartPos token.Pos
		var mainEndPos token.Pos

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "main" {
					mainStartPos = x.Pos()
					mainEndPos = x.End()
				}
			case *ast.CallExpr:
				if se, ok := x.Fun.(*ast.SelectorExpr); ok {
					if pkg, ok := se.X.(*ast.Ident); ok {
						if pkg.Name != "os" && se.Sel.Name != "Exit" {
							return true
						}
					}
					if mainStartPos < se.Pos() && se.Pos() < mainEndPos {
						pass.Reportf(se.Pos(), "forbidden direct call os.Exit")
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
