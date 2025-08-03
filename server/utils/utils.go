package utils

import (
	"sort"
	"strings"
)

func IsImageDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:")
}

func RemoveDuplicateKeys(keys []string) []string {
	sort.Strings(keys)

	uniqueKeys := make([]string, 0)
	for i, key := range keys {
		if i > 0 && keys[i] != keys[i-1] {
			uniqueKeys = append(uniqueKeys, key)
		}
	}
	return uniqueKeys
}
