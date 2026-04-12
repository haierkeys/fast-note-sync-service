package util

import (
	"strings"
	"unicode"
)

// Tokenize 对文本进行分词，支持中英文。
// 英文按非字母数字切分，中文按二元分词 (Bigram)。
func Tokenize(text string) []string {
	var tokens []string
	var currentToken strings.Builder

	// 转换为小写
	text = strings.ToLower(text)
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if isCJK(r) {
				// 处理中文/日文/韩文 (Bigram)
				tokens = append(tokens, string(r)) // 单字索引提高召回率
				if i+1 < len(runes) && isCJK(runes[i+1]) {
					tokens = append(tokens, string(runes[i:i+2]))
				}
			} else {
				// 处理普通字母数字
				currentToken.WriteRune(r)
			}
		} else {
			// 遇到分隔符，保存当前的普通单词
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
		}
	}

	// 处理最后一个单词
	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return UniqueStrings(tokens)
}

// isCJK 检查是否是中日韩字符
func isCJK(r rune) bool {
	return unicode.Is(unicode.Scripts["Han"], r) ||
		unicode.Is(unicode.Scripts["Hiragana"], r) ||
		unicode.Is(unicode.Scripts["Katakana"], r) ||
		unicode.Is(unicode.Scripts["Hangul"], r)
}

// UniqueStrings 字符串去重且过滤空字符串
func UniqueStrings(slice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
