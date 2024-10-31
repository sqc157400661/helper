package password

import (
	"crypto/rand"
	"math/big"
)

const (
	// 密码生成类型
	CreatePWDWithNum     = "num"
	CreatePWDWithChar    = "char"
	CreatePWDWithMix     = "mix"
	CreatePWDWithAdvance = "advance"
)

func CreatePasswd(length int, kind string) string {
	passwd := make([]rune, length)
	var codeModel []rune
	switch kind {
	case CreatePWDWithNum:
		codeModel = []rune("0123456789")
	case CreatePWDWithChar:
		codeModel = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	case CreatePWDWithMix:
		codeModel = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	case CreatePWDWithAdvance:
		codeModel = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+=-!@#$%*,.[]")
	default:
		codeModel = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	}
	for i := range passwd {
		index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(codeModel))))
		passwd[i] = codeModel[int(index.Int64())]
	}
	return string(passwd)

}
