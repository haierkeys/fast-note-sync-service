package fileurl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileUrlConversions(t *testing.T) {
	assert.Equal(t, ".png", GetFileExt("image.png"))
	assert.Equal(t, "image.png", GetFileName("image.png"))
	assert.Equal(t, "file.tar.gz", GetFileName("file.tar.gz"))
	
	// Check random name override for "image.png"
	randName := GetFileNameOrRandom("image.png")
	assert.NotEqual(t, "image.png", randName)
	assert.Contains(t, randName, "image.png")

	// Check non-image.png unchanged
	assert.Equal(t, "test.txt", GetFileNameOrRandom("test.txt"))
}

func TestPathSuffixCheckAdd(t *testing.T) {
	assert.Equal(t, "/abc/", PathSuffixCheckAdd("/abc", "/"))
	assert.Equal(t, "/abc/", PathSuffixCheckAdd("/abc/", "/"))
}

func TestUrlEscape(t *testing.T) {
	assert.Equal(t, "path/to/my%20file.txt", UrlEscape("path/to/my file.txt"))
	assert.Equal(t, "my%20file.txt", UrlEscape("my file.txt"))
}

func TestIsAbsPath(t *testing.T) {
	// These only apply strictly dynamically across platform, but we can test typical UNIX style if we assume it runs on Linux/Mac
	assert.True(t, IsAbsPath("/abs/path"))
	assert.False(t, IsAbsPath("rel/path"))
}
