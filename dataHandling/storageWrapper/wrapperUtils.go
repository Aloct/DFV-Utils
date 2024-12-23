package wrapperUtils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
)

type DBWrapper interface {
	Connect(ctx context.Context) error
	Close() error
	Query(ctx context.Context, query string) (interface{}, error)
	RateLimit(ctx context.Context) error
	rotatePW(ctx context.Context) error
	getData()
	setData()
}

type BaseWrapper struct {
	RateLimitEnabled bool
	PwRotationEnabled bool
	Subject string
	Container string
	Port int
	Password string
}

func (b *BaseWrapper) RateLimit(ctx context.Context) error {
	if b.RateLimitEnabled {

	}

	return nil
}

func (b *BaseWrapper) PwRotation(ctx context.Context) error {
	if b.PwRotationEnabled {

	}

	return nil
}

type RedisWrapper struct {
	BaseWrapper
	DbNum int
	DB *redis.Client
}

func NewRedisWrapper(rateLimit bool, pwRotation bool, subject string, container string, port int, dbnum int, password string) *RedisWrapper {
	redisClient := RedisWrapper{
		BaseWrapper: BaseWrapper{
			RateLimitEnabled: rateLimit,
			PwRotationEnabled: pwRotation,
			Subject: subject,
			Container: container,
			Port: port,
			Password: password,
		},
		DbNum: dbnum,
	}

	return &redisClient
}

func (r *RedisWrapper) Connect(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", r.Container, r.Port), 
		Password: r.Password, 
		DB: r.DbNum,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Fatalf("Redis connection failed %v", err)
	}
	fmt.Println("Connected to Redis")

	r.DB = rdb
	return nil
}

func (r *RedisWrapper) getData() {

}

func (r *RedisWrapper) setData() {

}


type MongoWrapper struct {
	BaseWrapper
}


type MySQLWrapper struct {
	BaseWrapper
	DB *sql.DB
	DBname string
	User string
}

func NewSQLWrapper(rateLimit bool, pwRotation bool, subject string, container string, port int, dbname string, user string, password string) *MySQLWrapper {
	sqlWrapper := MySQLWrapper{
		BaseWrapper: BaseWrapper{
			RateLimitEnabled: rateLimit,
			PwRotationEnabled: pwRotation,
			Subject: subject,
			Container: container,
			Port: port,
			Password: password,
		},
		DBname: dbname,
		User: user,
	}

	return &sqlWrapper
}

func (sr *MySQLWrapper) Connect() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", sr.User, sr.Password, sr.Container, sr.Port, sr.DBname))
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	sr.DB = db
}