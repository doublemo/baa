package dao

import (
	"errors"

	"gorm.io/gorm"
)

const (
	defaultUsersMaxRecord        = 10000000
	defaultUsersMaxTable         = 100
	defaultUsersIndexNoMaxRecord = 20000000
	defaultUsersIndexNoMaxTable  = 50
)

type (
	// Users 用户表
	Users struct {
		ID       uint64
		IndexNo  string
		Nickname string
		Headimg  string
		Age      int8
		Sex      int8
		Idcard   string
		Phone    string
	}

	// UsersIndexNo 索引表
	UsersIndexNo struct {
		ID      uint64 `gorm:"<-:create;primaryKey"`
		IndexNo string `gorm:"index"`
	}
)

// TableName 数据库表名称
func (users Users) TableName() string {
	return DBNamer("users")
}

// TableName 数据库表名称
func (usersIndexNo UsersIndexNo) TableName() string {
	return DBNamer("users", "index", "no")
}

// UseUsersTable 动态表名
func UseUsersTable(userID uint64) func(tx *gorm.DB) *gorm.DB {
	return useTable(userID, &Users{}, defaultUsersMaxRecord, defaultUsersMaxTable)
}

// UseUsersIndexNoTable 动态表名
func UseUsersIndexNoTable(name string) func(tx *gorm.DB) *gorm.DB {
	return useTable(name, &UsersIndexNo{}, defaultUsersIndexNoMaxRecord, defaultUsersIndexNoMaxTable)
}

// CreateUsers 创建用户信息
func CreateUsers(users *Users) error {
	if database == nil {
		return gorm.ErrInvalidDB
	}

	return database.Transaction(func(tx *gorm.DB) error {
		usersIndexNo := &UsersIndexNo{}
		ret := tx.Scopes(UseUsersIndexNoTable(users.IndexNo)).Where("index_no = ?", users.IndexNo).Take(&usersIndexNo)
		if ret.Error == nil {
			return ErrRecordIsFound
		}

		if !errors.Is(ret.Error, gorm.ErrRecordNotFound) {
			return ret.Error
		}

		if users.Phone != "" {
			ret = tx.Scopes(UseUsersIndexNoTable(users.Phone)).Where("index_no = ?", users.Phone).Take(&usersIndexNo)
			if errors.Is(ret.Error, gorm.ErrRecordNotFound) {
				ret = tx.Scopes(UseUsersIndexNoTable(users.Phone)).Create(&UsersIndexNo{ID: users.ID, IndexNo: users.Phone})
				if ret.Error != nil {
					return ret.Error
				}
			}
		}

		ret = tx.Scopes(UseUsersTable(users.ID)).Create(users)
		if ret.Error != nil {
			return ret.Error
		}

		ret = tx.Scopes(UseUsersIndexNoTable(users.IndexNo)).Create(&UsersIndexNo{ID: users.ID, IndexNo: users.IndexNo})
		return ret.Error
	})
}

// FindUsersByIndexNo 根据索引查询用户
func FindUsersByIndexNo(indexno string, cols ...string) (*Users, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	usersIndexNo := &UsersIndexNo{}
	tx := database.Scopes(UseUsersIndexNoTable(indexno)).Select("id", "index_no").Where("index_no = ?", indexno).First(usersIndexNo)
	if tx.Error != nil {
		return nil, tx.Error
	}

	users := &Users{}
	tx = database.Scopes(UseUsersTable(usersIndexNo.ID))
	if len(cols) > 0 {
		tx.Select(cols)
	}
	tx.Where("id = ?", usersIndexNo.ID).First(users)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return users, nil
}

// FindUsersByID 查询用户
func FindUsersByID(id uint64, cols ...string) (*Users, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	users := &Users{}
	tx := database.Scopes(UseUsersTable(id))
	if len(cols) > 0 {
		tx.Select(cols)
	}
	tx.Where("id = ?", id).First(users)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return users, nil
}

// FindUsersByMoreID 查询用户
func FindUsersByMoreID(users ...uint64) ([]*Users, error) {
	if database == nil {
		return nil, gorm.ErrInvalidDB
	}

	usersMap := make(map[string][]uint64)
	for _, id := range users {
		m := makeTablename(id, &Users{}, defaultUsersMaxRecord, defaultUsersMaxTable)
		if _, ok := usersMap[m]; !ok {
			usersMap[m] = make([]uint64, 0)
		}
		usersMap[m] = append(usersMap[m], id)
	}

	var data []*Users
	retData := make([]*Users, 0)
	for table, values := range usersMap {
		tx := database.Table(table).Select("id", "index_no", "nickname", "headimg", "age", "sex", "idcard", "phone").Where("id IN (?)", values).Find(&data)
		if tx.Error != nil {
			return nil, tx.Error
		}

		retData = append(retData, data...)
	}
	return retData, nil
}
