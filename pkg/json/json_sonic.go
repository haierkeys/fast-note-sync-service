// PATCH(#448): 默认不再使用 Sonic —— 它会原样保留非法 UTF-8 字节，
// 导致 gws: invalid text encoding 与 WebSocket 循环重连（见 issue #448）。
// 如确需 Sonic 的性能，请改为 json_std.go 的 build tag 方案并自行验证 UTF-8 合法性。
//go:build sonic

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
