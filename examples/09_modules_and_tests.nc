// 09_modules_and_tests.nc
// Демонстрация многофайловых модулей (import) и встроенных тестов (test / assert)

import { add, multiply, square } from "examples/math_utils.nc"

test "addition test from imported module" {
    let result = add(10, 25)
    assert(result == 35)
}

test "multiplication and square test" {
    assert(multiply(6, 7) == 42)
    assert(square(5) == 25)
}

test "simple logic test" {
    let flag = true
    assert(flag)
    assert(10 > 5)
}

fn main() {
    println("--- Necto v0.3.0 Modules and Unit Tests ---")
    println(f"add(15, 30) = {add(15, 30)}")
    println(f"multiply(7, 8) = {multiply(7, 8)}")
    println(f"square(9) = {square(9)}")
    println("Use 'necto test examples/09_modules_and_tests.nc' to run unit tests!")
}
