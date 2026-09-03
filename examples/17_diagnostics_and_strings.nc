// 17_diagnostics_and_strings.nc
// ==============================================================================
//        Necto v2.0.1 — Quality & Stabilization: String Methods & F-Strings
// ==============================================================================

fn main() {
    println("==================================================================")
    println("      Necto v2.0.1 — String Methods & Quality Milestone           ")
    println("==================================================================")

    // 1. starts_with and ends_with
    let filename = "project_spec.necto"
    let is_necto = filename.ends_with(".necto")
    let is_proj = filename.starts_with("project_")
    println(f"File: '{filename}'")
    println(f"  ends_with('.necto') : {is_necto}")
    println(f"  starts_with('project_'): {is_proj}")

    // 2. trim
    let messy = "   hello, world!   "
    let clean = messy.trim()
    println(f"Trim: '{messy}' -> '{clean}'")

    // 3. split
    let csv = "apple,banana,orange,grape"
    let fruits = csv.split(",")
    println(f"Split '{csv}' into {fruits.len()} items:")
    for i in 0..fruits.len() {
        let fruit = fruits[i]
        println(f"  [{i}] {fruit}")
    }

    println("==================================================================")
    println("   ✓ All string operations executed successfully!")
    println("==================================================================")
}

// ------------------------------------------------------------------
// Unit tests for v2.0.1 String features
// ------------------------------------------------------------------

test "str.starts_with returns true on matching prefix" {
    let s = "https://necto-lang.org"
    assert(s.starts_with("https://"))
    assert(!s.starts_with("ftp://"))
}

test "str.ends_with returns true on matching suffix" {
    let s = "archive.tar.gz"
    assert(s.ends_with(".tar.gz"))
    assert(s.ends_with(".gz"))
    assert(!s.ends_with(".zip"))
}

test "str.trim removes surrounding whitespace" {
    let s = "  \t\n  padded string  \n  "
    let t = s.trim()
    assert(t == "padded string")
}

test "str.split splits by single character delimiter" {
    let line = "user:1000:admin:/home/user"
    let parts = line.split(":")
    assert(parts.len() == 4)
    assert(parts[0] == "user")
    assert(parts[1] == "1000")
    assert(parts[2] == "admin")
    assert(parts[3] == "/home/user")
}

test "str.split returns single element when delimiter not found" {
    let single = "single_word"
    let parts = single.split(",")
    assert(parts.len() == 1)
    assert(parts[0] == "single_word")
}
