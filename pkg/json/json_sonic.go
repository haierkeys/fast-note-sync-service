//go:build amd64 || arm64

package json

import (
	"io"

	"github.com/bytedance/sonic"
)

var (
	// Marshal wraps sonic.Marshal
	// Marshal 包装了 sonic.Marshal
	Marshal = sonic.Marshal
	// Unmarshal wraps sonic.Unmarshal
	// Unmarshal 包装了 sonic.Unmarshal
	Unmarshal = sonic.Unmarshal
	// ConfigDefault wraps sonic.ConfigDefault
	// ConfigDefault 包装了 sonic.ConfigDefault
	ConfigDefault = sonicConfig{sonic.ConfigDefault}
)

type sonicConfig struct {
	sonic.API
}

func (s sonicConfig) NewDecoder(r io.Reader) Decoder {
	return s.API.NewDecoder(r)
}

func (s sonicConfig) NewEncoder(w io.Writer) Encoder {
	return s.API.NewEncoder(w)
}
