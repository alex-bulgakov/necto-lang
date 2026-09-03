package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

const VERSION = "2.0.3"

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
	fmt.Println("  necto bench [file.nc] [--runs] Run benchmarks in file or tests/ directory")
	fmt.Println("  necto doc [file/dir] [--serve] Generate interactive HTML documentation")
	fmt.Println("  necto install                  Install dependencies from necto.json")
	fmt.Println("  necto toolchain [status/inst]  Check or manage embedded C/LLVM backend compiler")
	fmt.Println("  necto build [file.nc] -o <out> Compile Necto program to native binary")
	fmt.Println("  necto bootstrap                Compile self-hosted compiler to bin/necto-native.exe (Stage 2)")
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

	case "bench":
		targetFile := ""
		runs := 0
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--runs" && i+1 < len(os.Args) {
				runs, _ = strconv.Atoi(os.Args[i+1])
				i++
			} else if strings.HasPrefix(os.Args[i], "--runs=") {
				runs, _ = strconv.Atoi(strings.TrimPrefix(os.Args[i], "--runs="))
			} else if !strings.HasPrefix(os.Args[i], "-") {
				targetFile = os.Args[i]
			}
		}
		if targetFile == "" {
			cfg, _ := findProjectConfig()
			if cfg != nil && cfg.Entry != "" {
				targetFile = cfg.Entry
			}
		}
		if targetFile == "" {
			fmt.Fprintln(os.Stderr, "Error: missing source file for benchmarks. Example: necto bench benchmark.nc")
			os.Exit(1)
		}
		benchFile(targetFile, runs)

	case "doc":
		target := "."
		outDir := "docs"
		serve := false
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--serve" {
				serve = true
			} else if (os.Args[i] == "-o" || os.Args[i] == "--output") && i+1 < len(os.Args) {
				outDir = os.Args[i+1]
				i++
			} else if !strings.HasPrefix(os.Args[i], "-") {
				target = os.Args[i]
			}
		}
		generateDocs(target, outDir, serve)

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

	case "bootstrap":
		runBootstrap()

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

	case "install":
		installDeps()

	case "toolchain":
		action := "status"
		if len(os.Args) >= 3 {
			action = os.Args[2]
		}
		handleToolchain(action)

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

func renderRichDiagnostic(filePath string, fileContent string, rawErr string) string {
	var line, col int
	var msg string
	n, _ := fmt.Sscanf(rawErr, "[%d:%d]", &line, &col)
	if n == 2 {
		idx := strings.Index(rawErr, "]")
		if idx != -1 {
			msg = strings.TrimSpace(rawErr[idx+1:])
		}
	} else {
		return "  • " + rawErr
	}

	lines := strings.Split(fileContent, "\n")
	if line <= 0 || line > len(lines) {
		return fmt.Sprintf("  • %s", rawErr)
	}

	sourceLine := strings.ReplaceAll(lines[line-1], "\r", "")
	linePrefix := fmt.Sprintf("%4d | ", line)
	if col > len(sourceLine)+1 {
		col = len(sourceLine) + 1
	}
	var caretSb strings.Builder
	for i := 0; i < len(linePrefix); i++ {
		caretSb.WriteByte(' ')
	}
	for i := 1; i < col; i++ {
		if i-1 < len(sourceLine) && sourceLine[i-1] == '\t' {
			caretSb.WriteByte('\t')
		} else {
			caretSb.WriteByte(' ')
		}
	}
	caretSb.WriteString("^")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("error: %s\n", msg))
	sb.WriteString(fmt.Sprintf("  --> %s:%d:%d\n", filePath, line, col))
	sb.WriteString("     |\n")
	sb.WriteString(fmt.Sprintf("%s%s\n", linePrefix, sourceLine))
	sb.WriteString(fmt.Sprintf("%s", caretSb.String()))

	return sb.String()
}

