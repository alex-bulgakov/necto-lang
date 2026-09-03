// 01_hello.Necto
// Базовый синтаксис, переменные, иммутабельность и интерполяция строк

fn main() {
    let language = "Necto"
    let version = "0.1.0-alpha"
    let release_year = 2026

    println(f"Welcome to {language} programming language (v{version})!")
    println(f"Built in {release_year} with modern language best practices.")

    let mut score: int = 100
    score += 50
    println(f"Current score: {score}")

    if score > 120 {
        println("Score threshold exceeded! Great job!")
    } else {
        println("Score is within normal limits.")
    }
}

