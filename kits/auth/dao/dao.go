package dao

import (
	"time"

	"github.com/doublemo/baa/internal/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

var db *gorm.DB

// Open 打开数据库
func Open(c conf.DBMySQLConfig) error {
	gormConfig := &gorm.Config{}
	if c.TablePrefix != "" {
		gormConfig.NamingStrategy = schema.NamingStrategy{
			TablePrefix: c.TablePrefix,
		}
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
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

	mdb, err := db.DB()
	if err != nil {
		return err
	}

	mdb.SetConnMaxIdleTime(time.Duration(c.ConnMaxIdleTime) * time.Second)
	mdb.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second)
	mdb.SetMaxIdleConns(c.MaxIdleConns)
	mdb.SetMaxOpenConns(c.MaxOpenConns)

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
	db.AutoMigrate(&Accounts{})
	return nil
}

// DB 获取数据库
func DB() *gorm.DB {
	return db
}
