package dao

import (
	"strings"
	"time"

	"github.com/doublemo/baa/internal/conf"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

var (
	db        *gorm.DB
	rdb       redis.UniversalClient
	rdbPrefix string
)

// Open 打开数据库
func Open(c conf.DBMySQLConfig, rc conf.Redis) error {
	gormConfig := &gorm.Config{}
	if c.TablePrefix != "" {
		gormConfig.NamingStrategy = schema.NamingStrategy{
			TablePrefix: c.TablePrefix,
		}
	}

	db0, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       c.DNS,
		SkipInitializeWithVersion: c.SkipInitializeWithVersion,
		DefaultStringSize:         c.DefaultStringSize,
		DefaultDatetimePrecision:  &c.DefaultDatetimePrecision,
		DisableDatetimePrecision:  c.DisableDatetimePrecision,
		DontSupportRenameIndex:    c.DontSupportRenameIndex,
		DontSupportRenameColumn:   c.DontSupportRenameColumn,
		DontSupportForShareClause: c.DontSupportForShareClause,
	}), gormConfig)
	if err != nil {
		return err
	}

	mdb, err := db0.DB()
	if err != nil {
		return err
	}

	mdb.SetConnMaxIdleTime(time.Duration(c.ConnMaxIdleTime) * time.Second)
	mdb.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second)
	mdb.SetMaxIdleConns(c.MaxIdleConns)
	mdb.SetMaxOpenConns(c.MaxOpenConns)
	db = db0

	if len(c.Resolver) > 0 {
		var res *dbresolver.DBResolver

		for _, r := range c.Resolver {
			rc := dbresolver.Config{
				Sources:  make([]gorm.Dialector, len(r.Sources)),
				Replicas: make([]gorm.Dialector, len(r.Replicas)),
			}

			for idx, source := range r.Sources {
				rc.Sources[idx] = mysql.Open(source)
			}

			for idx, source := range r.Replicas {
				rc.Replicas[idx] = mysql.Open(source)
			}

			if r.Policy == "random" {
				rc.Policy = &dbresolver.RandomPolicy{}
			}

			tables := make([]interface{}, len(r.Tables))
			for idx, source := range r.Tables {
				tables[idx] = source
			}

			if res == nil {
				res = dbresolver.Register(rc, tables...)
			} else {
				res.Register(rc, tables...)
			}
		}

		if res != nil {
			res.SetConnMaxIdleTime(time.Duration(c.ConnMaxIdleTime) * time.Second)
			res.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second)
			res.SetMaxIdleConns(c.MaxIdleConns)
			res.SetMaxOpenConns(c.MaxOpenConns)
			db.Use(res)
		}
	}

	// 迁移
	// db.AutoMigrate(&Accounts{})
	// db.Scopes(dao.UseContactsTable(344707147885449216)).Create(&dao.Contacts{ID: 1})

	// 连接redis
	rdbPrefix = rc.Prefix
	rdb, err = rc.Connect()
	return err
}

// DB 获取数据库
func DB() *gorm.DB {
	return db
}

// RDB 获取redis数据库
func RDB() redis.UniversalClient {
	return rdb
}

// RDBNamer 创建redis key
func RDBNamer(name ...string) string {
	prefix := rdbPrefix
	if prefix != "" {
		prefix += ":"
	}

	return prefix + strings.Join(name, ":")
}
