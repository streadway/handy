package breaker

import (
	"time"
)

const (
	// DefaultWindow is the default number of per-second buckets that will be
	// considered when calculating metrics on the circuit breaker.
	DefaultWindow = 5 * time.Second

	// DefaultCooldown is the default period a circuit will remain in the open
	// state before allowing a single sentinel request through.
	DefaultCooldown = 1 * time.Second

	// DefaultMinObservations is the default number of observations that must
	// be made before the circuit breaker
	DefaultMinObservations = 10
)

type states int

const (
	reset states = iota
	tripped

	closed
	open
	halfopen
)

// CircuitBreaker implements a circuit breaker state machine.
type CircuitBreaker struct {
	force   chan states
	allow   chan bool
	success chan time.Duration
	failure chan time.Duration

	config circuitBreakerConfig
}

type circuitBreakerConfig struct {
	FailureRatio float64 // normalized between 0.0 and 1.0

	Window          time.Duration // number of second buckets to observe
	Cooldown        time.Duration // time to wait before trying once when open
	MinObservations uint          // observations required to open when failing

	Now   func() time.Time                     // default time.Now
	After func(time.Duration) <-chan time.Time // default time.After
}

func newCircuitBreaker(c circuitBreakerConfig) CircuitBreaker {
	if c.FailureRatio < 0.0 {
		c.FailureRatio = 0.0
	}

	if c.FailureRatio > 1.0 {
		c.FailureRatio = 1.0
	}

	if c.Window == 0 {
		c.Window = DefaultWindow
	}

	if c.Cooldown == 0 {
		c.Cooldown = DefaultCooldown
	}

	if c.Now == nil {
		c.Now = time.Now
	}

	if c.After == nil {
		c.After = time.After
	}

	b := CircuitBreaker{
		force:   make(chan states),
		allow:   make(chan bool),
		success: make(chan time.Duration),
		failure: make(chan time.Duration),
		config:  c,
	}

	go b.run()

	return b
}

// NewCircuitBreaker constructs a new circuit breaker, initially closed.
// CircuitBreaker opens after failureRatio failures per success.
func NewCircuitBreaker(failureRatio float64) CircuitBreaker {
	return newCircuitBreaker(circuitBreakerConfig{
		MinObservations: DefaultMinObservations,
		FailureRatio:    failureRatio,
	})
}

func (b CircuitBreaker) shouldOpen(m *metric) bool {
	s := m.Summary()
	return s.total > b.config.MinObservations && s.rate > b.config.FailureRatio
}

/*
sed -n 's/^dot//p' | dot -Tsvg -ostates.svg
dot digraph {
dot  reset -> closed    [label="stats and time reset"]
dot  closed -> tripped	[label="failed and failure rate exceeded"]
dot  closed -> closed		[label="succeed and update stats"]
dot  closed -> closed		[label="failed and update stats"]
dot
dot  tripped -> open		[label="timeout scheduled"]
dot  open -> reset      [label="succeed"]
dot  open -> halfopen   [label="timeout expired"]
dot  halfopen -> open   [label="failed"]
dot  halfopen -> open   [label="allowed one"]
dot }
*/
func (b CircuitBreaker) run() {
	var (
		state   states
		timeout <-chan time.Time
		metrics *metric
	)

	for {
		//println(state, len(timeout), metrics)
		switch state {
		case reset:
			metrics = newMetric(b.config.Window, b.config.Now)
			timeout = nil
			state = closed

		case closed:
			select {
			case b.allow <- true:
			case d := <-b.success:
				metrics.Success(d)
			case d := <-b.failure:
				metrics.Failure(d)
				if b.shouldOpen(metrics) {
					state = tripped
				}
			case state = <-b.force:
			}

		case tripped:
			timeout = b.config.After(b.config.Cooldown)
			state = open

		case open:
			select {
			case b.allow <- false:
			case <-b.success:
				state = reset
			case <-b.failure:
			case <-timeout:
				state = halfopen
			case state = <-b.force:
			}

		case halfopen:
			select {
			case b.allow <- true:
				state = tripped
			case <-b.success:
				state = reset
			case <-b.failure:
				state = open
			case state = <-b.force:
			}

		}
	}
}

// Allow returns true if a new request should be allowed to proceed to the
// underlying resource.
func (b CircuitBreaker) Allow() bool {
	return <-b.allow
}

// Trip manually opens the circuit.
func (b CircuitBreaker) Trip() {
	b.force <- tripped
}

// Reset manually closes the circuit.
func (b CircuitBreaker) Reset() {
	b.force <- reset
}

// Success informs the circuit that a request to the underlying resource has
// completed successfully. Every Allowed request should signal either Success
// or Failure.
func (b CircuitBreaker) Success(d time.Duration) {
	b.success <- d
}

// Failure informs the circuit that a request to the underlying resource has
// failed. Every Allowed request should signal either Success or Failure.
func (b CircuitBreaker) Failure(d time.Duration) {
	b.failure <- d
}
