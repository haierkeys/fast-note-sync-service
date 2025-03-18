package util

func XorEncodeStr(msg []byte, key []byte) (out []byte) {
	ml := len(msg)
	kl := len(key)
	for i := 0; i < ml; i++ {
		out = append(out, (msg[i])^(key[i%kl]))
	}
	return out
}

func XorEncodeStrRune(msg []rune, key []rune) (out []rune) {
	ml := len(msg)
	kl := len(key)
	for i := 0; i < ml; i++ {
		out = append(out, (msg[i])^(key[i%kl]))
	}
	return out
}
