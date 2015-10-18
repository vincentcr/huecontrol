package services

import (
	"regexp"
	"testing"
)
import "github.com/stretchr/testify/assert"

var usersTestSvc *Services

func setupUsersTest(t *testing.T) *Services {

	testSuiteSetup(&usersTestSvc)

	return usersTestSvc
}

func TestNewID(t *testing.T) {
	ids := map[RecordID]struct{}{}
	for i := 0; i < 64; i++ {
		id := newID()
		assert.Len(t, id, 32)
		_, found := ids[id]
		assert.False(t, found)
		ids[id] = struct{}{}
	}
}

func TestUsersHashAndVerifyPassword(t *testing.T) {
	password := randWord(16)
	hash, err := hashPassword(password)
	assert.Nil(t, err)
	assertHashedPassword(t, password, hash)
	hash2, err := hashPassword(password)
	assert.NotEqual(t, hash, hash2)
	assertHashedPassword(t, password, hash2)

	assert.False(t, verifyPassword(password+"wrong", User{password: hash}))
	assert.False(t, verifyPassword("", User{password: hash}))
	assert.False(t, verifyPassword(hash, User{password: hash}))
}

func assertHashedPassword(t *testing.T, clear, hash string) {
	assert.NotEqual(t, clear, hash)
	assert.Regexp(t, regexp.MustCompile("^\\$2a\\$\\d+\\$.{10,}"), hash)
	assert.True(t, verifyPassword(clear, User{password: hash}))
}

func TestUsersNormalizeEmail(t *testing.T) {
	expected := "foo@bar"
	testCases := []string{
		"foo@bar",
		"FoO@bar",
		"FOO@BAR",
		" foo@bar ",
		"     FoO@bar ",
		" \t\nFoO@bar\n\n\n\n\n ",
	}

	for _, testCase := range testCases {
		assert.Equal(t, normalizeEmail(testCase), expected)
	}
}

func TestUsersCreate(t *testing.T) {
	svc := setupUsersTest(t)
	email := randEmail()
	password := randWord(16)
	u, err := svc.Users.Create(email, password)
	assert.Nil(t, err)
	assert.Equal(t, normalizeEmail(email), u.Email)
	assertHashedPassword(t, password, u.password)

	var foundCount int
	err = svc.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE id = $1 AND email = $2 AND password = $3",
		u.ID, u.Email, u.password).
		Scan(&foundCount)
	assert.Nil(t, err)
	assert.Equal(t, foundCount, 1)
}

func TestUsersCreateDuplicateEmailError(t *testing.T) {
	svc := setupUsersTest(t)
	email := randEmail()
	_, err := svc.Users.Create(email, randWord(16))
	assert.Nil(t, err)
	u2, err := svc.Users.Create(email, randWord(16))
	assert.Equal(t, err, ErrUniqueViolation)
	assert.Equal(t, u2, User{})

	var foundCount int
	err = svc.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE email = $1", normalizeEmail(email)).
		Scan(&foundCount)
	assert.Nil(t, err)
	assert.Equal(t, foundCount, 1)
}

func TestUsersGetByID(t *testing.T) {
	svc := setupUsersTest(t)
	users := make([]User, 0)
	for i := 0; i < 4; i++ {
		email := randEmail()
		password := randWord(16)
		u, err := svc.Users.Create(email, password)
		assert.Nil(t, err)
		users = append(users, u)
	}

	for _, u_c := range users {
		u_g, err := svc.Users.GetByID(u_c.ID)
		assert.Nil(t, err)

		assert.Equal(t, u_c.ID, u_g.ID)
		assert.Equal(t, u_c.Email, u_g.Email)
		assert.Equal(t, u_c.password, u_g.password)
	}
}

func TestUsersGetByIDNotFound(t *testing.T) {
	svc := setupUsersTest(t)
	_, err := svc.Users.Create(randEmail(), randWord(16))
	assert.Nil(t, err)

	_, err = svc.Users.GetByID(newID())
	assert.Equal(t, err, ErrNotFound)
}

func TestUsersAuthenticateWithPassword(t *testing.T) {
	svc := setupUsersTest(t)
	users := make([]User, 0)
	passwords := make([]string, 0)
	for i := 0; i < 4; i++ {
		email := randEmail()
		password := randWord(16)
		passwords = append(passwords, password)
		u, err := svc.Users.Create(email, password)
		assert.Nil(t, err)
		users = append(users, u)
	}

	for i, u_c := range users {
		u_g, err := svc.Users.AuthenticateWithPassword(u_c.Email, passwords[i])
		assert.Nil(t, err)

		assert.Equal(t, u_c.ID, u_g.ID)
		assert.Equal(t, u_c.Email, u_g.Email)
		assert.Equal(t, u_c.password, u_g.password)
	}

	bad_pairs := [][]string{
		{users[0].Email, passwords[0] + "nope"},
		{users[0].Email, passwords[1]},
		{"", passwords[0]},
		{users[0].Email, ""},
	}

	for _, bad_pair := range bad_pairs {
		u_not, err := svc.Users.AuthenticateWithPassword(bad_pair[0], bad_pair[1])
		assert.Equal(t, err, ErrNotFound)
		assert.Equal(t, u_not, User{})

	}

}
