fn fibo(n int) int {
    if n <= 1 {
        return n
    }
    return fibo(n-1)+fibo(n-2)
}
fn main() {
    println(fibo(40))
}
