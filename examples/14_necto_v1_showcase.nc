// 14_necto_v1_showcase.nc
// ==============================================================================
//           Necto Programming Language — Version 1.0.0 Golden Showcase
// ==============================================================================
// Этот файл демонстрирует все ключевые концепции языка Necto 1.0 в единой программе:
// 1. Алгебраические типы данных (Tagged Enums) и сопоставление с образцом (match)
// 2. Объектные методы (impl) и структуры (struct)
// 3. Безопасная обработка ошибок (Result[T, E] и оператор ?)
// 4. Умные указатели (Box[T])
// 5. C Foreign Function Interface (extern "C")
// 6. Встроенный сетевой стек и веб-сервер (http & net)
// 7. Встроенные юнит-тесты (test) и бенчмарки производительности (bench)
// 8. Doc-комментарии (///)

extern "C"{
    fn sqrt(x: float) -> float
    fn abs(x: int) -> int
}

/// Варианты маршрутизации HTTP-сервера
enum Route {
    Home,
    ApiStatus,
    ApiCompute(int),
    NotFound
}

/// Конфигурация веб-сервиса Necto
struct ServerConfig {
    host: str,
    port: int,
    service_name: str
}

impl ServerConfig {
    /// Формирует адрес для подключения
    fn address(self) -> str {
        return f"{self.host}:{self.port}"
    }

    /// Формирует текстовое описание сервиса
    fn summary(self) -> str {
        return f"Service '{self.service_name}' running on {self.address()}"
    }
}

/// Парсинг строкового пути в типизированный маршрут
fn route_request(path: str) -> Route {
    if path == "/"{
        return Route.Home
    }
    if path == "/api/status"{
        return Route.ApiStatus
    }
    if path == "/api/compute"{
        return Route.ApiCompute(42)
    }
    return Route.NotFound
}

/// Обработка входящего HTTP-запроса через pattern matching
fn handle_request(path: str) -> str {
    let r = route_request(path)
    match r {
        Route.Home => {
            return "<html><body><h1>Necto v1.0.0 Golden Release</h1><p>Welcome to the future of systems programming!</p></body></html>"
        },
        Route.ApiStatus => {
            return "{\"status\": \"operational\", \"version\": \"1.0.0\", \"language\": \"Necto\"}"
        },
        Route.ApiCompute(val) => {
            let root = sqrt(1764.0) // 42.0
            return "{\"input\": 42, \"sqrt\": 42.0}"
        },
        Route.NotFound => {
            return "404 Not Found"
        },
    }
}

/// Функция безопасного деления с возвратом Result[int, str]
fn safe_divide(a: int, b: int) -> Result[int, str] {
    if b == 0 {
        return Result.Err("division by zero")
    }
    return Result.Ok(a / b)
}

// ------------------------------------------------------------------
// Встроенный набор юнит-тестов (necto test)
// ------------------------------------------------------------------

test "ServerConfig and impl methods test"{
    let cfg = ServerConfig {
        host: "localhost",
        port: 9090,
        service_name: "NectoService"
    }
    assert(cfg.address() == "localhost:9090")
    assert(cfg.summary().contains("NectoService"))
}

test "C FFI interop test"{
    let s = sqrt(100.0)
    assert(s == 10.0)
    let a = abs(-123)
    assert(a == 123)
}

test "Result and Error Handling test" {
    let ok_res = safe_divide(100, 5)
    match ok_res {
        Result.Ok(val) => {
            assert(val == 20)
        },
        Result.Err(_) => {
            assert(false)
        },
    }

    let err_res = safe_divide(10, 0)
    match err_res {
        Result.Ok(_) => {
            assert(false)
        },
        Result.Err(msg) => {
            assert(msg == "division by zero")
        },
    }
}

test "Box smart pointer test" {
    let boxed = Box.new(999)
    assert(boxed.unwrap() == 999)
}

// ------------------------------------------------------------------
// Встроенные бенчмарки производительности (necto bench)
// ------------------------------------------------------------------

bench "method dispatch on struct" {
    let cfg = ServerConfig {
        host: "localhost",
        port: 8080,
        service_name: "Benchmark"
    }
    cfg.summary()
}

bench "C FFI sqrt calculation" {
    sqrt(144.0)
}

bench "router pattern matching" {
    handle_request("/api/status")
}

// ------------------------------------------------------------------
// Главная точка входа программы
// ------------------------------------------------------------------

fn main() {
    println("==================================================================")
    println("           NECTO PROGRAMMING LANGUAGE — VERSION 1.0.0            ")
    println("                      The Golden Release                          ")
    println("==================================================================")

    let cfg = ServerConfig {
        host: "localhost",
        port: 8080,
        service_name: "Necto-Core-Gateway"
    }
    println(cfg.summary())

    // 1. Демонстрация C FFI
    let c_val = sqrt(256.0)
    println(f"1. C Foreign Function Interface: sqrt(256.0) = {c_val}")

    // 2. Демонстрация Result
    let div_res = safe_divide(1000, 25)
    match div_res {
        Result.Ok(ans) => {
            println(f"2. Error Handling: 1000 / 25 = {ans}")
        },
        Result.Err(e) => {
            println(f"2. Error Handling Failed: {e}")
        },
    }

    // 3. Запуск встроенного веб-сервера
    println(f"\n3. Starting Embedded HTTP Web Server on port {cfg.port}...")
    http.listen(cfg.port, handle_request)
    println(f"   ✓ Server active at http://{cfg.address()}/")

    // 4. Клиентские запросы через http.get
    println("\n4. Querying endpoints with Necto HTTP Client:")
    let res_status = http.get(f"http://{cfg.address()}/api/status")
    match res_status {
        Result.Ok(body) => println(f"   GET /api/status -> {body}"),
        Result.Err(e) => println(f"   GET /api/status error: {e}"),
    }

    let res_compute = http.get(f"http://{cfg.address()}/api/compute")
    match res_compute {
        Result.Ok(body) => println(f"   GET /api/compute -> {body}"),
        Result.Err(e) => println(f"   GET /api/compute error: {e}"),
    }

    // 5. Проверка TCP сокетов
    let net_res = net.tcp_connect(cfg.address())
    match net_res {
        Result.Ok(msg) => println(f"   net.tcp_connect('{cfg.address()}') -> {msg}"),
        Result.Err(e) => println(f"   net.tcp_connect error: {e}"),
    }

    println("\n==================================================================")
    println("   ✓ Necto v1.0.0 Golden Release Verification Completed!         ")
    println("==================================================================")
}
