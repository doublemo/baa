package dao

import (
	"testing"

	"github.com/doublemo/baa/internal/conf"
	"github.com/doublemo/baa/kits/usrt/proto/pb"
)

func TestUserStatusChange(t *testing.T) {
	err := Open(conf.Redis{Addr: []string{"127.0.0.1:6379"}, Prefix: "baa:usrt"}, CacherConfig{})
	if err != nil {
		t.Fatal(err)
	}

	data := []*pb.USRT_User{
		&pb.USRT_User{ID: 344722248029966338, Type: "android", Value: "agent1.cn.sc.cd"},
		&pb.USRT_User{ID: 344722248029966338, Type: "macos", Value: "agent1.cn.sc.cd"},
		&pb.USRT_User{ID: 344722248029966338, Type: "pc", Value: "agent1.cn.sc.cd"},
		&pb.USRT_User{ID: 344722248029966338, Type: "iphone", Value: "agent1.cn.sc.cd"},
		&pb.USRT_User{ID: 344722248029966338, Type: "snid", Value: "snid1.cn.sc.cd"},
		&pb.USRT_User{ID: 344722248029966338, Type: "im", Value: "im1.cn.sc.cd"},
		&pb.USRT_User{ID: 344709394144956418, Type: "android", Value: "agent1.cn.sc.cd"},
		&pb.USRT_User{ID: 344709394144956418, Type: "macos", Value: "agent1.cn.sc.cd"},
		&pb.USRT_User{ID: 344709394144956418, Type: "pc", Value: "agent1.cn.sc.cd"},
		&pb.USRT_User{ID: 344709394144956418, Type: "iphone", Value: "agent1.cn.sc.cd"},
		&pb.USRT_User{ID: 344709394144956418, Type: "snid", Value: "snid1.cn.sc.cd"},
		&pb.USRT_User{ID: 344709394144956418, Type: "im", Value: "im1.cn.sc.cd"},
	}

	if noCompleted, err := UpdateStatusByUser(data...); err != nil || len(noCompleted) > 0 {
		t.Fatal(noCompleted, err)
	}

	t.Log(GetStatueCacheByUser(4, 5))
	t.Log(GetStatueCacheByUser(4, 5, 1, 2))

	// if err := RemoveStatusByUser(data...); err != nil {
	// 	t.Fatal(err)
	// }
}
