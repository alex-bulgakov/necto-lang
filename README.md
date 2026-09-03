# Necto Programming Language (v0.10.0-alpha)

**Necto** (от лат. *nectere / necto* — *«связывать воедино», «сплетать»*) — современный универсальный системный язык программирования, сплетающий воедино лучшие инженерные практики:
- **Синтез лучших концепций**:
  - **Встроенный бенчмаркер производительности (`necto bench`)**: блоки `bench "name" { ... }` с замером времени на итерацию (`ns/op`) и пропускной способности (`ops/s`).
  - **Автоматический генератор документации (`necto doc`)**: интерактивный современный HTML-сайт с поиском по doc-комментариям `///`.
  - **Встроенный сетевой стек и веб-сервер (`http` & `net`)**: легковесный HTTP-сервер (`http.listen`), HTTP-клиент (`http.get`, `http.post`) и сокеты (`net.tcp_connect`).
  - **C Foreign Function Interface (FFI)** через блоки `extern "C"` (`extern "C" { fn sqrt(x: float) -> float }`).
  - Объектные методы и блоки **`impl`** с поддержкой `self` (`impl Lexer { fn next_char(mut self) { ... } }`).
  - Эргономичная и безопасная обработка ошибок через **`Result[T, E]`** (`Result.Ok`, `Result.Err`) и оператор распространения ошибок **`?`** (`let data = fs.read_file("file.nc")?`).
  - Алгебраические типы данных с полезной нагрузкой (**Tagged Enums**) и сопоставление с образцом (`match`), вдохновлённые Rust и Swift.
  - Умные указатели **`Box[T]`** для рекурсивных структур данных (деревья синтаксиса, графы, списки).
  - Простота, встроенные коллекции, многофайловые модули и системный I/O (`[T]`, `Map[K, V]`, `import`, `fs`, `os`), вдохновлённые Go и Python.
  - Нулевая стоимость абстракций и компиляция через LLVM без пауз GC, вдохновлённая Zig и C.
- **Встроенный инструмент тестирования**: блоки `test "name" { assert(...) }` и команда `necto test`.
- **Расширение файлов**: **`.nc`** (лаконично, как `.rs`, `.go`, `.zig`, `.ts`).
- **Два режима выполнения**:
  - **Мгновенный интерпретатор** (`necto run script.nc`) для быстрой разработки, скриптинга и REPL.
  - **Native Compiler** (`necto build script.nc -o app.exe`) через Clang 17 / LLVM, компилирующий код в самостоятельные нативные исполняемые файлы `.exe`.
- **Самохостинговый компилятор и Полная автономность (Stage 1, 2, 3)**:
  - Компилятор Necto полностью реализован на самом языке Necto в пакете [**`compiler/`**](compiler/)!
  - Команда `necto bootstrap` автономно собирает нативный бинарник `bin/necto-native.exe`, компилирующий любые программы Necto без зависимости от Go!
- **Проектная система и инструменты**:
  - `necto init [name]` — создание нового проекта с манифестом `necto.json`.
  - `necto fmt` — автоматический форматировщик кода.
  - Официальное расширение для VS Code в каталоге [**`editors/vscode/`**](editors/vscode/).

- **Официальная документация:**
  - Полная спецификация языка: [**`SPECIFICATION.md`**](SPECIFICATION.md)
  - Официальное расширение для VS Code: [**`editors/vscode/`**](editors/vscode/)

---

## 🚀 Быстрый старт

### Сборка тулчейна Necto
```powershell
go build -o bin/necto.exe ./cmd/necto
```

### Команды CLI Necto

1. **Инициализация нового проекта (`necto init`):**
```powershell
.\bin\necto.exe init my_app
cd my_app
.\..\bin\necto.exe run
```

2. **Автоматическое форматирование кода (`necto fmt`):**
```powershell
.\bin\necto.exe fmt .
.\bin\necto.exe fmt --check
```

3. **Запуск встроенного HTTP веб-сервера:**
```powershell
.\bin\necto.exe run examples/12_http_server.nc
```

4. **Запуск C FFI примера:**
```powershell
.\bin\necto.exe run examples/11_c_interop.nc
.\bin\necto.exe build examples/11_c_interop.nc -o c_app.exe
```

5. **Автономная сборка компилятора (`necto bootstrap`):**
```powershell
.\bin\necto.exe bootstrap
.\bin\necto-native.exe
```

