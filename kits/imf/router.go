package imf

import (
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/kits/imf/errcode"
	"github.com/doublemo/baa/kits/imf/proto"
	"github.com/doublemo/baa/kits/imf/proto/pb"
	"github.com/doublemo/baa/kits/imf/segmenter"
	grpcproto "github.com/golang/protobuf/proto"
)

var (
	r  = router.New()
	nr = router.New()
)

// RouterConfig 路由配置
type RouterConfig struct {
}

// InitRouter init
func InitRouter(c FilterConfig) {
	// Register grpc load balance

	// 注册处理请求
	r.HandleFunc(proto.CheckCommand, func(req *corespb.Request) (*corespb.Response, error) { return check(req, c) })

	// 订阅处理
	nr.HandleFunc(proto.ReloadCommand, reloadDictionary)
	nr.HandleFunc(proto.DirtyWordsCommand, dirtyWords)
	nr.HandleFunc(proto.CheckCommand, func(req *corespb.Request) (*corespb.Response, error) { return check(req, c) })

}

func reloadDictionary(req *corespb.Request) (*corespb.Response, error) {
	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	if err := segmenter.Reload("", ""); err != nil {
		return errcode.Bad(w, errcode.ErrInternalServer, err.Error()), nil
	}

	w.Payload = &corespb.Response_Content{Content: make([]byte, 0)}
	return w, nil
}

func dirtyWords(req *corespb.Request) (*corespb.Response, error) {
	var frame pb.IMF_DirtyWords_Request
	{
		if err := grpcproto.Unmarshal(req.Payload, &frame); err != nil {
			return nil, err
		}
	}

	w := &corespb.Response{
		Command: req.Command,
		Header:  req.Header,
	}

	switch payload := frame.Payload.(type) {
	case *pb.IMF_DirtyWords_Request_Add:
		segmenter.AddDirtyWords(payload.Add)

	case *pb.IMF_DirtyWords_Request_Delete:
		segmenter.RemoveDirtyWords(payload.Delete)
	}

	bytes, _ := grpcproto.Marshal(&pb.IMF_DirtyWords_Reply{OK: true})
	w.Payload = &corespb.Response_Content{Content: bytes}
	return w, nil
}
