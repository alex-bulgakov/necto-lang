// 03_structs.Necto
// Пользовательские структуры, поля, композиция и мутации

struct Vector2 {
    x: int
    y: int
}

struct Player {
    id: int
    name: str
    hp: int
}

fn print_player(p: Player) {
    println(f"Player #{p.id} '{p.name}' | HP: {p.hp}")
}

fn main() {
    println("--- Testing Structs in Necto ---")
    
    let mut hero = Player {
        id: 1,
        name: "Arthur",
        hp: 100,
    }

    print_player(hero)

    // Изменение поля структуры
    hero.hp -= 25
    println("After receiving damage:")
    print_player(hero)

    let v1 = Vector2 { x: 5, y: 12 }
    println(f"Vector position: ({v1.x}, {v1.y})")
}