func parseAndCheck(filepath string) (*ast.Program, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not read file '%s': %w", filepath, err)
	}
	content := string(bytes)

	l := lexer.New(content)
	p := parser.New(l)
	prog := p.ParseProgram()

	if len(p.Errors()) > 0 {
		var sb strings.Builder
		sb.WriteString("Syntax Errors in " + filepath + ":\n\n")
		for _, e := range p.Errors() {
			sb.WriteString(renderRichDiagnostic(filepath, content, e))
			sb.WriteString("\n\n")
		}
		return nil, fmt.Errorf("%s", sb.String())
	}

	checker := types.NewChecker()
	checker.Check(prog)

	if len(checker.Errors()) > 0 {
		var sb strings.Builder
		sb.WriteString("Type Errors in " + filepath + ":\n\n")
		for _, e := range checker.Errors() {
			sb.WriteString(renderRichDiagnostic(filepath, content, e))
			sb.WriteString("\n\n")
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

func runBootstrap() {
	fmt.Println("==================================================================")
	fmt.Println("       Necto Bootstrap (Stage 2) — Building Native Compiler       ")
	fmt.Println("==================================================================")
	fmt.Println("Step 1: Running pure Necto compiler (compiler/main.nc)...")

	if _, err := os.Stat("compiler/main.nc"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: compiler/main.nc not found: %s\n", err)
		os.Exit(1)
	}

	runFile("compiler/main.nc")

	fmt.Println("\nStep 2: Compiling generated C code with Clang/LLVM...")
	tmpC := "stage1_output.tmp.c"
	if _, err := os.Stat(tmpC); err != nil {
		fmt.Fprintf(os.Stderr, "Error: generated C file '%s' not found: %s\n", tmpC, err)
		os.Exit(1)
	}
	defer os.Remove(tmpC)

	os.MkdirAll("bin", 0755)
	nativeOut := filepath.Join("bin", "necto-native.exe")
	compilerPath, sourceKind, err := codegen.FindClangCompiler()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: no C compiler backend found: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using backend compiler: %s (%s)\n", compilerPath, sourceKind)

	cmd := exec.Command(compilerPath, "-O2", tmpC, "-o", nativeOut)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: native compilation failed:\n%s\n%s\n", string(out), err)
		os.Exit(1)
	}

	fmt.Printf("✓ Step 3: Verified native build: '%s' is ready!\n", nativeOut)
	fmt.Println("==================================================================")
	fmt.Println("       Stage 2 Complete: Necto has successfully built itself!     ")
	fmt.Println("==================================================================")
}

func benchFile(filepath string, runs int) {
	prog, err := parseAndCheck(filepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	baseEnv := eval.NewEnvironment()
	for _, stmt := range prog.Statements {
		if _, isBench := stmt.(*ast.BenchmarkBlockStatement); !isBench {
			if _, isTest := stmt.(*ast.TestBlockStatement); !isTest {
				res := eval.Eval(stmt, baseEnv)
				if res != nil && res.Type() == eval.ERROR_OBJ {
					fmt.Fprintf(os.Stderr, "Initialization Error: %s\n", res.Inspect())
					os.Exit(1)
				}
			}
		}
	}

	var benchmarks []*ast.BenchmarkBlockStatement
	for _, stmt := range prog.Statements {
		if bb, isBench := stmt.(*ast.BenchmarkBlockStatement); isBench {
			benchmarks = append(benchmarks, bb)
		}
	}

	if len(benchmarks) == 0 {
		fmt.Printf("No benchmarks found in '%s'\n", filepath)
		return
	}

	fmt.Printf("Running %d benchmark(s) in %s...\n\n", len(benchmarks), filepath)
	fmt.Printf("%-35s %12s %15s %15s\n", "BENCHMARK", "ITERATIONS", "TIME / OP", "THROUGHPUT")
	fmt.Printf("%-35s %12s %15s %15s\n", strings.Repeat("-", 35), strings.Repeat("-", 12), strings.Repeat("-", 15), strings.Repeat("-", 15))

	for _, b := range benchmarks {
		benchEnv := eval.NewEnclosedEnvironment(baseEnv)

		targetRuns := runs
		if targetRuns <= 0 {
			// Калибровка на 500мс
			startSample := time.Now()
			sampleCount := 50
			for i := 0; i < sampleCount; i++ {
				eval.Eval(b.Body, benchEnv)
			}
			elapsedSample := time.Since(startSample)
			if elapsedSample <= 0 {
				elapsedSample = 1 * time.Microsecond
			}
			avgPerOp := float64(elapsedSample.Nanoseconds()) / float64(sampleCount)
			if avgPerOp <= 0 {
				avgPerOp = 1
			}
			targetRuns = int(float64(300*time.Millisecond.Nanoseconds()) / avgPerOp)
			if targetRuns < 100 {
				targetRuns = 100
			}
			if targetRuns > 1000000 {
				targetRuns = 1000000
			}
		}

		start := time.Now()
		for i := 0; i < targetRuns; i++ {
			eval.Eval(b.Body, benchEnv)
		}
		duration := time.Since(start)

		nsPerOp := duration.Nanoseconds() / int64(targetRuns)
		var opsPerSec float64
		if duration.Seconds() > 0 {
			opsPerSec = float64(targetRuns) / duration.Seconds()
		}

		timeStr := formatDuration(time.Duration(nsPerOp))
		throughputStr := formatThroughput(opsPerSec)

		fmt.Printf("%-35s %12d %15s %15s\n", b.Name, targetRuns, timeStr, throughputStr)
	}
	fmt.Println()
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns/op", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%.2f µs/op", float64(d.Nanoseconds())/1000.0)
	} else if d < time.Second {
		return fmt.Sprintf("%.2f ms/op", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2f s/op", d.Seconds())
}

func formatThroughput(ops float64) string {
	if ops >= 1_000_000 {
		return fmt.Sprintf("%.2f M ops/s", ops/1_000_000.0)
	} else if ops >= 1_000 {
		return fmt.Sprintf("%.2f K ops/s", ops/1_000.0)
	}
	return fmt.Sprintf("%.0f ops/s", ops)
}

type DocItem struct {
	Kind      string
	Name      string
	Signature string
	Doc       string
	File      string
}

func generateDocs(target, outDir string, serve bool) {
	var files []string
	info, err := os.Stat(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if !info.IsDir() {
		files = append(files, target)
	} else {
		filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
			if err == nil && !fi.IsDir() && (strings.HasSuffix(path, ".nc") || strings.HasSuffix(path, ".necto")) {
				if !strings.Contains(path, ".git") && !strings.Contains(path, "bin") {
					files = append(files, path)
				}
			}
			return nil
		})
	}

	if len(files) == 0 {
		fmt.Printf("No Necto source files (.nc) found in '%s'\n", target)
		return
	}

	var items []DocItem
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		l := lexer.New(string(content))
		p := parser.New(l)
		prog := p.ParseProgram()

		for _, stmt := range prog.Statements {
			switch s := stmt.(type) {
			case *ast.FnDeclaration:
				params := make([]string, len(s.Parameters))
				for i, param := range s.Parameters {
					params[i] = param.Name.Value + ": " + param.Type
				}
				sig := fmt.Sprintf("fn %s(%s)", s.Name.Value, strings.Join(params, ", "))
				if s.ReturnType != "" {
					sig += " -> " + s.ReturnType
				}
				items = append(items, DocItem{
					Kind:      "Function",
					Name:      s.Name.Value,
					Signature: sig,
					Doc:       s.DocComment,
					File:      filepath.Base(f),
				})

			case *ast.StructDeclaration:
				var fields []string
				for _, field := range s.Fields {
					fields = append(fields, fmt.Sprintf("%s: %s", field.Name.Value, field.Type))
				}
				sig := fmt.Sprintf("struct %s {\n    %s\n}", s.Name.Value, strings.Join(fields, ",\n    "))
				items = append(items, DocItem{
					Kind:      "Struct",
					Name:      s.Name.Value,
					Signature: sig,
					Doc:       s.DocComment,
					File:      filepath.Base(f),
				})

			case *ast.EnumDeclaration:
				var variants []string
				for _, v := range s.Variants {
					if len(v.Types) > 0 {
						variants = append(variants, fmt.Sprintf("%s(%s)", v.Name.Value, strings.Join(v.Types, ", ")))
					} else {
						variants = append(variants, v.Name.Value)
					}
				}
				sig := fmt.Sprintf("enum %s {\n    %s\n}", s.Name.Value, strings.Join(variants, ",\n    "))
				items = append(items, DocItem{
					Kind:      "Enum",
					Name:      s.Name.Value,
					Signature: sig,
					Doc:       s.DocComment,
					File:      filepath.Base(f),
				})
			}
		}
	}

	projectName := "Necto Documentation"
	projectVer := VERSION
	if cfg, _ := findProjectConfig(); cfg != nil {
		if cfg.Name != "" {
			projectName = cfg.Name
		}
		if cfg.Version != "" {
			projectVer = cfg.Version
		}
	}

	htmlContent := renderDocsHTML(projectName, projectVer, items)
	os.MkdirAll(outDir, 0755)
	outPath := filepath.Join(outDir, "index.html")
	if err := os.WriteFile(outPath, []byte(htmlContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing documentation to %s: %s\n", outPath, err)
		os.Exit(1)
	}

	fmt.Printf("✓ Documentation generated successfully: %s (%d symbols documented)\n", outPath, len(items))

	if serve {
		port := 8080
		fmt.Printf("Starting interactive doc server at http://localhost:%d/ (Press Ctrl+C to stop)\n", port)
		http.Handle("/", http.FileServer(http.Dir(outDir)))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %s\n", err)
		}
	}
}

