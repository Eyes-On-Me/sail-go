package triple_des

import (
	"crypto/cipher"
	"crypto/des"
	"github.com/sail-services/sail-go/com/data/crypt/pkcs5"
)

// 3DES 加密, key 只能24位长度
func Encrypt(origData, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	origData = pkcs5.Padding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:8])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// 3DES 解密, key 只能24位长度
func Decrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:8])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	return pkcs5.Unpadding(origData, block.BlockSize())
}
