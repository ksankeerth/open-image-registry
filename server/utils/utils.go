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
	"time"
	"unicode"
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

func ParseSqliteTimestamp(timeStr string) (*time.Time, error) {
	// Strip monotonic clock part if present (e.g. "2025-09-05 11:41:39.976848 +0530 +0530 m=+339.588692899")
	timeStr = strings.SplitN(timeStr, " m=", 2)[0]

	// Try parsing with timezone offset and repeated offset
	t, err := time.Parse("2006-01-02 15:04:05.999999 -0700 -0700", timeStr)
	if err == nil {
		return &t, nil
	}

	// Fallback: SQLite may store as ISO8601 (e.g. "2025-09-05T11:41:39Z")
	t, err = time.Parse(time.RFC3339Nano, timeStr)
	if err == nil {
		return &t, nil
	}

	// Fallback: SQLite default (no TZ, no T)
	t, err = time.Parse("2006-01-02 15:04:05", timeStr)
	if err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("cannot parse SQLite timestamp: %s", timeStr)
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,32}$`)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsValidUsername(username string) bool {
	return usernameRegex.Match([]byte(username))
}

func IsValidEmail(email string) bool {
	return emailRegex.Match([]byte(email))
}

func ValidatePassword(pw string) (bool, string) {
	// Length check
	if len(pw) < 12 {
		return false, "Password must be at least 12 characters long"
	}
	if len(pw) > 64 {
		return false, "Password cannot exceed 64 characters"
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool

	for _, ch := range pw {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case isAllowedSymbol(ch):
			hasSymbol = true
		}
	}

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}
	if !hasDigit {
		return false, "Password must contain at least one number"
	}
	if !hasSymbol {
		return false, "Password must contain at least one symbol (!@#$%^&*)"
	}

	return true, ""
}

func isAllowedSymbol(ch rune) bool {
	symbols := "!@#$%^&*"
	for _, s := range symbols {
		if ch == s {
			return true
		}
	}
	return false
}