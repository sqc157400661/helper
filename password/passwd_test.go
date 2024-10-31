package password

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreatePasswd(t *testing.T) {
	var pwd string
	pwd = CreatePasswd(8, CreatePWDWithNum)
	assert.Equal(t, len(pwd), 8)
	fmt.Println(pwd)
	pwd = CreatePasswd(8, CreatePWDWithChar)
	assert.Equal(t, len(pwd), 8)
	fmt.Println(pwd)
	pwd = CreatePasswd(8, CreatePWDWithMix)
	assert.Equal(t, len(pwd), 8)
	fmt.Println(pwd)
	pwd = CreatePasswd(8, CreatePWDWithAdvance)
	assert.Equal(t, len(pwd), 8)
	fmt.Println(pwd)

}

func TestEncrypt(t *testing.T) {
	a := NewEncryptor()
	aa := "sqctest"
	fmt.Println(a.Encrypt(aa))

	bb := "T15+2TXdFynaKaRXGrk8vA=="

	bb1, _ := a.Decrypt(bb)

	fmt.Println(bb1)
	//fmt.Println(a.Decrypt(string(decodedBytes)))
}
