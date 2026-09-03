// 04_option_safety.Necto
// Безопасность памяти и типов: Option[T] вместо null и Pattern Matching

fn find_user_role(user_id: int) -> Option[str] {
    if user_id == 1 {
        return Some("Administrator")
    }
    if user_id == 2 {
        return Some("Developer")
    }
    return None
}

fn describe_role(user_id: int) {
    let opt = find_user_role(user_id)

    match opt {
        Some(role) => println(f"User #{user_id} has role: {role}"),
        None => println(f"User #{user_id} was not found in database!"),
    }
}

fn main() {
    println("--- Option & Null-Safety in Necto ---")
    describe_role(1)
    describe_role(2)
    describe_role(99)
}

