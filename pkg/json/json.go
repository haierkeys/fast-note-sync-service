package json

import "io"

// Config defines the interface for JSON configuration
// Config 定义了 JSON 配置的接口
type Config interface {
	NewDecoder(io.Reader) Decoder
	NewEncoder(io.Writer) Encoder
}

// Decoder defines the interface for JSON decoder
// Decoder 定义了 JSON 解码器的接口
type Decoder interface {
	Decode(any) error
}

// Encoder defines the interface for JSON encoder
// Encoder 定义了 JSON 编码器的接口
type Encoder interface {
	Encode(any) error
}
