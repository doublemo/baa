package interceptor

import (
	"errors"
	"fmt"
	"strconv"

	corespb "github.com/doublemo/baa/cores/proto/pb"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	irouter "github.com/doublemo/baa/internal/router"
	"github.com/doublemo/baa/kits/agent/router"
	grpcproto "github.com/golang/protobuf/proto"
)

func AllowCommands(commands ...int32) func(router.RequestInterceptor) router.RequestInterceptor {
	commandsMap := make(map[int32]bool)
	for _, v := range commands {
		commandsMap[v] = true
	}

	return func(next router.RequestInterceptor) router.RequestInterceptor {
		return router.RequestInterceptorFunc(func(args router.RequestInterceptorArgs) error {
			if args.Request == nil {
				return next.Process(args)
			}

			if !commandsMap[args.Request.Command] {
				return errors.New("Unsupported command, unable to pass verification")
			}

			return next.Process(args)
		})
	}
}

func Authenticate(skip ...int32) func(router.RequestInterceptor) router.RequestInterceptor {
	skipMap := make(map[int32]bool)
	for _, v := range skip {
		skipMap[v] = true
	}

	return func(next router.RequestInterceptor) router.RequestInterceptor {
		return router.RequestInterceptorFunc(func(args router.RequestInterceptorArgs) error {
			if args.Peer == nil || args.Request == nil {
				return next.Process(args)
			}

			if skipMap[args.Request.Command] {
				return next.Process(args)
			}

			peer := args.Peer
			accountID, ok := peer.Params("AccountID")
			if !ok {
				return errors.New("Authorization failed, unable to carry out follow-up work")
			}

			userID, ok := peer.Params("UserID")
			if !ok {
				return errors.New("Authorization failed, unable to carry out follow-up work")
			}

			args.Request.Header["AccountID"] = accountID.(string)
			args.Request.Header["UserID"] = userID.(string)
			return next.Process(args)
		})
	}
}

func OnLogin(next router.ResponseInterceptor) router.ResponseInterceptor {
	return router.ResponseInterceptorFunc(func(args router.ResponseInterceptorArgs) error {
		if args.Peer == nil || args.Response == nil {
			return next.Process(args)
		}

		w := args.Response
		peer := args.Peer
		if w.Command != command.AuthLogin.Int32() {
			return next.Process(args)
		}

		var content []byte
		switch payload := w.Payload.(type) {
		case *corespb.Response_Content:
			content = payload.Content
		default:
			return next.Process(args)
		}

		var frame pb.Authentication_Form_LoginReply
		{
			if err := grpcproto.Unmarshal(content, &frame); err != nil {
				return err
			}
		}

		switch payload := frame.Payload.(type) {
		case *pb.Authentication_Form_LoginReply_Account:
			peer.SetParams("Token", payload.Account.Token)
		default:
		}

		if w.Header == nil {
			return errors.New("Invalid login return data, causing the gateway to fail to complete the login action")
		}

		accountID, ok1 := w.Header["AccountID"]
		unionID, ok2 := w.Header["UnionID"]
		userID, ok3 := w.Header["UserID"]
		if !ok1 || !ok2 || !ok3 {
			return errors.New("Invalid login return data, causing the gateway to fail to complete the login action")
		}

		peer.SetParams("AccountID", accountID)
		peer.SetParams("AccountUnionID", unionID)
		peer.SetParams("UserID", userID)
		return next.Process(args)
	})
}

func AuthenticateToken(mux *irouter.Mux, skip ...int32) func(router.RequestInterceptor) router.RequestInterceptor {
	skipMap := make(map[int32]bool)
	for _, v := range skip {
		skipMap[v] = true
	}

	return func(next router.RequestInterceptor) router.RequestInterceptor {
		return router.RequestInterceptorFunc(func(args router.RequestInterceptorArgs) error {
			if args.Request == nil {
				return next.Process(args)
			}

			r := args.Request
			if skipMap[r.Command] {
				return next.Process(args)
			}

			if r.Header == nil {
				return errors.New("Token verification failed, header is nil")
			}

			token, ok := r.Header["X-Session-Token"]
			if !ok || token == "" {
				return errors.New("Token cannot be empty, verification failed")
			}

			req := &corespb.Request{
				Header:  make(map[string]string),
				Command: int32(command.AuthorizedToken),
			}

			req.Payload, _ = grpcproto.Marshal(&pb.Authentication_Form_Authorized_Token{Token: token})
			resp, err := mux.Handler(kit.Auth.Int32(), req)
			if err != nil {
				return err
			}

			switch payload := resp.Payload.(type) {
			case *corespb.Response_Content:
				var info pb.Authentication_Form_Authorized_Info
				{
					if err := grpcproto.Unmarshal(payload.Content, &info); err != nil {
						return err
					}
				}

				r.Header["UserID"] = strconv.FormatUint(info.UserID, 10)
				r.Header["AccountID"] = strconv.FormatUint(info.ID, 10)
			case *corespb.Response_Error:
				return fmt.Errorf("authenticateToken: code %d error:%s", payload.Error.Code, payload.Error.Message)
			}

			return next.Process(args)
		})
	}
}
