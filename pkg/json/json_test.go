package json

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON_MarshalUnmarshal(t *testing.T) {
	data := map[string]interface{}{
		"key": "value",
		"num": float64(123),
	}

	// Test Marshal
	bytesData, err := Marshal(data)
	assert.NoError(t, err)
	assert.NotEmpty(t, bytesData)

	// Test Unmarshal
	var unmarshaled map[string]interface{}
	err = Unmarshal(bytesData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, data, unmarshaled)
}

func TestJSON_ConfigDefault(t *testing.T) {
	data := map[string]interface{}{
		"hello": "world",
	}

	// Test Encoder
	var buf bytes.Buffer
	encoder := ConfigDefault.NewEncoder(&buf)
	err := encoder.Encode(data)
	assert.NoError(t, err)

	encodedStr := buf.String()
	assert.Contains(t, encodedStr, `"hello"`)
	assert.Contains(t, encodedStr, `"world"`)

	// Test Decoder
	var output map[string]interface{}
	decoder := ConfigDefault.NewDecoder(&buf)
	err = decoder.Decode(&output)
	assert.NoError(t, err)
	assert.Equal(t, "world", output["hello"])
}
