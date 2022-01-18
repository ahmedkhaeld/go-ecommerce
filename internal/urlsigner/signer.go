package urlsigner

import (
	"fmt"
	goalone "github.com/bwmarrin/go-alone"
	"strings"
	"time"
)

type Signer struct {
	Secret []byte
}

// GenerateTokenFromString take a string and sign it and hand back the signed string
func (s *Signer) GenerateTokenFromString(data string) string {
	var urlToSign string

	// sign new urls that already have url params, else it contains a question mark
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	if strings.Contains(data, "?") {
		urlToSign = fmt.Sprintf("%s&hash=", data)
	} else {
		urlToSign = fmt.Sprintf("%s?hash=", data)
	}
	// signing the email
	tokenBytes := crypt.Sign([]byte(urlToSign))
	token := string(tokenBytes)
	return token
	// the token is a fully signed url that end with hash equals plus the bit to validate the url
}

// VerifyToken verify that the link click  hasn't been changed in any way.
func (s *Signer) VerifyToken(token string) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	_, err := crypt.Unsign([]byte(token))

	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// Expired check to see if token is expired
func (s *Signer) Expired(token string, minutesUntilExpire int) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	ts := crypt.Parse([]byte(token))

	return time.Since(ts.Timestamp) > time.Duration(minutesUntilExpire)*time.Minute
}
