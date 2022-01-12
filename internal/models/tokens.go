package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

// ScopeAuthentication
//when working with an api, there are different kinds of scope
// a good practice to identify scope for some particular part of code
const (
	ScopeAuthentication = "authentication"
)

// Token type for authentication tokens
type Token struct {
	PlainText string    `json:"token"`
	UserID    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// GenerateToken creates a token that last for ttl
//take in user id, time to life, and scope
// returns a pointer to Token and potentially an error
func GenerateToken(userID int, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: int64(userID),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}
	// make sure token is secured, by assigning random bytes
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	// create the token plainText that being sent to the end user
	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]
	return token, nil
}
