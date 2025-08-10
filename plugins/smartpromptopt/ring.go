package smartpromptopt

// RingBuffer is a ring buffer.
// It is used to store a fixed number of values.
type RingBuffer[T any] struct {
	buf  []T
	idx  int
	full bool
}

// NewRingBuffer creates a new RingBuffer.
// It returns the RingBuffer.
func NewRingBuffer[T any](size int) *RingBuffer[T] {
	if size <= 0 {
		size = 16
	}
	return &RingBuffer[T]{buf: make([]T, size)}
}

// Add adds a value to the RingBuffer.
// It returns the RingBuffer.
func (r *RingBuffer[T]) Add(v T) {
	if len(r.buf) == 0 {
		return
	}
	r.buf[r.idx] = v
	r.idx = (r.idx + 1) % len(r.buf)
	if r.idx == 0 {
		r.full = true
	}
}

// Snapshot returns a snapshot of the RingBuffer.
// It returns the snapshot.
func (r *RingBuffer[T]) Snapshot() []T {
	if len(r.buf) == 0 {
		return nil
	}
	if !r.full {
		cp := make([]T, r.idx)
		copy(cp, r.buf[:r.idx])
		return cp
	}
	cp := make([]T, len(r.buf))
	copy(cp, r.buf[r.idx:])
	copy(cp[len(r.buf)-r.idx:], r.buf[:r.idx])
	return cp
}
