package kit

import (
	"fmt"
)

func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}
	return values
}

func WrapError(err error, format string, a ...any) error {
	errorMsg := fmt.Sprintf(format, a...)
	return fmt.Errorf(errorMsg+": %w", err)
}
