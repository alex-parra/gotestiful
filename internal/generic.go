package internal

import "fmt"

// sf is short-hand for string format and calls fmt.Sprintf if arguments are passed
func sf(str string, args ...any) string {
	if len(args) == 0 {
		return str
	}
	return fmt.Sprintf(str, args...)
}

// isZeroValue checks if 'val' is zero-value of it's type
func isZeroValue[T comparable](val T) bool {
	return val == *new(T)
}

// zvfb (zero-value fallback) returns the 'val' if it's not zero-value, else returns 'fallback'
func zvfb[T comparable](val T, fallback T) T {
	if isZeroValue(val) {
		return fallback
	}
	return val
}

// ifelse returns 'trueVal' if 'cond' is true, else returns 'falseVal'
func ifelse[T any](cond bool, trueVal T, falseVal T) T {
	if cond {
		return trueVal
	}
	return falseVal
}
