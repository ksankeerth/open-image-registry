package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const NameRegex = "^[a-zA-Z0-9_-]+$"

func IsImageDigest(value string) bool {
	return strings.HasPrefix(value, "sha256:")
}

func RemoveDuplicateKeys(keys []string) []string {
	if len(keys) == 0 {
		return keys
	}
	sort.Strings(keys)

	uniqueKeys := []string{keys[0]}
	for i, key := range keys {
		if i > 0 && keys[i] != keys[i-1] {
			uniqueKeys = append(uniqueKeys, key)
		}
	}
	return uniqueKeys
}

func CalcuateDigest(content []byte) string {
	hash := sha256.New()
	hash.Write(content)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(hash.Sum(nil)))
}

func isValidName(name string) bool {
	matched, _ := regexp.MatchString(NameRegex, name)
	return matched
}

func IsValidNamespace(namespace string) bool {
	return isValidName(namespace)
}

func IsValidRegistry(registry string) bool {
	return isValidName(registry)
}

func IsValidRepository(repository string) bool {
	return isValidName(repository)
}

func ParseImageBlobContentRangeFromRequest(headerValue string) (start, end int64, err error) {
	if headerValue == "" {
		return 0, 0, nil
	}
	re := regexp.MustCompile(`^(\d+)-(\d+)$`)
	matches := re.FindStringSubmatch(headerValue)
	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("invalid content range: %s", headerValue)
	}

	start, err = strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start: %w", err)
	}

	end, err = strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end: %w", err)
	}

	return start, end, nil
}

func StorageLocation(args ...string) string {
	return filepath.Clean(filepath.Join(args...))
}

func CombineAndCalculateSHA256Digest(inputs ...string) string {
	hash := sha256.New()
	hash.Write([]byte(strings.Join(inputs, ":")))
	return fmt.Sprintf("sha256:%x", hash.Sum(nil))
}
