package utils

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsImageDigest(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"sha256adecf...", false},
		{"sha256:1d34ffeaf190be23d3de5a8de0a436676b758f48f835c3a2d4768b798c15a7f1", true},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, IsImageDigest(tt.value))
	}
}

func TestRemoveDuplicateKeys(t *testing.T) {
	tests := []struct {
		keys       []string
		uniqueKeys []string
	}{
		{[]string{"a", "a", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{"abc", "xyz", "xyz", "xyz", "XYZ", "lml"}, []string{"XYZ", "abc", "lml", "xyz"}},
		{nil, nil},
		{[]string{"", "", ""}, []string{""}},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.uniqueKeys, RemoveDuplicateKeys(tt.keys))
	}
}

func TestCalcuateDigest(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"Empty content", []byte(""), "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"Simple string", []byte("hello"), "sha256:2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"Another string", []byte("golang"), "sha256:d754ed9f64ac293b10268157f283ee23256fb32a4f8dedb25c8446ca5bcb0bb3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, CalcuateDigest(tt.input))
		})
	}
}

func TestIsValidRegistry(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"myregistry", true},
		{"my-registry", true},
		{"my_registry", true},
		{"reg123", true},
		{"my.registry", false},
		{"my/registry", false},
		{"reg!stry", false},
		{"", false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, IsValidRegistry(tt.input))
	}
}

func TestIsValidRepository(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"repo", true},
		{"repo-1", true},
		{"repo_2", true},
		{"repo.name", false},
		{"org/repo", false},
		{"repo name", false},
		{"", false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, IsValidRepository(tt.input))
	}
}

func TestIsValidNamespace(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"namespace", true},
		{"my-namespace", true},
		{"my_namespace", true},
		{"ns123", true},
		{"ns.name", false},
		{"ns/name", false},
		{"ns!", false},
		{"", false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, IsValidNamespace(tt.input))
	}
}

func TestParseImageBlobContentRangeFromRequest(t *testing.T) {
	tests := []struct {
		header    string
		wantStart int64
		wantEnd   int64
		wantErr   bool
	}{
		{"", 0, 0, false},
		{"0-100", 0, 100, false},
		{"abc-def", 0, 0, true},
		{"100", 0, 0, true},
		{"0-100-200", 0, 0, true},
	}

	for _, tt := range tests {
		start, end, err := ParseImageBlobContentRangeFromRequest(tt.header)
		if tt.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		}
	}
}

func TestStorageLocation(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"tmp"}, filepath.Clean("tmp")},
		{[]string{"var", "lib", "docker"}, filepath.Clean("var/lib/docker")},
		{[]string{"var", "lib", "..", "etc"}, filepath.Clean("var/etc")},
		{[]string{}, "."},
		{[]string{"/", "usr", "local"}, filepath.Clean("/usr/local")},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, StorageLocation(tt.args...))
	}
}

func TestCombineAndCalculateSHA256Digest(t *testing.T) {
	tests := []struct {
		inputs []string
		want   string
	}{
		{[]string{"hello"}, fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("hello")))},
		{[]string{"foo", "bar"}, fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("foo:bar")))},
		{[]string{}, fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("")))},
		{[]string{"foo:bar", "baz"}, fmt.Sprintf("sha256:%x", sha256.Sum256([]byte("foo:bar:baz")))},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, CombineAndCalculateSHA256Digest(tt.inputs...))
	}
}
