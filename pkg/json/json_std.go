// PATCH(#448): 标准库 encoding/json 保证非法 UTF-8 被替换为 U+FFFD，
// 从而 WebSocket text frame 始终为合法 UTF-8。原先本文件仅在非 amd64/arm64 平台生效，
// Apple Silicon 走 json_sonic.go（Sonic）会保留非法字节导致 gws 拒绝 text frame 并断连。
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
