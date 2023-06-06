package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	up "github.com/upper/db/v4"
)

type Token struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	FirstName string    `db:"first_name" json:"first_name"`
	Email     string    `db:"email" json:"email"`
	PlainText string    `db:"token" json:"token"`
	Hash      []byte    `db:"token_hash" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Expires   time.Time `db:"expiry" json:"expiry"`
}

func (t *Token) Table() string {
	return "tokens"
}

func (t *Token) GetUserByToken(token string) (*User, error) {
	var u User
	var uToken Token

	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": token})
	err := res.One(&uToken)
	if err != nil {
		return nil, err
	}

	collection = upper.Collection("users")
	res = collection.Find(up.Cond{"id": uToken.UserID})
	err = res.One(&u)
	if err != nil {
		return nil, err
	}

	u.Token = uToken

	return &u, nil
}

func (t *Token) GetUserTokens(id int) ([]*Token, error) {
	var tokens []*Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"user_id": id})
	err := res.All(&tokens)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (t *Token) GetTokenByID(id int) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"id": id})
	err := res.One(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (t *Token) GetToken(plainText string) (*Token, error) {
	var token Token
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": plainText})
	err := res.One(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (t *Token) DeleteToken(plainText string) error {
	collection := upper.Collection(t.Table())
	res := collection.Find(up.Cond{"token": plainText})
	err := res.Delete()
	if err != nil {
		return err
	}

	return nil
}

func (t *Token) Insert(token Token, u User) error {
	collection := upper.Collection(t.Table())

	// delete existing tokens
	res := collection.Find(up.Cond{"user_id": u.ID})
	err := res.Delete()
	if err != nil {
		return err
	}

	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	token.FirstName = u.FirstName
	token.Email = u.Email

	_, err = collection.Insert(token)
	if err != nil {
		return err
	}
	return nil
}

func (t *Token) GenerateToken(userID int, ttl time.Duration) (*Token, error) {
	token := &Token{
		UserID:  userID,
		Expires: time.Now().Add(ttl),
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	token.PlainText = base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

func (t *Token) AuthenticateToken(r *http.Request) (*User, error) {
	headers := r.Header.Get("Authorization")
	if headers == "" {
		return nil, errors.New("no authorization header received")
	}
	headerParts := strings.Split(headers, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, errors.New("no authorization header received")
	}

	token := headerParts[1]

	if len(token) != 26 {
		return nil, errors.New("token wrong size")
	}

	tkn, err := t.GetToken(token)
	if err != nil {
		return nil, errors.New("no matching token found")
	}

	if tkn.Expires.Before(time.Now()) {
		return nil, errors.New("expired token")
	}

	user, err := t.GetUserByToken(token)
	if err != nil {
		return nil, errors.New("no matching user found")
	}
	return user, nil
}

func (t *Token) ValidToken(token string) (bool, error) {
	user, err := t.GetUserByToken(token)
	if err != nil {
		return false, errors.New("no matching user found")
	}

	if user.Token.PlainText == "" {
		return false, errors.New("no matching token found")
	}

	if user.Token.Expires.Before(time.Now()) {
		return false, errors.New("expired token")
	}

	return true, nil
}
