package urlsigner

import (
	"fmt"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

type Signer struct {
	Secret []byte
}

func (s *Signer) GenerateTokenFromString(data string) string {
	var urlToSign string

	signer := goalone.New(s.Secret, goalone.Timestamp)
	if strings.Contains(data, "?") {
		urlToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		urlToSign = fmt.Sprintf("%s&hash=", data)
	}

	tokenBytes := signer.Sign([]byte(urlToSign))
	token := string(tokenBytes)

	return token

}

func (s *Signer) VerifyToken(token string) bool {
	signer := goalone.New(s.Secret, goalone.Timestamp)
	_, err := signer.Unsign([]byte(token))
	return err == nil
}

func (s *Signer) Expired(token string, minTilExpire int) bool {
	signer := goalone.New(s.Secret, goalone.Timestamp)
	ts := signer.Parse([]byte(token))

	return time.Since(ts.Timestamp) > time.Duration(minTilExpire)*time.Minute
}
