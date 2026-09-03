package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"necto/pkg/ast"
	"necto/pkg/codegen"
	"necto/pkg/eval"
	"necto/pkg/lexer"
	"necto/pkg/parser"
	"necto/pkg/types"
)

const VERSION = "0.4.0-alpha"

func printUsage() {
	fmt.Println("Necto Programming Language Compiler & Runtime")
	fmt.Printf("Version: %s\n\n", VERSION)
	fmt.Println("Usage:")
	fmt.Println("  necto run <file.nc>            Run a Necto source file directly")
	fmt.Println("  necto test <file.nc>           Run all unit tests in a Necto source file")
	fmt.Println("  necto build <file.nc> -o <out> Compile Necto program to native binary")
	fmt.Println("  necto check <file.nc>          Type check program without executing")
	fmt.Println("  necto repl                     Start interactive Necto REPL")
	fmt.Println("  necto version                  Show Necto version")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: missing source file to run. Example: necto run main.nc")
			os.Exit(1)
		}
		runFile(os.Args[2])

	case "test":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: missing source file to test. Example: necto test main.nc")
			os.Exit(1)
		}
		testFile(os.Args[2])

	case "build":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: missing source file to build. Example: necto build main.nc -o app.exe")
			os.Exit(1)
		}
		sourceFile := os.Args[2]
		outFile := "output.exe"
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "-o" && i+1 < len(os.Args) {
				outFile = os.Args[i+1]
				break
			}
		}
		buildFile(sourceFile, outFile)

	case "check":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: missing source file to check.")
			os.Exit(1)
		}
		checkFile(os.Args[2])

	case "repl":
		startRepl()

	case "version", "--version", "-v":
		fmt.Printf("Necto Language Toolchain v%s\n", VERSION)

	default:
		// Если передан просто файл: necto file.nc
		if strings.HasSuffix(command, ".nc") || strings.HasSuffix(command, ".necto") || strings.HasSuffix(command, ".aura") {
			runFile(command)
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
			printUsage()
			os.Exit(1)
		}
	}
}

func parseAndCheck(filepath string) (*ast.Program, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not read file '%s': %w", filepath, err)
	}

	l := lexer.New(string(bytes))
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		var sb strings.Builder
		sb.WriteString("Syntax Errors:\n")
		for _, e := range p.Errors() {
			sb.WriteString(fmt.Sprintf("  • %s\n", e))
		}
		return nil, fmt.Errorf("%s", sb.String())
	}

	checker := types.NewChecker()
	checker.Check(prog)

	if len(checker.Errors()) > 0 {
		var sb strings.Builder
		sb.WriteString("Type Errors:\n")
		for _, e := range checker.Errors() {
			sb.WriteString(fmt.Sprintf("  • %s\n", e))
		}
		return nil, fmt.Errorf("%s", sb.String())
	}

	return prog, nil
}

func runFile(filepath string) {
	prog, err := parseAndCheck(filepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	env := eval.NewEnvironment()
	res := eval.Eval(prog, env)

	if res != nil && res.Type() == eval.ERROR_OBJ {
		fmt.Fprintf(os.Stderr, "%s\n", res.Inspect())
		os.Exit(1)
	}

	// Если определена функция main(), вызываем её
	if mainFn, ok := env.Get("main"); ok {
		if _, isFn := mainFn.(*eval.Function); isFn {
			mainCall := &ast.CallExpression{
				Function: &ast.Identifier{Value: "main"},
			}
			mainRes := eval.Eval(mainCall, env)
			if mainRes != nil && mainRes.Type() == eval.ERROR_OBJ {
				fmt.Fprintf(os.Stderr, "%s\n", mainRes.Inspect())
				os.Exit(1)
			}
		}
	}
}

func checkFile(filepath string) {
	_, err := parseAndCheck(filepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("✓ Syntax and type check passed successfully!")
}

func buildFile(sourceFile, outFile string) {
	prog, err := parseAndCheck(sourceFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Compiling %s to %s via Clang/LLVM...\n", sourceFile, outFile)
	err = codegen.BuildNativeExecutable(prog, outFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Build Error: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Successfully created native binary: %s\n", outFile)
}

func testFile(filepath string) {
	prog, err := parseAndCheck(filepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Сначала выполняем глобальные объявления (функции, структуры, enums) в базовом окружении
	baseEnv := eval.NewEnvironment()
	for _, stmt := range prog.Statements {
		if _, isTest := stmt.(*ast.TestBlockStatement); !isTest {
			res := eval.Eval(stmt, baseEnv)
			if res != nil && res.Type() == eval.ERROR_OBJ {
				fmt.Fprintf(os.Stderr, "Initialization Error: %s\n", res.Inspect())
				os.Exit(1)
			}
		}
	}

	// Собираем все тесты
	var tests []*ast.TestBlockStatement
	for _, stmt := range prog.Statements {
		if tb, isTest := stmt.(*ast.TestBlockStatement); isTest {
			tests = append(tests, tb)
		}
	}

	if len(tests) == 0 {
		fmt.Printf("No tests found in '%s'\n", filepath)
		return
	}

	fmt.Printf("Running %d test(s) in %s...\n\n", len(tests), filepath)
	passed := 0
	failed := 0

	for _, t := range tests {
		testEnv := eval.NewEnclosedEnvironment(baseEnv)
		start := time.Now()
		res := eval.Eval(t.Body, testEnv)
		duration := time.Since(start)

		if res != nil && res.Type() == eval.ERROR_OBJ {
			fmt.Printf("  ✗ FAIL: test \"%s\" (took %v)\n", t.Name, duration)
			fmt.Printf("    %s\n\n", res.Inspect())
			failed++
		} else {
			fmt.Printf("  ✓ PASS: test \"%s\" (took %v)\n", t.Name, duration)
			passed++
		}
	}

	fmt.Printf("\nTest Results: %d passed, %d failed\n", passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func startRepl() {
	fmt.Printf("Necto Programming Language Interactive REPL (v%s)\n", VERSION)
	fmt.Println("Type 'exit' to quit.")
	fmt.Println("")

	scanner := bufio.NewScanner(os.Stdin)
	env := eval.NewEnvironment()

	for {
		fmt.Print("necto> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "exit" || trimmed == "quit" {
			break
		}
		if trimmed == "" {
			continue
		}

		l := lexer.New(line)
		p := parser.New(l)
		prog := p.ParseProgram()

		if len(p.Errors()) > 0 {
			for _, e := range p.Errors() {
				fmt.Printf("Syntax Error: %s\n", e)
			}
			continue
		}

		result := eval.Eval(prog, env)
		if result != nil && result.Type() != eval.NULL_OBJ {
			fmt.Println(result.Inspect())
		}
	}
}
