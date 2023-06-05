package fenix

import (
	"crypto/rand"
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
