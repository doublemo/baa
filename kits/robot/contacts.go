package robot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/doublemo/baa/cores/crypto/id"
	"github.com/doublemo/baa/cores/crypto/token"
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/robot/session"
	"github.com/golang/protobuf/jsonpb"
)

// syncContacts 同步联系人
func syncContacts(peer session.Peer) {

}

// doCheckFriendRequest 检查好友请求
func doCheckFriendRequest(peer session.Peer, page int32, c RobotConfig) error {
	userid, ok := peer.Params("UserID")
	if !ok {
		return errors.New("invalid UserID")
	}

	agent, ok := peer.Params("AgentHttp")
	if !ok {
		return errors.New("invalid agent addr")
	}

	tk, ok := peer.Params("Token")
	if !ok {
		return errors.New("invalid token")
	}

	cmd, _ := id.Encrypt(uint64(command.UserContactsRequest), []byte(c.CommandSecret))
	frame := &pb.User_Contacts_FriendRequestList{
		UserId:  userid.(string),
		Page:    page,
		Size:    10,
		Version: 0,
	}

	pm := jsonpb.Marshaler{}
	data, err := pm.MarshalToString(frame)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	url := agent.(string) + "/v1/x/user/%s?t=%d"
	t := time.Now().Unix()
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf(url, cmd, t), bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}

	tks := token.NewTKS()
	tks.Push("command=" + cmd)
	tks.Push(fmt.Sprintf("t=%d", t))

	defer req.Body.Close()
	req.Header.Set("X-CSRF-Token", tks.Marshal(c.CSRFSecret))
	req.Header.Set("X-Session-Token", tk.(string))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body), req.URL.String())
	return err
}

func friendRequest(peer session.Peer, w coresproto.Response) error {
	if w.StatusCode() != 0 {
		return errors.New(string(w.Body()))
	}
	return nil
}
