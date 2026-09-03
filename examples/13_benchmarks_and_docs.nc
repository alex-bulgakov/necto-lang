// 13_benchmarks_and_docs.nc
// Necto v0.10.0-alpha Milestone:
// Документирование кода (/// doc-комментарии) и замеры производительности (bench)

/// Структура точки в двумерном пространстве
struct Point {
    x: int,
    y: int
}

/// Вычисляет N-ое число Фибоначчи рекурсивным способом.
/// Внимание: имеет экспоненциальную сложность O(2^n).
fn fib_recursive(n: int) -> int {
    if n <= 1 {
        return n
    }
    return fib_recursive(n - 1) + fib_recursive(n - 2)
}

/// Вычисляет N-ое число Фибоначчи линейным итеративным алгоритмом.
/// Имеет линейную сложность O(n) и минимальное потребление памяти.
fn fib_iterative(n: int) -> int {
    if n <= 1 {
        return n
    }
    let mut a = 0
    let mut b = 1
    for i in 2..(n + 1) {
        let temp = a + b
        a = b
        b = temp
    }
    return b
}

/// Проверка корректности математических функций
test "Fibonacci correctness"{
    assert(fib_recursive(10) == 55)
    assert(fib_iterative(10) == 55)
    assert(fib_iterative(20) == 6765)
}

// ------------------------------------------------------------------
// Блоки встроенных бенчмарков производительности
// Запуск: necto bench examples/13_benchmarks_and_docs.nc
// ------------------------------------------------------------------

bench "recursive fibonacci (n=12)"{
    fib_recursive(12)
}

bench "iterative fibonacci (n=30)"{
    fib_iterative(30)
}

bench "loop accumulation (1..100)"{
    let mut sum = 0
    for i in 1..101 {
        sum += i
    }
}

fn main() {
    println("==================================================================")
    println("       Necto v0.10.0-alpha: Benchmarking & Doc Generation         ")
    println("==================================================================")
    println("This file contains documentation comments (///) and benchmarks.")
    println("\nTo run performance benchmarks:")
    println("  necto bench examples/13_benchmarks_and_docs.nc")
    println("\nTo generate interactive HTML documentation:")
    println("  necto doc examples/13_benchmarks_and_docs.nc --output docs")
    println("==================================================================")

    let f10 = fib_iterative(10)
    println(f"Calculated fib(10) = {f10}")
}