func renderDocsHTML(projectName, version string, items []DocItem) string {
	itemsJSON, _ := json.Marshal(items)

	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + projectName + ` v` + version + ` — Documentation</title>
    <style>
        :root {
            --bg: #0d1117;
            --surface: #161b22;
            --border: #30363d;
            --text: #c9d1d9;
            --text-heading: #f0f6fc;
            --accent: #58a6ff;
            --badge-fn: #238636;
            --badge-struct: #1f6feb;
            --badge-enum: #a371f7;
            --code-bg: #090d13;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
            background: var(--bg);
            color: var(--text);
            display: flex;
            height: 100vh;
            overflow: hidden;
        }
        aside {
            width: 320px;
            background: var(--surface);
            border-right: 1px solid var(--border);
            display: flex;
            flex-direction: column;
        }
        .sidebar-header {
            padding: 20px;
            border-bottom: 1px solid var(--border);
        }
        .sidebar-header h1 {
            font-size: 1.25rem;
            color: var(--text-heading);
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .sidebar-header .version-tag {
            font-size: 0.75rem;
            background: #21262d;
            padding: 2px 8px;
            border-radius: 12px;
            color: var(--accent);
        }
        .search-box {
            padding: 12px 20px;
            border-bottom: 1px solid var(--border);
        }
        .search-box input {
            width: 100%;
            padding: 8px 12px;
            background: var(--bg);
            border: 1px solid var(--border);
            border-radius: 6px;
            color: var(--text);
            font-size: 0.875rem;
            outline: none;
        }
        .search-box input:focus { border-color: var(--accent); }
        .symbol-list {
            flex: 1;
            overflow-y: auto;
            padding: 10px;
        }
        .symbol-item {
            padding: 8px 12px;
            border-radius: 6px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 2px;
            font-size: 0.875rem;
        }
        .symbol-item:hover { background: #21262d; color: var(--text-heading); }
        .badge {
            font-size: 0.65rem;
            padding: 2px 6px;
            border-radius: 4px;
            font-weight: 600;
            text-transform: uppercase;
        }
        .badge-Function { background: rgba(35, 134, 54, 0.2); color: #3fb950; border: 1px solid rgba(35,134,54,0.4); }
        .badge-Struct { background: rgba(31, 111, 235, 0.2); color: #58a6ff; border: 1px solid rgba(31,111,235,0.4); }
        .badge-Enum { background: rgba(163, 113, 247, 0.2); color: #bc8cff; border: 1px solid rgba(163,113,247,0.4); }
        main {
            flex: 1;
            overflow-y: auto;
            padding: 40px;
        }
        .card {
            background: var(--surface);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 24px;
            margin-bottom: 24px;
        }
        .card-header {
            display: flex;
            align-items: center;
            gap: 12px;
            margin-bottom: 12px;
        }
        .card-title {
            font-size: 1.35rem;
            color: var(--text-heading);
            font-family: monospace;
        }
        .card-file {
            font-size: 0.75rem;
            color: #8b949e;
            margin-left: auto;
        }
        .signature-block {
            background: var(--code-bg);
            border: 1px solid var(--border);
            border-radius: 6px;
            padding: 12px 16px;
            font-family: monospace;
            font-size: 0.9rem;
            color: #e6edf3;
            white-space: pre-wrap;
            margin-bottom: 14px;
        }
        .doc-text {
            color: #8b949e;
            font-size: 0.95rem;
            line-height: 1.6;
            white-space: pre-wrap;
        }
    </style>
</head>
<body>
    <aside>
        <div class="sidebar-header">
            <h1>` + projectName + ` <span class="version-tag">v` + version + `</span></h1>
        </div>
        <div class="search-box">
            <input type="text" id="searchInput" placeholder="Search symbols..." autocomplete="off">
        </div>
        <div class="symbol-list" id="symbolList"></div>
    </aside>
    <main id="mainContent">
        <h2 style="margin-bottom: 20px; color: var(--text-heading);">API Reference</h2>
        <div id="cardsList"></div>
    </main>

    <script>
        const items = ` + string(itemsJSON) + `;
        const searchInput = document.getElementById('searchInput');
        const symbolList = document.getElementById('symbolList');
        const cardsList = document.getElementById('cardsList');

        function render(filter = '') {
            symbolList.innerHTML = '';
            cardsList.innerHTML = '';
            const lower = filter.toLowerCase();

            items.filter(it => it.Name.toLowerCase().includes(lower) || it.Kind.toLowerCase().includes(lower)).forEach((it, idx) => {
                const navItem = document.createElement('div');
                navItem.className = 'symbol-item';
                navItem.innerHTML = '<span>' + it.Name + '</span><span class="badge badge-' + it.Kind + '">' + it.Kind + '</span>';
                navItem.onclick = () => {
                    const el = document.getElementById('item-' + idx);
                    if (el) el.scrollIntoView({ behavior: 'smooth' });
                };
                symbolList.appendChild(navItem);

                const card = document.createElement('div');
                card.className = 'card';
                card.id = 'item-' + idx;
                card.innerHTML = 
                    '<div class="card-header">' +
                        '<span class="badge badge-' + it.Kind + '">' + it.Kind + '</span>' +
                        '<span class="card-title">' + it.Name + '</span>' +
                        '<span class="card-file">' + it.File + '</span>' +
                    '</div>' +
                    '<div class="signature-block">' + it.Signature + '</div>' +
                    '<div class="doc-text">' + (it.Doc ? it.Doc : '<em style="color: #484f58">No documentation provided.</em>') + '</div>';
                cardsList.appendChild(card);
            });
        }

        searchInput.addEventListener('input', (e) => render(e.target.value));
        render();
    </script>
</body>
</html>`
}

