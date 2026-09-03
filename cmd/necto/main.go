package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"necto/pkg/ast"
	"necto/pkg/codegen"
	"necto/pkg/eval"
	"necto/pkg/format"
	"necto/pkg/lexer"
	"necto/pkg/parser"
	"necto/pkg/types"
)

const VERSION = "0.6.0-alpha"

type ProjectConfig struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Entry        string            `json:"entry"`
	Description  string            `json:"description,omitempty"`
	Authors      []string          `json:"authors,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

func printUsage() {
	fmt.Println("Necto Programming Language Compiler & Runtime")
	fmt.Printf("Version: %s\n\n", VERSION)
	fmt.Println("Usage:")
	fmt.Println("  necto init [name]              Initialize a new Necto project with necto.json")
	fmt.Println("  necto fmt [file/dir] [--check] Format Necto source code to canonical style")
	fmt.Println("  necto run [file.nc]            Run Necto program (or project entry from necto.json)")
	fmt.Println("  necto test [file.nc]           Run unit tests in file or tests/ directory")
	fmt.Println("  necto build [file.nc] -o <out> Compile Necto program to native binary")
	fmt.Println("  necto check [file.nc]          Type check program without executing")
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
	case "init":
		name := "my_project"
		if len(os.Args) >= 3 {
			name = os.Args[2]
		}
		initProject(name)

	case "fmt":
		checkOnly := false
		target := "."
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--check" {
				checkOnly = true
			} else if !strings.HasPrefix(os.Args[i], "-") {
				target = os.Args[i]
			}
		}
		runFormatter(target, checkOnly)

	case "run":
		sourceFile := ""
		if len(os.Args) >= 3 {
			sourceFile = os.Args[2]
		} else {
			cfg, _ := findProjectConfig()
			if cfg != nil && cfg.Entry != "" {
				sourceFile = cfg.Entry
			}
		}
		if sourceFile == "" {
			fmt.Fprintln(os.Stderr, "Error: missing source file to run. Run 'necto run <file.nc>' or create a necto.json project.")
			os.Exit(1)
		}
		runFile(sourceFile)

	case "test":
		if len(os.Args) >= 3 {
			testFile(os.Args[2])
		} else {
			// Автоматический поиск тестов в tests/ или necto.json
			ranAny := false
			if info, err := os.Stat("tests"); err == nil && info.IsDir() {
				filepath.Walk("tests", func(path string, fi os.FileInfo, err error) error {
					if err == nil && !fi.IsDir() && strings.HasSuffix(path, ".nc") {
						testFile(path)
						ranAny = true
					}
					return nil
				})
			}
			if !ranAny {
				cfg, _ := findProjectConfig()
				if cfg != nil && cfg.Entry != "" {
					testFile(cfg.Entry)
					ranAny = true
				}
			}
			if !ranAny {
				fmt.Fprintln(os.Stderr, "Error: missing source file to test. Example: necto test main.nc")
				os.Exit(1)
			}
		}

	case "build":
		sourceFile := ""
		if len(os.Args) >= 3 && !strings.HasPrefix(os.Args[2], "-") {
			sourceFile = os.Args[2]
		} else {
			cfg, _ := findProjectConfig()
			if cfg != nil && cfg.Entry != "" {
				sourceFile = cfg.Entry
			}
		}
		if sourceFile == "" {
			fmt.Fprintln(os.Stderr, "Error: missing source file to build. Example: necto build main.nc -o app.exe")
			os.Exit(1)
		}
		outFile := "output.exe"
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "-o" && i+1 < len(os.Args) {
				outFile = os.Args[i+1]
				break
			}
		}
		buildFile(sourceFile, outFile)

	case "check":
		sourceFile := ""
		if len(os.Args) >= 3 {
			sourceFile = os.Args[2]
		} else {
			cfg, _ := findProjectConfig()
			if cfg != nil && cfg.Entry != "" {
				sourceFile = cfg.Entry
			}
		}
		if sourceFile == "" {
			fmt.Fprintln(os.Stderr, "Error: missing source file to check.")
			os.Exit(1)
		}
		checkFile(sourceFile)

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

func findProjectConfig() (*ProjectConfig, string) {
	candidates := []string{"necto.json", "../necto.json"}
	for _, c := range candidates {
		if data, err := os.ReadFile(c); err == nil {
			var cfg ProjectConfig
			if err := json.Unmarshal(data, &cfg); err == nil {
				return &cfg, c
			}
		}
	}
	return nil, ""
}

func initProject(name string) {
	dir := name
	if dir == "" || dir == "." {
		dir = "."
		name = filepath.Base(filepath.Clean("."))
	} else {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory '%s': %s\n", dir, err)
			os.Exit(1)
		}
	}

	cfgPath := filepath.Join(dir, "necto.json")
	if _, err := os.Stat(cfgPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: project already exists in '%s' (necto.json found)\n", dir)
		os.Exit(1)
	}

	cfg := ProjectConfig{
		Name:        name,
		Version:     "0.1.0",
		Entry:       "src/main.nc",
		Description: fmt.Sprintf("A Necto project for %s", name),
		Authors:     []string{},
	}
	cfgData, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(cfgPath, cfgData, 0644)

	srcDir := filepath.Join(dir, "src")
	os.MkdirAll(srcDir, 0755)
	mainCode := `// src/main.nc
fn main() {
    println("Hello from Necto!");
}
`
	os.WriteFile(filepath.Join(srcDir, "main.nc"), []byte(mainCode), 0644)

	testsDir := filepath.Join(dir, "tests")
	os.MkdirAll(testsDir, 0755)
	testCode := `// tests/main_test.nc
test "sanity check" {
    assert(1 + 1 == 2);
}
`
	os.WriteFile(filepath.Join(testsDir, "main_test.nc"), []byte(testCode), 0644)

	gitignore := "bin/\n*.exe\n*.tmp.c\n"
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644)

	fmt.Printf("✓ Initialized new Necto project '%s' in '%s'\n", name, dir)
	fmt.Println("\nTo run the project:")
	if dir != "." {
		fmt.Printf("  cd %s\n", dir)
	}
	fmt.Println("  necto run")
}

func runFormatter(target string, checkOnly bool) {
	if target == "" {
		target = "."
	}

	info, err := os.Stat(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	hasUnformatted := false
	formattedCount := 0

	formatOneFile := func(path string) {
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %s\n", path, err)
			return
		}
		formatted, err := format.Format(string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Format error in %s: %s\n", path, err)
			return
		}
		if string(content) != formatted {
			hasUnformatted = true
			if checkOnly {
				fmt.Printf("✗ %s is not formatted\n", path)
			} else {
				if err := os.WriteFile(path, []byte(formatted), 0644); err == nil {
					fmt.Printf("✓ Formatted %s\n", path)
					formattedCount++
				}
			}
		}
	}

	if !info.IsDir() {
		formatOneFile(target)
	} else {
		filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
			if err == nil && !fi.IsDir() && (strings.HasSuffix(path, ".nc") || strings.HasSuffix(path, ".necto")) {
				if !strings.Contains(path, ".git") && !strings.Contains(path, "bin") {
					formatOneFile(path)
				}
			}
			return nil
		})
	}

	if checkOnly {
		if hasUnformatted {
			fmt.Println("\nSome files require formatting. Run 'necto fmt' to fix.")
			os.Exit(1)
		} else {
			fmt.Println("✓ All files are properly formatted!")
		}
	} else {
		if formattedCount == 0 {
			fmt.Println("✓ All files are already formatted!")
		}
	}
}
