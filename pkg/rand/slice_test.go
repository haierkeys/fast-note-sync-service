package rand

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomStrSliceOne(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}
	
	// Ensure the returned string is one of the slice elements
	// 确保返回的字符串是切片中的元素之一
	for i := 0; i < 10; i++ {
		res := RandomStrSliceOne(slice)
		assert.Contains(t, slice, res)
	}
}

func TestGetRandString(t *testing.T) {
	// Test negative/zero length
	assert.Equal(t, "", GetRandString(0))
	assert.Equal(t, "", GetRandString(-5))

	// Test correct length
	assert.Len(t, GetRandString(10), 10)
	assert.Len(t, GetRandString(50), 50)

	// Since it's random, running it multiple times shouldn't result in exactly the same string easily
	// (Though possible, highly improbable for len 20)
	str1 := GetRandString(20)
	str2 := GetRandString(20)
	assert.NotEqual(t, str1, str2)
}