6. **Запуск встроенных юнит-тестов (`necto test`):**
```powershell
.\bin\necto.exe test examples/13_benchmarks_and_docs.nc
```

7. **Запуск бенчмарков производительности (`necto bench`):**
```powershell
.\bin\necto.exe bench examples/13_benchmarks_and_docs.nc
```

8. **Генерация интерактивной документации (`necto doc`):**
```powershell
.\bin\necto.exe doc . --output docs
# Для запуска локального веб-сервера документации:
.\bin\necto.exe doc . --serve
```

9. **Компиляция Necto-файла в нативный исполняемый файл `.exe` через Clang/LLVM:**
```powershell
.\bin\necto.exe build examples/02_fibonacci.nc -o fib.exe
.\fib.exe
```

10. **Статическая проверка типов (Type Check):**
```powershell
.\bin\necto.exe check compiler/main.nc
```

11. **Интерактивная консоль (REPL):**
```powershell
.\bin\necto.exe repl
```

---

## 📖 Новые возможности Necto (v0.4.0)

### 1. Методы структур (`impl` блоки)
```necto
struct Lexer {
    source: str
    cursor: int
}

impl Lexer {
    fn new(src: str) -> Lexer {
        return Lexer { source: src, cursor: 0 }
    }

    fn next_char(mut self) -> int {
        let c = self.source.char_at(self.cursor)
        self.cursor += 1
        return c
    }
}

let mut lex = Lexer.new("fn main() {}")
let c = lex.next_char()
```

### 2. Тип `Result[T, E]` и оператор распространения ошибок `?`
```necto
fn compile_source(path: str) -> Result[str, str] {
    let source = fs.read_file(path)?  // При ошибке функция немедленно вернет Result.Err
    let tokens = tokenize(source)?
    return Result.Ok("Compilation successful")
}
```

### 3. Tagged Enums и сопоставление с образцом
```necto
enum Token {
    Number(int)
    Ident(str)
    Eof
}

let tok = Token.Number(42)

match tok {
    Token.Number(val) => println(f"Number: {val}"),
    Token.Ident(name) => println(f"Ident: {name}"),
    Token.Eof         => println("EOF"),
}
```

### 4. Умный указатель `Box[T]` для рекурсивных структур данных
```necto
enum Expr {
    Num(int)
    Binary(op: str, left: Box[Expr], right: Box[Expr])
}
```

---

## 📂 Структура демонстрационных примеров

| Файл | Описание |
| :--- | :--- |
| [`01_hello.nc`](file:///d:/Projects/newLanguage/examples/01_hello.nc) | Базовый синтаксис, переменные, ветвления |
| [`02_fibonacci.nc`](file:///d:/Projects/newLanguage/examples/02_fibonacci.nc) | Рекурсия, производительность циклов, нативная сборка |
| [`03_structs.nc`](file:///d:/Projects/newLanguage/examples/03_structs.nc) | Структуры данных и изменяемость полей |
| [`04_option_safety.nc`](file:///d:/Projects/newLanguage/examples/04_option_safety.nc) | Безопасность `Option[T]` без `null` |
| [`05_file_io.nc`](file:///d:/Projects/newLanguage/examples/05_file_io.nc) | Системный файловый I/O (`fs.read_file`, `fs.write_file`) |
| [`06_collections.nc`](file:///d:/Projects/newLanguage/examples/06_collections.nc) | Динамические списки `[T]`, `Map[K, V]` и срезы строк |
| [`07_mini_lexer.nc`](file:///d:/Projects/newLanguage/examples/07_mini_lexer.nc) | Лексер языка Necto на чистом Necto |
| [`08_mini_parser.nc`](file:///d:/Projects/newLanguage/examples/08_mini_parser.nc) | AST-парсер и вычислитель деревьев на чистом Necto |
| [`09_modules_and_tests.nc`](file:///d:/Projects/newLanguage/examples/09_modules_and_tests.nc) | Многофайловый `import` и запуск юнит-тестов (`necto test`) |
| [`10_self_compiler_pipeline.nc`](file:///d:/Projects/newLanguage/examples/10_self_compiler_pipeline.nc) | **Сквозной компилятор: методы (impl), Result, оператор ?, Codegen** |

---

## 🧪 Запуск полного набора тестов компилятора
```powershell
go test -v ./...
```
