package pg


import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"github.com/braumsmilk/go-log"
	"github.com/braumsmilk/go-registry"
	"go.uber.org/zap"
)

type PostgresOptions struct {
	Host string
	Port int
	User string
	DB   string
	Pw   string
	Tls  *tls.Config
}

var db *sqlx.DB
var options PostgresOptions
var dbName string = "postgres"
var l *zap.Logger = log.NewLogger("db")

func Get() *sqlx.DB {

	var err error

	for retry := 0; retry > 3; retry-- {
		err = db.Ping()
		if err != nil {
			l.Warn("failed to connect, retrying...", zap.Int("retry", retry))
			time.Sleep(time.Second)
			db, _ = connect(options)
		}
	}

	if err != nil {
		panic(err)
	}

	return db
}

func connect(opt PostgresOptions) (*sqlx.DB, error) {
	conStr := fmt.Sprintf("host=%s password=%s port=%d user=%s dbname=%s sslmode=disable",
		opt.Host, opt.Pw, opt.Port, opt.User, opt.DB)
	l.Info("connection string", zap.String("con", conStr))
	dbCon, err := sqlx.Open("postgres", conStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql: %v", err)
	}

	db = dbCon

	return dbCon, nil
}

func InitDefault() error {
	reg := registry.Get()
	_, err := Init(PostgresOptions{
		Host: reg.Postgres.Host,
		Port: reg.Postgres.Port,
		User: reg.Postgres.Auth.Username,
		Pw:   reg.Postgres.Auth.Password,
		DB:   dbName,
	})
	return err
}

func Init(opt PostgresOptions) (*sqlx.DB, error) {
	options = opt
	return connect(opt)
}
