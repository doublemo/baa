package robot

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/doublemo/baa/cores/crypto/id"
	"github.com/doublemo/baa/cores/crypto/token"
	coresproto "github.com/doublemo/baa/cores/proto"
	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/cores/types"
	"github.com/doublemo/baa/kits/robot/errcode"
	"github.com/golang/protobuf/jsonpb"
)

func RequestPostWithContext(ctx context.Context, cmd coresproto.Command, url string, data []byte, commandSecret []byte, tk, csrfSecret string) ([]byte, *types.ErrCode) {
	m, err := id.Encrypt(uint64(cmd), commandSecret)
	if err != nil {
		return nil, types.NewErrCode(errcode.ErrInternalServer.Code(), err.Error())
	}

	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	url += m
	t := time.Now().Unix()
	url = fmt.Sprintf(url+"?t=%d", t)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, types.NewErrCode(errcode.ErrInternalServer.Code(), err.Error())
	}

	defer req.Body.Close()

	tks := token.NewTKS()
	tks.Push("command=" + m)
	tks.Push(fmt.Sprintf("t=%d", t))

	req.Header.Set("X-CSRF-Token", tks.Marshal(csrfSecret))
	req.Header.Set("X-Session-Token", tk)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, types.NewErrCode(errcode.ErrInternalServer.Code(), err.Error())
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewErrCode(errcode.ErrInternalServer.Code(), err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, types.NewErrCode(errcode.ErrInternalServer.Code(), string(body))
	}

	var errmessage corespb.Error
	{
		if err := jsonpb.UnmarshalString(string(body), &errmessage); err == nil {
			return nil, types.NewErrCode(errmessage.Code, errmessage.Message)
		}
	}
	return body, nil
}
