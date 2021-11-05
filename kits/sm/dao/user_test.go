package dao

import (
	"context"
	"testing"

	"github.com/doublemo/baa/internal/conf"
)

func TestOnline(t *testing.T) {
	err := Open(conf.Redis{Addr: []string{"127.0.0.1:6379"}, Prefix: "baa:sm"}, CacherConfig{})
	if err != nil {
		t.Fatal(err)
	}

	err = Online(context.Background(), &Users{
		ID:          1,
		AgentServer: "aggg",
		Platform:    "web",
		Token:       "xxx",
		OnlineAt:    10000,
		IMServer:    "im",
		IDServer:    "id",
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestOffline(t *testing.T) {
	err := Open(conf.Redis{Addr: []string{"127.0.0.1:6379"}, Prefix: "baa:sm"}, CacherConfig{})
	if err != nil {
		t.Fatal(err)
	}

	err = Offline(context.Background(), 1, "pc")

	if err != nil {
		t.Fatal(err)
	}
}

func TestGetUsers(t *testing.T) {
	err := Open(conf.Redis{Addr: []string{"127.0.0.1:6379"}, Prefix: "baa:sm"}, CacherConfig{})
	if err != nil {
		t.Fatal(err)
	}

	data, err := GetMultiUsers(context.Background(), 1)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(data)
}

func TestUpdateUsersServer(t *testing.T) {
	err := Open(conf.Redis{Addr: []string{"127.0.0.1:6379"}, Prefix: "baa:sm"}, CacherConfig{})
	if err != nil {
		t.Fatal(err)
	}

	err = UpdateUsersServer(context.Background(), 1, "agent", "xx.xx.xx.xx")

	if err != nil {
		t.Fatal(err)
	}
}
