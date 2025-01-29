package wrapperUtils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type DBWrapper interface {
	// Connect(ctx context.Context) error
	// Close() error
	// // RateLimit(ctx context.Context) error
	// // rotatePW(ctx context.Context) error
	// getData(query string) map[int]map[string]any 
	// setData(query string) 
}

type DBPool struct {
	pool map[string]DBWrapper
}

type BaseWrapper struct {
	RateLimitEnabled bool
	PwRotationEnabled bool
	Subject string
	Container string
	Port int
	Password string
}

// func (b *BaseWrapper) RateLimit(ctx context.Context) error {
// 	if b.RateLimitEnabled {

// 	}

// 	return nil
// }

// func (b *BaseWrapper) PwRotation(ctx context.Context) error {
// 	if b.PwRotationEnabled {

// 	}

// 	return nil
// }

type RedisWrapper struct {
	BaseWrapper
	DbNum int
	DB *redis.Client
}

func NewDBPool() *DBPool {
	return &DBPool{
		pool: make(map[string]DBWrapper, 0),
	}
}

func (p *DBPool) NewRedisWrapper(dbName string) (*RedisWrapper, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	dbNum, err := strconv.Atoi(os.Getenv(dbName + "_NUM"))
	if err != nil {
		return nil, err
	}

	dbPort, err := strconv.Atoi(os.Getenv(dbName + "_PORT"))

	redisClient := RedisWrapper{
		BaseWrapper: BaseWrapper{			
			// RateLimitEnabled: rateLimit,
			// PwRotationEnabled: pwRotation,
			Subject: os.Getenv(dbName  + "_SUBJECT"),
			Container: os.Getenv(dbName  + "_CONTAINER"),
			Port: dbPort,
			Password: os.Getenv(dbName  + "_PASSWORD"),
		},
		DbNum: dbNum,
	}

	return &redisClient, err
}

func (r *RedisWrapper) Connect(ctx context.Context) error {
	fmt.Println("test")
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", r.Container, r.Port), 
		Password: "", 
		DB: r.DbNum,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Fatalf("Redis connection failed %v", err)
	}
	fmt.Println("Connected to" + r.Container)

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

func (p DBPool) NewSQLWrapper(dbName string) *MySQLWrapper{
	sqlWrapper := MySQLWrapper{
		BaseWrapper: BaseWrapper{
			// RateLimitEnabled: rateLimit,
			// PwRotationEnabled: pwRotation,
			Subject: os.Getenv(dbName  + "_SUBJECT"),
			Container: os.Getenv(dbName  + "_CONTAINER"),
			Port: 3306,
			Password: os.Getenv(dbName  + "_PASSWORD"),
		},
		DBname: os.Getenv(dbName  + "_NAME"),
		User: os.Getenv(dbName  + "_USER"),
	}

	p.pool[dbName] = sqlWrapper

	return &sqlWrapper
}

func (sr *MySQLWrapper) Connect(ctx context.Context) error {
	fmt.Println("test")
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", sr.User, sr.Password, sr.Container, sr.Port, sr.DBname))
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	fmt.Println("Connected to" + sr.Container)

	sr.DB = db

	return nil
}

func (sr *MySQLWrapper) Close() error {
	return nil
}

// expectedDts: map[string]any
func (sr *MySQLWrapper) getData(query string) map[int]map[string]any {
	rows, err := sr.DB.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	returnedMap := make(map[int]map[string]any)
	uncastedRowValues := make([]any, len(columns))
	emptyMap := make(map[string]any)

	for rows.Next() {
		if err := rows.Scan(uncastedRowValues...); err != nil {
			log.Fatal(err)
		}

		for i, v := range columns {
			emptyMap[v] = uncastedRowValues[i]
		}

		returnedMap[uncastedRowValues[0].(int)] = emptyMap 
	}

	return returnedMap
}

func (sr *MySQLWrapper) setData(query string) {
	result, err := sr.DB.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	// audit relevant
	fmt.Println(result)
}