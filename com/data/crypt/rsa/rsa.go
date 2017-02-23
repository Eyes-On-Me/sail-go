package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// 随机生成个私钥和公钥 [中间包括公钥]
func NewPrivKey(bits int) *rsa.PrivateKey {
	priv, _ := rsa.GenerateKey(rand.Reader, bits)
	return priv
}

// 加密
func Encrypt(src []byte, pub_key *rsa.PublicKey) (enc []byte, err error) {
	return rsa.EncryptOAEP(sha1.New(), rand.Reader, pub_key, src, nil)
}

// 解密
func Decrypt(src []byte, priv_key *rsa.PrivateKey) (plain []byte, err error) {
	return rsa.DecryptOAEP(sha1.New(), rand.Reader, priv_key, src, nil)
}

// 加密
func EncryptB(orig_data, pub_key []byte) ([]byte, error) {
	block, _ := pem.Decode(pub_key)
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, orig_data)
}

// 解密
func DecryptB(cipher_text, priv_key []byte) ([]byte, error) {
	block, _ := pem.Decode(priv_key)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, cipher_text)
}
