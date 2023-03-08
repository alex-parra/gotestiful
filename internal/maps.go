package internal

// mapHasKey returns true if a map has the key provided
func mapHasKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}
