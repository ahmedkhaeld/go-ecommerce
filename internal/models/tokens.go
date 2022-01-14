package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"log"
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

func (m *DBModel) InsertToken(t *Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// delete existing tokens, no need to keep the old tokens when user login again
	stmt := `delete from tokens where user_id = ?`
	_, err := m.DB.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	stmt = `insert into tokens (user_id, name, email, token_hash, expiry, created_at, updated_at)
			values (?, ?, ?, ?, ?, ?, ?)`

	_, err = m.DB.ExecContext(ctx, stmt,
		u.ID,
		u.LastName,
		u.Email,
		t.Hash,
		t.Expiry,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *DBModel) GetUserForToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))

	var user User

	query := ` 
		select 
			u.id, u.first_name, u.last_name, u.email
		from
			users u
			inner join tokens t on (u.id = t.user_id)
		where
			t.token_hash = ?
		and t.expiry > ?
	`
	// convert slice to an array to handle the token hash
	err := m.DB.QueryRowContext(ctx, query, tokenHash[:], time.Now()).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
	)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &user, nil
}
