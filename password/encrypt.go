package password

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"strings"
)

type Encryptor struct {
	// 需要使用的加密算法
	algorithm string
	//密钥
	key string
	//初始化向量
	iv string
	//字符替换表
	replaceTable string
	replaceWith  string
}

var DefaultEncryptor = Encryptor{
	algorithm:    "AES/CBC/PKCS5Padding",
	key:          "4gJnZ9dRkNlWl1Lp",
	iv:           "2sTcY7rMePfBh8Nq",
	replaceTable: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	replaceWith:  "BAZYXWVUTSRQPONMLKJIHGFEDCbazyxwvutsrqponmlkjihgfedc0123456789",
}

// Option is an application option.
type OptionFunc func(o *Encryptor)

func Algorithm(algorithm string) OptionFunc {
	return func(o *Encryptor) { o.algorithm = algorithm }
}
func Key(key string) OptionFunc {
	return func(o *Encryptor) { o.key = key }
}
func Iv(iv string) OptionFunc {
	return func(o *Encryptor) { o.iv = iv }
}

func ReplaceWith(replaceWith string) OptionFunc {
	return func(o *Encryptor) { o.replaceWith = replaceWith }
}
func ReplaceTable(replaceTable string) OptionFunc {
	return func(o *Encryptor) { o.replaceTable = replaceTable }
}

func NewEncryptor(opts ...OptionFunc) *Encryptor {
	encryptor := DefaultEncryptor
	for _, opt := range opts {
		opt(&encryptor)
	}
	return &encryptor
}

// 对明文进行字符替换
func (e *Encryptor) ReplaceCharacters(plaintext string) string {
	var sb strings.Builder
	for _, c := range plaintext {
		index := strings.IndexRune(e.replaceTable, c)
		if index >= 0 {
			sb.WriteRune(rune(e.replaceWith[index]))
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// 对密文进行字符替换
func (e *Encryptor) RestoreCharacters(ciphertext string) string {
	var sb strings.Builder
	for _, c := range ciphertext {
		index := strings.IndexRune(e.replaceWith, c)
		if index >= 0 {
			sb.WriteRune(rune(e.replaceTable[index]))
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// 对明文进行对称加密
func (e *Encryptor) EncryptSymmetric(plaintext string) (string, error) {
	key := []byte(e.key)
	iv := []byte(e.iv)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	plaintext = e.ReplaceCharacters(plaintext)
	padded := pkcs5Padding([]byte(plaintext), block.BlockSize())
	ciphertext := make([]byte, len(padded))
	mode.CryptBlocks(ciphertext, padded)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// 对密文进行对称解密
func (e *Encryptor) DecryptSymmetric(ciphertext string) (string, error) {
	key := []byte(e.key)
	iv := []byte(e.iv)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	plaintext := make([]byte, len(data))
	mode.CryptBlocks(plaintext, data)
	plaintext = pkcs5UnPadding(plaintext)
	return e.RestoreCharacters(string(plaintext)), nil
}

// 对明文进行多重加密
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	switch e.algorithm {
	case "AES/CBC/PKCS5Padding":
		return e.EncryptSymmetric(plaintext)
	default:
		return "", errors.New("unsupported algorithm: " + e.algorithm)
	}
}

// 对密文进行多重解密
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	switch e.algorithm {
	case "AES/CBC/PKCS5Padding":
		return e.DecryptSymmetric(ciphertext)
	default:
		return "", errors.New("unsupported algorithm: " + e.algorithm)
	}
}

func pkcs5Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func pkcs5UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}
