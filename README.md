# Necto Programming Language (v0.2.0-alpha)

**Necto** (от лат. *nectere / necto* — *«связывать воедино», «сплетать»*) — современный универсальный системный язык программирования, сплетающий воедино лучшие инженерные практики:
- **Синтез лучших концепций**:
  - Безопасность памяти и типов без `null` (`Option[T]`, `match`, иммутабельность `let` vs `let mut`), вдохновлённая Rust и Swift.
  - Простота, встроенные коллекции и системный I/O (`[T]`, `Map[K, V]`, `fs`, `os`), вдохновлённые Go и Python.
  - Нулевая стоимость абстракций и компиляция через LLVM без стоп-зе-ворлд пауз, вдохновлённая Zig и C.
- **Расширение файлов**: **`.nc`** (лаконично, как `.rs`, `.go`, `.zig`, `.ts`).
- **Два режима выполнения**:
  - **Мгновенный интерпретатор** (`necto run script.nc`) для быстрой разработки, скриптинга и REPL.
  - **Native Compiler** (`necto build script.nc -o app.exe`) через Clang 17 / LLVM, компилирующий код в самостоятельные нативные исполняемые файлы `.exe`.
- **Готовность к самохостингу (Self-Hosting)**: первый компонент будущего компилятора — лексический анализатор — уже успешно написан на самом языке Necto ([`examples/07_mini_lexer.nc`](file:///d:/Projects/newLanguage/examples/07_mini_lexer.nc))!

---

## 🚀 Быстрый старт

### Сборка тулчейна Necto
```powershell
go build -o bin/necto.exe ./cmd/necto
```

### Команды CLI Necto

1. **Запуск программы напрямую через интерпретатор:**
```powershell
.\bin\necto.exe run examples/01_hello.nc
```

2. **Работа с файлами и системным I/O:**
```powershell
.\bin\necto.exe run examples/05_file_io.nc
```

3. **Коллекции (динамические массивы и Map):**
```powershell
.\bin\necto.exe run examples/06_collections.nc
```

4. **Запуск лексера, написанного на чистом Necto:**
```powershell
.\bin\necto.exe run examples/07_mini_lexer.nc
```

5. **Компиляция в нативный бинарный файл (`.exe`) через Clang/LLVM:**
```powershell
.\bin\necto.exe build examples/02_fibonacci.nc -o fib.exe
.\fib.exe
```

6. **Статическая проверка типов (Type Check):**
```powershell
.\bin\necto.exe check examples/07_mini_lexer.nc
```

7. **Интерактивная консоль (REPL):**
```powershell
.\bin\necto.exe repl
necto> let mut m = Map.new()
necto> m.set("lang", "Necto")
necto> m["lang"]
"Necto"
```

---

## 📖 Обзор синтаксиса Necto

### 1. Переменные и иммутабельность
```necto
let pi = 3.14159             // Неизменяемая переменная (защита компилятором)
let mut counter: int = 0      // Изменяемая переменная
counter += 1
```

### 2. Динамические массивы (`[T]`)
```necto
let mut list: [int] = [10, 20]
list.push(30)
println(f"Length: {list.len()}") // 3

let last = list.pop()
match last {
    Some(val) => println(f"Popped: {val}"),
    None => println("Empty"),
}
```

### 3. Хеш-таблицы (`Map[K, V]`)
```necto
let mut scores = Map.new()
scores.set("Alice", 100)
scores.set("Bob", 85)

if scores.has("Alice") {
    let score = scores["Alice"]
    println(f"Alice's score: {score}")
}
```

### 4. Файловый I/O
```necto
// Запись в файл
let ok = fs.write_file("data.txt", "Hello Necto!")

// Безопасное чтение файла без исключений
match fs.read_file("data.txt") {
    Some(text) => println(f"Read: {text}"),
    None => println("File not found!"),
}
```

### 5. Строковые срезы и символы
```necto
let s = "Hello, Necto!"
println(f"Sub: {s.sub(7, 12)}")   // "Necto"
println(f"Code: {s.char_at(0)}")   // 72 (ASCII 'H')
```

### 6. Пользовательские структуры (Structs)
```necto
struct Player {
    id: int
    name: str
    hp: int
}

let mut hero = Player {
    id: 1,
    name: "Arthur",
    hp: 100,
}

hero.hp -= 25
println(f"Hero {hero.name} HP: {hero.hp}") // Hero Arthur HP: 75
```

---

## 📂 Структура проекта

```
d:/Projects/newLanguage/
├── bin/
│   └── necto.exe                 # Исполняемый файл CLI компилятора и рантайма
├── cmd/
│   └── necto/
│       └── main.go               # Точка входа CLI (run, build, check, repl)
├── pkg/
│   ├── token/
│   │   └── token.go              # Определения токенов языка
│   ├── lexer/
│   │   ├── lexer.go              # Лексический анализатор
│   │   └── lexer_test.go         # Тесты лексера
│   ├── ast/
│   │   └── ast.go                # Абстрактное синтаксическое дерево (AST)
│   ├── parser/
│   │   ├── parser.go             # Pratt Parser / рекурсивный спуск
│   │   └── parser_test.go        # Тесты парсера
│   ├── types/
│   │   ├── types.go              # Система типов Necto (Int, Float, Str, Option, Map, Struct)
│   │   ├── checker.go            # Type Checker и статический анализатор
│   │   └── checker_test.go       # Тесты системы типов
│   ├── eval/
│   │   ├── object.go             # Модель объектов среды выполнения
│   │   ├── environment.go        # Области видимости
│   │   ├── evaluator.go          # Tree-walk интерпретатор и встроенные модули (fs, os, Map)
│   │   └── evaluator_test.go     # Тесты интерпретатора
│   └── codegen/
│       └── codegen.go            # Генератор нативного C/LLVM кода под Clang
├── examples/
│   ├── 01_hello.nc               # Базовый синтаксис
│   ├── 02_fibonacci.nc           # Рекурсия и циклы (нативная сборка)
│   ├── 03_structs.nc             # Структуры данных и мутации
│   ├── 04_option_safety.nc       # Безопасность Option[T] и match
│   ├── 05_file_io.nc             # Системный ввод/вывод файлов
│   ├── 06_collections.nc         # Динамические массивы и Map
│   └── 07_mini_lexer.nc          # Лексер языка Necto, написанный на самом Necto!
├── go.mod                        # Модуль проекта necto
└── README.md                     # Документация проекта
```

---

## 🧪 Запуск всех тестов
```powershell
go test -v ./...
```
