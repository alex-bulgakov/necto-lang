// 11_c_interop.nc
// Necto v0.7.0-alpha Milestone:
// Прямой вызов функций Си-библиотек через блоки extern "C" (C FFI)

extern "C"{
    fn sqrt(x: float) -> float
    fn abs(x: int) -> int
    fn sin(x: float) -> float
    fn pow(base: float, exp: float) -> float
}

test "C FFI sqrt and pow calculation"{
    let r = sqrt(144.0)
    assert(r == 12.0)

    let p = pow(2.0, 8.0)
    assert(p == 256.0)
}

test "C FFI abs calculation"{
    let negative = -42
    let positive = abs(negative)
    assert(positive == 42)
}

fn main() {
    println("==================================================================")
    println("       Necto v0.7.0-alpha: C Foreign Function Interface (FFI)     ")
    println("==================================================================")

    let num = 625.0
    let root = sqrt(num)
    println(f"1. sqrt({num}) from C libc = {root}")

    let val = -999
    let absolute = abs(val)
    println(f"2. abs({val}) from C libc = {absolute}")

    let base = 3.0
    let exp = 4.0
    let power = pow(base, exp)
    println(f"3. pow({base}, {exp}) from C libc = {power}")

    let angle = 0.0
    let s = sin(angle)
    println(f"4. sin({angle}) from C libc = {s}")

    println("\n✓ Direct C Foreign Function Interface successfully verified!")
}