func installDeps() {
	cfg, _ := findProjectConfig()
	if cfg == nil {
		fmt.Fprintln(os.Stderr, "Error: could not find necto.json in current directory.")
		fmt.Fprintln(os.Stderr, "Run 'necto init <name>' to create a project first.")
		os.Exit(1)
	}

	if len(cfg.Dependencies) == 0 {
		fmt.Println("No dependencies defined in necto.json.")
		fmt.Println("Add dependencies to your necto.json:")
		fmt.Println(`  "dependencies": {`)
		fmt.Println(`    "my-lib": "https://github.com/user/my-lib.git@main"`)
		fmt.Println(`  }`)
		return
	}

	depsDir := ".necto" + string(os.PathSeparator) + "deps"
	if err := os.MkdirAll(depsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not create dependencies directory '%s': %v\n", depsDir, err)
		os.Exit(1)
	}

	fmt.Printf("Installing %d dependency(ies)...\n\n", len(cfg.Dependencies))

	installed := 0
	failed := 0

	for name, source := range cfg.Dependencies {
		// Parse source: "https://github.com/user/repo.git@branch" or "https://github.com/user/repo.git"
		repoURL := source
		branch := "main"
		if idx := strings.LastIndex(source, "@"); idx > 0 {
			repoURL = source[:idx]
			branch = source[idx+1:]
		}

		depPath := filepath.Join(depsDir, name)

		if info, err := os.Stat(depPath); err == nil && info.IsDir() {
			// Dependency already exists — pull latest
			fmt.Printf("  ↻ Updating '%s' (branch: %s)...\n", name, branch)
			cmd := exec.Command("git", "-C", depPath, "pull", "--ff-only")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "    ✗ Failed to update '%s': %v\n", name, err)
				failed++
				continue
			}
			fmt.Printf("    ✓ Updated '%s'\n", name)
		} else {
			// Clone fresh
			fmt.Printf("  ↓ Cloning '%s' from %s (branch: %s)...\n", name, repoURL, branch)
			cmd := exec.Command("git", "clone", "--depth", "1", "--branch", branch, repoURL, depPath)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "    ✗ Failed to clone '%s': %v\n", name, err)
				failed++
				continue
			}
			fmt.Printf("    ✓ Installed '%s'\n", name)
		}
		installed++
	}

	fmt.Printf("\nDone: %d installed, %d failed.\n", installed, failed)
	if installed > 0 {
		fmt.Printf("Dependencies are available in: %s/\n", depsDir)

		// Count .nc files in deps
		ncCount := 0
		filepath.Walk(depsDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(path, ".nc") {
				ncCount++
			}
			return nil
		})
		if ncCount > 0 {
			fmt.Printf("Found %d Necto source file(s) in dependencies.\n", ncCount)
		}
	}
}

