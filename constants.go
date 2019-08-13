package navvy

import (
	"errors"
)

const (
	// CLOSED represents that the pool is closed.
	CLOSED = 1
)

var (
	// ErrInvalidPoolSize will be returned when setting a negative number as pool capacity.
	ErrInvalidPoolSize = errors.New("invalid size for pool")

	// ErrPoolClosed will be returned when submitting task to a closed pool.
	ErrPoolClosed = errors.New("this pool has been closed")
)
