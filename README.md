# Benchmarking Interpreters

What is the performance difference between an AST interpreter and a
bytecode VM? Build and run both to see!

Example:

```bash
$ go build -o
$ cat tests/fib.lisp
(def fib (n)
     (if (<= n 2)
         1
       (+ (fib (- n 1)) (fib (- n 2)))))

(def main ()
     (fib 5))
$ ./main tests/fib.lisp --mode ast
Result: 5, Time: 110.489µs
$ ./main tests/fib.lisp --mode vm
Result: 5, Time: 38.419µs
$ cat tests/fib.lisp
(def fib (n)
     (if (<= n 2)
         1
       (+ (fib (- n 1)) (fib (- n 2)))))

(def main ()
     (fib 30))
$ ./main tests/fib.lisp --mode ast
Result: 832040, Time: 1.852528s
$ ./main tests/fib.lisp --mode vm
Result: 832040, Time: 101.971038ms
```
