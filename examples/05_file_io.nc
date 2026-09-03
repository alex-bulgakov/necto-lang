// 05_file_io.Necto
// Системный ввод/вывод: создание, запись и чтение файлов

fn main() {
    println("--- File I/O Demonstration in Necto v0.2.0 ---")

    let filepath = "sample_data.txt"
    let content = "Hello from Necto v0.2.0 File System!\nNecto is moving closer to self-hosting."

    println(f"Writing to '{filepath}'...")
    let ok = fs.write_file(filepath, content)
    if ok {
        println("File written successfully!")
    } else {
        println("Failed to write file.")
        return
    }

    println(f"Reading from '{filepath}'...")
    let read_result = fs.read_file(filepath)

    match read_result {
        Some(text) => {
            println("File Content:")
            println(text)
            println(f"Total characters read: {text.len()}")
        },
        None => println("Could not read file!"),
    }
}

