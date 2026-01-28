package util

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	tokenLimit = 10
	tokenStart = 8
	tokenEnd   = 18
)

func AuthCodeEncrypt(token string, action string, key string) (out string, err error) {
	var (
		strauth   string
		tokenByte []byte
		keyByte   []byte
	)
	if len(token) == 0 {
		return out, errors.New("token is not allowed to be empty")
	}
	if len(action) == 0 {
		action = "EN"
	}
	if action == "DE" {
		token = strings.Replace(token, "[a]", "+", -1)
		token = strings.Replace(token, "[b]", "&", -1)
		token = strings.Replace(token, "[c]", "/", -1)
	}

	tokenLen := len(token)
	if tokenLen <= tokenLimit {
		return out, errors.New("The token length does not meet the requirements")
	}
	if action == "EN" {
		strauth = EncodeMD5(token)[tokenStart:tokenEnd]
	} else {
		strauth = token[tokenLen-tokenLimit : tokenLen]
		tokenByte, _ = base64.StdEncoding.DecodeString(token[0 : tokenLen-tokenLimit])
		token = string(tokenByte)
	}

	key = EncodeMD5(strauth + key)

	tokenByte = []byte(token)
	keyByte = []byte(key)
	tmpCode := XorEncodeStr(tokenByte, keyByte)
	code := string(tmpCode)
	if action == "DE" {
		if EncodeMD5(code)[tokenStart:tokenEnd] == strauth {
			out = code
		}
	} else {
		out = base64.StdEncoding.EncodeToString([]byte(code + strauth))
		out = strings.Replace(out, "[a]", "+", -1)
		out = strings.Replace(out, "[b]", "&", -1)
		out = strings.Replace(out, "[c]", "/", -1)
	}
	return out, nil
}

/**
 * str: plaintext or ciphertext // 明文或密文
 * operation: encryption ENCODE or decryption DECODE // 加密ENCODE或解密DECODE
 * key: secret key // 密钥
 * expiry: secret key validity period // 密钥有效期
 */
