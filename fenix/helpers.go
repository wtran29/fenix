package fenix

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"math/big"
	"os"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

func (f *Fenix) RandomString(n int) (string, error) {
	result := make([]byte, n)
	length := big.NewInt(int64(len(charset)))

	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, length)
		if err != nil {
			return "", err
		}

		result[i] = charset[idx.Int64()]
	}
	return string(result), nil
}

func (f *Fenix) CreateDirIfNotExist(path string) error {
	const mode = 0755
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Fenix) CreateFileIfNotExists(path string) error {
	var _, err = os.Stat(path)
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	}
	return nil
}

type Encryption struct {
	Key []byte
}

func (e *Encryption) Encrypt(text string) (string, error) {
	plaintext := []byte(text)

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	// Generate new random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Add padding to the plaintext
	plaintext = paddingText(plaintext, block.BlockSize())

	ciphertext := make([]byte, len(iv)+len(plaintext))
	copy(ciphertext, iv)

	// Create the cipher stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Encrypt the plaintext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (e *Encryption) Decrypt(encText string) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("invalid ciphertext")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Create cipher stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt the ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	// Remove padding from plaintext
	plaintext, err := unpaddingText(ciphertext, block.BlockSize())
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// paddingText applies PKCS7 padding to the plaintext
func paddingText(plaintext []byte, blockSize int) []byte {
	padding := blockSize - (len(plaintext) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padText...)
}

// unpaddingText removes PKCS7 padding from the plaintext
func unpaddingText(plaintext []byte, blockSize int) ([]byte, error) {
	padding := int(plaintext[len(plaintext)-1])
	if padding < 1 || padding > blockSize {
		return nil, errors.New("invalid padding")
	}
	if len(plaintext) < padding {
		return nil, errors.New("invalid padding")
	}
	for i := len(plaintext) - 1; i > len(plaintext)-padding-1; i-- {
		if int(plaintext[i]) != padding {
			return nil, errors.New("invalid padding")
		}
	}
	return plaintext[:len(plaintext)-padding], nil
}
