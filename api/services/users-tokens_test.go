package services

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var tokensTestSvc *Services

func setupTokensTest(t *testing.T) *Services {

	testSuiteSetup(&tokensTestSvc)
	flushRedis(tokensTestSvc) //for this suite we want to flush between each test

	return tokensTestSvc
}

func TestTokensGenerateToken(t *testing.T) {
	count := 1024
	userID := newID()
	size := 24
	tokens := map[Token]struct{}{}

	for i := 0; i < count; i++ {
		token, err := tokenGenerate(userID, size)
		assert.Nil(t, err)

		buf, err := base64.URLEncoding.DecodeString(base64Pad(token))
		assert.Nil(t, err)

		assert.Equal(t, len(buf), len(userID)+1+size, "size mismatched")

		assert.True(t, strings.HasPrefix(string(buf), string(userID)), "token should start with user ID")

		_, exists := tokens[token]
		assert.False(t, exists, "duplicate token: "+token)

		tokens[token] = struct{}{}

	}
}

func base64Pad(token Token) string {
	padded := string(token)
	if m := len(padded) % 4; m != 0 {
		padded += strings.Repeat("=", 4-m)
	}
	return padded
}

func TestTokensCreateAndGet(t *testing.T) {
	svc := setupTokensTest(t)

	tokensPerUser := 8
	tokensByUser := map[RecordID][]Token{}
	users := mockUsers(4)

	//create tokens
	for userID, user := range users {
		tokensByUser[userID] = make([]Token, tokensPerUser)
		for i := 0; i < tokensPerUser; i++ {

			token, err := svc.Users.tokenCreate(user)
			assert.Nil(t, err)

			tokensByUser[userID][i] = token
		}
	}

	//get users from tokens
	for userID, userTokens := range tokensByUser {
		expected := users[userID]
		for _, token := range userTokens {
			actual, err := svc.Users.tokenGetUser(token)
			assert.Nil(t, err)
			assert.EqualValues(t, expected, actual, "users should be same")
		}
	}
}

func TestTokensExpire(t *testing.T) {
	svc := setupTokensTest(t)
	user := mockUser()
	duration := time.Millisecond * 100
	token, err := svc.Users.tokenCreateWithOptions(user, TokenOptions{Duration: duration})
	assert.Nil(t, err)
	actual, err := svc.Users.tokenGetUser(token)
	assert.Nil(t, err)
	assert.EqualValues(t, user, actual, "users should be same")

	time.Sleep(duration / 2)

	actual, err = svc.Users.tokenGetUser(token)
	assert.Nil(t, err)
	assert.EqualValues(t, user, actual, "users should be same")

	time.Sleep(duration/2 + 1)
	actual, err = svc.Users.tokenGetUser(token)
	assert.Equal(t, err, ErrNotFound)
	assert.Equal(t, actual, User{})
}

func TestTokensDelete(t *testing.T) {
	svc := setupTokensTest(t)
	user := mockUser()
	token, err := svc.Users.tokenCreate(user)
	assert.Nil(t, err)
	actual, err := svc.Users.tokenGetUser(token)
	assert.Nil(t, err)
	assert.EqualValues(t, user, actual, "users should be same")

	err = svc.Users.tokenDelete(user.ID, token)
	assert.Nil(t, err)
	actual, err = svc.Users.tokenGetUser(token)
	assert.Equal(t, err, ErrNotFound)
	assert.Equal(t, actual, User{})
}

func TestTokensDeleteAll(t *testing.T) {
	svc := setupTokensTest(t)
	tokensPerUser := 8
	tokensByUser := map[RecordID][]Token{}
	users := mockUsers(4)

	//create tokens
	for userID, user := range users {
		tokensByUser[userID] = make([]Token, tokensPerUser)
		for i := 0; i < tokensPerUser; i++ {

			token, err := svc.Users.tokenCreate(user)
			assert.Nil(t, err)

			tokensByUser[userID][i] = token
		}
	}

	chosenUser := randomUser(users)
	err := svc.Users.tokenDeleteAll(chosenUser.ID)
	assert.Nil(t, err)

	for _, user := range users {
		for _, token := range tokensByUser[user.ID] {
			actual, err := svc.Users.tokenGetUser(token)
			if user == chosenUser {
				assert.Equal(t, err, ErrNotFound)
				assert.Equal(t, actual, User{})
			} else {
				assert.EqualValues(t, user, actual, "user should be the same")
			}
		}
	}

}
