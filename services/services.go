package services

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/satori/go.uuid"

	"gopkg.in/redis.v3"
)

type Services struct {
	Users *Users
	db    *sql.DB
	redis *redis.Client
}

var (
	ErrUniqueViolation = fmt.Errorf("unique_violation")
	ErrNotFound        = fmt.Errorf("not_found")
)

func New(env string) (*Services, error) {
	config, err := loadConfig(env)
	if err != nil {
		return nil, err
	} else {
		fmt.Printf("using config: %#v\n", config)
	}

	db, err := setupDB(config)
	if err != nil {
		return nil, err
	}
	redisClient, err := setupRedis(config)
	if err != nil {
		return nil, err
	}

	users, err := newUsers(config, db, redisClient)
	if err != nil {
		return nil, err
	}

	svc := &Services{Users: users, db: db, redis: redisClient}

	return svc, nil
}

func setupDB(config Config) (*sql.DB, error) {
	connParams := config.PostgresURL
	db, err := sql.Open("postgres", connParams)
	if err != nil {
		return nil, fmt.Errorf("unable to create db driver with params %v: %v", connParams, err)
	}

	_, err = db.Query("SELECT 1")
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db with params %v: %v", connParams, err)
	}

	return db, nil
}

func setupRedis(config Config) (*redis.Client, error) {
	addr := config.RedisURL
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("unable to ping redis server at %v: %v", addr, err)
	}
	return client, nil
}

type RecordID string

func (id RecordID) String() string {
	return string(id)
}

func (id RecordID) Value() (driver.Value, error) {
	return string(id), nil
}

func (id *RecordID) Scan(val interface{}) error {
	bytes, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("Cast error: expected RecordID bytes, got %v", val)
	}
	str := string(bytes)
	*id = RecordID(strings.Replace(str, "-", "", -1))
	return nil
}

func newID() RecordID {
	u4 := uuid.NewV4()

	u4str := strings.ToLower(strings.Replace(u4.String(), "-", "", -1))
	return RecordID(u4str)
}
