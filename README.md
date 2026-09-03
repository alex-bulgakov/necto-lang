# Necto Programming Language (v0.3.0-alpha)

**Necto** (от лат. *nectere / necto* — *«связывать воедино», «сплетать»*) — современный универсальный системный язык программирования, сплетающий воедино лучшие инженерные практики:
- **Синтез лучших концепций**:
  - Алгебраические типы данных с полезной нагрузкой (**Tagged Enums**) и сопоставление с образцом (`match`), вдохновлённые Rust и Swift.
  - Умные указатели **`Box[T]`** для рекурсивных структур данных (деревья синтаксиса, графы, списки).
  - Простота, встроенные коллекции, многофайловые модули и системный I/O (`[T]`, `Map[K, V]`, `import`, `fs`, `os`), вдохновлённые Go и Python.
  - Нулевая стоимость абстракций и компиляция через LLVM без пауз GC, вдохновлённая Zig и C.
- **Встроенный инструмент тестирования**: блоки `test "name" { assert(...) }` и команда `necto test`.
- **Расширение файлов**: **`.nc`** (лаконично, как `.rs`, `.go`, `.zig`, `.ts`).
- **Два режима выполнения**:
  - **Мгновенный интерпретатор** (`necto run script.nc`) для быстрой разработки, скриптинга и REPL.
  - **Native Compiler** (`necto build script.nc -o app.exe`) через Clang 17 / LLVM, компилирующий код в самостоятельные нативные исполняемые файлы `.exe`.
- **Путь к самохостингу (Self-Hosting)**:
  - Компонент 1: Лексер языка Necto на чистом Necto ([`examples/07_mini_lexer.nc`](file:///d:/Projects/newLanguage/examples/07_mini_lexer.nc)).
  - Компонент 2: **Синтаксический AST-парсер и вычислитель дерева выражений на чистом Necto** ([`examples/08_mini_parser.nc`](file:///d:/Projects/newLanguage/examples/08_mini_parser.nc))!

---

## 🚀 Быстрый старт

### Сборка тулчейна Necto
```powershell
go build -o bin/necto.exe ./cmd/necto
```

### Команды CLI Necto

1. **Запуск программы напрямую через интерпретатор:**
```powershell
.\bin\necto.exe run examples/08_mini_parser.nc
```

2. **Запуск встроенных юнит-тестов (`necto test`):**
```powershell
.\bin\necto.exe test examples/09_modules_and_tests.nc
```

3. **Работа с файлами и системным I/O:**
```powershell
.\bin\necto.exe run examples/05_file_io.nc
```

4. **Коллекции (динамические массивы и Map):**
```powershell
.\bin\necto.exe run examples/06_collections.nc
```

5. **Компиляция в нативный бинарный файл (`.exe`) через Clang/LLVM:**
```powershell
.\bin\necto.exe build examples/02_fibonacci.nc -o fib.exe
.\fib.exe
```

6. **Статическая проверка типов (Type Check):**
```powershell
.\bin\necto.exe check examples/08_mini_parser.nc
```

7. **Интерактивная консоль (REPL):**
```powershell
.\bin\necto.exe repl
necto> let b = Box.new(42)
necto> b.unwrap()
42
```

---

## 📖 Новые возможности Necto (v0.3.0)

### 1. Tagged Enums (Алгебраические типы данных)
```necto
enum Token {
    Number(int)
    Ident(str)
    Plus
    Eof
}

let tok = Token.Number(42)

match tok {
    Token.Number(val) => println(f"Number: {val}"),
    Token.Ident(name) => println(f"Identifier: {name}"),
    Token.Plus        => println("Plus operator"),
    Token.Eof         => println("End of file"),
}
```

### 2. Умный указатель `Box[T]` для рекурсивных структур (AST)
```necto
enum Expr {
    Number(int)
    Binary(op: str, left: Box[Expr], right: Box[Expr])
}

// Построение дерева: 2 + (3 * 4)
let mul = Expr.Binary("*", Box.new(Expr.Number(3)), Box.new(Expr.Number(4)))
let root = Expr.Binary("+", Box.new(Expr.Number(2)), Box.new(mul))
```

### 3. Модульная система (`import`)
```necto
// math_utils.nc
fn add(a: int, b: int) -> int { return a + b }

// main.nc
import { add } from "examples/math_utils.nc"

fn main() {
    println(f"Result: {add(10, 20)}")
}
```

### 4. Встроенное тестирование (`test` и `assert`)
```necto
test "addition works" {
    assert(2 + 2 == 4)
    assert(10 > 5)
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
| [`07_mini_lexer.nc`](file:///d:/Projects/newLanguage/examples/07_mini_lexer.nc) | Лексер языка Necto, написанный на 100% чистом Necto |
| [`08_mini_parser.nc`](file:///d:/Projects/newLanguage/examples/08_mini_parser.nc) | **AST-парсер и вычислитель деревьев на чистом Necto** |
| [`09_modules_and_tests.nc`](file:///d:/Projects/newLanguage/examples/09_modules_and_tests.nc) | **Многофайловый `import` и запуск встроенных `test`** |

---

## 🧪 Запуск всех тестов компилятора
```powershell
go test -v ./...
```
