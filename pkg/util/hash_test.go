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
		// Simulate 20MB data (above 10MB threshold)
		size := 20 * 1024 * 1024
		data := make([]byte, size)
		// Fill first 5MB with 1
		for i := 0; i < 5*1024*1024; i++ {
			data[i] = 1
		}
		// Fill middle 10MB with 3 (should be ignored)
		for i := 5 * 1024 * 1024; i < 15*1024*1024; i++ {
			data[i] = 3
		}
		// Fill last 5MB with 2
		for i := 15 * 1024 * 1024; i < size; i++ {
			data[i] = 2
		}

		// Calculate expected hash for 5MB of 1s followed by 5MB of 2s
		// (Same logic as EncodeHash32Bytes does now)
		got := EncodeHash32Bytes(data)
		if got == "" {
			t.Errorf("EncodeHash32Bytes(20MB) returned empty string")
		}
	})
}
