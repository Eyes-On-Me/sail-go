package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"github.com/sail-services/sail-go/com/data/crypt/pkcs5"
)

// AES 加密 - key长度 16, 24, 32 对应 AES-128, AES-192, AES-256
func Encrypt(orig_data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	block_size := block.BlockSize()
	orig_data = pkcs5.Padding(orig_data, block_size)
	block_mode := cipher.NewCBCEncrypter(block, key[:block_size])
	crypted := make([]byte, len(orig_data))
	block_mode.CryptBlocks(crypted, orig_data)
	return crypted, nil
}

// AES 解密
func Decrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	block_size := block.BlockSize()
	block_mode := cipher.NewCBCDecrypter(block, key[:block_size])
	orig_data := make([]byte, len(crypted))
	block_mode.CryptBlocks(orig_data, crypted)
	return pkcs5.Unpadding(orig_data, block_size)
}
