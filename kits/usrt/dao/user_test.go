package dao

import (
	"testing"

	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/kits/usrt/proto/pb"
)

func TestUserStatusChange(t *testing.T) {
	err := Open(conf.Redis{Addr: []string{"127.0.0.1:6379"}, Prefix: "dao_test"}, CacherConfig{})
	if err != nil {
		t.Fatal(err)
	}

	data := []*pb.USRT_User{
		&pb.USRT_User{ID: 1, LoginType: "android", Addr: "agent01.ac.cb"},
		&pb.USRT_User{ID: 1, LoginType: "macos", Addr: "agent02.ac.cb"},
		&pb.USRT_User{ID: 1, LoginType: "pc", Addr: "agent03.ac.cb"},
		&pb.USRT_User{ID: 1, LoginType: "iphone", Addr: "agent04.ac.cb"},
		&pb.USRT_User{ID: 2, LoginType: "android", Addr: "agent01.ac.cb"},
		&pb.USRT_User{ID: 3, LoginType: "android", Addr: "agent01.ac.cb"},
		&pb.USRT_User{ID: 4, LoginType: "android", Addr: "agent01.ac.cb"},
		&pb.USRT_User{ID: 4, LoginType: "macos", Addr: "agent01.ac.cb"},
		&pb.USRT_User{ID: 5, LoginType: "android", Addr: "agent01.ac.cb"},
		&pb.USRT_User{ID: 5, LoginType: "macos", Addr: "agent02.ac.cb"},
		&pb.USRT_User{ID: 5, LoginType: "pc", Addr: "agent03.ac.cb"},
		&pb.USRT_User{ID: 5, LoginType: "iphone", Addr: "agent04.ac.cb"},
	}

	if noCompleted, err := UpdateStatusByUser(data...); err != nil || len(noCompleted) > 0 {
		t.Fatal(noCompleted, err)
	}

	t.Log(GetStatueCacheByUser(4, 5))
	t.Log(GetStatueCacheByUser(4, 5, 1, 2))

	if err := RemoveStatusByUser(data...); err != nil {
		t.Fatal(err)
	}
}
