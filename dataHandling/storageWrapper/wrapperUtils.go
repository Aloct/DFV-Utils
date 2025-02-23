package wrapperUtils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/awnumar/memguard"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type stk func(keyRaw any) ([]byte, error)
type kts func(keyRaw any) (string, error)


type DBWrapper interface {
	Connect(ctx context.Context, retrys int) error
	// RateLimit(ctx context.Context) error
	// rotatePW(ctx context.Context) error
	GetData(query string,  values []any) (any, error)
	SetData(query string, values []any, duration *time.Duration) error
	GetKey(id string, stringToKey interface{}) (any, error)
	SetKey(id string, key any, d *time.Duration, keyToString interface{}) error
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
		err := p.Pool[dbName].Connect(context.Background(), 2)
		if err != nil {
			return nil, err
		}

		return p.Pool[dbName].(*RedisWrapper), nil
	}

	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	dbNum, err := strconv.Atoi(os.Getenv(dbName + "_NUM"))
	if err != nil {
		return nil, err
	}

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

func (r *RedisWrapper) Connect(ctx context.Context, retrys int) error {
	var rdb *redis.Client
	var err error
	for i := 0; i < retrys; i++ {
		rdb = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d", r.Container, r.Port), 
			Password: "", 
			DB: r.DbNum,
		})
	
		_, err = rdb.Ping(context.Background()).Result()

		if err == nil {
			i = retrys
		} else {
			time.Sleep(time.Second * 5)
		}
	}
	if err != nil {
		return err
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

func (r *RedisWrapper) GetKey(id string, stringToKey interface{}) (any, error) {
	val := r.DB.Get(context.Background(), id)

	returnedData, err := func () (*memguard.Enclave, error) {
		valRaw, err := val.Result()
		if err != nil {
			return nil, err
		}

		valKey, err := (stringToKey.(stk))(valRaw)
		if err != nil {
			return nil, err
		}

		return memguard.NewEnclave(valKey), err
	}()
	if err != nil {
		return nil, err
	}

	return returnedData, nil
}

func (r *RedisWrapper) SetKey(id string, key any, duration *time.Duration, keyToString interface{}) error {
	keyLocked, err := key.(*memguard.Enclave).Open()
	if err != nil {
		return err
	}

	ret, err := func() (*redis.StatusCmd, error) {
		keyString, err := (keyToString.(kts))(keyLocked.Bytes())
		if err != nil {
			return nil, err
		}

		return r.DB.Set(context.Background(), id, keyString, *duration), nil
	}()
	keyLocked.Destroy()

	if err != nil {
		return err
	}

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
		err := p.Pool[dbName].Connect(context.Background(), 2)
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

	p.Pool[dbName] = &sqlWrapper

	return &sqlWrapper, nil
}

func (sr *MySQLWrapper) Connect(ctx context.Context, retrys int) error {
	var err error
	var db *sql.DB
	for i := 0; i < retrys; i++ {
		db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", sr.User, sr.Password, sr.Container, sr.Port, sr.DBname))

		if err == nil {
			i = retrys
		} else {
			time.Sleep(time.Second * 5)
		}
	}
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


// key is not handled in a enclave cause its already encrypted
func (sr *MySQLWrapper) GetKey(id string, stringToKey interface{}) (any, error) {
	var returnedValue any
	err := sr.DB.QueryRow("SELECT k_val FROM kstore WHERE id = ?", id).Scan(&returnedValue)
	if err != nil {
		return nil, err
	}

	keySlice, err := (stringToKey.(stk))(returnedValue)
	// keySlice, err := hex.DecodeString(string(returnedValue.([]byte)))
	if err != nil {
		return nil, err
	}

	return keySlice, nil
}

// key is not handled in a enclave cause its already encrypted
func (sr *MySQLWrapper) SetKey(id string, key any, d *time.Duration, keyToString interface{}) error {
	keyString, err := (keyToString.(kts))(key)
	if err != nil {
		return err
	}

	_, err = sr.DB.Exec("INSERT INTO kstore (id, k_val) VALUES (?, ?)", keyString)
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