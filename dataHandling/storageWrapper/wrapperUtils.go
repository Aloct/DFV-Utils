package wrapperUtils

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
)

type DBWrapper interface {
	Connect(ctx context.Context) error
	Close() error
	Query(ctx context.Context, query string) (interface{}, error)
	RateLimit(ctx context.Context) error
	rotatePW(ctx context.Context) error
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
	DB *redis.Client
}

func NewRedisWrapper(rateLimit bool, pwRotation bool, subject string, container string, port int, password string) *RedisWrapper {
	redisClient := RedisWrapper{
		BaseWrapper: BaseWrapper{
			RateLimitEnabled: rateLimit,
			PwRotationEnabled: pwRotation,
			Subject: subject,
			Container: container,
			Port: port,
			Password: password,
		},
	}

	return &redisClient
}

func (r *RedisWrapper) Connect(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", r.Container, r.Port), 
		Password: r.Password, 
		DB: 0,
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

type MongoWrapper struct {
	BaseWrapper
}

type MySQLWrapper struct {
	BaseWrapper
}