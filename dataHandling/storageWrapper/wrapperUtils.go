package wrapperUtils

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/awnumar/memguard"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type DBWrapper interface {
	Connect(ctx context.Context) error
	// RateLimit(ctx context.Context) error
	// rotatePW(ctx context.Context) error
	GetData(query string,  values []any) (any, error)
	SetData(query string, values []any, duration *time.Duration) error
	GetKey(id string) (any, error)
	SetKey(id string, key any, d *time.Duration) error
}

type DBPool struct {
	Pool map[string]DBWrapper
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
		Pool: make(map[string]DBWrapper, 0),
	}
}

func (p *DBPool) NewRedisWrapper(dbName string) (*RedisWrapper, error) {
	if p.Pool[dbName] != nil {
		err := p.Pool[dbName].Connect(context.Background())
		if err != nil {
			return nil, err
		}

		return p.Pool[dbName].(*RedisWrapper), nil
	}

	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	fmt.Println(dbName)
	dbNum, err := strconv.Atoi(os.Getenv(dbName + "_NUM"))
	if err != nil {
		return nil, err
	}

	fmt.Println(dbName)
	dbPort, err := strconv.Atoi(os.Getenv(dbName + "_PORT"))
	if err != nil {
		return nil, err
	}

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

	p.Pool[dbName] = &redisClient

	return &redisClient, err
}

func (r *RedisWrapper) Connect(ctx context.Context) error {
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
	fmt.Println("Connected to " + r.Container)

	r.DB = rdb
	return nil
}

func (r *RedisWrapper) GetData(query string, values []any) (any, error) {
	val := r.DB.Get(context.Background(), values[0].(string))

	returnedData, err := val.Result()
	if err != nil {
		return nil, err
	}

	return returnedData, nil
}

func (r *RedisWrapper) SetData(query string, values []any, duration *time.Duration) error {
	ret := r.DB.Set(context.Background(), values[0].(string), values[1], *duration)

	if ret.Err() != nil {
		return ret.Err()
	}

	return nil
}

func (r *RedisWrapper) GetKey(id string) (any, error) {
	val := r.DB.Get(context.Background(), id)
	fmt.Println(val)

	returnedData, err := func () (*memguard.Enclave, error) {
		valRaw, err := val.Result()
		if err != nil {
			return nil, err
		}

		fmt.Println(valRaw)
		valKey, err := base64.StdEncoding.DecodeString(valRaw)

		fmt.Println(valKey)
		return memguard.NewEnclave(valKey), err
	}()
	if err != nil {
		return nil, err
	}

	fmt.Println(reflect.TypeOf(returnedData))

	return returnedData, nil
}

func (r *RedisWrapper) SetKey(id string, key any, duration *time.Duration) error {
	keyLocked, err := key.(*memguard.Enclave).Open()
	if err != nil {
		return err
	}

	ret := r.DB.Set(context.Background(), id, keyLocked.Bytes(), *duration)
	keyLocked.Destroy()

	if ret.Err() != nil {
		return ret.Err()
	}

	return nil
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

func (p DBPool) NewSQLWrapper(dbName string) (*MySQLWrapper, error){
	if p.Pool[dbName] != nil {
		err := p.Pool[dbName].Connect(context.Background())
		if err != nil {
			return nil, err
		}

		return p.Pool[dbName].(*MySQLWrapper), nil
	}

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

	fmt.Println(dbName)
	p.Pool[dbName] = &sqlWrapper

	return &sqlWrapper, nil
}

func (sr *MySQLWrapper) Connect(ctx context.Context) error {
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

// key is stored encrypted in its raw form (no b64 encoding etc.)
func (sr *MySQLWrapper) GetKey(id string) (any, error) {
	var returnedValue any
	err := sr.DB.QueryRow("SELECT k_val FROM kstore WHERE id = ?", id).Scan(&returnedValue)
	if err != nil {
		return nil, err
	}


	return returnedValue, nil
}

func (sr *MySQLWrapper) SetKey(id string, key any, d *time.Duration) error {
	_, err := sr.DB.Exec("INSERT INTO kstore (id, k_val) VALUES (?, ?)", key.(string))
	if err != nil {
		return err
	}

	return nil
}

// expectedDts: map[string]any
func (sr *MySQLWrapper) GetData(query string, values []any) (any, error) {
	rows, err := sr.DB.Query(query, values...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("failed to get data (no match in db)")
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	returnedMap := make(map[int]map[string]any)
	uncastedRowValues := make([]any, len(columns))
	emptyMap := make(map[string]any)

	for rows.Next() {
		if err := rows.Scan(uncastedRowValues...); err != nil {
			return nil, err
		}

		for i, v := range columns {
			emptyMap[v] = uncastedRowValues[i]
		}

		returnedMap[uncastedRowValues[0].(int)] = emptyMap 
	}

	return returnedMap, nil
}

func (sr *MySQLWrapper) SetData(query string, values []any, n *time.Duration) error {
	result, err := sr.DB.Exec(query, values...)
	if err != nil {
		return err
	}

	// audit relevant
	fmt.Println(result)

	return nil
}