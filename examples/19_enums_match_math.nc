// 19_enums_match_math.nc
// ==============================================================================
//       Necto v2.0.3 Milestone: Tagged Enums, Match, Math Module & String Ops
// ==============================================================================

enum HttpStatus {
    Ok(int)
    NotFound(str)
    ServerError(str)
}

fn status_message(s: HttpStatus) -> str {
    match s {
        HttpStatus.Ok(code) => {
            return f"Status OK: code {code}"
        },
        HttpStatus.NotFound(msg) => {
            return f"Not Found: {msg}"
        },
        HttpStatus.ServerError(err) => {
            return f"Internal Error: {err}"
        },
    }
}

fn main() {
    println("==================================================================")
    println("      Necto v2.0.3 — Tagged Enums, Math & String Methods          ")
    println("==================================================================")

    // 1. Math module
    println("\n1. Standard Library 'math' Module:")
    let neg = -42
    let pos = math.abs(neg)
    let m_min = math.min(10, 25)
    let m_max = math.max(10, 25)
    let power = math.pow(2.0, 8.0)
    let root = math.sqrt(81.0)
    let rounded = math.round(3.7)

    println(f"   math.abs({neg})    = {pos}")
    println(f"   math.min(10, 25)  = {m_min}")
    println(f"   math.max(10, 25)  = {m_max}")
    println(f"   math.pow(2, 8)    = {power}")
    println(f"   math.sqrt(81)     = {root}")
    println(f"   math.round(3.7)   = {rounded}")

    // 2. String transformation methods
    println("\n2. String Transformations:")
    let raw = "Hello, NECTO Language!"
    let lower = raw.to_lower()
    let upper = raw.to_upper()
    let replaced = raw.replace("NECTO", "World")

    println(f"   original : '{raw}'")
    println(f"   to_lower : '{lower}'")
    println(f"   to_upper : '{upper}'")
    println(f"   replace  : '{replaced}'")

    // 3. Tagged Enums & Pattern Matching
    println("\n3. Tagged Enums & Match:")
    let s1 = HttpStatus.Ok(200)
    let s2 = HttpStatus.NotFound("/api/missing")
    println(f"   s1: {status_message(s1)}")
    println(f"   s2: {status_message(s2)}")

    println("\n==================================================================")
    println("   ✓ All v2.0.3 operations executed successfully!")
    println("==================================================================")
}

// ------------------------------------------------------------------
// Unit tests for v2.0.3 features
// ------------------------------------------------------------------

test "math.abs works on negative and positive values" {
    assert(math.abs(-100) == 100)
    assert(math.abs(50) == 50)
}

test "math.min and math.max work correctly" {
    assert(math.min(5, 12) == 5)
    assert(math.max(5, 12) == 12)
}

test "math.pow and math.sqrt calculations" {
    assert(math.pow(2.0, 4.0) == 16.0)
    assert(math.sqrt(49.0) == 7.0)
}

test "str.to_lower and str.to_upper transformations" {
    let s = "Hello World"
    assert(s.to_lower() == "hello world")
    assert(s.to_upper() == "HELLO WORLD")
}

test "str.replace replaces all matching substrings" {
    let s = "foo bar foo"
    let res = s.replace("foo", "baz")
    assert(res == "baz bar baz")
}

test "tagged enum pattern matching" {
    let ok_status = HttpStatus.Ok(200)
    assert(status_message(ok_status) == "Status OK: code 200")
    let err_status = HttpStatus.NotFound("page")
    assert(status_message(err_status) == "Not Found: page")
}
