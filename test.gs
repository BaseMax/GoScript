print("hey")

fn f(n) {
    if n <= 1 { 1 }
    n * f(n-1)
}

println(f(5))
