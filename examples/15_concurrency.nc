// 15_concurrency.nc
// ==============================================================================
//         Necto v1.1.0 — Concurrency: spawn, Channel[T], and sleep
// ==============================================================================
// Demonstrates lightweight concurrent tasks and typed channels for safe
// message passing between spawned goroutines.

/// Worker function that computes a square and sends it to the channel
fn compute_and_send(ch: Channel, value: int) {
    let result = value * value
    ch.send(result)
}

/// Sends a greeting message into a channel
fn greet_task(ch: Channel) {
    ch.send(42)
}

/// Producer: sends N sequential values into a channel, then closes it
fn produce_values(ch: Channel, count: int) {
    let mut i = 0
    while i < count {
        ch.send(i * 10)
        i = i + 1
    }
    ch.close()
}

fn main() {
    println("==================================================================")
    println("    NECTO v1.1.0 — Concurrency: spawn, Channel[T], and sleep    ")
    println("==================================================================")

    // 1. Basic spawn + channel handshake
    println("\n1. Basic spawn + Channel handshake:")
    let ch1 = Channel.new(1)
    spawn(greet_task, ch1)
    sleep(50)
    let msg = ch1.recv()
    match msg {
        Some(val) => {
            println(f"   [main] Received from spawned task: {val}")
        },
        None => {
            println("   [main] Channel was closed")
        },
    }

    // 2. Fan-out: multiple workers sending results
    println("\n2. Fan-out: 5 workers computing squares:")
    let ch2 = Channel.new(10)

    spawn(compute_and_send, ch2, 1)
    spawn(compute_and_send, ch2, 2)
    spawn(compute_and_send, ch2, 3)
    spawn(compute_and_send, ch2, 4)
    spawn(compute_and_send, ch2, 5)

    sleep(100)

    let mut received = 0
    while received < 5 {
        let r = ch2.recv()
        match r {
            Some(val) => {
                println(f"   Worker result: {val}")
                received = received + 1
            },
            None => {
                received = 5
            },
        }
    }

    // 3. Producer-consumer with channel close
    println("\n3. Producer-Consumer with close:")
    let pipe = Channel.new(20)
    spawn(produce_values, pipe, 5)
    sleep(100)

    let mut reading = true
    while reading {
        let item = pipe.recv()
        match item {
            Some(val) => {
                println(f"   Consumer received: {val}")
            },
            None => {
                println("   Consumer: channel closed, done.")
                reading = false
            },
        }
    }

    println("\n==================================================================")
    println("   ✓ Concurrency demo completed successfully!")
    println("==================================================================")
}

// ------------------------------------------------------------------
// Unit tests for concurrency primitives
// ------------------------------------------------------------------

test "Channel.new creates a buffered channel" {
    let ch = Channel.new(5)
    ch.send(42)
    let val = ch.recv()
    match val {
        Some(v) => {
            assert(v == 42)
        },
        None => {
            assert(false)
        },
    }
}

test "Channel buffered send and recv ordering" {
    let ch = Channel.new(3)
    ch.send(10)
    ch.send(20)
    ch.send(30)

    let v1 = ch.recv()
    let v2 = ch.recv()
    let v3 = ch.recv()

    match v1 {
        Some(val) => {
            assert(val == 10)
        },
        None => {
            assert(false)
        },
    }
    match v2 {
        Some(val) => {
            assert(val == 20)
        },
        None => {
            assert(false)
        },
    }
    match v3 {
        Some(val) => {
            assert(val == 30)
        },
        None => {
            assert(false)
        },
    }
}

test "Channel close returns None on recv" {
    let ch = Channel.new(1)
    ch.send(99)
    ch.close()
    let v1 = ch.recv()
    match v1 {
        Some(val) => {
            assert(val == 99)
        },
        None => {
            assert(false)
        },
    }
    let v2 = ch.recv()
    match v2 {
        Some(_) => {
            assert(false)
        },
        None => {
            assert(true)
        },
    }
}

test "spawn runs function concurrently via channel" {
    let ch = Channel.new(1)
    spawn(greet_task, ch)
    sleep(50)
    let val = ch.recv()
    match val {
        Some(v) => {
            assert(v == 42)
        },
        None => {
            assert(false)
        },
    }
}

// ------------------------------------------------------------------
// Benchmarks
// ------------------------------------------------------------------

bench "channel send-recv round trip" {
    let ch = Channel.new(1)
    ch.send(42)
    ch.recv()
}
