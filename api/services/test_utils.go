// common methods for setting up tests, generating data, etc.
package services

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
)

func testSuiteSetup(psvc **Services) {
	if *psvc != nil {
		return
	}

	svc, err := New("test")
	if err != nil {
		panic(err.Error())
	}
	*psvc = svc
	if err := flushRedis(svc); err != nil {
		panic(err.Error())
	}
	if err := truncateAllTables(svc); err != nil {
		panic(err.Error())
	}
}

func flushRedis(svc *Services) error {
	if err := svc.redis.FlushAll().Err(); err != nil {
		return fmt.Errorf("flushRedis failed: %v", err)
	}
	return nil
}

func truncateAllTables(svc *Services) error {
	var tx *sql.Tx
	done := func(err error) error {
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("truncateAllTables failed: %v", err)
		} else {
			return nil
		}
	}

	tx, err := svc.db.Begin()
	if err != nil {
		return done(err)
	}
	names, err := getTableNames(tx)
	if err != nil {
		return done(err)
	}

	for _, name := range names {
		stmt := fmt.Sprintf("TRUNCATE TABLE %v CASCADE", name)
		if _, err := tx.Exec(stmt); err != nil {
			return done(err)
		}
	}

	return done(nil)
}

func getTableNames(tx *sql.Tx) ([]string, error) {
	names := make([]string, 0)
	cursor, err := tx.Query("SELECT tablename FROM pg_tables WHERE tableowner = (SELECT current_user)")
	if err != nil {
		return nil, err
	}

	defer cursor.Close()
	for cursor.Next() {
		var name string
		if err := cursor.Scan(&name); err != nil {
			log.Fatal(err)
		}
		names = append(names, name)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return names, nil
}

func mockUsers(n int) map[RecordID]User {
	users := map[RecordID]User{}
	for i := 0; i < n; i++ {
		user := mockUser()
		users[user.ID] = user
	}
	return users
}

func randomUser(users map[RecordID]User) User {
	randomIdx := rand.Intn(len(users))
	idx := 0
	for _, user := range users {
		if idx == randomIdx {
			return user
		}
		idx++
	}
	panic("should never reach here")
}

func mockUser() User {
	return User{
		ID:    RecordID(randString(32, alphanum)),
		Email: randEmail(),
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var digits = []rune("0123456789")
var alphanum = append(append(make([]rune, 0), digits...), letters...)

func randEmail() string {
	return randWord(10) + "@" + randWord(8) + ".com"
}

func randWord(n int) string {
	return randString(n, letters)
}

func randString(n int, chars []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(letters))]
	}
	return string(b)
}

//
// func setupTokensTest(t *testing.T) *Users {
// 	flushAll()
//
// 	Users, err := newUsers(nil, redisClient)
// 	assert.Nil(t, err)
// 	return Users
// }
