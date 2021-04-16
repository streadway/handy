package report

// Filter passes events through selectively.
type Filter struct {
	selector func(e Event) bool
	next     Reporter
}

// Initialize a new Filter that reports events to the next Reporter only if the
// selector function returns true.
func NewFilter(selector func(e Event) bool, next Reporter) Filter {
	return Filter{
		selector: selector,
		next:     next,
	}
}

// Report implements the Reporter interface.
func (f Filter) Report(e Event) {
	if f.selector(e) {
		f.next.Report(e)
	}
}
