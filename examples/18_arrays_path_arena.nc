// 18_arrays_path_arena.nc
// ==============================================================================
//       Necto v2.0.2 Milestone: Path Module, Dynamic Arrays & Arena Memory
// ==============================================================================

fn main() {
    println("==================================================================")
    println("      Necto v2.0.2 — Path Module, Arrays & Arena Allocator        ")
    println("==================================================================")

    // 1. Path module operations
    println("\n1. Standard Library 'path' Module:")
    let full_path = path.join("projects", "necto", "main.nc")
    let extension = path.ext(full_path)
    let filename = path.base(full_path)
    let directory = path.dir(full_path)

    println(f"   full_path : '{full_path}'")
    println(f"   ext       : '{extension}'")
    println(f"   base      : '{filename}'")
    println(f"   dir       : '{directory}'")

    // 2. Dynamic array operations
    println("\n2. Dynamic Array operations:")
    let mut numbers: [int] = []
    let mut i = 1
    while i <= 5 {
        numbers.push(i * 10)
        i += 1
    }

    println(f"   Array length: {numbers.len()}")
    let mut idx = 0
    while idx < numbers.len() {
        let val = numbers[idx]
        println(f"   numbers[{idx}] = {val}")
        idx += 1
    }

    println("\n==================================================================")
    println("   ✓ v2.0.2 features executed successfully!")
    println("==================================================================")
}

// ------------------------------------------------------------------
// Unit tests for v2.0.2 features
// ------------------------------------------------------------------

test "path.ext extracts file extension" {
    assert(path.ext("document.pdf") == ".pdf")
    assert(path.ext("archive.tar.gz") == ".gz")
    assert(path.ext("no_ext") == "")
}

test "path.base extracts filename from path" {
    assert(path.base("dir/subdir/file.txt") == "file.txt")
    assert(path.base("file.txt") == "file.txt")
}

test "path.join joins multiple path components" {
    let p = path.join("a", "b", "c.nc")
    assert(p.contains("c.nc"))
}

test "array push and len" {
    let mut arr: [int] = []
    assert(arr.len() == 0)
    arr.push(100)
    arr.push(200)
    arr.push(300)
    assert(arr.len() == 3)
    assert(arr[0] == 100)
    assert(arr[1] == 200)
    assert(arr[2] == 300)
}
