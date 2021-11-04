package dao

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	defaultAccountsMaxRecord              = 10000000
	defaultAccountsMaxTable               = 100
	defaultAccountsSchemaNameIdxMaxRecord = 20000000
	defaultAccountsSchemaNameIdxMaxTable  = 25
)

type (
	// Accounts 账户
	Accounts struct {
		ID        uint64 `gorm:"<-:create;primaryKey"`
		UnionID   uint64 `gorm:"<-:create;index"`
		UserID    uint64 `gorm:"<-:create;index"`
		Schema    string `gorm:"<-:create;size:50;index:schema_name"`
		Name      string `gorm:"<-:create;index:schema_name"`
		Secret    string
		Status    int
		ExpiresAt int64
		PeerID    string
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt `gorm:"index"`
	}

	// AccountsSchemaNameIdx 账户索引表
	AccountsSchemaNameIdx struct {
		ID    uint64 `gorm:"<-:create;primaryKey"`
		Table uint32
	}
)

// TableName 数据库表名称
func (accounts Accounts) TableName() string {
	return DBNamer("accounts")
}

// TableName 数据库表名称
func (idx AccountsSchemaNameIdx) TableName() string {
	return DBNamer("accounts", "scheme", "name", "idx")
}

// GetAccoutsBySchemaAndName 根据条件 scheme, name 获取信息
func GetAccoutsBySchemaAndName(schema, name string, cols ...string) (*Accounts, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	idx, err := GetAccountsSchemeNameIdx(schema, name)
	if err != nil {
		return nil, err
	}

	accounts := &Accounts{}
	table := DBNamer(accounts.TableName(), strconv.FormatUint(uint64(idx.Table), 10))
	tx := database.Table(table)
	if len(cols) > 0 {
		tx.Select(cols)
	}
	tx.Where("scheme = ? AND name = ?", schema, name).First(accounts)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return accounts, nil
}

func GetAccoutsByID(id uint64, cols ...string) (*Accounts, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}
	accounts := &Accounts{}
	r := database.Scopes(UseAccountsTableFromUint64(id))
	if len(cols) > 0 {
		r.Select(cols)
	}
	r.Where("id = ?", id).First(accounts)
	if r.Error != nil {
		return nil, r.Error
	}
	return accounts, nil
}

func CreateAccount(accounts *Accounts) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	return database.Transaction(func(tx *gorm.DB) error {
		r := tx.Scopes(UseAccountsTableFromUint64(accounts.ID)).Create(accounts)
		if r.Error != nil {
			return r.Error
		}

		if r.RowsAffected != 1 {
			return errors.New("CreateFailed")
		}

		idx := makeTablenameFromString(strings.ToLower(accounts.Schema+accounts.Name), defaultAccountsSchemaNameIdxMaxRecord, defaultAccountsSchemaNameIdxMaxTable)
		r = tx.Scopes(UseAccountsSchemeNameIdxTableFromString(accounts.Schema, accounts.Name)).Create(&AccountsSchemaNameIdx{ID: accounts.ID, Table: idx})
		if r.Error != nil || r.RowsAffected != 1 {
			return errors.New("CreateFailed")
		}

		return nil
	})
}

func UpdatesAccountByID(id uint64, col string, value interface{}) (int64, error) {
	if database == nil {
		return 0, gorm.ErrInvalidDB
	}

	r := database.Scopes(UseAccountsTableFromUint64(id)).Where("id = ?", id).Update(col, value)
	return r.RowsAffected, r.Error
}

func GetAccountsSchemeNameIdx(schema, name string) (*AccountsSchemaNameIdx, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	idx := &AccountsSchemaNameIdx{}
	tx := database.Scopes(UseAccountsSchemeNameIdxTableFromString(schema, name)).Select("id", "table").Where("scheme = ? AND name = ?", schema, name).First(idx)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return idx, nil
}

// UseAccountsTableFromUint64 动态表名
func UseAccountsTableFromUint64(id uint64) func(tx *gorm.DB) *gorm.DB {
	return useTable(id, &Accounts{}, defaultAccountsMaxRecord, defaultAccountsMaxTable)
}

// UseAccountsSchemeNameIdxTableFromString 动态表名
func UseAccountsSchemeNameIdxTableFromString(schema, name string) func(tx *gorm.DB) *gorm.DB {
	return useTable(strings.ToLower(schema+name), &AccountsSchemaNameIdx{}, defaultAccountsSchemaNameIdxMaxRecord, defaultAccountsSchemaNameIdxMaxTable)
}
