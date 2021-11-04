package dao

import (
	"context"
	"testing"

	"github.com/doublemo/baa/internal/conf"
)

func TestUserStatusChange(t *testing.T) {
	err := Open(conf.DBMySQLConfig{
		DNS:         "root:mlh520@tcp(127.0.0.1:3306)/baav2_auth?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true",
		TablePrefix: "bba_",
	}, conf.Redis{Addr: []string{"127.0.0.1:6379"}, Prefix: "baa:auth"})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(AssignServers(context.Background(), 1, "auth", "im", "user"))
}
