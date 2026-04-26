package util

import (
	"testing"
)

func TestHashConsistency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ASCII", "Hello", "69609650"},
		{"Unicode", "你好", "652829"},
		{"Complex", "Fast Note Sync 🚀", "475362430"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeHash32(tt.input)
			if got != tt.expected {
				t.Errorf("EncodeHash32(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHashBytesConsistency(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"Binary1", []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}, "69609650"}, // "Hello" as bytes
		{"Binary2", []byte{0xff, 0x00, 0xaa, 0x55}, "7602060"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeHash32Bytes(tt.input)
			if got != tt.expected {
				t.Errorf("EncodeHash32Bytes(%v) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}

	t.Run("LargeData", func(t *testing.T) {
		// Simulate 110MB data
		size := 110 * 1024 * 1024
		data := make([]byte, size)
		// Fill first 50MB with 1
		for i := 0; i < 50*1024*1024; i++ {
			data[i] = 1
		}
		// Fill middle 10MB with 3 (should be ignored)
		for i := 50 * 1024 * 1024; i < 60*1024*1024; i++ {
			data[i] = 3
		}
		// Fill last 50MB with 2
		for i := 60 * 1024 * 1024; i < size; i++ {
			data[i] = 2
		}

		expected := "1442840576"
		got := EncodeHash32Bytes(data)
		if got != expected {
			t.Errorf("EncodeHash32Bytes(110MB) = %v, want %v", got, expected)
		}
	})
}
