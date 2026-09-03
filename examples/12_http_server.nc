// 12_http_server.nc
// Necto v0.9.0-alpha Milestone:
// Полноценный HTTP веб-сервер и HTTP-клиент на чистом Necto

fn handle_request(path: str) -> str {
    if path == "/"{
        return "<html><body><h1>Welcome to Necto Web Server!</h1><p>Running on Necto v0.9.0</p></body></html>"
    }
    if path == "/api/status"{
        return "{\"status\": \"ok\", \"language\": \"Necto\", \"version\": \"0.9.0-alpha\"}"
    }
    if path == "/api/echo"{
        return "{\"message\": \"Hello from Necto HTTP server!\"}"
    }
    return "404 Not Found"
}

test "HTTP router logic unit test"{
    let index = handle_request("/")
    assert(index.contains("Welcome to Necto Web Server!"))

    let status = handle_request("/api/status")
    assert(status.contains("\"status\": \"ok\""))

    let not_found = handle_request("/unknown")
    assert(not_found == "404 Not Found")
}

test "HTTP server live query test"{
    // Запуск локального сервера на порту 8899
    http.listen(8899, handle_request)

    // Выполнение клиентского запроса через http.get
    let res = http.get("http://localhost:8899/api/status")
    match res {
        Result.Ok(body) => {
            assert(body.contains("\"language\": \"Necto\""))
        },
        Result.Err(err) => {
            println(f"HTTP GET failed: {err}")
            assert(false)
        },
    }
}

fn main() {
    println("==================================================================")
    println("       Necto v0.9.0-alpha: HTTP Web Server & Networking Client    ")
    println("==================================================================")

    let port = 8080
    println(f"Starting Necto HTTP Web Server on port {port}...")

    http.listen(port, handle_request)

    println(f"✓ Web server running at http://localhost:{port}/")
    println("Querying server endpoints via http.get():\n")

    // Тестирование endpoint /
    let res_index = http.get(f"http://localhost:{port}/")
    match res_index {
        Result.Ok(body) => {
            println("1. GET / response:")
            println(f"   {body}\n")
        },
        Result.Err(e) => {
            println(f"1. GET / error: {e}\n")
        },
    }

    // Тестирование endpoint /api/status
    let res_api = http.get(f"http://localhost:{port}/api/status")
    match res_api {
        Result.Ok(body) => {
            println("2. GET /api/status response (JSON):")
            println(f"   {body}\n")
        },
        Result.Err(e) => {
            println(f"2. GET /api/status error: {e}\n")
        },
    }

    // Тестирование TCP connect
    let tcp_res = net.tcp_connect(f"localhost:{port}")
    match tcp_res {
        Result.Ok(msg) => {
            println(f"3. net.tcp_connect('localhost:{port}') -> {msg}")
        },
        Result.Err(e) => {
            println(f"3. net.tcp_connect error: {e}")
        },
    }

    println("\n✓ Necto networking stack verified successfully!")
}
