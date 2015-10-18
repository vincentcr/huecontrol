package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/imdario/mergo"

	"gopkg.in/redis.v3"
)

const tokenKeyFormat = "token.%s"
const tokenListKeyFormat = "tokenlist.%s"

type Token string

type TokenOptions struct {
	Duration   time.Duration
	SecretSize int
}

var DefaultTokenOptions = TokenOptions{Duration: 0, SecretSize: 32}

func (users *Users) tokenCreate(user User) (Token, error) {
	return users.tokenCreateWithOptions(user, TokenOptions{})
}

func (users *Users) tokenCreateWithOptions(user User, options TokenOptions) (Token, error) {
	mergo.Merge(&options, DefaultTokenOptions)

	userJson, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("unable to json-encode user %v: %v", user, err)
	}

	token, err := tokenGenerate(user.ID, options.SecretSize)
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf(tokenKeyFormat, token)
	err = users.redis.Set(key, string(userJson), options.Duration).Err()
	if err != nil {
		return "", fmt.Errorf("redis.Set(%v, %s, %v) failed: %v", key, userJson, options.Duration, err)
	}

	tokenListKey := fmt.Sprintf(tokenListKeyFormat, user.ID)
	err = users.redis.SAdd(tokenListKey, key).Err()
	if err != nil {
		return "", fmt.Errorf("redis.Sadd(%v, %s) failed: %v", tokenListKey, key)
	}

	return token, nil
}

func tokenGenerate(userID RecordID, size int) (Token, error) {
	secretOffset := len(userID) + 1
	buf := make([]byte, secretOffset+size)
	copy(buf, userID+":")
	nRandBytes, err := rand.Read(buf[secretOffset:])
	if err != nil {
		return "", fmt.Errorf("unable to generate %v bytes of randomness: %v", size, err)
	} else if nRandBytes < size {
		return "", fmt.Errorf("got %d bytes from rand instead of requested %v", nRandBytes, size)
	}

	encoded := base64.URLEncoding.EncodeToString(buf)

	//URLEncode still might contain the '=', which is not very URL-friendly.
	encodedForURL := strings.Replace(encoded, "=", "", -1)

	return Token(encodedForURL), nil
}

func (users *Users) tokenGetUser(token Token) (User, error) {
	key := fmt.Sprintf(tokenKeyFormat, token)
	userJson, err := users.redis.Get(key).Result()
	if err == redis.Nil {
		return User{}, ErrNotFound
	} else if err != nil {
		return User{}, fmt.Errorf("unable to get key %v: %v", token, err)
	}

	user := User{}
	err = json.Unmarshal([]byte(userJson), &user)
	if err != nil {
		return User{}, fmt.Errorf("unable to unmarshall user from json %v: %v", userJson, err)
	}

	return user, nil
}

func (users *Users) tokenDelete(userID RecordID, token Token) error {
	key := fmt.Sprintf(tokenKeyFormat, token)
	err := users.redis.Del(key).Err()
	if err != nil {
		return fmt.Errorf("unable to delete key %v: %v", key, err)
	}

	tokenListKey := fmt.Sprintf(tokenListKeyFormat, userID)
	err = users.redis.SRem(tokenListKey, key).Err()
	if err != nil {
		return fmt.Errorf("unable to remove entry %v from set %v: %v", key, tokenListKey, err)
	}

	return nil
}

func (users *Users) tokenDeleteAll(userID RecordID) error {
	tokenListKey := fmt.Sprintf(tokenListKeyFormat, userID)
	keys, err := users.redis.SMembers(tokenListKey).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("unable to get members of set %v: %v", tokenListKey, err)
	}

	err = users.redis.Del(keys...).Err()
	if err != nil {
		return fmt.Errorf("unable to delete keys %v: %v", keys, err)
	}

	return nil
}
