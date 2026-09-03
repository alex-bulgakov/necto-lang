// 06_collections.Necto
// Динамические массивы, хеш-таблицы (Map), строковые операции и аргументы ОС

fn test_dynamic_arrays() {
    println("--- 1. Dynamic Arrays ---")
    let mut numbers: [int] = [10, 20]
    println(f"Initial length: {numbers.len()}")

    numbers.push(30)
    numbers.push(40)
    println(f"After push(30, 40), length: {numbers.len()}")

    let popped = numbers.pop()
    match popped {
        Some(val) => println(f"Popped value: {val}"),
        None => println("Array was empty"),
    }
    println(f"Final length: {numbers.len()}")
}

fn test_hash_maps() {
    println("--- 2. Hash Maps (Map[K, V]) ---")
    let mut scores = Map.new()

    scores.set("Alice", 100)
    scores.set("Bob", 85)
    scores.set("Charlie", 92)

    println(f"Map size: {scores.len()}")
    let bob_score = scores["Bob"]
    println(f"Bob's score via index: {bob_score}")

    if scores.has("Alice") {
        println("Alice is in the leaderboard!")
    }

    let lookup = scores.get("Dave")
    match lookup {
        Some(score) => println(f"Dave's score: {score}"),
        None => println("Dave was not found in map (handled safely)"),
    }
}

fn test_string_manipulation() {
    println("--- 3. String Manipulation ---")
    let sentence = "Compiler in Necto is coming soon!"
    println(f"Length: {sentence.len()}")

    let first_word = sentence.sub(0, 8)
    println(f"Sub(0, 8): '{first_word}'")

    let ascii_code = sentence.char_at(0)
    println(f"First character ASCII code: {ascii_code}")

    if sentence.contains("Necto") {
        println("The sentence contains 'Necto'!")
    }
}

fn main() {
    test_dynamic_arrays()
    test_hash_maps()
    test_string_manipulation()

    println("--- 4. OS CLI Arguments ---")
    let args = os.args()
    println(f"Process executed with {args.len()} arguments")
}