func AuthDzCodeEncrypt(str, operation, key string, expiry int64) (string, error) {
	// Dynamic secret key length, the same plaintext will generate different ciphertext depending on the dynamic key
	// 动态密匙长度，相同的明文会生成不同密文就是依靠动态密匙
	// Adding a random key can make the ciphertext have no pattern, even if the original text and key are exactly the same, the encryption result will be different every time, increasing the difficulty of cracking.
	// 加入随机密钥，可以令密文无任何规律，即便是原文和密钥完全相同，加密结果也会每次不同，增大破解难度。
	// The larger the value, the greater the change pattern of the ciphertext, ciphertext change = 16 to the power of ckeyLength
	// 取值越大，密文变动规律越大，密文变化 = 16 的 ckeyLength 次方
	// When this value is 0, no random key is generated
	// 当此值为 0 时，则不产生随机密钥
	ckeyLength := 4

	// Secret key
	// 密匙
	if key == "" {
		key = "STARFISSION_AUTH_KEY"
	}

	key = EncodeMD5(key)

	// Key a will participate in encryption and decryption
	// 密匙a会参与加解密
	keya := EncodeMD5(key[:16])
	// Key b will be used for data integrity verification
	// 密匙b会用来做数据完整性验证
	keyb := EncodeMD5(key[16:])
	// Key c is used to change the generated ciphertext
	// 密匙c用于变化生成的密文
	keyc := ""
	if ckeyLength != 0 {
		if operation == "DECODE" {
			keyc = str[:ckeyLength]
		} else {
			sTime := EncodeMD5(time.Now().String())
			sLen := 32 - ckeyLength
			keyc = sTime[sLen:]
		}
	}

	// Key involved in the operation
	// 参与运算的密匙
	cryptKey := fmt.Sprintf("%s%s", keya, EncodeMD5(keya+keyc))
	keyLength := len(cryptKey)

	// Plaintext, the first 10 bits are used to save the timestamp, the validity of the data is verified during decryption, and 10 to 26 bits are used to save $keyb (key b), and the data integrity will be verified through this key during decryption
	// 明文，前10位用来保存时间戳，解密时验证数据有效性，10到26位用来保存$keyb(密匙b)，解密时会通过这个密匙验证数据完整性
	// If it is decoding, it will start from the $ckeyLength bit, because the first $ckeyLength bits of the ciphertext save the dynamic key to ensure correct decryption
	// 如果是解码的话，会从第$ckeyLength位开始，因为密文前$ckeyLength位保存 动态密匙，以保证解密正确
	if operation == "DECODE" {
		str = strings.Replace(str, "[a]", "+", -1)
		str = strings.Replace(str, "[b]", "&", -1)
		str = strings.Replace(str, "[c]", "/", -1)

		strByte, err := base64.StdEncoding.DecodeString(str[ckeyLength:])
		if err != nil {
			return "", err
		}
		str = string(strByte)
	} else {

		if expiry != 0 {
			expiry = expiry + time.Now().Unix()
		}
		tmpMd5 := EncodeMD5(str + keyb)
		str = fmt.Sprintf("%010d%s%s", expiry, tmpMd5[:16], str)
	}

	stringLength := len(str)
	resdata := make([]byte, 0, stringLength)
	var rndkey, box [256]int
	// Generate secret key book
	// 产生密匙簿
	j := 0
	a := 0
	i := 0
	tmp := 0
	for i = 0; i < 256; i++ {
		rndkey[i] = int(cryptKey[i%keyLength])
		box[i] = i
	}
	// Use a fixed algorithm to mess up the key book to increase randomness.
	// 用固定的算法，打乱密匙簿，增加随机性
	for i = 0; i < 256; i++ {
		j = (j + box[i] + rndkey[i]) % 256
		tmp = box[i]
		box[i] = box[j]
		box[j] = tmp
	}
	// Core encryption and decryption part
	// 核心加解密部分
	a = 0
	j = 0
	tmp = 0
	for i = 0; i < stringLength; i++ {
		a = (a + 1) % 256
		j = (j + box[a]) % 256
		tmp = box[a]
		box[a] = box[j]
		box[j] = tmp
		// Get the secret key from the key book for XOR, and then convert it to a character
		// 从密匙簿得出密匙进行异或，再转成字符
		resdata = append(resdata, byte(int(str[i])^box[(box[a]+box[j])%256]))
	}
	result := string(resdata)

	if operation == "DECODE" {
		// Verify data validity and integrity
		// 验证数据有效性、完整性
		frontTen, _ := strconv.ParseInt(result[:10], 10, 0)
		if (frontTen == 0 || frontTen-time.Now().Unix() > 0) && result[10:26] == EncodeMD5(result[26:] + keyb)[:16] {
			return result[26:], nil
		} else {
			return "", errors.New("AuthCode Encrypt error")
		}
	} else {
		// Save the dynamic key in the ciphertext, which is why the same plaintext can be decrypted after producing different ciphertexts
		// 把动态密匙保存在密文里，这也是为什么同样的明文，生产不同密文后能解密的原因
		// Because the encrypted ciphertext may be some special characters, which may be lost during the copying process, use base64 encoding
		// 因为加密后的密文可能是一些特殊字符，复制过程可能会丢失，所以用base64编码
		result = keyc + base64.StdEncoding.EncodeToString([]byte(result))

		result = strings.Replace(result, "+", "[a]", -1)
		result = strings.Replace(result, "&", "[b]", -1)
		result = strings.Replace(result, "/", "[c]", -1)

		return result, nil
	}
}

// base64Encode base64 encodes a string
// base64Encode 对字符串进行Base64编码
// s: the string to be encoded // 待编码的字符串
// return: encoded string // 返回值: Base64编码后的字符串
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// base64Decode decodes a base64 encoded string
// base64Decode 对Base64编码的字符串进行解码
// s: the base64 string to be decoded // 待解码的Base64字符串
// return: original string // 返回值: 解码后的原始字符串
func base64Decode(s string) string {
	sByte, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return string(sByte)
	} else {
		return ""
	}
}

// XorEncodeStr encrypts a byte slice using XOR operation
// XorEncodeStr 使用异或操作对字节切片进行加密
// msg: byte slice to be encrypted // 要加密的字节切片
// key: key byte slice // 加密密钥的字节切片
// return: encrypted byte slice // 返回值: 加密后的字节切片
func XorEncodeStr(msg []byte, key []byte) (out []byte) {
	ml := len(msg)
	kl := len(key)
	for i := 0; i < ml; i++ {
		out = append(out, (msg[i])^(key[i%kl]))
	}
	return out
}

// XorEncodeStrRune encrypts a rune slice using XOR operation
// XorEncodeStrRune 使用异或操作对rune切片进行加密
// msg: rune slice to be encrypted // 要加密的rune切片
// key: key rune slice // 加密密钥的rune切片
// return: encrypted rune slice // 返回值: 加密后的rune切片
func XorEncodeStrRune(msg []rune, key []rune) (out []rune) {
	ml := len(msg)
	kl := len(key)
	for i := 0; i < ml; i++ {
		out = append(out, (msg[i])^(key[i%kl]))
	}
	return out
}
