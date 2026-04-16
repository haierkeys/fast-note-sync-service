//go:build !amd64 && !arm64

package json

import (
	"encoding/json"
	"io"
)

var (
	// Marshal wraps json.Marshal
	// Marshal 包装了 json.Marshal
	Marshal = json.Marshal
	// Unmarshal wraps json.Unmarshal
	// Unmarshal 包装了 json.Unmarshal
	Unmarshal = json.Unmarshal
	// ConfigDefault is the default config using standard library
	// ConfigDefault 是使用标准库的默认配置
	ConfigDefault = stdConfig{}
)

type stdConfig struct{}

func (s stdConfig) NewDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

func (s stdConfig) NewEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}