func handleToolchain(action string) {
	switch action {
	case "status":
		path, source, err := codegen.FindClangCompiler()
		fmt.Println("==================================================================")
		fmt.Println("             Necto C/LLVM Backend Toolchain Status                ")
		fmt.Println("==================================================================")
		if err != nil {
			fmt.Println("Status: ✗ No C compiler backend found")
			fmt.Println("Searched locations:")
			fmt.Println("  1. Bundled toolchain: ./bin/clang/bin/clang.exe or ./toolchain/clang/")
			fmt.Println("  2. Isolated user toolchain: ~/.necto/toolchain/clang/bin/clang.exe")
			fmt.Println("  3. System PATH: clang or gcc")
			fmt.Println("\nRun 'necto toolchain install' to setup an isolated portable toolchain.")
		} else {
			fmt.Println("Status:    ✓ Ready for native AOT compilation (necto build)")
			fmt.Printf("Compiler:  %s\n", path)
			fmt.Printf("Source:    %s\n", source)
		}
		fmt.Println("==================================================================")

	case "install", "setup":
		installToolchain()

	default:
		fmt.Printf("Unknown toolchain action '%s'. Supported: status, install\n", action)
	}
}

func installToolchain() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining user home directory: %s\n", err)
		return
	}

	targetDir := filepath.Join(userHome, ".necto", "toolchain", "clang")
	os.MkdirAll(targetDir, 0755)

	fmt.Println("==================================================================")
	fmt.Println("             Necto Embedded Toolchain Setup                       ")
	fmt.Println("==================================================================")
	fmt.Printf("Target directory: %s\n\n", targetDir)

	// Check if already available
	if currentPath, source, err := codegen.FindClangCompiler(); err == nil {
		fmt.Printf("✓ Active compiler already detected: %s (%s)\n", currentPath, source)
		fmt.Println("Your system is already configured for native compilation!")
		return
	}

	fmt.Println("Configuring isolated toolchain directory...")
	fmt.Println("To complete standalone offline toolchain setup, copy or symlink")
	fmt.Printf("your portable clang distribution to: %s\n", targetDir)
	fmt.Println("==================================================================")
}

