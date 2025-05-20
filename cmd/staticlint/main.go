// Package main запускает набор статических анализаторов для проверки кода проекта.
// Он использует фреймворк multichecker из пакета golang.org/x/tools/go/analysis,
// который позволяет объединять несколько анализаторов и применять их к Go-коду.
//
// Для запуска используется функция multichecker.Main, которой передаётся список анализаторов.
//
// Включены:
//
// 1. Стандартные анализаторы из пакета golang.org/x/tools/go/analysis/passes:
//   - inspect: предоставляет общий механизм обхода AST для других анализаторов.
//   - printf: проверяет соответствие форматных строк и аргументов в функциях печати.
//   - shadow: обнаруживает случаи затенения переменных.
//   - structtag: проверяет правильность оформления struct-тегов (например, JSON).
//   - unusedresult: находит вызовы функций, возвращающие значения, которые игнорируются.
//
// 2. Анализаторы из пакета staticcheck.io (honnef.co/go/tools/staticcheck):
//   - Это расширенный набор проверок: потенциальные ошибки, антипаттерны, устаревшие конструкции и прочее.
//   - Включает анализаторы SA, ST и другие из пакета staticcheck.
//
// 3. Пользовательский анализатор:
//   - noexit: реализован в пакете github.com/lenarlenar/go-my-metrics-service/cmd/staticlint/noexit.
//     Проверяет использование os.Exit в функции main и запрещает его.
//
// Использование:
//
//	go run ./cmd/staticlint ./...
package main

import (
	"github.com/lenarlenar/go-my-metrics-service/cmd/staticlint/noexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

// main запускает анализаторы и передаёт их multichecker.
func main() {
	var mychecks []*analysis.Analyzer

	// Добавление стандартных анализаторов
	mychecks = append(mychecks,
		inspect.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		unusedresult.Analyzer,
	)

	// Добавление анализаторов из пакета staticcheck.io
	for _, v := range staticcheck.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	// Добавление собственного анализатора
	mychecks = append(mychecks, noexit.Analyzer)

	multichecker.Main(
		mychecks...,
	)
}
