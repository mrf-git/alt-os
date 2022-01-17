package limits

import "fmt"

// ErrLimitExceeded represents an error that occurs due to
// exceeding a limit value.
type ErrLimitExceeded struct {
	LimitName   string
	LimitValue  OsLimit
	ActualValue int
}

func (err ErrLimitExceeded) Error() string {
	return fmt.Sprintf("limit %s exceeded - limit: %d, actual: %d",
		err.LimitName, err.LimitValue, err.ActualValue)
}
