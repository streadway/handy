package retry

import (
	"math"
	"time"
)

// Constant sleeps for delta duration
func Constant(delta time.Duration) Delayer {
	return func(a Attempt) {
		time.Sleep(delta)
	}
}

// Linear sleeps for delta * the number of attempts
func Linear(delta time.Duration) Delayer {
	return func(a Attempt) {
		time.Sleep(delta * time.Duration(a.Count))
	}
}

// Linear sleeps for delta * 2^attempts
func Exponential(base time.Duration) Delayer {
	return func(a Attempt) {
		time.Sleep(time.Duration(float64(base) * math.Exp(float64(a.Count))))
	}
}

// Fibonacci sleeps for delta * fib(attempts)
func Fibonacci(delta time.Duration) Delayer {
	fib := func(i int) int {
		var n0, n1, j int
		for n0, n1, j = 0, 1, 0; j < i; j++ {
			n0, n1 = n1, n0+n1
		}
		return n0
	}
	return func(a Attempt) {
		time.Sleep(delta * time.Duration(fib(int(a.Count))))
	}
}