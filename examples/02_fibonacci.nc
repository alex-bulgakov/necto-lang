// 02_fibonacci.Necto
// Вычисления, функции, рекурсия и циклы

fn fib(n: int) -> int {
    if n <= 1 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

fn main() {
    println("--- Fibonacci Sequence in Necto ---")
    
    // Цикл for in range
    for i in 0..10 {
        let val = fib(i)
        println(f"fib({i}) = {val}")
    }

    println("--- Loop Performance Check ---")
    let mut acc = 0
    let mut k = 1
    while k <= 100 {
        acc += k
        k += 1
    }
    println(f"Sum of 1..100 = {acc}")
}

