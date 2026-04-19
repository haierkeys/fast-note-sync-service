package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool2Int(t *testing.T) {
	assert.Equal(t, int64(1), Bool2Int(true))
	assert.Equal(t, int64(0), Bool2Int(false))
}

func TestStrTo_Conversions(t *testing.T) {
	s := StrTo("123")
	assert.Equal(t, "123", s.String())

	val, err := s.Int()
	assert.NoError(t, err)
	assert.Equal(t, 123, val)
	assert.Equal(t, 123, s.MustInt())

	uVal, err := s.UInt32()
	assert.NoError(t, err)
	assert.Equal(t, uint32(123), uVal)
	assert.Equal(t, uint32(123), s.MustUInt32())

	i64Val, err := s.Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(123), i64Val)
	assert.Equal(t, int64(123), s.MustInt64())
}

func TestStrTo_ToSize(t *testing.T) {
	tests := []struct {
		input       string
		expected    int64
		expectError bool
	}{
		{"", 0, false},
		{"1024", 1024, false},
		{"1 KB", 1024, false},
		{"2MB", 2 * 1024 * 1024, false},
		{"10 B", 10, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			s := StrTo(tt.input)
			val, err := s.ToSize()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, val)
			}
			// Test MustToSize behavior
			if tt.expectError {
				assert.Equal(t, int64(999), s.MustToSize(999))
			} else {
				var mustVal int64
				if tt.expected > 0 {
					mustVal = tt.expected
				} else {
					mustVal = 999
				}
				assert.Equal(t, mustVal, s.MustToSize(999))
			}
		})
	}
}

type TestSrc struct {
	ID   int
	Name string
}

type TestDst struct {
	ID   int
	Name string
	Desc string
}

func TestStructAssign(t *testing.T) {
	src := TestSrc{ID: 1, Name: "test"}
	dst := TestDst{Desc: "desc"}

	StructAssign(&src, &dst)

	assert.Equal(t, 1, dst.ID)
	assert.Equal(t, "test", dst.Name)
	assert.Equal(t, "desc", dst.Desc)
}

func TestStructAssign2(t *testing.T) {
	src := TestSrc{ID: 1, Name: "test"}
	dst := &TestDst{Desc: "desc"}

	StructAssign2(&src, dst)

	assert.Equal(t, 1, dst.ID)
	assert.Equal(t, "test", dst.Name)
	assert.Equal(t, "desc", dst.Desc)
}

func TestStructToMap(t *testing.T) {
	src := TestSrc{ID: 1, Name: "test"}
	data := make(map[string]interface{})
	err := StructToMap(src, data)

	assert.NoError(t, err)
	assert.Equal(t, float64(1), data["ID"]) // JSON unmarshals numbers to float64 by default
	assert.Equal(t, "test", data["Name"])
}

func TestStructToMapByReflect(t *testing.T) {
	src := TestSrc{ID: 1, Name: "test"}
	res := StructToMapByReflect(src)

	assert.NotNil(t, res)
	assert.Equal(t, 1, res["ID"])
	assert.Equal(t, "test", res["Name"])
}

func TestCamelCaseConversions(t *testing.T) {
	assert.Equal(t, "hello_world", Camel2Case("HelloWorld"))
	assert.Equal(t, "HelloWorld", Case2Camel("hello_world"))
	assert.Equal(t, "helloWorld", Case2LowerCamel("hello_world"))
	assert.Equal(t, "Hello", Ucfirst("hello"))
	assert.Equal(t, "hELLO", Lcfirst("HELLO"))
}

func TestMapAnyToMapStr(t *testing.T) {
	input := map[string]interface{}{
		"id":   123,
		"name": "test",
		"flag": true,
	}

	result := MapAnyToMapStr(input)
	assert.Equal(t, "123", result["id"])
	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "true", result["flag"])
}
