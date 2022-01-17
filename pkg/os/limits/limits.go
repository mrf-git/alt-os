package limits

// OsLimit represents a limit enforced by the operating system
// and is an alias for an integer.
type OsLimit = int

// Represents an unlimited value.
const UNLIMITED = -1

// OS limits.
const (
	_ OsLimit = 0
	// Maximum size of executable code in memory.
	MAX_EXECUTABLE_SIZE = 0xFFFFF000
)
