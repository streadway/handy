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

// Circuit implements a circuit breaker state machine.
type Circuit struct {
	force   chan states
	allow   chan bool
	success chan time.Duration
	failure chan time.Duration

	config circuitConfig
}

type circuitConfig struct {
	FailureRatio float64 // normalized between 0.0 and 1.0

	Window          time.Duration // number of second buckets to observe
	Cooldown        time.Duration // time to wait before trying once when open
	MinObservations uint          // observations required to open when failing

	Now   func() time.Time                     // default time.Now
	After func(time.Duration) <-chan time.Time // default time.After
}

func newCircuit(config circuitConfig) Circuit {
	if config.FailureRatio < 0.0 {
		config.FailureRatio = 0.0
	}

	if config.FailureRatio > 1.0 {
		config.FailureRatio = 1.0
	}

	if config.Window == 0 {
		config.Window = DefaultWindow
	}

	if config.Cooldown == 0 {
		config.Cooldown = DefaultCooldown
	}

	if config.Now == nil {
		config.Now = time.Now
	}

	if config.After == nil {
		config.After = time.After
	}

	circuit := Circuit{
		force:   make(chan states),
		allow:   make(chan bool),
		success: make(chan time.Duration),
		failure: make(chan time.Duration),
		config:  config,
	}

	go circuit.run()

	return circuit
}

// NewCircuit constructs a new circuit breaker, initially closed.
// Circuit opens after failureRatio failures per success.
func NewCircuit(failureRatio float64) Circuit {
	return newCircuit(circuitConfig{
		MinObservations: DefaultMinObservations,
		FailureRatio:    failureRatio,
	})
}

func (c Circuit) shouldOpen(m *metric) bool {
	s := m.Summary()
	return s.total > c.config.MinObservations && s.rate > c.config.FailureRatio
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
func (c Circuit) run() {
	var (
		state   states
		timeout <-chan time.Time
		metrics *metric
	)

	for {
		//println(state, len(timeout), metrics)
		switch state {
		case reset:
			metrics = newMetric(c.config.Window, c.config.Now)
			timeout = nil
			state = closed

		case closed:
			select {
			case c.allow <- true:
			case d := <-c.success:
				metrics.Success(d)
			case d := <-c.failure:
				metrics.Failure(d)
				if c.shouldOpen(metrics) {
					state = tripped
				}
			case state = <-c.force:
			}

		case tripped:
			timeout = c.config.After(c.config.Cooldown)
			state = open

		case open:
			select {
			case c.allow <- false:
			case <-c.success:
				state = reset
			case <-c.failure:
			case <-timeout:
				state = halfopen
			case state = <-c.force:
			}

		case halfopen:
			select {
			case c.allow <- true:
				state = tripped
			case <-c.success:
				state = reset
			case <-c.failure:
				state = open
			case state = <-c.force:
			}

		}
	}
}

// Allow returns true if a new request should be allowed to proceed to the
// underlying resource.
func (c Circuit) Allow() bool {
	return <-c.allow
}

// Trip manually opens the circuit.
func (c Circuit) Trip() {
	c.force <- tripped
}

// Reset manually closes the circuit.
func (c Circuit) Reset() {
	c.force <- reset
}

// Success informs the circuit that a request to the underlying resource has
// completed successfully. Every Allowed request should signal either Success
// or Failure.
func (c Circuit) Success(d time.Duration) {
	c.success <- d
}

// Failure informs the circuit that a request to the underlying resource has
// failed. Every Allowed request should signal either Success or Failure.
func (c Circuit) Failure(d time.Duration) {
	c.failure <- d
}
