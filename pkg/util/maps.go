package util

// ShallowCloneMap makes a shallow clone of a given map.
func ShallowCloneMap[K comparable, V any](orig map[K]V) map[K]V {
	new := make(map[K]V, len(orig))
	for key, value := range orig {
		new[key] = value
	}

	return new
}
